package loanrequests

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Repo interface {
	Insert(loanRequest *LoanRequest) error
	Get(loanRequestID, userID int64) (*LoanRequest, error)
	UpdateTx(loanRequestID, userID int64, newStatus string) (*LoanRequest, error)
}

type UserService interface {
	GetUser(userID int64) (*user.User, error)
	UpdateUser(
		userID int64, userName, userEmail string, userPasswordHash []byte,
		userAccountBalance float64, userActivated bool,
	) (*user.User, error)
}

type LoanService interface {
	GetLoan(u *user.User, amount, dailyInterestRate float64) error
}

type Service struct {
	Repo        Repo
	UserService UserService
	LoanService LoanService
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
		return nil, validator.ErrFailedValidation
	}

	err := s.Repo.Insert(&loanRequest)
	if err != nil {
		return nil, err
	}

	return &loanRequest, nil
}

func (s *Service) AcceptLoanRequest(loanRequestID, userID int64) (*LoanRequest, error) {
	loanRequest, err := s.Repo.Get(loanRequestID, userID)
	if err != nil {
		return nil, err
	}

	if loanRequest.Status != "PENDING" {
		return nil, user.ErrNoRecord
	}

	loanRequest, err = s.Repo.UpdateTx(loanRequestID, userID, "ACCEPTED")
	if err != nil {
		return nil, err
	}

	// update the user account, add the loan to the account balance
	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return nil, err
	}

	u.AccountBalance += loanRequest.Amount
	_, err = s.UserService.UpdateUser(
		userID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	// record the loan on the loans table
	err = s.LoanService.GetLoan(u, loanRequest.Amount, loanRequest.DailyInterestRate)
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
