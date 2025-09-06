package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/transaction"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) DepositMoney(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID      int64   `json:"user_id"`
		Amount      float64 `json:"amount"`
		PerformedBy string  `json:"performed_by"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	transactionService := transaction.Service{
		Repo: &transaction.Repository{DB: app.DB},
	}
	tr, err := transactionService.Deposit(
		v, input.UserID, input.Amount, input.PerformedBy,
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNoRecord):
			app.NotFoundResponse(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !v.IsValid() {
		app.FailedValidationResponse(w, v.Errors)
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated, jsonutil.Envelope{
			"message":     "transaction completed successfully",
			"transaction": tr,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *Application) WithdrawMoney(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID      int64   `json:"user_id"`
		Amount      float64 `json:"amount"`
		PerformedBy string  `json:"performed_by"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	transactionService := transaction.Service{
		Repo: &transaction.Repository{DB: app.DB},
	}
	tr, err := transactionService.Withdraw(
		v, input.UserID, input.Amount, input.PerformedBy,
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNoRecord):
			app.NotFoundResponse(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !v.IsValid() {
		app.FailedValidationResponse(w, v.Errors)
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated, jsonutil.Envelope{
			"message":     "transaction completed successfully",
			"transaction": tr,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
