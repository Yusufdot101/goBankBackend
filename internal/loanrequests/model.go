package loanrequests

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type LoanRequest struct {
	ID                int64
	CreatedAt         time.Time
	UserID            int64
	Amount            float64
	DailyInterestRate float64
	Status            string
}

func ValidateLoanRequest(v *validator.Validator, loanRequest *LoanRequest) {
	v.CheckAddError(loanRequest.Amount != 0, "amount", "must be given")
	v.CheckAddError(loanRequest.Amount > 0, "amount", "must be more than 0")

	v.CheckAddError(loanRequest.DailyInterestRate >= 0, "dialy interest rate", "cannot be negative")
	// v.CheckAddError(loanRequest.DailyInterestRate != 0, "amount", "must be given")
	// v.CheckAddError(loanRequest.DailyInterestRate >= 0, "amount", "cannot be less than 0")
}
