package app

import (
	"context"
	"net/http"

	"github.com/Yusufdot101/goBankBackend/internal/user"
)

// contextKey is custom type to avoid conflicts when setting request contexts
type contextKey string

var userContextKey = contextKey("user")

// get the user identity, whether anonymous or real, we panic in case the assertion fails because
// we expect the key to be there by the time this is called
func (app *Application) getUserContext(r *http.Request) *user.User {
	user, ok := r.Context().Value(userContextKey).(*user.User)
	if !ok {
		panic("user key missing in request context")
	}

	return user
}

// store the user, anonymous or otherwise, to the request context, so that other elements have
// access to it
func (app *Application) setUserContext(r *http.Request, u *user.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, u)
	return r.WithContext(ctx)
}
