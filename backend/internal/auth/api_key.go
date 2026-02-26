package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func NewAPIKey() (string, error) {
	return NewAPIKeyFromReader(rand.Reader)
}

func NewAPIKeyFromReader(r io.Reader) (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return "mmk_" + base64.RawURLEncoding.EncodeToString(b), nil
}
