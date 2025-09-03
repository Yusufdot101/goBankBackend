package token

import "time"

const (
	ScopeActivation    = "activation"
	ScopeAuthorization = "authorization"
)

type Token struct {
	ID        int64
	CreatedAt time.Time
	Expiry    time.Time
	UserID    int64
	Plaintext string
	hash      []byte
	Scope     string
}
