package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/mailer"
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
		Mailer: mailer.New(
			app.Config.SMTP.Host,
			app.Config.SMTP.Port,
			app.Config.SMTP.Username,
			app.Config.SMTP.Password,
			app.Config.SMTP.Sender,
		),
		Repo: &user.Repository{DB: app.DB},
	}

	transferService := transfer.Service{
		Repo: &transfer.Repository{DB: app.DB},
	}

	fromUser := app.getUserContext(r)
	v := validator.New()
	tr, fromUser, err := transferService.TransferMoney(
		v, &userService, fromUser, input.ToEmail, input.Amount,
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNoRecord):
			app.TransferFailedResponse(w, http.StatusNotFound, "to email not found")
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !v.IsValid() {
		app.FailedValidationResponse(w, v.Errors)
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
