package app

import (
	"aloh-tui/internal/tui"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/logger"
	"io"
	l "log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
)

func Run(env, version string) {
	null, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	os.Stderr = null

	l.SetOutput(io.Discard)

	netwLogFilePath, appLogFilePath, nickFilePath, keysPath, err := utils.SetupFiles()
	if err != nil {
		l.Fatalf("failed to setup files: %v", err)
	}

	log := logger.NewLogger(env, appLogFilePath, version)

	model, err := tui.NewModel(netwLogFilePath, nickFilePath, keysPath, log)
	if err != nil {
		l.Fatalf("failed to create model: %v", err)
	}
	if err := clipboard.Init(); err != nil {
		l.Fatalf("failed to init clipboard: %v", err)
	}
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithFPS(120))
	if _, err := p.Run(); err != nil {
		l.Fatalf("failed to run model: %v", err)
	}
	model.Clean()
	if err := null.Close(); err != nil {
		l.Fatalf("failed to close null: %v", err)
	}
}
