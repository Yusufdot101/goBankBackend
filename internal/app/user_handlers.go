package app

import (
	"errors"
	"fmt"
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

	tokenService := token.Service{Repo: &token.Repository{DB: app.DB}}
	userService := user.Service{
		Mailer:       mailer.NewMailerFromEnv(),
		Repo:         &user.Repository{DB: app.DB},
		TokenService: &tokenService,
	}

	v := validator.New()
	u, token, err := userService.Register(v, input.Name, input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrFailedValidation):
			app.FailedValidationResponse(w, v.Errors)

		case errors.Is(err, user.ErrDuplicateEmail):
			v.AddError("email", "user with this email already exists")
			app.FailedValidationResponse(w, v.Errors)

		default:
			app.ServerError(w, r, err)
		}
		return
	}

	// send the email to the user the token
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.ServerError(w, r, fmt.Errorf("%s", err))
			}
		}()
		data := map[string]any{
			"userName": u.Name,
			"userID":   u.ID,
			"token":    token.Plaintext,
		}
		_ = userService.Mailer.Send(u.Email, "user_welcome.html", data)
	}()

	err = jsonutil.WriteJSON(
		w, http.StatusAccepted, jsonutil.Envelope{
			"message": "account created successfully, please follow the instructions sent to your email to activate your account",
			"user":    u,
		},
	)
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

	v := validator.New()
	if token.ValidateToken(v, input.TokenPlaintext); !v.IsValid() {
		app.FailedValidationResponse(w, v.Errors)
		return
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
