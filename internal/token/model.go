package token

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

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

func ValidateToken(v *validator.Validator, tokenPlaintext string) {
	v.CheckAddError(tokenPlaintext != "", "token", "must be provided")
	v.CheckAddError(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
