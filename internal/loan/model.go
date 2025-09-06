package loan

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Loan struct {
	ID                int64
	CreatedAt         time.Time
	UserID            int64
	Amount            float64
	Action            string
	DailyInterestRate float64
	RemainingAmount   float64
	LastUpdatedAt     time.Time
	Version           int32
}

type LoanDeletion struct {
	ID                int64
	CreatedAt         time.Time
	LoanCreatedAt     time.Time
	LoanLastUpdatedAt time.Time
	LoanID            int64
	DebtorID          int64
	DeletedByID       int64
	Amount            float64
	DailyInterestRate float64
	RemainingAmount   float64
	Reason            string
}

func ValidateLoan(v *validator.Validator, loan *Loan) {
	v.CheckAddError(loan.Amount != 0, "amount", "must be given")
	v.CheckAddError(loan.Amount > 0, "amount", "must be more than 0")

	// v.CheckAddError(loan.DailyInterestRate != 0, "amount", "must be given")
	// v.CheckAddError(loan.DailyInterestRate >= 0, "amount", "cannot be less than 0")

	safeActions := []string{"took", "paid"}
	v.CheckAddError(validator.ValueInList(loan.Action, safeActions...), "action", "invaild")
}

func ValidateLoanDeletion(v *validator.Validator, loanDeletion *LoanDeletion) {
	v.CheckAddError(loanDeletion.Reason != "", "reason", "must be given")
}
