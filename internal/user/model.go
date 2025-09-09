package user

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// User is custom struct to hold the user information and details
type User struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Password       password  `json:"-"`
	Activated      bool      `json:"activated"`
	AccountBalance float64   `json:"account_balance"`
	Version        int32     `json:"version"`
}

// AnonymousUser is for use not signed in
var AnonymousUser = &User{}

func (user *User) IsAnonymous() bool {
	return user == AnonymousUser
}

type password struct {
	plaintext *string
	Hash      []byte
}

func (p *password) Set(plaintext string, cost int) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.Hash = hash

	return nil
}

func (p password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintext))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func ValidateUser(v *validator.Validator, user *User) {
	v.CheckAddError(user.Name != "", "name", "must be given")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.Hash == nil {
		panic("user missing hash")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.CheckAddError(email != "", "email", "must be given")
	v.CheckAddError(validator.Matches(email, validator.EmailRX), "email", "must be vaild email")
}

func ValidatePasswordPlaintext(v *validator.Validator, passwordPlaintext string) {
	v.CheckAddError(passwordPlaintext != "", "password", "must be given")
	v.CheckAddError(len(passwordPlaintext) >= 8, "password", "must be at least 8 bytes")
	v.CheckAddError(len(passwordPlaintext) <= 72, "password", "cannot be more than 72 bytes")
}
