package tui

import (
	"aloh-tui/internal/tui/components/states"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) focusInputs() {
	m.unfocusInputs()
	ps := lipgloss.NewStyle().Foreground(lipgloss.Color(m.themeColor))
	switch m.state {
	case states.REG_STATE:
		m.regTextInputs[m.cursor].Focus()
		m.regTextInputs[m.cursor].PromptStyle = ps
	case states.CONN_STATE:
		if m.sideState == states.RIGHT_STATE {
			m.friendsInputs[0].Focus()
			m.friendsInputs[0].PromptStyle = ps
		}
	case states.FRIEND_STATE:
		if m.sideState == states.RIGHT_STATE {
			m.friendsInputs[m.cursor].Focus()
			m.friendsInputs[m.cursor].PromptStyle = ps
		}
	case states.CHAT_STATE:
		m.chatTextInput.Focus()
		m.chatTextInput.PromptStyle = ps
	case states.LOGIN_STATE:
		m.logingInput[m.cursor].Focus()
		m.logingInput[m.cursor].PromptStyle = ps
	case states.PROFILE_STATE:
		if m.cursor < len(m.profileInputs) {
			m.profileInputs[m.cursor].Focus()
			m.profileInputs[m.cursor].PromptStyle = ps
		}
	}
}

func (m *Model) unfocusInputs() {
	for i := range m.regTextInputs {
		m.regTextInputs[i].Blur()
		m.regTextInputs[i].PromptStyle = lipgloss.NewStyle()
		m.regTextInputs[i].TextStyle = lipgloss.NewStyle()
	}

	for i := range m.friendsInputs {
		m.friendsInputs[i].Blur()
		m.friendsInputs[i].PromptStyle = lipgloss.NewStyle()
		m.friendsInputs[i].TextStyle = lipgloss.NewStyle()
	}

	for i := range m.logingInput {
		m.logingInput[i].Blur()
		m.logingInput[i].PromptStyle = lipgloss.NewStyle()
		m.logingInput[i].TextStyle = lipgloss.NewStyle()
	}

	m.chatTextInput.Blur()
	m.chatTextInput.PromptStyle = lipgloss.NewStyle()
	m.chatTextInput.TextStyle = lipgloss.NewStyle()

	for i := range m.profileInputs {
		m.profileInputs[i].Blur()
		m.profileInputs[i].PromptStyle = lipgloss.NewStyle()
		m.profileInputs[i].TextStyle = lipgloss.NewStyle()
	}

}
