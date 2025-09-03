package app

import (
	"fmt"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/jsonutil"
)

// LogError uses the jsonlog to log the error
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

func (app *Application) EditConflictResponse(w http.ResponseWriter) {
	message := "an error occured and your edit could not go through, please try again"

	app.ErrorResponse(w, http.StatusConflict, message)
}

func (app *Application) FailedValidationResponse(w http.ResponseWriter, err map[string]string) {
	app.ErrorResponse(w, http.StatusBadRequest, err)
}
