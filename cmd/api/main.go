package main

import (
	"flag"
	"os"

	"github.com/Yusufdot101/goBankBackend/internal/app"
	"github.com/Yusufdot101/goBankBackend/internal/jsonlog"
)

func main() {
	var config app.Config

	// create command line flags to customize the application at runtime
	flag.IntVar(&config.Port, "addr", 4000, "api server port")

	flag.StringVar(&config.DB.DSN, "db-dsn", os.Getenv("GOBANK_BACKEND_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&config.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(
		&config.DB.IdleConnTimout, "db-idle-conn-timout", "15m",
		"PostgreSQL idle connection timout",
	)

	flag.StringVar(&config.SMTP.Host, "smpt-host", "sandbox.smtp.mailtrap.io", "SMPT host")
	flag.IntVar(&config.SMTP.Port, "smpt-port", 25, "SMPT port")
	flag.StringVar(&config.SMTP.Username, "smpt-username", "3b009b986e9a42", "SMPT username")
	flag.StringVar(&config.SMTP.Password, "smpt-password", "5554cb8d083921", "SMPT password")
	flag.StringVar(&config.SMTP.Sender, "smpt-sender", "ym <noreply@ym.net>", "SMPT sender")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, 0)

	db, err := app.OpenDB(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("Connection to the database established", nil)

	application := &app.Application{
		Config: config,
		Logger: logger,
		DB:     db,
	}

	err = application.Serve()
	if err != nil {
		application.Logger.PrintFatal(err, nil)
	}
}
