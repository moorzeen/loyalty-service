package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func generateHash(pass, secret string) string {
	data := []byte(pass)
	hash := make([]byte, 0)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	hash = h.Sum(hash)

	return hex.EncodeToString(hash)
}
