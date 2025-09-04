package user

import (
	"errors"
	"log"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Mailer interface {
	Send(to, template string, data map[string]any) error
}

type Service struct {
	Repo   *Repository
	Mailer Mailer
}

func (s *Service) Register(
	v *validator.Validator, name, email, passwordPlaintext string,
) (*User, string, error) {
	user := &User{
		Name:  name,
		Email: email,
	}
	err := user.Password.Set(passwordPlaintext)
	if err != nil {
		return nil, "", err
	}

	if ValidateUser(v, user); !v.IsValid() {
		return nil, "", nil
	}

	err = s.Repo.Insert(user)
	if err != nil {
		return nil, "", err
	}

	// get the activation token and send it to the user
	tokenService := token.Service{Repository: &token.Repository{DB: s.Repo.DB}}
	t, err := tokenService.New(user.ID, 3*24*time.Hour, token.ScopeActivation)

	// send the email with the account activation, is async because we dont want to wait for it as
	// it might take a long time
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ERROR: %s", err)
			}
		}()
		data := map[string]any{
			"userID":   user.ID,
			"userName": user.Name,
			"token":    t.Plaintext,
		}
		err = s.Mailer.Send(user.Email, "user_welcome.html", data)
		if err != nil {
			panic(err.Error())
		}
	}()

	if err != nil {
		return nil, t.Plaintext, err
	}

	return user, t.Plaintext, err
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

	err = s.Repo.Update(u)
	if err != nil {
		return u, err
	}

	tokenService := token.Service{Repository: &token.Repository{DB: s.Repo.DB}}
	err = tokenService.Repository.DeleteAllForUser(u.ID, token.ScopeActivation)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) TransferMoney(fromUser, toUser *User, amount float64) (*User, error) {
	fromUser.AccountBalance -= amount
	err := s.Repo.Update(fromUser)
	if err != nil {
		return nil, err
	}

	toUser.AccountBalance += amount
	err = s.Repo.Update(toUser)
	// if no error, return the updated state of the sender account
	if err == nil {
		return fromUser, nil
	}

	// in case of an error transferring the money, try to put the amount taken from the sender back
	fromUser.AccountBalance += amount
	err = s.Repo.Update(fromUser)
	return nil, err
}
