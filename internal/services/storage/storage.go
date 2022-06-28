package storage

import (
	"errors"
)

type User struct {
	Login        string
	PasswordHash string
}

var (
	ErrLoginTaken  = errors.New("username is already taken")
	ErrInvalidUser = errors.New("invalid login or password")
)

type Storage interface {
	AddUser(user, passHash string) error
	IsUser(login, hash string) error
}
