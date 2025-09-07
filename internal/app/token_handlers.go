package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/user"
)

func (app *Application) GetAuthorizationToken(w http.ResponseWriter, r *http.Request) {
	// the inputs expected from the client
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	userService := user.Service{Repo: &user.Repository{DB: app.DB}}

	u, err := userService.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNoRecord):
			app.InvalidCredentialsResponse(w)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	// check if the password matches
	mathes, err := u.Password.Matches(input.Password)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	if !mathes {
		app.InvalidCredentialsResponse(w)
		return
	}

	tokenService := token.Service{Repo: &token.Repository{DB: app.DB}}
	t, err := tokenService.AuthorizationToken(u.ID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated,
		jsonutil.Envelope{
			"token":  t.Plaintext,
			"expiry": t.Expiry,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
