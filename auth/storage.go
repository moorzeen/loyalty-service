package auth

import (
	"context"
)

type User struct {
	ID           uint64
	Username     string
	PasswordHash []byte
}

type Session struct {
	UserID  uint64
	SignKey []byte
}

type Storage interface {
	AddUser(ctx context.Context, username string, passwordHash []byte) error
	GetUser(ctx context.Context, username string) (*User, error)
	SetSession(ctx context.Context, userID uint64, signKey []byte) error
	GetSession(ctx context.Context, userID uint64) (*Session, error)
}
