package main

import (
	_ "net/http/pprof"
	"aloh-tui/internal/app"
	"aloh-tui/pkg/logger"
	"net/http"
)

var (
	//env     string = "dev"
	version string = "1.0.1"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	app.Run(logger.DevEnv, version)
}