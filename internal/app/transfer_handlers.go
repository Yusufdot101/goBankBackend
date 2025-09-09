package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/transfer"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) TransferMoney(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ToEmail string  `json:"to_email"`
		Amount  float64 `json:"amount"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	userService := user.Service{
		Repo: &user.Repository{DB: app.DB},
	}

	transferService := transfer.Service{
		Repo:        &transfer.Repository{DB: app.DB},
		UserService: &userService,
	}

	fromUser := app.getUserContext(r)
	v := validator.New()
	tr, fromUser, err := transferService.TransferMoney(v, fromUser, input.ToEmail, input.Amount)
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrFailedValidation):
			app.FailedValidationResponse(w, v.Errors)

		case errors.Is(err, user.ErrNoRecord):
			app.TransferFailedResponse(w, http.StatusNotFound, "to email not found")

		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{
		"message":  "money transferred successfuly",
		"transfer": tr,
		"user":     fromUser,
	})
	if err != nil {
		app.ServerError(w, r, err)
	}
}
