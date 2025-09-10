package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/permission"
	"github.com/Yusufdot101/goBankBackend/internal/token"
	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// this function will always be run, in case of panics or otherwise
		defer func() {
			// check if panic occured

			if err := recover(); err != nil {
				// recover() returns any type so we normalize it to error type with fmt.Errorf
				app.ServerError(w, r, fmt.Errorf("%s", err))
			}
		}()
		// call the next handler in the chain
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (app *Application) rateLimit(next http.Handler) http.Handler {
	// client will hold client info used in rate limiting so that each IP has its own rate limit
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)

	// do cleanup every minute so that we dont waste resources on IPs that dont vist
	go func() {
		for {
			// after every minute, delete clients that didn't visit in the last 3 mins
			time.Sleep(1 * time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	fn := func(w http.ResponseWriter, r *http.Request) {
		if !app.Config.Limiter.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip := realip.FromRequest(r)
		if _, ok := clients[ip]; !ok {
			mu.Lock()
			clients[ip] = &client{
				limiter: rate.NewLimiter(
					rate.Limit(app.Config.Limiter.RequestsPerSecond), app.Config.Limiter.Burst,
				),
			}
			mu.Unlock()
		}

		// update the lastSeen
		clients[ip].lastSeen = time.Now()

		// if not permitted; rate limit exceeded, send appropriate message and info
		if !clients[ip].limiter.Allow() {
			app.RateLimitExceededResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.setUserContext(r, user.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// we expect the Authorization header to be in the format, "Bearer <token>", so we split
		// the it into two parts
		headParts := strings.Split(authorizationHeader, " ")
		if len(headParts) != 2 || headParts[0] != "Bearer" {
			app.InvalidAuthorizationTokenResponse(w)
			return
		}

		authorizationToken := headParts[1]
		v := validator.New()
		if token.ValidateToken(v, authorizationToken); !v.IsValid() {
			app.InvalidAuthorizationTokenResponse(w)
			return
		}

		s := user.Service{
			Repo: &user.Repository{DB: app.DB},
		}
		// try to get the user for the provided token
		u, err := s.GetUserForToken(authorizationToken, token.ScopeAuthorization)
		if err != nil {
			app.InvalidAuthorizationTokenResponse(w)
			return
		}

		r = app.setUserContext(r, u)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (app *Application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := app.getUserContext(r)
		if !u.Activated {
			app.RequireActivatedUserResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requireAuthorizedUser(fn)
}

func (app *Application) requireAuthorizedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := app.getUserContext(r)
		if u.IsAnonymous() {
			app.RequireAuthorizedUserResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (app *Application) requirePermission(next http.HandlerFunc, code ...string) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := app.getUserContext(r)
		permissionService := permission.Service{
			Repo: &permission.Repository{DB: app.DB},
		}

		v := validator.New()
		for _, c := range code {
			has, err := permissionService.UserHas(v, u, c)
			if err != nil {
				switch {
				case errors.Is(err, validator.ErrFailedValidation):
					app.FailedValidationResponse(w, v.Errors)
				default:
					app.ServerError(w, r, err)
				}
				return
			}
			if has {
				next.ServeHTTP(w, r)
			}
		}

		app.RequirePermissionResponse(w)
	}

	// also needs to be authorized and activated
	return app.requireActivatedUser(fn)
}
