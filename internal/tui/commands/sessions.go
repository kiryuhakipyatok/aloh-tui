package commands

import (
	"aloh-tui/internal/networking"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SessionsUpdateMsg struct {
	Time     string
	Sessions []string
	Err      error
}

func FetchSessionsCmd(netw networking.Networking, nickname string) tea.Cmd {
	return func() tea.Msg {
		sessionsUpdateMsg := SessionsUpdateMsg{}
		Sessions, Err := netw.FetchCurrentConnects(nickname)
		if Err != nil {
			sessionsUpdateMsg.Err = Err
		}

		t := time.Now().Format("15:04:05")

		sessionsUpdateMsg.Time = t
		sessionsUpdateMsg.Sessions = Sessions

		return sessionsUpdateMsg

	}
}
