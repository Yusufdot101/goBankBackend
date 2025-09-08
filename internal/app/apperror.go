package app

import (
	"fmt"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
)

// LogError uses the app's loggger to log the error for debugging
func (app *Application) LogError(err error) {
	app.Logger.PrintError(err, nil)
}

// ErrorResponse is a function that writes an error to the response using WriteJSON
func (app *Application) ErrorResponse(w http.ResponseWriter, statusCode int, message any) {
	err := jsonutil.WriteJSON(w, statusCode, jsonutil.Envelope{"error": message})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ServerError is for errors that aren't caused by the client
func (app *Application) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.LogError(err)

	message := "the server encountered and error and could not resolve your request"
	app.ErrorResponse(w, http.StatusInternalServerError, message)
}

func (app *Application) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the resource you requested for could not be found"
	app.ErrorResponse(w, http.StatusNotFound, message)
}

func (app *Application) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not allowed for this resource", r.Method)
	app.ErrorResponse(w, http.StatusMethodNotAllowed, message)
}

func (app *Application) BadRequestResponse(w http.ResponseWriter, err error) {
	app.ErrorResponse(w, http.StatusBadRequest, err.Error())
}

func (app *Application) FailedValidationResponse(w http.ResponseWriter, err map[string]string) {
	app.ErrorResponse(w, http.StatusBadRequest, err)
}

func (app *Application) InvalidCredentialsResponse(w http.ResponseWriter) {
	message := "invaild credentials"
	app.ErrorResponse(w, http.StatusBadRequest, message)
}

func (app *Application) RateLimitExceededResponse(w http.ResponseWriter) {
	message := "rate limit exceeded"
	app.ErrorResponse(w, http.StatusTooManyRequests, message)
}

func (app *Application) InvalidAuthorizationTokenResponse(w http.ResponseWriter) {
	// to let the user know the format required
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "token invalid or missing"
	app.ErrorResponse(w, http.StatusBadRequest, message)
}

func (app *Application) RequireActivatedUserResponse(w http.ResponseWriter) {
	message := "you need an activated account to access this resource"
	app.ErrorResponse(w, http.StatusUnauthorized, message)
}

func (app *Application) RequireAuthorizedUserResponse(w http.ResponseWriter) {
	message := "you need need to be authorized to access this resource"
	app.ErrorResponse(w, http.StatusForbidden, message)
}

func (app *Application) TransferFailedResponse(
	w http.ResponseWriter, statusCode int, reason string,
) {
	message := fmt.Sprintf("transfer failed: %s", reason)
	app.ErrorResponse(w, statusCode, message)
}

func (app *Application) RequirePermissionResponse(w http.ResponseWriter) {
	message := "You do not have the necessary permission to access this resource"
	app.ErrorResponse(w, http.StatusForbidden, message)
}
