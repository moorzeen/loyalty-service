package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/moorzeen/loyalty-service/internal/services/storage"
)

const secret = "some secret string"

type Auth struct {
	storage storage.Storage
}

var (
	ErrShortPassword = errors.New("the password is too short, requires more than 7 characters")
)

type Service interface {
	Register(login, pass string) error
	Login(login, pass string) (http.Cookie, error)
	Validate(token string) error
}

func NewService(storage storage.Storage) Service {
	return Auth{storage}
}

func (a Auth) Register(login, pass string) error {
	err := passComplexity(pass)
	if err != nil {
		return err
	}

	hash := generateHash(pass)

	err = a.storage.AddUser(login, hash)
	if err != nil {
		return err
	}

	return nil
}

func (a Auth) Login(login, pass string) (http.Cookie, error) {
	authCookie := http.Cookie{}
	user := storage.User{}

	user, err := a.storage.GetUser(login)
	if err != nil {
		return authCookie, err
	}

	hash := generateHash(pass)

	if user.PasswordHash != hash {
		return authCookie, storage.ErrInvalidUser
	}

	user.SessionUUID = uuid.New()

	err = a.storage.SetSession(user.Login, user.SessionUUID)
	if err != nil {
		return http.Cookie{}, err
	}

	authCookie = http.Cookie{Name: "authToken", Value: user.SessionUUID.String()}

	return authCookie, nil
}

func (a Auth) Validate(token string) error {
	return nil
}

func passComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return ErrShortPassword
	}
	return nil
}

func generateHash(pass string) string {
	data := []byte(pass)
	hash := make([]byte, 0)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	hash = h.Sum(hash)

	return hex.EncodeToString(hash)
}
