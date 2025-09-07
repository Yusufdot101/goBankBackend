package app

import (
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
)

func (app *Application) Healthcheck(w http.ResponseWriter, r *http.Request) {
	env := jsonutil.Envelope{
		"status": "available",
		"app_info": map[string]string{
			"Environment": app.Config.Environment,
			"version":     app.Config.Version,
		},
	}

	err := jsonutil.WriteJSON(w, http.StatusOK, env)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
