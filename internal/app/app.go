package app

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/jsonlog"
	_ "github.com/lib/pq"
)

type Config struct {
	Version           string
	Port              int
	Environment       string
	DailyInterestRate float64
	DB                struct {
		DSN            string
		MaxOpenConns   int
		MaxIdleConns   int
		IdleConnTimout string
	}
	Limiter struct {
		Enabled           bool
		RequestsPerSecond float64
		Burst             int
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

type Application struct {
	Config Config
	Logger *jsonlog.Logger
	DB     *sql.DB
	wg     sync.WaitGroup
}

func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	// set the maximum open (in-use + idle) connections to the database
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)

	// set the maximum idle connections to the database
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	// convert the timeout string to time.Duration type because thats is needed
	duration, err := time.ParseDuration(cfg.DB.IdleConnTimout)
	if err != nil {
		return nil, err
	}

	// set the maximum idle connection time
	db.SetConnMaxIdleTime(duration)

	// create a 5 sec context to test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// test the connection, if a connection is not established in 5 secs, it will raise an error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
