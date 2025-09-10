package transaction

import (
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Repo interface {
	Insert(transaction *Transaction) error
}

type UserService interface {
	GetUser(userID int64) (*user.User, error)
	UpdateUser(
		userID int64, userName, userEmail string, userPasswordHash []byte,
		userAccountBalance float64, userActivated bool,
	) (*user.User, error)
}

type Service struct {
	Repo        Repo
	UserService UserService
}

func (s *Service) Deposit(
	v *validator.Validator, userID int64, amount float64, performedBy string,
) (*Transaction, error) {
	transaction := &Transaction{
		UserID:      userID,
		Amount:      amount,
		Action:      "DEPOSIT",
		PerformedBy: performedBy,
	}
	if ValidateTransaction(v, transaction); !v.IsValid() {
		return nil, validator.ErrFailedValidation
	}

	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return nil, err
	}

	err = s.Repo.Insert(transaction)
	if err != nil {
		return nil, err
	}

	u.AccountBalance += transaction.Amount
	_, err = s.UserService.UpdateUser(
		u.ID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *Service) Withdraw(
	v *validator.Validator, userID int64, amount float64, performedBy string,
) (*Transaction, error) {
	transaction := &Transaction{
		UserID:      userID,
		Amount:      amount,
		Action:      "WITHDRAW",
		PerformedBy: performedBy,
	}
	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return nil, err
	}

	v.CheckAddError(u.AccountBalance >= amount, "account balance", "insufficient funds")
	if ValidateTransaction(v, transaction); !v.IsValid() {
		return nil, validator.ErrFailedValidation
	}

	err = s.Repo.Insert(transaction)
	if err != nil {
		return nil, err
	}

	u.AccountBalance -= transaction.Amount
	_, err = s.UserService.UpdateUser(
		u.ID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
