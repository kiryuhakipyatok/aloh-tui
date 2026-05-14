package main

import (
	"aloh-tui/internal/app"
	"aloh-tui/pkg/logger"
)

var (
	//env     string = "dev"
	version string = "1.0.1"
)

func main() {
	app.Run(logger.DevEnv, version)
}
