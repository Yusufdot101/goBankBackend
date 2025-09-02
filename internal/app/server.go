package app

import (
	"fmt"
	"net/http"
	"time"
)

func (app *Application) Serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	app.Logger.PrintInfo("server running", map[string]string{"addr": srv.Addr})

	return srv.ListenAndServe()
}
