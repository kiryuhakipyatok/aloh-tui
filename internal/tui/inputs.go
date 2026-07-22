package tui

import (
	"aloh-tui/internal/tui/components/states"
	"sync"

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
		if m.connected {
			m.chatTextInput.Focus()
			m.chatTextInput.PromptStyle = ps
		}
	case states.LOGIN_STATE:
		m.logingInput[m.cursor].Focus()
		m.logingInput[m.cursor].PromptStyle = ps
	case states.PROFILE_STATE:
		if m.sideState == states.RIGHT_STATE && m.cursor < len(m.appereanceInputs) {
			m.appereanceInputs[m.cursor].Focus()
			m.appereanceInputs[m.cursor].PromptStyle = ps
		}
	case states.NICKNAME_STATE:
		m.nicknameInputs[m.cursor].Focus()
		m.nicknameInputs[m.cursor].PromptStyle = ps
	case states.TAGLINE_STATE:
		m.taglineInput.Focus()
		m.taglineInput.PromptStyle = ps
	case states.COLOR_STATE:
		m.colorInput.Focus()
		m.colorInput.PromptStyle = ps
	case states.PASSWORD_STATE:
		m.passwordInputs[m.cursor].Focus()
		m.passwordInputs[m.cursor].PromptStyle = ps
	}
}

func (m *Model) unfocusInputs() {
	var wg sync.WaitGroup

	wg.Go(func() {
		for i := range m.regTextInputs {
			m.regTextInputs[i].Blur()
			m.regTextInputs[i].PromptStyle = lipgloss.NewStyle()
			m.regTextInputs[i].TextStyle = lipgloss.NewStyle()
		}
	})

	wg.Go(func() {
		for i := range m.friendsInputs {
			m.friendsInputs[i].Blur()
			m.friendsInputs[i].PromptStyle = lipgloss.NewStyle()
			m.friendsInputs[i].TextStyle = lipgloss.NewStyle()
		}

	})

	wg.Go(func() {
		for i := range m.logingInput {
			m.logingInput[i].Blur()
			m.logingInput[i].PromptStyle = lipgloss.NewStyle()
			m.logingInput[i].TextStyle = lipgloss.NewStyle()
		}
	})

	wg.Go(func() {
		m.chatTextInput.Blur()
		m.chatTextInput.PromptStyle = lipgloss.NewStyle()
		m.chatTextInput.TextStyle = lipgloss.NewStyle()
	})

	wg.Go(func() {
		for i := range m.appereanceInputs {
			m.appereanceInputs[i].Blur()
			m.appereanceInputs[i].PromptStyle = lipgloss.NewStyle()
			m.appereanceInputs[i].TextStyle = lipgloss.NewStyle()
		}
	})

	wg.Go(func() {
		for i := range m.nicknameInputs {
			m.nicknameInputs[i].Blur()
			m.nicknameInputs[i].PromptStyle = lipgloss.NewStyle()
			m.nicknameInputs[i].TextStyle = lipgloss.NewStyle()
		}
	})

	wg.Go(func() {
		for i := range m.passwordInputs {
			m.passwordInputs[i].Blur()
			m.passwordInputs[i].PromptStyle = lipgloss.NewStyle()
			m.passwordInputs[i].TextStyle = lipgloss.NewStyle()
		}
	})

	wg.Go(func() {
		m.colorInput.Blur()
		m.colorInput.PromptStyle = lipgloss.NewStyle()
		m.colorInput.TextStyle = lipgloss.NewStyle()
	})

	wg.Go(func() {
		m.taglineInput.Blur()
		m.taglineInput.PromptStyle = lipgloss.NewStyle()
		m.taglineInput.TextStyle = lipgloss.NewStyle()
	})

	wg.Wait()

}
