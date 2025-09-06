package transaction

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Transaction struct {
	ID          int64
	CreatedAt   time.Time
	UserID      int64
	Action      string
	Amount      float64
	PerformedBy string
}

func ValidateTransaction(v *validator.Validator, transaction *Transaction) {
	v.CheckAddError(transaction.Amount != 0, "amount", "must be given")
	v.CheckAddError(transaction.Amount > 0, "amount", "must be more than 0")

	safeActions := []string{
		"DEPOSIT",
		"WITHDRAW",
	}
	v.CheckAddError(validator.ValueInList(transaction.Action, safeActions...), "action", "invalid")

	v.CheckAddError(transaction.PerformedBy != "", "performed by", "must be given")
}
