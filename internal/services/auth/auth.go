package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/moorzeen/loyalty-service/storage"
)

const secret = "some secret string"
const AuthCookieName = "userAuth"

type Auth struct {
	storage storage.Service
}

type Credentials struct {
	Login    string
	Password string
}

type User struct {
	ID           int64
	Login        string
	PasswordHash []byte
}

type SignedCredentials struct {
	ID        int64
	Signature []byte
}

type Session struct {
	UserID       int64
	SignatureKey []byte
}

var (
	ErrShortPassword  = errors.New("the password is too short, requires more than 7 characters")
	ErrWrongUser      = errors.New("wrong login or password")
	ErrInvalidSession = errors.New("invalid auth token")
)

type Service interface {
	Register(cred Credentials) error
	Login(cred Credentials) (SignedCredentials, error)
	Validate(signedCred SignedCredentials) error
}

func NewService(storage storage.Service) Service {
	return Auth{storage}
}

func (a Auth) Register(cred Credentials) error {

	if err := passComplexity(cred.Password); err != nil {
		return ErrShortPassword
	}

	passwordHash := generateHash(cred.Password)

	if err := a.storage.AddUser(cred.Login, passwordHash); err != nil {
		return err
	}

	return nil
}

func (a Auth) Login(cred Credentials) (SignedCredentials, error) {

	signedCred := SignedCredentials{}

	user, err := a.storage.GetUserByLogin(cred.Login)
	if err != nil {
		return signedCred, ErrWrongUser
	}

	if !hmac.Equal(generateHash(cred.Password), user.PasswordHash) {
		return signedCred, fmt.Errorf("invalid password")
	}

	signKey, err := generateKey()
	if err != nil {
		return signedCred, fmt.Errorf("failed to generate signature key: %w", err)
	}

	session := storage.Session{UserID: user.ID, SignatureKey: signKey}

	err = a.storage.AddSession(session)
	if err != nil {
		return signedCred, err
	}

	signedCred = signUserID(session)

	return signedCred, nil
}

func (a Auth) Validate(signedCred SignedCredentials) error {
	return nil
}

/* HELPERS */

func passComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return ErrShortPassword
	}
	return nil
}

func generateHash(pass string) []byte {
	data := []byte(pass)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	hash := h.Sum(nil)

	return hash
}

func generateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

func signUserID(session Session) SignedCredentials {
	h := hmac.New(sha256.New, session.SignatureKey)
	h.Write([]byte(strconv.FormatInt(session.UserID, 10)))
	sum := h.Sum(nil)

	return SignedCredentials{
		ID:        session.UserID,
		Signature: sum,
	}
}

func MakeAuthCookie(u SignedCredentials) http.Cookie {
	v := fmt.Sprintf("%d|%x", u.ID, u.Signature)
	return http.Cookie{
		Name:  AuthCookieName,
		Value: v,
	}
}
