package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time

type AnimTickMsg time.Time

type UpdateTickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*2000, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AnimTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
		return AnimTickMsg(t)
	})
}

func UpdateTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return UpdateTickMsg(t)
	})
}
