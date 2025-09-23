package common

import (
	"fmt"
	"testing"
)

func TestPassword2Hash(t *testing.T) {
	pass := "ucloud123"
	hash, err := Password2Hash(pass)
	if err != nil {
		t.Errorf("Password2Hash failed: %v", err)
	}
	if !ValidatePasswordAndHash(pass, hash) {
		t.Errorf("ValidatePasswordAndHash failed")
	}
	fmt.Println(hash)
}
