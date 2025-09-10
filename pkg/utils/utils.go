package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

var (
	gzipMagic = []byte{0x1f, 0x8b} // GZIP 标识
)

// IsGzip 检查数据是否为GZIP格式
func IsGzip(body []byte) bool {
	return len(body) >= 2 && bytes.Equal(body[:2], gzipMagic)
}

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
