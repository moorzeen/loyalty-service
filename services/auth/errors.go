package auth

import (
	"errors"
)

var (
	ErrShortPassword    = errors.New("the password is too short, requires more than 7 characters")
	ErrUsernameTaken    = errors.New("username is already taken")
	ErrInvalidUser      = errors.New("invalid login or password")
	ErrInvalidAuthToken = errors.New("invalid authToken")
	ErrNoUser           = errors.New("login not found")
	ErrWrongPassword    = errors.New("wrong password")
)
