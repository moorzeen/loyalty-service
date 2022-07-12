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
	"github.com/moorzeen/loyalty-service/internal/storage"
)

const (
	passwordHashKey    = "super secret key for user passwords hash"
	UserAuthCookieName = "authToken"
)

type Credentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type Service struct {
	storage storage.Service
}

func NewService(str storage.Service) Service {
	return Service{storage: str}
}

func (a *Service) SignUp(ctx context.Context, cred Credentials) error {

	if err := passComplexity(cred.Password); err != nil {
		return ErrShortPassword
	}

	passwordHash := generateHash(cred.Password, passwordHashKey)

	userID, err := a.storage.AddUser(ctx, cred.Username, passwordHash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrUsernameTaken
	}
	if err != nil {
		return err
	}

	err = a.storage.AddAccount(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Service) SignIn(ctx context.Context, cred Credentials) (string, error) {

	// get user by username from BD
	user, err := a.storage.GetUser(ctx, cred.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoUser
	}
	if err != nil {
		return "", err
	}

	// compare hash of entered password with hash from DB
	passwordHash := generateHash(cred.Password, passwordHashKey)
	if !hmac.Equal(passwordHash, user.PasswordHash) {
		return "", ErrWrongPassword
	}

	// generate user signKey for session token
	signKey, err := generateKey()
	if err != nil {
		return "", err
	}

	// add userID and signKey to session DB table
	err = a.storage.SetSession(ctx, user.ID, signKey)
	if err != nil {
		return "", err
	}

	// generate userID signature
	sign := generateHash(strconv.FormatUint(user.ID, 10), string(signKey))

	// make authToken
	authToken := fmt.Sprintf("%d|%x", user.ID, sign)

	return authToken, nil
}

func (a *Service) ValidateToken(ctx context.Context, authToken string) (uint64, error) {
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
		return 0, ErrInvalidAuthToken
	}
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(sign, generateHash(strconv.FormatUint(userID, 10), string(session.SignKey))) {
		return 0, ErrInvalidAuthToken
	}

	return session.UserID, nil
}
