package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const secret = "some Good Secret"

func PassComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return errors.New("the password is too short, requires more than 7 characters")
	}
	return nil
}

func GenerateHash(pass string) string {
	data := []byte(pass)
	hash := make([]byte, 0)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	hash = h.Sum(hash)

	return hex.EncodeToString(hash)
}
