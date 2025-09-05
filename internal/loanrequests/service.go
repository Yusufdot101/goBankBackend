package loanrequests

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/loan"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Service struct {
	Repo *Repository
}

func (s *Service) New(
	v *validator.Validator, u *user.User, amount, dailyInterestRate float64,
) (*LoanRequest, error) {
	loanRequest := LoanRequest{
		CreatedAt:         time.Now(),
		UserID:            u.ID,
		Amount:            amount,
		DailyInterestRate: dailyInterestRate,
		Status:            "PENDING",
	}

	if ValidateLoanRequest(v, &loanRequest); !v.IsValid() {
		return nil, nil
	}

	err := s.Repo.Insert(&loanRequest)
	if err != nil {
		return nil, err
	}

	return &loanRequest, nil
}

func (s *Service) AcceptLoanRequest(loanRequestID, userID int64) (*LoanRequest, error) {
	loanRequest, err := s.Repo.UpdateTx(loanRequestID, userID, "ACCEPTED")
	if err != nil {
		return nil, err
	}

	if loanRequest.Status != "PENDING" {
		return nil, user.ErrNoRecord
	}

	// update the user account, add the loan to the account balance
	userService := user.Service{
		Repo: &user.Repository{DB: s.Repo.DB},
	}

	u, err := userService.Repo.Get(userID)
	if err != nil {
		return nil, err
	}

	u.AccountBalance += loanRequest.Amount
	_, err = userService.Repo.UpdateTx(
		userID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	// record the loan on the loans table
	loanService := loan.Service{
		Repo: &loan.Repository{DB: s.Repo.DB},
	}

	err = loanService.GetLoan(u, loanRequest.Amount, loanRequest.DailyInterestRate)
	if err != nil {
		return nil, err
	}

	return loanRequest, nil
}

func (s *Service) DeclineLoanRequest(loanRequestID, userID int64) (*LoanRequest, error) {
	loanRequest, err := s.Repo.UpdateTx(loanRequestID, userID, "DECLINED")
	if err != nil {
		return nil, err
	}

	return loanRequest, nil
}
