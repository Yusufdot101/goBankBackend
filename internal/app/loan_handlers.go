package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/loan"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) PayLoan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LoadID int64   `json:"loan_id"`
		Amount float64 `json:"amount"`
	}
	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	loanService := loan.Service{
		Repo: &loan.Repository{DB: app.DB},
	}

	v := validator.New()
	u := app.getUserContext(r)
	l, err := loanService.MakePayment(v, input.LoadID, u.ID, input.Amount)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNoRecord):
			app.NotFoundResponse(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{
		"message": "loan payment success",
		"loan":    l,
	})
	if err != nil {
		app.ServerError(w, r, err)
	}
}
