package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

type Service struct {
	Repository *Repository
}

func generateToken(userID int64, timeToLive time.Duration, scope string) (*Token, error) {
	token := Token{
		UserID:    userID,
		CreatedAt: time.Now(),
		Expiry:    time.Now().Add(timeToLive),
		Scope:     scope,
	}

	// generate 16 random bytes, this will be the plaintext
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// encode it to base 32, it might have '=' at the end so we remove it with base32.NoPadding
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// hash the plaintext with sha256
	hash := sha256.Sum256([]byte(token.Plaintext))

	// make it slice and store it
	token.hash = hash[:]

	return &token, nil
}

func (s *Service) New(userID int64, timeToLive time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, timeToLive, scope)
	if err != nil {
		return nil, err
	}

	err = s.Repository.Insert(token)
	return token, err
}

func (s *Service) AuthorizationToken(userID int64) (*Token, error) {
	token, err := s.New(userID, 24*time.Hour, ScopeAuthorization)
	if err != nil {
		return nil, err
	}

	return token, nil
}
