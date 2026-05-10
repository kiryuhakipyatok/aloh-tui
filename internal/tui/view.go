package tui

import (
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/tui/tuierrs"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	cDim       = lipgloss.Color("#75715E")
	
	cAccent    = lipgloss.Color("#FD971F")
	cErr       = lipgloss.Color("#F92672")
	cSubtext   = lipgloss.Color("#66D9EF")

	cText = lipgloss.AdaptiveColor{Light: "#1A1919", Dark: "#F8F8F2"}

	whitePulse = []string{
		"#444444",
		"#666666",
		"#888888",
		"#AAAAAA",
		"#CCCCCC",
		"#FFFFFF",
		"#CCCCCC",
		"#AAAAAA",
		"#888888",
		"#666666",
	}

	blackPulse = []string{
		"#BBBBBB",
		"#999999",
		"#777777",
		"#555555",
		"#333333",
		"#000000",
		"#333333",
		"#555555",
		"#777777",
		"#999999",
	}

	headerInactive = lipgloss.NewStyle().Foreground(cDim).Bold(true)
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m Model) isLoggedIn() bool {
	return m.user != nil && m.user.Networking != nil && m.user.Data.Personal.Nickname != ""
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return m.spinner.View() + "loading..."
	}

	padW, padH := 2, 1
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	if usableW < 60 || usableH < 18 {
		return "terminal too small"
	}

	if m.curWindow == windows.START_WINDOW {
		footer := lipgloss.NewStyle().Foreground(cDim).Width(usableW).Render("ALT+Q: quit")
		uiContent := m.renderStartView(usableW, usableH-lipgloss.Height(footer)-1)
		finalLayout := lipgloss.JoinVertical(lipgloss.Left, uiContent, "", footer)
		return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)))
	}

	rightPart := lipgloss.NewStyle().Foreground(m.themeColor).Render(titles.LL)

	rightW := lipgloss.Width(rightPart)

	footerData := "ALT+Q: quit | TAB/ARRS: switch tabs | ESC: back | ALT+H: help"

	leftPart := lipgloss.NewStyle().
		Foreground(cDim).
		Width(usableW - rightW).
		Align(lipgloss.Left).
		Render(footerData)

	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, leftPart, rightPart)
	footerH := lipgloss.Height(footer)

	logo := lipgloss.NewStyle().Foreground(m.themeColor).Render(titles.A)
	logoH := lipgloss.Height(logo)

	var tabs []string
	if m.isLoggedIn() {
		tabs = m.defTabs
	} else {
		tabs = m.regTabs
	}

	if m.activeTab >= len(tabs) {
		m.activeTab = 0
	}

	gridH := usableH - footerH - logoH
	if gridH < 10 {
		return "terminal too small"
	}

	inactiveBorder := tabBorderWithBottom("┴", "─", "┴")
	activeBorder := tabBorderWithBottom("┘", " ", "└")

	inactiveTabStyle := lipgloss.NewStyle().
		Border(inactiveBorder, true).
		BorderForeground(cDim).
		BorderBottomForeground(cDim).
		Foreground(cDim).
		Align(lipgloss.Center)

	activeTabStyle := lipgloss.NewStyle().
		Border(activeBorder, true).
		BorderForeground(cDim).
		Foreground(m.themeColor).
		Bold(true).
		Align(lipgloss.Center)

	windowStyle := lipgloss.NewStyle().
		BorderForeground(cDim).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()

	var renderedTabs []string
	baseWidth := usableW / len(tabs)
	remainder := usableW % len(tabs)

	for i, t := range tabs {
		isFirst, isLast, isActive := i == 0, i == len(tabs)-1, i == m.activeTab
		style := inactiveTabStyle
		if isActive {
			style = activeTabStyle
		}

		tabWidth := baseWidth
		if i < remainder {
			tabWidth++
		}
		innerWidth := max(0, tabWidth-2)
		style = style.Width(innerWidth)

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		}
		if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}

		styledT := style.Border(border).Render(t)
		var mark string
		if !m.isLoggedIn() {
			switch i {
			case 0:
				mark = "registerT"
			case 1:
				mark = "loginT"
			}
		} else {
			switch i {
			case 0:
				mark = "friendsT"
			case 1:
				mark = "chatT"
			case 2:
				mark = "voiceT"
			case 3:
				mark = "videoT"
			case 4:
				mark = "profileT"
			case 5:
				mark = "settingsT"
			}
		}

		renderedTabs = append(renderedTabs, m.zone.Mark(mark, styledT))
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	windowInnerW := usableW - 4
	windowInnerH := gridH - lipgloss.Height(tabsRow) - 2

	var uiContent string
	switch m.state {
	case states.ERR_STATE:
		uiContent = m.renderErrView(windowInnerW, windowInnerH, cText)
	case states.HELP_STATE:
		uiContent = m.renderHelpView(windowInnerW, windowInnerH, cText)
	default:
		if !m.isLoggedIn() {
			switch m.activeTab {
			case 0:
				uiContent = m.renderRegistrationTab(windowInnerW, windowInnerH)
			case 1:
				uiContent = m.renderLoginTab(windowInnerW, windowInnerH)
			}
		} else {
			switch m.activeTab {
			case 0:
				uiContent = m.renderConnectTab(windowInnerW, windowInnerH)
			case 1:
				uiContent = m.renderChatTab(windowInnerW, windowInnerH, cText)
			case 2:
				uiContent = m.renderVoiceTab(windowInnerW, windowInnerH, cText)
			case 3:
				uiContent = m.renderVideoTab(windowInnerW, windowInnerH)
			case 4:
				uiContent = m.renderProfileView(windowInnerW, windowInnerH, cText)
			case 5:
				uiContent = m.renderSettingsView(windowInnerW, windowInnerH)
			}
		}
	}

	contentBox := windowStyle.Width(usableW - 2).Height(windowInnerH).Render(lipgloss.NewStyle().Padding(0, 1).Render(uiContent))

	tabWindow := lipgloss.JoinVertical(lipgloss.Left, tabsRow, contentBox)

	finalLayout := lipgloss.JoinVertical(lipgloss.Left, "", logo, tabWindow, "", footer)
	screen := lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)

	return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(screen))
}

func (m Model) renderVoiceTab(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► active connections")

	states := make([]string, 0, len(m.connections))

	speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
	mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

	for _, c := range m.connections {
		state := ""
		clearNick := ansi.Strip(c)
		if _, ok := mutedUsers[clearNick]; ok {
			state = " 🔇"
		} else if _, ok := speakingUsers[clearNick]; ok {
			state = " 🔊"
		}
		states = append(states, state)
	}
	rawStatesStr := strings.Join(states, "\n\n\n")
	rawStatesW := lipgloss.Width(rawStatesStr)
	m.connectionsList.SetSize(leftW-rawStatesW-4, h-2)
	connsView := lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().PaddingLeft(2).Render(m.connectionsList.View()),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(rawStatesStr))

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", connsView)
	leftPanel := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► details & controls")
	statusText := "offline"
	statusColor := cDim
	if m.connected {
		statusText = "connected"
		statusColor = m.themeColor
	}

	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight, "",
		safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("network status: ")+lipgloss.NewStyle().Foreground(statusColor).
		Render(statusText), rightW),
		"",
		safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("user management:"), rightW),
		safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.NewStyle().Foreground(m.themeColor).
			Render("ALT+UP/DN")+lipgloss.NewStyle().Foreground(cText).Render(" - adjust user volume")), rightW),
		safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.NewStyle().Foreground(m.themeColor).
			Render("ALT+Z")+lipgloss.NewStyle().Foreground(cText).Render("     - mute/mnmute user")), rightW),
		"",
	)
	if m.connected {
		rightBox = lipgloss.JoinVertical(lipgloss.Left, rightBox,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cErr).Render("ENTER to disconnect"),
		)
	}

	rightPanel := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPanel := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("voiceW", lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, dividerPanel, rightPanel))
}

func (m Model) renderVideoTab(w, h int) string {
	return m.zone.Mark("videoW", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, safeTruncate(lipgloss.NewStyle().Foreground(cDim).Render("~ video functionality coming soon ~"), w)))
}

func (m Model) renderConnectTab(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► online friends")
	m.onlineList.SetSize(leftW-2, h-2)
	onlineView := lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().PaddingLeft(2).Render(m.onlineList.View()))
	if len(m.online) == 0 {
		onlineView = lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().PaddingLeft(2).Render("zero friends online"))
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", lipgloss.NewStyle().PaddingLeft(2).Render(onlineView))
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► connect")
	var rightContent string
	if m.state == states.LOAD_STATE && m.prState == states.CONN_STATE {
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(m.spinner.View() + "connecting to friend...")
	} else if m.connected {
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Foreground(m.themeColor).Render("you are already connected.\ngo to voice tab to manage")
	} else {
		m.connTextInputs.Width = max(1, rightW-4)
		rightContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("enter nickname:"), "",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.connTextInputs.View()),
		)
	}
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", rightContent)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("friendsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderSettingsView(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► app settings")

	m.settingsList.SetSize(rightW-2, h-2)
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.View()))
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► microphones")
	m.microphonesList.SetSize(leftW-2, h-2)
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", lipgloss.NewStyle().PaddingLeft(2).Render(m.microphonesList.View()))
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("settingsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderProfileView(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► profile info")
	leftBox := lipgloss.JoinVertical(lipgloss.Left,
		lblLeft, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("nickname:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(m.user.Data.Personal.Nickname),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("registered:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(m.user.Data.Personal.RegisterTime),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("theme color: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.ThemeColor)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.themeColorInput.View()),
	)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► statistics")
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(safeTruncate("no statistics available", rightW)))
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("profileW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderRegistrationTab(w, h int) string {
	inputPad := lipgloss.NewStyle().PaddingLeft(2).Width(w)
	var content string
	if m.state == states.LOAD_STATE && m.prState == states.REG_STATE {
		content = lipgloss.NewStyle().Foreground(cDim).Render(m.spinner.View() + "processing registration...")
	} else {
		for i := range m.regTextInputs {
			m.regTextInputs[i].Width = max(1, w-10)
		}
		inputs := lipgloss.JoinVertical(lipgloss.Left,
			m.regTextInputs[0].View(), "",
			m.regTextInputs[1].View(), "",
			m.regTextInputs[2].View(),
		)
		content = lipgloss.JoinVertical(lipgloss.Left, m.headerActiveStyle.Render("► create new account"), "", inputs)
	}
	return m.zone.Mark("registerW", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inputPad.Render(content)))
}

func (m Model) renderLoginTab(w, h int) string {
	inputPad := lipgloss.NewStyle().PaddingLeft(2).Width(w)
	var content string
	if m.state == states.LOAD_STATE && (m.prState == states.CONN_STATE || m.prState == states.LOGIN_STATE) {
		content = lipgloss.NewStyle().Foreground(cDim).Render(m.spinner.View() + "login proccesing...")
	} else {
		for i := range m.logingInput {
			m.logingInput[i].Width = max(1, w-10)
		}
		inputs := lipgloss.JoinVertical(lipgloss.Left,
			m.logingInput[0].View(), "",
			m.logingInput[1].View(),
		)
		content = lipgloss.JoinVertical(lipgloss.Left, m.headerActiveStyle.Render("► login"), "", inputs)
	}
	return m.zone.Mark("loginW", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, inputPad.Render(content)))
}

func (m Model) renderChatTab(w, h int, c lipgloss.AdaptiveColor) string {
	m.chatTextInput.Width = max(1, w-4)
	inputView := lipgloss.NewStyle().PaddingLeft(2).Render(m.chatTextInput.View())

	cons := make([]string, 0, len(m.connections))
	speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
	mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

	for _, c := range m.connections {
		state := ""
		clearNick := ansi.Strip(c)
		if _, ok := mutedUsers[clearNick]; ok {
			state = " 🔇"
		} else if _, ok := speakingUsers[clearNick]; ok {
			state = " 🔊"
		}
		cons = append(cons, c+state)
	}

	rawConnStr := strings.Join(cons, "  ·  ")
	if rawConnStr == "" {
		rawConnStr = "no active connections"
	}
	connsView := lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(safeTruncate(rawConnStr, w-2))

	coloredUserNickname := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(m.user.Data.Personal.Nickname)

	usersAudioState := lipgloss.NewStyle().PaddingLeft(2).Render(coloredUserNickname)

	var chatParts []string
	if m.user.Engines.AudioEngine.UserIsSpeaking() {
		usersAudioState = lipgloss.NewStyle().PaddingLeft(2).Render(coloredUserNickname + " 🔊")
	}

	if m.muteState != "" {
		usersAudioState += lipgloss.NewStyle().PaddingLeft(2).Render(m.muteState)
	}
	chatParts = append(chatParts, usersAudioState, connsView, "")

	usedH := 0
	for _, p := range chatParts {
		usedH += lipgloss.Height(p)
	}
	usedH += lipgloss.Height(inputView) + 1

	historyMaxH := max(0, h-usedH)
	var historyBox string

	if historyMaxH > 0 {
		var allMsgsLines []string
		msgStyle := lipgloss.NewStyle().Width(w - 2).MaxWidth(w - 2).PaddingLeft(2)

		for _, msg := range m.messages {
			t := lipgloss.NewStyle().Foreground(cDim).Render(msg.Time)
			n := lipgloss.NewStyle().Foreground(cSubtext).Bold(true).Render(msg.Nickname + ":")

			var renderedMsg string

			if strings.HasPrefix(msg.Text, "\n") {
				txt := msg.Text
				renderedMsg = lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("%s %s %s", t, n, txt))
			} else {
				txt := lipgloss.NewStyle().Foreground(c).Render(msg.Text)
				renderedMsg = msgStyle.Render(fmt.Sprintf("%s %s %s", t, n, txt))
			}
			allMsgsLines = append(allMsgsLines, strings.Split(renderedMsg, "\n")...)
		}

		lenAll := len(allMsgsLines)
		m.chatOffset = min(m.chatOffset, max(0, lenAll-historyMaxH))
		startIdx := max(0, lenAll-historyMaxH-m.chatOffset)
		endIdx := max(0, lenAll-m.chatOffset)

		visibleMsgs := allMsgsLines[startIdx:endIdx]
		historyBox = lipgloss.Place(w, historyMaxH, lipgloss.Left, lipgloss.Bottom, strings.Join(visibleMsgs, "\n"))
	}

	chatParts = append(chatParts, historyBox, "", inputView)
	return m.zone.Mark("chatW", lipgloss.JoinVertical(lipgloss.Left, chatParts...))
}

func (m Model) renderStartView(w, h int) string {
	pulse := whitePulse
	if !lipgloss.HasDarkBackground() {
		pulse = blackPulse
	}

	currentColor := pulse[m.animFrame%len(pulse)]
	titleLogo := m.logoAnim[m.animFrame%len(m.logoAnim)]
	if w < 80 || h < 25 {
		titleLogo = titles.LITTLE_LOGO
	}

	logo := lipgloss.NewStyle().Foreground(m.themeColor).Render(titleLogo)
	msg := lipgloss.NewStyle().Foreground(lipgloss.Color(currentColor)).Bold(true).Render(titles.START)
	msgCentered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(msg)
	content := lipgloss.JoinVertical(lipgloss.Center, logo, "", "", msgCentered)
	return m.zone.Mark("start", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content))
}

func (m Model) renderErrView(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := lipgloss.NewStyle().Foreground(cErr).Bold(true).Render("► error detected")
	errText := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(tuierrs.CastError(m.err))
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", errText)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := lipgloss.NewStyle().Foreground(cErr).Bold(true).Render("► resolution")
	var rightContent string
	if m.user.Networking == nil && m.user.Data.Personal.Nickname != "" {
		m.regTextInputs[1].Width = max(1, rightW-4)
		rightContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("please re-enter credentials"),
		)
	} else {
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("check logs")
	}
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", rightContent)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane)
}

func (m Model) renderHelpView(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► general & navigation")
	lblRight := m.headerActiveStyle.Render("► audio & voice controls")

	renderShortcut := func(keys, desc string) string {
		k := lipgloss.NewStyle().Foreground(m.themeColor).Width(22).Render(keys)
		d := lipgloss.NewStyle().Foreground(cText).Render(desc)
		return safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Render(k+d), leftW)
	}
	sectionLbl := lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim)

	leftRows := []string{
		lblLeft, "",
		sectionLbl.Render("app controls:"),
		renderShortcut("ALT+Q", "- quit application"),
		renderShortcut("ESC", "- go back / close error"),
		renderShortcut("ENTER", "- confirm / connect / send msg"),
		"",
		sectionLbl.Render("navigation:"),
		renderShortcut("TAB / ALT+RIGHT", "- next tab"),
		renderShortcut("SHIFT+TAB / ALT+LEFT", "- previous tab"),
		renderShortcut("UP / DN", "- move cursor in lists/inputs"),
		"",
		sectionLbl.Render("quick jump:"),
		renderShortcut("ALT+C", "- go to connect tab"),
		renderShortcut("ALT+A", "- go to voice tab"),
		renderShortcut("ALT+S", "- go to settings tab"),
		renderShortcut("ALT+H", "- go to / close help menu"),
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, leftRows...)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	rightRows := []string{
		lblRight, "",
		sectionLbl.Render("voice controls:"),
		renderShortcut("ALT+V", "- toggle mic mute"),
		renderShortcut("ALT+B", "- toggle full mute"),
		renderShortcut("ALT+Z", "- toggle user's mute"),
		renderShortcut("ALT+UP/DN", "- increase / decrease user's volume"), "",
		sectionLbl.Render("chat controls:"),
		renderShortcut("CTRL+P", "- paste image"),
	}

	rightBox := lipgloss.JoinVertical(lipgloss.Left, rightRows...)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("profileW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func vertLine(h int) string {
	if h <= 0 {
		return ""
	}
	return strings.Repeat("│\n", h-1) + "│"
}

func safeTruncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	cleanStr := ansi.Strip(s)
	runes := []rune(cleanStr)
	if len(runes) > maxW {
		return string(runes[:maxW-2]) + ".."
	}
	return s
}
