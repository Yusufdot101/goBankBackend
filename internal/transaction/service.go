package transaction

import (
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Service struct {
	Repo *Repository
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
		return nil, nil
	}

	userService := user.Service{
		Repo: &user.Repository{DB: s.Repo.DB},
	}
	u, err := userService.Repo.Get(userID)
	if err != nil {
		return nil, err
	}

	err = s.Repo.Insert(transaction)
	if err != nil {
		return nil, err
	}

	u.AccountBalance += transaction.Amount
	_, err = userService.Repo.UpdateTx(
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
	if ValidateTransaction(v, transaction); !v.IsValid() {
		return nil, nil
	}

	userService := user.Service{
		Repo: &user.Repository{DB: s.Repo.DB},
	}
	u, err := userService.Repo.Get(userID)
	if err != nil {
		return nil, err
	}

	if u.AccountBalance < amount {
		v.AddError("account balance", "insufficient funds")
		return nil, nil
	}

	err = s.Repo.Insert(transaction)
	if err != nil {
		return nil, err
	}

	u.AccountBalance -= transaction.Amount
	_, err = userService.Repo.UpdateTx(
		u.ID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
