package storage

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	Login        string
	PasswordHash string
	SessionUUID  uuid.UUID
}

var (
	ErrLoginTaken  = errors.New("username is already taken")
	ErrInvalidUser = errors.New("invalid login or password")
)

type Storage interface {
	AddUser(login, passHash string) error
	GetUser(login string) (User, error)
	SetSession(login string, token uuid.UUID) error
}
