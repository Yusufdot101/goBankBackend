package main

import (
	"flag"
	"os"

	"github.com/Yusufdot101/goBankBackend/internal/app"
	"github.com/Yusufdot101/goBankBackend/internal/jsonlog"
)

func main() {
	var config app.Config

	flag.IntVar(&config.Port, "addr", 4000, "api server port")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, 0)

	application := &app.Application{
		Config: config,
		Logger: logger,
	}

	err := application.Serve()
	if err != nil {
		application.Logger.PrintFatal(err, nil)
	}
}
