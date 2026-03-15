package oauth

import (
	"crypto/rand"
	"encoding/base64"
)

func newCodeVerifierWithSize(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
