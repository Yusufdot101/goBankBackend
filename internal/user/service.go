package user

import (
	"errors"
	"sync"
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
	v *validator.Validator, name, email, passwordPlaintext string, deferredFunc func(), wg *sync.WaitGroup,
) (*User, error) {
	user := &User{
		Name:  name,
		Email: email,
	}
	err := user.Password.Set(passwordPlaintext)
	if err != nil {
		return nil, err
	}

	if ValidateUser(v, user); !v.IsValid() {
		return nil, nil
	}

	err = s.Repo.Insert(user)
	if err != nil {
		return nil, err
	}

	// get the activation token and send it to the user
	tokenService := token.Service{Repository: &token.Repository{DB: s.Repo.DB}}
	t, err := tokenService.New(user.ID, 3*24*time.Hour, token.ScopeActivation)

	// send the email with the account activation, is async because we dont want to wait for it as
	// it might take a long time
	func() {
		wg.Add(1)
		defer wg.Done()
		go func() {
			defer deferredFunc()
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
	}()

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

	tokenService := token.Service{Repository: &token.Repository{DB: s.Repo.DB}}
	err = tokenService.Repository.DeleteAllForUser(u.ID, token.ScopeActivation)
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
