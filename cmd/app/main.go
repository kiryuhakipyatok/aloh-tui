package main

import (
	"aloh-tui/internal/app"
)

var (
	env     string = "local"
	version string = "1.0.0"
)

func main() {
	app.Run(env, version)
}
