package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

func passComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return errors.New("required at least 8 characters")
	}
	return nil
}

func generateHash(input string, key string) []byte {
	data := []byte(input)
	signKey := []byte(key)

	h := hmac.New(sha256.New, signKey)
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
