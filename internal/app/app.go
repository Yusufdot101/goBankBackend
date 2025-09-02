package app

import (
	"github.com/Yusufdot101/goBankBackend/internal/jsonlog"
)

const version = "1.0.0"

type Config struct {
	Port int
}

type Application struct {
	Config Config
	Logger *jsonlog.Logger
}
