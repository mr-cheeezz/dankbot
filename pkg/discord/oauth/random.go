package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func randomToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be greater than 0")
	}

	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("read random token bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
