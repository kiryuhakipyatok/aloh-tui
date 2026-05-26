package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	TickMsg       time.Time
	AnimTickMsg   time.Time
	PulseTickMsg  time.Time
	UpdateTickMsg time.Time
)

func TickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*1000, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AnimTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*210, func(t time.Time) tea.Msg {
		return AnimTickMsg(t)
	})
}

func PulseTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*140, func(t time.Time) tea.Msg {
		return PulseTickMsg(t)
	})
}

func TimeTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}