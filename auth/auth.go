package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/moorzeen/loyalty-service/auth/helpers"
	"github.com/moorzeen/loyalty-service/auth/storage"
)

const (
	passwordHashKey    = "super secret key for user passwords hash"
	UserAuthCookieName = "authToken"
)

type Auth struct {
	storage storage.Storage
}

func NewAuth(str storage.Storage) Service {
	return &Auth{storage: str}
}

func (a *Auth) SignUp(ctx context.Context, username, password string) error {
	if err := helpers.PassComplexity(password); err != nil {
		return ErrShortPassword
	}

	passwordHash := helpers.GenerateHash(password, []byte(passwordHashKey))

	err := a.storage.AddUser(ctx, username, passwordHash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrUsernameTaken
	}
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (a *Auth) SignIn(ctx context.Context, username, password string) (string, error) {

	// get user from BD
	user, err := a.storage.GetUser(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoUser
	}
	if err != nil {
		log.Println(err)
		return "", err
	}

	// compare incoming password with passwordHash from DB
	passwordHash := helpers.GenerateHash(password, []byte(passwordHashKey))
	if !hmac.Equal(passwordHash, user.PasswordHash) {
		return "", ErrInvalidUser
	}

	// generate user signKey for session token
	signKey, err := helpers.GenerateKey()
	if err != nil {
		log.Println(err)
		return "", err
	}

	// add userID and signKey to session DB table
	err = a.storage.SetSession(ctx, user.ID, signKey)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// generate userID signature
	sign := helpers.GenerateHash(strconv.FormatUint(user.ID, 10), signKey)

	// make authToken
	authToken := fmt.Sprintf("%d|%x", user.ID, sign)

	return authToken, nil
}

func (a *Auth) TokenCheck(ctx context.Context, authToken string) error {
	var (
		userID uint64
		sign   []byte
	)
	_, err := fmt.Sscanf(authToken, "%d|%x", &userID, &sign)
	if err != nil {
		log.Printf("failed to parse authentication cookie \"%s\": %s", authToken, err.Error())
	}

	session, err := a.storage.GetSession(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidAuthToken
	}
	if err != nil {
		log.Println(err)
		return err
	}

	if !bytes.Equal(sign, helpers.GenerateHash(strconv.FormatUint(userID, 10), session.SignKey)) {
		return ErrInvalidAuthToken
	}

	return nil
}
