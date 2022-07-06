package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

func PassComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return errors.New("required at least 8 characters")
	}
	return nil
}

func GenerateHash(input string, key []byte) []byte {
	data := []byte(input)

	h := hmac.New(sha256.New, key)
	h.Write(data)
	hash := h.Sum(nil)

	return hash
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}
