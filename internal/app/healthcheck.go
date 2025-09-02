package app

import (
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
)

func (app *Application) Healthcheck(w http.ResponseWriter, r *http.Request) {
	env := jsonutil.Envelope{
		"status":  "available",
		"version": version,
	}

	err := jsonutil.WriteJSON(w, http.StatusOK, env)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
