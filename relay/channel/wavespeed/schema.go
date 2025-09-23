package wavespeed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// SchemaProperty represents a property in OpenAPI schema
type SchemaProperty struct {
	Type        string      `json:"type"`
	Default     interface{} `json:"default"`
	Enum        []string    `json:"enum,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
	MinLength   *int        `json:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"-"` // Set during parsing
}

// OpenAPISchema represents the OpenAPI schema structure we need
type OpenAPISchema struct {
	Components struct {
		Schemas struct {
			Input struct {
				Type       string                    `json:"type"`
				Required   []string                  `json:"required"`
				Properties map[string]SchemaProperty `json:"properties"`
			} `json:"Input"`
		} `json:"schemas"`
	} `json:"components"`
}

// SchemaManager manages loading and caching of model schemas
type SchemaManager struct {
	schemas map[string]*OpenAPISchema
}

var schemaManager *SchemaManager

func init() {
	schemaManager = &SchemaManager{
		schemas: make(map[string]*OpenAPISchema),
	}
}

// GetSchemaManager returns the global schema manager instance
func GetSchemaManager() *SchemaManager {
	return schemaManager
}

// LoadSchema loads schema for a specific model
func (sm *SchemaManager) LoadSchema(modelPath string) (*OpenAPISchema, error) {
	// Check if already cached
	if schema, exists := sm.schemas[modelPath]; exists {
		return schema, nil
	}

	// Determine schema file path
	var schemaFile string
	if IsVideoModel(modelPath) {
		schemaFile = filepath.Join("schemas", "video", modelPathToFileName(modelPath))
	} else {
		schemaFile = filepath.Join("schemas", "image", modelPathToFileName(modelPath))
	}

	// Load schema from file
	schema, err := sm.loadSchemaFromFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema for model %s: %w", modelPath, err)
	}

	// Mark required fields
	for _, requiredField := range schema.Components.Schemas.Input.Required {
		if prop, exists := schema.Components.Schemas.Input.Properties[requiredField]; exists {
			prop.Required = true
			schema.Components.Schemas.Input.Properties[requiredField] = prop
		}
	}

	// Cache the schema
	sm.schemas[modelPath] = schema
	return schema, nil
}

// loadSchemaFromFile loads schema from JSON file
func (sm *SchemaManager) loadSchemaFromFile(filePath string) (*OpenAPISchema, error) {
	// Get the absolute path to the schema file
	// This assumes schemas are stored relative to the wavespeed package
	fullPath := filepath.Join("relay", "channel", "wavespeed", filePath)

	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var schema OpenAPISchema
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// modelPathToFileName converts model path to schema file name
func modelPathToFileName(modelPath string) string {
	// Convert path separators to hyphens
	fileName := strings.ReplaceAll(modelPath, "/", "-")
	return fileName + ".json"
}

// BuildRequestFromSchema builds a request map based on schema and input parameters
func (sm *SchemaManager) BuildRequestFromSchema(modelPath string, params map[string]interface{}) (map[string]interface{}, error) {
	schema, err := sm.LoadSchema(modelPath)
	if err != nil {
		return nil, err
	}

	request := make(map[string]interface{})
	inputSchema := schema.Components.Schemas.Input

	// Process each property in the schema
	for propName, propDef := range inputSchema.Properties {
		value := sm.resolvePropertyValue(propName, propDef, params)
		if value != nil {
			request[propName] = value
		}
	}

	// Validate required fields
	for _, requiredField := range inputSchema.Required {
		if _, exists := request[requiredField]; !exists {
			return nil, fmt.Errorf("required field '%s' is missing", requiredField)
		}
	}

	return request, nil
}

// resolvePropertyValue resolves the value for a property based on schema and input
func (sm *SchemaManager) resolvePropertyValue(propName string, propDef SchemaProperty, params map[string]interface{}) interface{} {
	// First check if value is provided in params
	if value, exists := params[propName]; exists {
		// Validate and convert the value based on property type
		return sm.validateAndConvertValue(value, propDef)
	}

	// Use default value if available
	if propDef.Default != nil {
		return propDef.Default
	}

	// Return nil for optional fields without default
	if !propDef.Required {
		return nil
	}

	// For required fields without default, this will be caught in validation
	return nil
}

// validateAndConvertValue validates and converts a value according to schema property definition
func (sm *SchemaManager) validateAndConvertValue(value interface{}, propDef SchemaProperty) interface{} {
	switch propDef.Type {
	case "string":
		if str, ok := value.(string); ok {
			// Check enum constraint
			if len(propDef.Enum) > 0 {
				for _, enumVal := range propDef.Enum {
					if str == enumVal {
						return str
					}
				}
				// If not in enum, return default or first enum value
				if propDef.Default != nil {
					return propDef.Default
				}
				if len(propDef.Enum) > 0 {
					return propDef.Enum[0]
				}
			}
			// Check length constraints
			if propDef.MaxLength != nil && len(str) > *propDef.MaxLength {
				return str[:*propDef.MaxLength]
			}
			return str
		}
		// Try to convert to string
		return fmt.Sprintf("%v", value)

	case "integer":
		if intVal, ok := value.(int); ok {
			return sm.validateIntegerConstraints(intVal, propDef)
		}
		if floatVal, ok := value.(float64); ok {
			return sm.validateIntegerConstraints(int(floatVal), propDef)
		}
		// Return default if conversion fails
		if propDef.Default != nil {
			return propDef.Default
		}
		return 0

	case "number":
		if floatVal, ok := value.(float64); ok {
			return sm.validateNumberConstraints(floatVal, propDef)
		}
		if intVal, ok := value.(int); ok {
			return sm.validateNumberConstraints(float64(intVal), propDef)
		}
		// Return default if conversion fails
		if propDef.Default != nil {
			return propDef.Default
		}
		return 0.0

	case "boolean":
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
		// Return default if conversion fails
		if propDef.Default != nil {
			return propDef.Default
		}
		return false

	default:
		return value
	}
}

// validateIntegerConstraints validates integer value against min/max constraints
func (sm *SchemaManager) validateIntegerConstraints(value int, propDef SchemaProperty) int {
	if propDef.Minimum != nil && float64(value) < *propDef.Minimum {
		return int(*propDef.Minimum)
	}
	if propDef.Maximum != nil && float64(value) > *propDef.Maximum {
		return int(*propDef.Maximum)
	}
	return value
}

// validateNumberConstraints validates number value against min/max constraints
func (sm *SchemaManager) validateNumberConstraints(value float64, propDef SchemaProperty) float64 {
	if propDef.Minimum != nil && value < *propDef.Minimum {
		return *propDef.Minimum
	}
	if propDef.Maximum != nil && value > *propDef.Maximum {
		return *propDef.Maximum
	}
	return value
}

// GetModelEndpoint returns the API endpoint for a model
func GetModelEndpoint(modelPath string) string {
	config := GetModelConfig(modelPath)
	return fmt.Sprintf("/api/v3/%s", config.Endpoint)
}