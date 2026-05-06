package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time

type AnimTickMsg time.Time

type UpdateTickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*1500, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AnimTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*175, func(t time.Time) tea.Msg {
		return AnimTickMsg(t)
	})
}

func UpdateTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*300, func(t time.Time) tea.Msg {
		return UpdateTickMsg(t)
	})
}