package main

import (
	"aloh-tui/internal/app"
	"aloh-tui/pkg/logger"
	_ "net/http/pprof"
	"os"

	gapp "gioui.org/app"
)

var (
	//env     string = "dev"
	version string = "1.0.1"
)

func main() {
	go func() {
		app.Run(logger.DevEnv, version)
		os.Exit(1)
	}()

	gapp.Main()
}
