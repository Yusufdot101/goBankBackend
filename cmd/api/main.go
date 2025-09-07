package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Yusufdot101/goBankBackend/internal/app"
	"github.com/Yusufdot101/goBankBackend/internal/jsonlog"
)

// declare the variables. we will use the -X linker flag of the go build to burn-in the
// value at build time
var (
	buildTime string
	version   = "1.0.0"
)

func main() {
	config := app.Config{
		Version:     version,
		Environment: os.Getenv("ENV"),
	}

	// create command line flags to customize the application at runtime
	flag.IntVar(&config.Port, "addr", 4000, "API server port")
	flag.Float64Var(&config.DailyInterestRate, "interest-rate", 5, "Bank daily interest rate")

	flag.StringVar(&config.DB.DSN, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&config.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(
		&config.DB.IdleConnTimout, "db-idle-conn-timout", "15m",
		"PostgreSQL idle connection timout",
	)

	flag.Float64Var(
		&config.Limiter.RequestsPerSecond, "limiter-rps", 2,
		"Rate limiter maximum requests per second",
	)
	flag.IntVar(&config.Limiter.Burst, "limiter-burst", 4, "Rate limiter burst")
	flag.BoolVar(&config.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")

	displayVersion := flag.Bool("version", false, "Display application version and exit")
	flag.Parse()
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

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
