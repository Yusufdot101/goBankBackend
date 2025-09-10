package app

import (
	"errors"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
	"github.com/Yusufdot101/goBankBackend/internal/permission"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

func (app *Application) GrantPermission(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64  `json:"user_id"`
		Code   string `json:"code"`
	}
	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	userService := &user.Service{
		Repo: &user.Repository{DB: app.DB},
	}
	permissionService := permission.Service{
		Repo:        &permission.Repository{DB: app.DB},
		UserService: userService,
	}

	v := validator.New()
	err = permissionService.GrantUser(v, input.UserID, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrFailedValidation):
			app.FailedValidationResponse(w, v.Errors)

		case errors.Is(err, user.ErrNoRecord):
			app.NotFoundResponse(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated,
		jsonutil.Envelope{
			"message":    "permisison granted",
			"user_id":    input.UserID,
			"permisison": input.Code,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
}

func (app *Application) AddNewPermisison(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code string `json:"code"`
	}
	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, err)
		return
	}

	permissionService := permission.Service{
		Repo: &permission.Repository{DB: app.DB},
	}

	v := validator.New()
	err = permissionService.AddNewPermission(v, input.Code)
	switch {
	}
	if err != nil {
		switch {
		case errors.Is(err, validator.ErrFailedValidation):
			app.FailedValidationResponse(w, v.Errors)

		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(
		w, http.StatusCreated, jsonutil.Envelope{
			"message": "new permisison add",
			"code":    input.Code,
		},
	)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
}
