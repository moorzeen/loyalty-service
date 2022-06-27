package storage

import (
	"errors"
)

type User struct {
	Login        string
	PasswordHash string
}

var (
	ErrLoginTaken = errors.New("username is already taken")
)

type Storage interface {
	Register(login, passHash string) error
}
