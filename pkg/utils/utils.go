package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

// GenerateDeviceID 根据用户名生成设备ID（16字符十六进制字符串）
func GenerateDeviceID(username string) string {
	raw := username
	hash := sha1.Sum([]byte(raw))
	return hex.EncodeToString(hash[:8]) // 截断成 16 字符
}

// JsonPrint 美化打印JSON
func JsonPrint(v any) string {
	if v == nil {
		return "{}"
	}

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(b)
}
