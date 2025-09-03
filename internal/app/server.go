package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// channel to hold the error, if an error occured durinng shutdown
	shutdownError := make(chan error)
	go func() {
		// listen for the terminate and interrupt signals and put them in the quit channel
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		// this will hold till something is in the quit channel
		s := <-quit

		app.Logger.PrintInfo("server shutting down", map[string]string{
			"addr":   srv.Addr,
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.PrintInfo("finishing background tasks", nil)
		app.wg.Wait()
		shutdownError <- err
	}()
	app.Logger.PrintInfo("server running", map[string]string{"addr": srv.Addr})

	err := srv.ListenAndServe()
	// if the error is http.ErrServerClosed, it means the shutdown worked
	if err != http.ErrServerClosed {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.PrintInfo("stopped server", nil)
	return nil
}
