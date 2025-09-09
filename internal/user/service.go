package user

import (
	"errors"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type UserRepo interface {
	Insert(user *User) error
	Get(userID int64) (*User, error)
	GetByEmail(email string) (*User, error)
	GetForToken(tokenPlaintext, scope string) (*User, error)
	UpdateTx(
		userID int64, name, email string, passwordHash []byte,
		accountBalance float64, activated bool,
	) (*User, error)
}

type Mailer interface {
	Send(to, template string, data map[string]any) error
}

type TokenService interface {
	New(userID int64, timeToLive time.Duration, scope string) (*token.Token, error)
	DeleteAllForUser(userID int64, scope string) error
}

type Service struct {
	Repo         UserRepo
	Mailer       Mailer
	TokenService TokenService
}

func (s *Service) GetUser(userID int64) (*User, error) {
	user, err := s.Repo.Get(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetUserByEmail(email string) (*User, error) {
	user, err := s.Repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUser(
	userID int64, name, email string, passwordHash []byte, accountBalance float64, activated bool,
) (*User, error) {
	user, err := s.Repo.UpdateTx(userID, name, email, passwordHash, accountBalance, activated)
	return user, err
}

func (s *Service) Register(
	v *validator.Validator,
	name, email, passwordPlaintext string,
) (*User, *token.Token, error) {
	user := &User{
		Name:  name,
		Email: email,
	}

	err := user.Password.Set(passwordPlaintext, 12)
	if err != nil {
		return nil, nil, err
	}

	if ValidateUser(v, user); !v.IsValid() {
		return nil, nil, validator.ErrFailedValidation
	}

	err = s.Repo.Insert(user)
	if err != nil {
		return nil, nil, err
	}

	// get the activation token and send it to the user
	t, err := s.TokenService.New(user.ID, 3*24*time.Hour, token.ScopeActivation)
	if err != nil {
		return nil, nil, err
	}

	return user, t, nil
}

func (s *Service) GetUserForToken(tokenPlaintext, scope string) (*User, error) {
	user, err := s.Repo.GetForToken(tokenPlaintext, scope)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Activate(tokenPlaintext string) (*User, error) {
	u, err := s.Repo.GetForToken(tokenPlaintext, token.ScopeActivation)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoRecord):
			return nil, token.ErrInvaildToken

		default:
			return nil, err
		}
	}

	u.Activated = true

	u, err = s.Repo.UpdateTx(u.ID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated)
	if err != nil {
		return u, err
	}

	err = s.TokenService.DeleteAllForUser(u.ID, token.ScopeActivation)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) TransferMoney(fromUser, toUser *User, amount float64) (*User, error) {
	fromUser.AccountBalance -= amount
	fromUser, err := s.Repo.UpdateTx(
		fromUser.ID, fromUser.Name, fromUser.Email, fromUser.Password.Hash,
		fromUser.AccountBalance, fromUser.Activated,
	)
	if err != nil {
		return nil, err
	}

	// update the recipient account
	toUser.AccountBalance += amount
	_, err = s.Repo.UpdateTx(
		toUser.ID, toUser.Name, toUser.Email, toUser.Password.Hash,
		toUser.AccountBalance, toUser.Activated,
	)
	// if no error, return the updated state of the sender account
	if err == nil {
		return fromUser, nil
	}

	// in case of an error transferring the money, try to put the amount taken from the sender back
	fromUser.AccountBalance += amount
	_, err = s.Repo.UpdateTx(
		fromUser.ID, fromUser.Name, fromUser.Email, fromUser.Password.Hash,
		fromUser.AccountBalance, fromUser.Activated,
	)
	return nil, err
}
