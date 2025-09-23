# 链接替换功能

## 功能说明

这个功能允许将API响应中的链接替换为指定域名，原链接通过Base64编码后放在新域名的路径中，用于隐藏实际调用的channel。

## 配置

在 `common/constants.go` 中有两个配置变量：

- `LinkReplacementEnabled`: 是否启用链接替换功能
- `ReplacementDomain`: 替换用的域名

## 使用方法

1. 启用功能：
```go
common.LinkReplacementEnabled = true
common.ReplacementDomain = "proxy.example.com"
```

2. 当功能启用时，所有API响应中的HTTP/HTTPS链接都会被替换为：
```
https://proxy.example.com/redirect/{base64_encoded_original_url}
```

## 替换示例

### 普通链接
```
原始: "Visit https://api.openai.com for details"
替换后: "Visit https://proxy.example.com/redirect/fa49b77a0791fcf50414c9d40174df2b19ce72d401f65ab3f3209afd46784848526ca154cfda2667f8a3811b9f345f4eb0c0 for details"
```

### Markdown链接
```
原始: "Check [OpenAI API](https://api.openai.com/v1/models) documentation"
替换后: "Check [OpenAI API](https://proxy.example.com/redirect/28b86b868077949df739e4e5a4f487b6c86fc3d15141f94250e02999ed3796c8eb4f5c3a7853ad03ac130df4e95d63a2ff36967c8b2142c8b9999a2121) documentation"
```

### JSON中的链接
```json
// 原始
{"url": "https://example.com/api", "text": "Visit https://test.com"}

// 替换后
{"url": "https://proxy.example.com/redirect/4ccc2f4c5c15278e5df2bc72eeabf3f36d574498a299129d58c4e158ee4de2ddd5ef04a0bcc2db2459450e39c19c18", "text": "Visit https://proxy.example.com/redirect/7e702444d143a986431270c3334d57d1fce6273984e81ee33812e94d3489b0df2a974beefbe56eb4a01e0cad5d5ee5fd5b6d35b4"}
```

### JSON中的Markdown链接
```json
// 原始
{"message": "Check [OpenAI API](https://api.openai.com) for details"}

// 替换后
{"message": "Check [OpenAI API](https://proxy.example.com/redirect/b3b723a0756866d0648242f2f3b40be8768b18ef57eb9e6c94464784c61fc1379fecaf03d047cda4eca1edce1113ee23c446) for details"}
```

## 工作原理

1. **统一入口**: 通过 `relaycommon.ApplyLinkReplacement()` 函数统一处理所有链接替换
2. **响应处理**: 在各个channel的DoResponse方法中调用链接替换函数
3. **智能检测**: 自动检测内容格式（JSON/文本/Markdown）并应用相应的替换规则
4. **安全编码**: 使用AES-GCM加密原始链接，再转换为十六进制字符串，看起来像随机ID
5. **重定向处理**: 访问替换后的链接时，系统会解密并重定向到原始链接

## API接口

### 核心函数

- `relaycommon.ApplyLinkReplacement(content string) string` - 统一的链接替换入口
- `relaycommon.ApplyLinkReplacementToBytes(data []byte) []byte` - 字节数据的链接替换
- `common.ReplaceLinksInText(text string) string` - 文本和Markdown链接替换
- `common.ReplaceLinksInJSON(jsonStr string) string` - JSON格式的链接替换

## 支持的场景

- JSON响应中的链接
- JSON响应中的Markdown格式链接
- 文本响应中的链接
- 文本响应中的Markdown格式链接
- 流式响应中的链接（包括Markdown格式）
- 复杂嵌套JSON结构中的链接

## 已集成的Channel

- OpenAI (包括流式和非流式)
- Claude (包括流式和非流式)
- 其他channel可以通过类似方式集成

## 测试

运行测试验证功能：
```bash
go test ./common -v
```

## 注意事项

- 功能默认关闭，需要手动启用
- 只有当 `LinkReplacementEnabled` 为 `true` 且 `ReplacementDomain` 不为空时才会生效
- 链接替换会轻微增加响应处理时间
- 确保替换域名能够正确解析并提供重定向服务
- 系统会自动避免重复编码已经被替换的链接
- JSON中的Markdown链接会保持格式完整性，只替换URL部分

## 安全特性

- **AES-GCM加密**: 使用高强度加密算法保护原始链接
- **随机性**: 每次加密同一链接都会产生不同结果，增强安全性
- **不可逆推**: 编码后的链接看起来像随机十六进制字符串，难以识别原始内容
- **密钥管理**: 基于系统密钥和域名生成唯一加密密钥
- **向后兼容**: 支持解码旧的Base64编码链接