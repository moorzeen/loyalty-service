package auth

import (
	"context"
)

type Service interface {
	SignUp(ctx context.Context, username, password string) error
	SignIn(ctx context.Context, username, password string) (string, error)
	TokenCheck(ctx context.Context, authToken string) error
}
