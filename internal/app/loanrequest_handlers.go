package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/loan"
	"github.com/Yusufdot101/goBankBackend/internal/loanrequests"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) NewLoanRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Amount float64 `json:"amount"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	loanRequestService := loanrequests.Service{
		Repo: &loanrequests.Repository{DB: app.DB},
	}

	v := validator.New()
	u := app.getUserContext(r)
	loanRequest, err := loanRequestService.New(v, u, input.Amount, app.Config.DailyInterestRate)
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrFailedValidation):
			app.FailedValidationResponse(w, v.Errors)

		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{
		"message": "your request was sent, we will inform you if it was accepted",
		"loan":    loanRequest,
	})
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *Application) RespondToLoanRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LoanRequestID int64  `json:"loan_request_id"`
		UserID        int64  `json:"user_id"`
		Status        string `json:"status"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	userService := user.Service{
		Repo: &user.Repository{DB: app.DB},
	}
	loanService := loan.Service{
		Repo: &loan.Repository{DB: app.DB},
	}
	loanRequestService := loanrequests.Service{
		Repo:        &loanrequests.Repository{DB: app.DB},
		UserService: &userService,
		LoanService: &loanService,
	}

	var message string
	var loanRequest *loanrequests.LoanRequest
	switch input.Status {
	case "ACCEPTED":
		message = "your loan was accepted"
		loanRequest, err = loanRequestService.AcceptLoanRequest(input.LoanRequestID, input.UserID)
	case "DECLINED":
		message = "your loan was declined"
		loanRequest, err = loanRequestService.DeclineLoanRequest(input.LoanRequestID, input.UserID)
	default:
		return
	}

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
		"message":      message,
		"loan_request": loanRequest,
	})
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
}
