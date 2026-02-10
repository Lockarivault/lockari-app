package encryption

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// EncodeBase64 encodes bytes into a URL-safe base64 string without padding.
func EncodeBase64(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeBase64 decodes a URL-safe base64 string without padding.
func DecodeBase64(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return decoded, nil
}

// EncodeHex encodes bytes into hexadecimal string.
func EncodeHex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return hex.EncodeToString(b)
}

// DecodeHex decodes hexadecimal string into bytes.
func DecodeHex(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode hex: %w", err)
	}
	return decoded, nil
}
