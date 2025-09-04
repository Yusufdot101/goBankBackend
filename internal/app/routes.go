package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.NotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.MethodNotAllowedResponse)

	// returns application inforamation
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.requireActivatedUser(app.Healthcheck))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.CreateUser)

	router.HandlerFunc(http.MethodPut, "/v1/users/activation", app.ActivateUser)

	// get authorization token for an account
	router.HandlerFunc(http.MethodPut, "/v1/tokens/authorization", app.GetAuthorizationToken)

	router.HandlerFunc(http.MethodPut, "/v1/transfer", app.requireActivatedUser(app.TransferMoney))

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
