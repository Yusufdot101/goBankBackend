package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.NotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.MethodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.Healthcheck)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.CreateUser)

	router.HandlerFunc(http.MethodPut, "/v1/users/activation", app.ActivateUser)

	return router
}
