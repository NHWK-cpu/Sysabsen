package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSecureToken membuat string acak sepanjang n byte
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}