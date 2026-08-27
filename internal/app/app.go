package app

import (
	"aloh-tui/internal/tui"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/logger"
	l "log"

	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
)

func Run(env, version string) {

	fp, err := utils.SetupFiles()
	if err != nil {
		l.Fatalf("failed to setup files: %v", err)
	}

	null, err := redirectErr(fp.ErrorsLog)

	log := logger.NewLogger(env, fp.AppLog, version)

	model, err := tui.NewModel(fp.NetworkingLog, fp.UserdataFile, fp.KeysDir, log)
	if err != nil {
		l.Fatalf("failed to create model: %v", err)
	}
	if err := clipboard.Init(); err != nil {
		l.Fatalf("failed to init clipboard: %v", err)
	}
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithFPS(30))
	if _, err := p.Run(); err != nil {
		l.Fatalf("failed to run model: %v", err)
	}
	model.Clean()
	if err := null.Close(); err != nil {
		l.Fatalf("failed to close null: %v", err)
	}
}
