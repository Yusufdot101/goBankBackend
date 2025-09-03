package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/mailer"
	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	s := user.Service{
		Mailer: mailer.New(
			app.Config.SMTP.Host,
			app.Config.SMTP.Port,
			app.Config.SMTP.Username,
			app.Config.SMTP.Password,
			app.Config.SMTP.Sender,
		),
		Repo: &user.Repository{DB: app.DB},
	}

	v := validator.New()
	u, token, err := s.Register(v, input.Name, input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateEmail):
			v.AddError("email", "user with this email already exists")
			app.FailedValidationResponse(w, v.Errors)

		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !v.IsValid() {
		app.FailedValidationResponse(w, v.Errors)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"user": u, "token": token})
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *Application) ActivateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	s := user.Service{
		Mailer: mailer.New(
			app.Config.SMTP.Host,
			app.Config.SMTP.Port,
			app.Config.SMTP.Username,
			app.Config.SMTP.Password,
			app.Config.SMTP.Sender,
		),
		Repo: &user.Repository{DB: app.DB},
	}

	u, err := s.Activate(input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, token.ErrInvaildToken):
			app.BadRequestResponse(w, err)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated,
		jsonutil.Envelope{
			"message": "account activated successfully",
			"user":    u,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
