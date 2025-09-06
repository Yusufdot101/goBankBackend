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

	router.HandlerFunc(http.MethodPut, "/v1/loans/get", app.requireActivatedUser(app.NewLoanRequest))

	router.HandlerFunc(http.MethodPut, "/v1/loans/pay", app.requireActivatedUser(app.PayLoan))

	router.HandlerFunc(
		http.MethodPut, "/v1/loans/respond",
		app.requirePermission(app.RespondToLoanRequest, "APPROVE_LOANS", "ADMIN", "SUPERUSER"),
	)

	router.HandlerFunc(
		http.MethodPut, "/v1/loans/delete",
		app.requirePermission(app.DeleteLoan, "DELETE_LOANS", "ADMIN", "SUPERUSER"),
	)

	router.HandlerFunc(
		http.MethodPut, "/v1/permissions/grant",
		app.requirePermission(app.GrantPermission, "SUPERUSER"),
	)

	router.HandlerFunc(
		http.MethodPut, "/v1/permissions/add",
		app.requirePermission(app.AddNewPermisison, "SUPERUSER"),
	)

	router.HandlerFunc(
		http.MethodPut, "/v1/deposit",
		app.requirePermission(app.DepositMoney, "DEPOSIT", "ADMIN", "SUPERUSER"),
	)

	router.HandlerFunc(
		http.MethodPut, "/v1/withdraw",
		app.requirePermission(app.WithdrawMoney, "WITHDDRAW", "ADMIN", "SUPERUSER"),
	)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
