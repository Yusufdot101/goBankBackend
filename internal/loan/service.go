package loan

import (
	"math"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Service struct {
	Repo *Repository
}

func (s *Service) GetLoan(
	u *user.User, amount, dailyInterestRate float64,
) error {
	loan := Loan{
		UserID:            u.ID,
		Amount:            amount,
		Action:            "took",
		DailyInterestRate: dailyInterestRate,
		RemainingAmount:   amount,
		LastUpdatedAt:     time.Now(),
	}

	err := s.Repo.Insert(&loan)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) MakePayment(v *validator.Validator, loanID, useID int64, payment float64) (*Loan, error) {
	if payment <= 0 {
		v.AddError("amount", "must be more than 0")
		return nil, nil
	}
	loan, err := s.Repo.GetByID(loanID)
	if err != nil {
		return nil, err
	}

	if loan.RemainingAmount == 0 {
		v.AddError("loan", "payment is completed")
		return nil, nil
	}

	// get the time since last payment was made, we use LastUpdatedAt instead of created_at to
	// avoid charging in partial payments.
	elapsedTimeDays := time.Since(loan.LastUpdatedAt).Hours() / 24
	interest := elapsedTimeDays * (loan.RemainingAmount * (loan.DailyInterestRate / 100))
	totalOwed := loan.RemainingAmount + interest

	loan, err = s.Repo.MakePaymentTx(loan.ID, useID, payment, totalOwed)
	if err != nil {
		return nil, err
	}

	loanPayment := Loan{
		UserID: loan.UserID,
		Amount: math.Min(payment, totalOwed),
		Action: "paid",
	}

	err = s.Repo.Insert(&loanPayment)
	if err != nil {
		return nil, err
	}

	return loan, nil
}

func (s *Service) AcceptLoanRequest(loanID, UserID int64) error {
	return nil
}
