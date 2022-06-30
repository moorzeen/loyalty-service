package storage

import (
	"errors"
)

type user struct {
	ID           int64
	Login        string
	PasswordHash []byte
}

type Session struct {
	UserID       int64
	SignatureKey []byte
}

var (
	ErrLoginTaken  = errors.New("username is already taken")
	ErrInvalidUser = errors.New("invalid login or password")
)

type Service interface {
	AddUser(login string, passHash []byte) error
	GetUserByLogin(login string) (user, error)
	AddSession(session Session) error
}
