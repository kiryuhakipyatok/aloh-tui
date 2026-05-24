package tui

import (
	"aloh-tui/internal/tui/components/titles"
	"fmt"
	"maps"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) getChatSizes() (int, int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0, 0
	}

	padW, padH := 2, 1
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	footerH := lipgloss.Height("X")
	logo := lipgloss.NewStyle().Render(titles.A)
	logoH := lipgloss.Height(logo)

	gridH := usableH - footerH - logoH
	activeBorder := tabBorderWithBottom("┘", " ", "└")

	activeTabStyle := lipgloss.NewStyle().
		Border(activeBorder, true).
		Bold(true).
		Align(lipgloss.Center)

	tab := m.defTabs[0]
	style := activeTabStyle
	border, _, _, _, _ := style.GetBorder()
	border.BottomLeft = "│"
	border.BottomRight = "│"
	styledT := style.Border(border).Render(tab)
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, styledT)
	w := usableW - 4
	h := gridH - lipgloss.Height(tabsRow) - 2

	m.chatTextInput.Width = max(1, w-4)
	inputView := lipgloss.NewStyle().PaddingLeft(2).Render(m.chatTextInput.View())

	rawConnStr := "X"
	connsView := lipgloss.NewStyle().PaddingLeft(2).Render(safeTruncate(rawConnStr, w-2))

	usersAudioState := "X"

	var chatParts []string

	chatParts = append(chatParts, usersAudioState, connsView, "")

	usedH := 0
	for _, p := range chatParts {
		usedH += lipgloss.Height(p)
	}
	usedH += lipgloss.Height(inputView) + 1

	historyMaxH := max(0, h-usedH)

	var maxOffset int
	if historyMaxH > 0 {
		var allMsgsLines []string
		msgStyle := lipgloss.NewStyle().Width(w - 2).MaxWidth(w - 2).PaddingLeft(2)

		for _, msg := range m.messages {
			t := lipgloss.NewStyle().Render(msg.Time)
			n := lipgloss.NewStyle().Bold(true).Render(msg.Nickname + ":")
			txt := lipgloss.NewStyle().Render(msg.Text)
			renderedMsg := msgStyle.Render(fmt.Sprintf("%s %s %s", t, n, txt))
			allMsgsLines = append(allMsgsLines, strings.Split(renderedMsg, "\n")...)
		}

		lenAll := len(allMsgsLines)

		maxOffset = lenAll - historyMaxH
	}

	return maxOffset, w, historyMaxH
}

func (m *Model) coloredNickname(nickname string) string {
	for _, v := range m.connections {
		if nickname == ansi.Strip(v) {
			nickname = v
		}
	}
	return nickname
}

func (m *Model) updateFreindsReqs(nickname string) (tea.Model, tea.Cmd) {
	for i, fr := range m.friendsReqs {
		if fr.Nickname == nickname {
			m.friendsReqs = append(m.friendsReqs[:i], m.friendsReqs[i+1:]...)
			break
		}
	}
	cmd := m.updateFriendsReqList()
	return m, cmd
}

func isEqualOnline(online1, online2 map[string][]string) bool {
	return maps.EqualFunc(online1, online2, func(arr1 []string, arr2 []string) bool {
		slices.Sort(arr1)
		slices.Sort(arr2)
		return slices.Equal(arr1, arr2)
	})
}
