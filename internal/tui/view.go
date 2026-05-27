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
	cDim  = lipgloss.Color("#75715E")
	cGray = lipgloss.Color("#5f5f5f")

	cAccent  = lipgloss.Color("#FD971F")
	cErr     = lipgloss.Color("#F92672")
	cSubtext = lipgloss.Color("#66D9EF")

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

	padW, padH := 2, 0
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	if usableW < 76 || usableH < 24 {
		return "terminal too small"
	}

	if m.curWindow == windows.START_WINDOW {
		footer := lipgloss.NewStyle().Foreground(cDim).Width(usableW).Render("ALT+Q: quit")
		uiContent := m.renderStartView(usableW, usableH-lipgloss.Height(footer)-1)
		finalLayout := lipgloss.JoinVertical(lipgloss.Left, uiContent, "", footer)
		return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)))
	}

	rightFooterPart := lipgloss.NewStyle().Foreground(m.themeColor).Render(titles.LL)

	rightFooterW := lipgloss.Width(rightFooterPart)

	footerData := "ALT+Q: quit | TAB/ARRS: switch tabs | ESC: back | ALT+H: help"

	leftFooterPart := lipgloss.NewStyle().
		Foreground(cDim).
		Width(usableW - rightFooterW).
		Align(lipgloss.Left).
		Render(footerData)

	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, leftFooterPart, rightFooterPart)
	footerH := lipgloss.Height(footer)

	rightHeaderData := make([]string, 0, 3)
	curTime := m.curTime
	if m.user.GetShowDateState() {
		curDate := curTime.Format("2006-01-02")
		rightHeaderData = append(rightHeaderData, curDate)
	}

	if m.user.GetShowTimeState() {
		curTime := curTime.Format("15:04:05")
		rightHeaderData = append(rightHeaderData, curTime)
	}

	if m.user.GetShowZoneState() {
		curZone := curTime.Format("-07:00")
		rightHeaderData = append(rightHeaderData, curZone)
	}

	rightHeaderPart := lipgloss.NewStyle().Foreground(cDim).Padding(1, 1, 0, 0).Render(strings.Join(rightHeaderData, " "))
	rightHeaderW := lipgloss.Width(rightHeaderPart)

	leftHeaderPart := lipgloss.NewStyle().Width(usableW-rightHeaderW).Foreground(m.themeColor).Padding(1, 0, 0, 1).Render(titles.A)

	header := lipgloss.JoinHorizontal(lipgloss.Bottom, leftHeaderPart, rightHeaderPart)
	headerH := lipgloss.Height(header)

	var tabs []string
	if m.isLoggedIn() {
		tabs = m.defTabs
	} else {
		tabs = m.regTabs
	}

	if m.activeTab >= len(tabs) {
		m.activeTab = 0
	}

	gridH := usableH - footerH - headerH
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
		if _, ok := m.tabsNotifications[t]; ok && m.user.GetAppNotificationsState() {
			t += fmt.Sprintf(" %s", m.user.Data.Setup.Appereance.NotificaionTag)
		}
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
	windowInnerH := gridH - lipgloss.Height(tabsRow) - 1

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
				uiContent = m.renderFriendsTab(windowInnerW, windowInnerH)
			case 1:
				uiContent = m.renderChatTab(windowInnerW, windowInnerH, cText)
			case 2:
				uiContent = m.renderVoiceTab(windowInnerW, windowInnerH)
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

	finalLayout := lipgloss.JoinVertical(lipgloss.Left, header, tabWindow, footer)
	screen := lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)

	return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(screen))
}

func (m Model) renderVoiceTab(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► active connections")

	states := make([]string, 0, len(m.connections))
	bars := make([]string, 0, len(m.connections))
	speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
	mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

	for _, c := range m.connections {
		state := ""
		clearNick := ansi.Strip(c)
		userColors := m.usersColors[clearNick]
		bar := lipgloss.NewStyle().Foreground(userColors.subColor).Render(fmt.Sprintf("voice power %.1f: ", 0.0))
		if _, ok := mutedUsers[clearNick]; ok {
			state = " 🔇"
		} else if rms, ok := speakingUsers[clearNick]; ok {
			state = " 🔊"
			bar = lipgloss.NewStyle().Foreground(userColors.mainColor).Render(fmt.Sprintf("voice power %.1f:", rms)) +
				lipgloss.NewStyle().Foreground(userColors.subColor).Render(strings.Repeat("▐", int(rms)/100))
		}
		bars = append(bars, bar)
		states = append(states, state)
	}
	rawStatesStr := strings.Join(states, "\n\n\n")
	rawBarsStr := strings.Join(bars, "\n\n\n")
	leftMiddleHeight := h - lipgloss.Height(lblLeft) - 2
	if leftMiddleHeight < 0 {
		leftMiddleHeight = 0
	}
	var leftMiddleSect string
	if !m.connected {
		pulse := whitePulse
		if !lipgloss.HasDarkBackground() {
			pulse = blackPulse
		}
		aloneLogo := m.aloneAnim[m.animFrame%len(m.aloneAnim)]
		currentColor := pulse[m.pulseFrame%len(pulse)]
		textStyle := lipgloss.NewStyle().MaxWidth(leftW).Foreground(lipgloss.Color(currentColor)).Render(aloneLogo)
		leftMiddleSect = lipgloss.Place(
			leftW,
			leftMiddleHeight,
			lipgloss.Center,
			lipgloss.Center,
			textStyle,
		)
	} else {
		rawStatesW := lipgloss.Width(rawStatesStr)
		m.connectionsList.lipList.SetSize(leftW-rawStatesW-4, h-2)
		textStyle := lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(lipgloss.Left, m.connectionsList.lipList.View(), rawStatesStr))
		leftMiddleSect = lipgloss.Place(
			leftW,
			leftMiddleHeight,
			lipgloss.Left,
			lipgloss.Top,
			textStyle,
		)
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", leftMiddleSect)
	leftPanel := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► details")
	voicePowerLbl := ""
	statusText := "not connected"
	statusColor := cDim
	disc := ""
	if m.connected {
		voicePowerLbl = lipgloss.NewStyle().PaddingLeft(2).Width(rightW).Align(lipgloss.Center).Foreground(cText).Render("voice powers")
		statusText = "connected"
		statusColor = m.themeColor
		disc = lipgloss.NewStyle().Foreground(cErr).PaddingRight(1).Render("ENTER to disconnect")
	}

	bottomRightLPart := lipgloss.NewStyle().Foreground(statusColor).Render("status: " + statusText)

	bottomRightLWidth := lipgloss.Width(bottomRightLPart)

	bottomRightRPart := lipgloss.NewStyle().
		Width(rightW - bottomRightLWidth).
		Align(lipgloss.Right).
		Render(disc)

	bottomRight := lipgloss.JoinHorizontal(lipgloss.Bottom, bottomRightLPart, bottomRightRPart)
	rightMiddleHeight := h - lipgloss.Height(lblRight) - lipgloss.Height(bottomRight) - lipgloss.Height(voicePowerLbl)
	if rightMiddleHeight < 0 {
		rightMiddleHeight = 0
	}
	var rightMiddleSect string
	if !m.connected {
		pulse := whitePulse
		if !lipgloss.HasDarkBackground() {
			pulse = blackPulse
		}
		currentColor := pulse[m.pulseFrame%len(pulse)]
		notConnLogo := m.notConnAnim[m.animFrame%len(m.notConnAnim)]
		textStyle := lipgloss.NewStyle().MaxWidth(rightW).Foreground(lipgloss.Color(currentColor)).Render(notConnLogo)
		rightMiddleSect = lipgloss.Place(
			rightW,
			rightMiddleHeight,
			lipgloss.Center,
			lipgloss.Center,
			textStyle,
		)
	} else {
		barsContent := safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Render(rawBarsStr), rightW)

		rightMiddleSect = lipgloss.Place(
			rightW,
			rightMiddleHeight,
			lipgloss.Left,
			lipgloss.Top,
			barsContent,
		)
	}

	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight,
		voicePowerLbl,
		rightMiddleSect,
		bottomRight,
	)

	rightPanel := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPanel := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("voiceW", lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, dividerPanel, rightPanel))
}

func (m Model) renderVideoTab(w, h int) string {
	return m.zone.Mark("videoW", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, safeTruncate(lipgloss.NewStyle().Foreground(cDim).Render("~ video functionality coming soon ~"), w)))
}

func (m Model) renderFriendsTab(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblTopLeft := m.headerActiveStyle.Render("► online friends")

	leftTopContent := lipgloss.NewStyle().PaddingLeft(2).Height(h - 4).Render("zero friends online")
	if len(m.onlineList.lipList.Items()) > 0 {
		m.onlineList.lipList.SetSize(leftW-2, h-4)

		leftTopContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.onlineList.lipList.View())
	}

	leftBotContent := "offline: "

	offline := m.getOfflineUsers()
	if len(offline) > 0 {
		leftBotContent = lipgloss.NewStyle().Foreground(cDim).Render(leftBotContent + strings.Join(m.getOfflineUsers(), " · "))
	} else {
		leftBotContent = lipgloss.NewStyle().Foreground(cDim).Render(leftBotContent + "all friends are online")
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblTopLeft, "", leftTopContent, "", leftBotContent)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblTopRight := m.headerActiveStyle.Render("► connect to friend")

	rightTopH := (h - 1) / 2
	rightBotH := h - rightTopH - 1
	m.friendsInputs[0].Width = max(1, rightW-4)
	var rightTopContent string
	if m.state == states.LOAD_STATE && m.prState == states.CONN_STATE {
		rightTopContent = lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(m.spinner.View() + "connecting to friend...")
	} else if m.connected {
		alr := lipgloss.NewStyle().Foreground(m.themeColor).Render("you're already ")
		curMins := lipgloss.NewStyle().Foreground(cText).Render(fmt.Sprintf("%d min", m.user.GetMinutesInCurrentConenction()))
		withWho := lipgloss.NewStyle().Foreground(m.themeColor).Render(fmt.Sprintf(" in connection with: %s", strings.Join(m.connections, " · ")))
		rightTopContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("enter nickname to:"),
			"",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.friendsInputs[0].View()),
			"", "",
			safeTruncate(lipgloss.NewStyle().PaddingLeft(2).
				Render(alr+curMins+withWho), rightW),
		)
	} else {
		rightTopContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("enter nickname to:"),
			"",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.friendsInputs[0].View()),
		)
	}

	lblBottomRight := m.headerActiveStyle.Render("► manage friends")
	var rightBottomContent string
	if m.state == states.LOAD_STATE && m.prState == states.FRIEND_STATE {
		rightBottomContent = safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(m.spinner.View()+"sending friend request..."), rightW)
	} else {
		inputWidth := max(1, (rightW-4)/2)
		for i := range m.friendsInputs {
			m.friendsInputs[i].Width = inputWidth
		}
		frReq := safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Width(inputWidth).Render(m.friendsInputs[1].View()), inputWidth)
		delFr := safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Width(inputWidth).Render(m.friendsInputs[2].View()), inputWidth)
		bl := safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Width(inputWidth).Render(m.friendsInputs[3].View()), inputWidth)
		unb := safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Width(inputWidth).Render(m.friendsInputs[4].View()), inputWidth)
		if rightBotH > 10 {
			rightBottomContent = lipgloss.JoinVertical(lipgloss.Left,
				safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("enter nickname to:"), rightW),
				"",
				frReq,
				"",
				delFr,
				"",
				bl,
				"",
				unb,
			)
		} else {
			firstLine := lipgloss.JoinHorizontal(lipgloss.Left, frReq, bl)
			secondLine := lipgloss.JoinHorizontal(lipgloss.Left, delFr, unb)
			rightBottomContent = lipgloss.JoinVertical(lipgloss.Left,
				safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("enter nickname to:"), rightW),
				"",
				firstLine,
				"",
				secondLine,
			)
		}

	}

	rightTopBox := lipgloss.JoinVertical(lipgloss.Left, lblTopRight, "", rightTopContent)
	rightTop := lipgloss.Place(rightW, rightTopH, lipgloss.Left, lipgloss.Top, rightTopBox)
	rightBotBox := lipgloss.JoinVertical(lipgloss.Left, lblBottomRight, "", rightBottomContent)
	rightBot := lipgloss.Place(rightW, rightBotH, lipgloss.Left, lipgloss.Top, rightBotBox)
	dividerHorizPane := lipgloss.Place(rightW, 1, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(cDim).Render(horizLine(rightW)))
	rightBox := lipgloss.JoinVertical(lipgloss.Left, rightTop, dividerHorizPane, rightBot)

	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerVertPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))

	return m.zone.Mark("friendsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerVertPane, rightPane))
}

func (m Model) renderSettingsView(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► app settings")

	m.settingsList.lipList.SetSize(leftW-2, h-2)
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.lipList.View()))
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► microphones")
	m.microphonesList.lipList.SetSize(rightW-2, h-2)
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", lipgloss.NewStyle().PaddingLeft(2).Render(m.microphonesList.lipList.View()))
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, lipgloss.NewStyle().Foreground(cDim).Render(vertLine(h)))
	return m.zone.Mark("settingsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderProfileView(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	leftTopH := (2 * (h - 1)) / 3
	lblTopLeft := m.headerActiveStyle.Render("► profile info")

	friends := m.user.GetFriends()
	friendsLen := len(friends)

	leftLTopBoxWidth := 17

	rightStatsW := leftW - leftLTopBoxWidth
	if rightStatsW < 0 {
		rightStatsW = 0
	}
	renderStatisctic := func(title, data string) string {
		if rightStatsW <= 0 {
			return ""
		}

		kw := 27
		if rightStatsW < 35 {
			kw = rightStatsW - 10
			if kw < 5 {
				kw = 5
			}
		}

		kTrunc := title
		if len(title) > kw {
			kTrunc = safeTruncate(title, kw)
		}

		k := lipgloss.NewStyle().Foreground(cDim).Width(kw).Render(kTrunc)
		d := lipgloss.NewStyle().Foreground(cText).Render(data)
		return safeTruncate(lipgloss.NewStyle().PaddingRight(2).Render(k+d), rightStatsW)
	}
	bf := m.user.GetBestFriend()

	leftRTopBox := lipgloss.JoinVertical(lipgloss.Left, "", "",
		renderStatisctic("amount of friends: ", fmt.Sprintf("%d", friendsLen)),
		"",
		renderStatisctic("max time in connection: ", fmt.Sprintf("%d minutes", m.user.GetMaxTimeInConnetion())),
		"",
		renderStatisctic("amount of messages: ", fmt.Sprintf("%d", m.user.GetAmountOfMessages())),
		"",
		renderStatisctic("minutes in connections: ", fmt.Sprintf("%d", m.user.GetAmountOfMinutesInConnections())),
		"",
		renderStatisctic("amount of connections: ", fmt.Sprintf("%d", m.user.GetAmountOfConnections())),
		"",
		renderStatisctic("best friend: ", fmt.Sprintf("%s (%d connections)", bf.Nickname,
			bf.AmountOfConnections)))

	leftRTopBoxH := lipgloss.Height(leftRTopBox)
	leftRTopBox = lipgloss.Place(rightStatsW, leftRTopBoxH, lipgloss.Right, lipgloss.Top, leftRTopBox)

	// leftRTopBoxWidth := lipgloss.Width(leftRTopBox)

	// leftLTopBoxWidth := leftW - leftRTopBoxWidth

	nick := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().PaddingLeft(2).Width(leftLTopBoxWidth).Foreground(cDim).Render("nickname:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(m.user.Data.Personal.Nickname))
	reg := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().PaddingLeft(2).Width(leftLTopBoxWidth).Foreground(cDim).Render("registered:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(m.user.Data.Personal.RegisterTime))
	personalInfo := lipgloss.JoinVertical(lipgloss.Left, nick, "", reg)
	leftLTopBox := lipgloss.JoinVertical(lipgloss.Left, lblTopLeft, "",
		personalInfo,
		// "",
		// lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("theme color: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.ThemeColor)),
		// lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[0].View()),
		// "",
		// lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("best friend tag: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.BestFriendTag)),
		// lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[1].View()),
		// "",
		// lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("notification sign: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.NotificaionSign)),
		// lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[2].View()),
	)

	leftTopBox := lipgloss.JoinHorizontal(lipgloss.Left, leftLTopBox, leftRTopBox)

	actualLeftTopH := lipgloss.Height(leftTopBox)
	if leftTopH < actualLeftTopH {
		leftTopH = actualLeftTopH
	}

	leftTopBox = lipgloss.NewStyle().MaxHeight(leftTopH).Render(leftTopBox)

	lblBotLeft := m.headerActiveStyle.Render("► friend requests")

	leftBotH := h - leftTopH - 1
	if leftBotH < 0 {
		leftBotH = 0
	}
	var leftBotBox string
	//if leftBotH >= 3 {
	var leftBotContent string
	if m.state == states.LOAD_STATE && m.prState == states.PROFILE_STATE && m.sideState == states.LEFT_STATE {
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).Render(m.spinner.View() + "accepting friend request...")
	} else if len(m.friendsReqsList.lipList.Items()) <= 0 {
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).Render("zero friend request")
	} else {
		m.friendsReqsList.lipList.SetSize(leftW-2, leftBotH-2)
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).Render(m.friendsReqsList.lipList.View())
	}
	leftBotBox = lipgloss.JoinVertical(lipgloss.Left, lblBotLeft, "", leftBotContent)
	// } else {
	// 	leftBotBox = lblBotLeft
	// }
	leftBotBox = lipgloss.NewStyle().MaxHeight(leftBotH).Render(leftBotBox)
	leftBot := lipgloss.Place(leftW, leftBotH, lipgloss.Left, lipgloss.Top, leftBotBox)
	leftTop := lipgloss.Place(leftW, leftTopH, lipgloss.Left, lipgloss.Top, leftTopBox)
	dividerHorizPane := lipgloss.Place(leftW, 1, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(cDim).Render(horizLine(leftW)))

	leftBox := lipgloss.JoinVertical(lipgloss.Left, leftTop, dividerHorizPane, leftBot)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)
	lblRight := m.headerActiveStyle.Render("► apereance")

	for i := range m.profileInputs {
		m.profileInputs[i].Width = max(1, rightW-6)
	}
	colors := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("colors")
	tags := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("tags")
	datetimes := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("datetimes")

	rightTopBox := lipgloss.JoinVertical(lipgloss.Left, colors, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("theme color: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.ThemeColor)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[0].View()),
		"", tags, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("best friend tag: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.BestFriendTag)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[1].View()),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("notification sign: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(m.user.Data.Setup.Appereance.NotificaionTag)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.profileInputs[2].View()),
	)

	rightTopH := lipgloss.Height(rightTopBox)
	rightBotH := h - rightTopH - 4
	var rightBotBox string
	//if rightBotH >= 16{
	m.apearenceList.lipList.SetSize(rightW-2, rightBotH-2)
	rightBotBox = lipgloss.NewStyle().PaddingLeft(2).MaxHeight(rightBotH).Render(m.apearenceList.lipList.View())
	// } else {
	// 	rightBotBox = lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("other apreance settings...")
	// }

	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "",
		rightTopBox, "", datetimes, "",
		rightBotBox,
	)

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
	coloredUserNickname := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(m.user.Data.Personal.Nickname)
	usersAudioState := coloredUserNickname

	if m.user.Engines.AudioEngine.UserIsSpeaking() {
		usersAudioState = coloredUserNickname + " 🔊"
	}
	if m.muteState != "" {
		usersAudioState += m.muteState
	}
	topSect := usersAudioState

	chatW := w - 4

	m.chatTextInput.Width = max(1, chatW)

	bottomSect := lipgloss.NewStyle().MaxWidth(chatW).PaddingLeft(2).Render(m.chatTextInput.View())

	middleH := h - lipgloss.Height(topSect) - lipgloss.Height(bottomSect) - 1
	if middleH < 0 {
		middleH = 0
	}

	var middleSect string

	if !m.connected {
		pulse := whitePulse
		if !lipgloss.HasDarkBackground() {
			pulse = blackPulse
		}
		aloneLogo := m.aloneAnim[m.animFrame%len(m.aloneAnim)]
		currentColor := pulse[m.pulseFrame%len(pulse)]

		textStyle := lipgloss.NewStyle().
			MaxWidth(w).
			Foreground(lipgloss.Color(currentColor)).
			Bold(true).
			Render(aloneLogo)

		middleSect = lipgloss.Place(
			w,
			middleH,
			lipgloss.Center,
			lipgloss.Center,
			textStyle,
		)
	} else {
		cons := make([]string, 0, len(m.connections))
		speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
		mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

		for _, c := range m.connections {
			var (
				state string
				bf    string
				res   string
			)
			clearNick := ansi.Strip(c)
			if _, ok := mutedUsers[clearNick]; ok {
				state = " 🔇"
			} else if _, ok := speakingUsers[clearNick]; ok {
				state = " 🔊"
			}

			if clearNick == m.user.Data.Statistics.BestFriend.Nickname {
				bf = m.user.Data.Setup.Appereance.BestFriendTag + " "
			}
			res = bf + c + state
			cons = append(cons, res)
		}

		rawConnStr := lipgloss.NewStyle().Foreground(cDim).Render("with: ") + strings.Join(cons, "  ·  ")

		connView := safeTruncate(lipgloss.NewStyle().Foreground(cDim).Render(rawConnStr), w-2)

		historyMaxH := middleH - lipgloss.Height(connView) - 1
		if historyMaxH < 0 {
			historyMaxH = 0
		}

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

		middleContent := lipgloss.JoinVertical(lipgloss.Left, connView, "", historyBox)
		middleSect = lipgloss.Place(w, middleH, lipgloss.Left, lipgloss.Top, middleContent)
	}

	finalView := lipgloss.JoinVertical(lipgloss.Left, topSect, middleSect, "", bottomSect)

	return m.zone.Mark("chatW", finalView)
}

func (m Model) renderStartView(w, h int) string {
	pulse := whitePulse
	if !lipgloss.HasDarkBackground() {
		pulse = blackPulse
	}

	currentColor := pulse[m.pulseFrame%len(pulse)]
	titleLogo := m.logoAnim[m.animFrame%len(m.logoAnim)]

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
	if !m.isLoggedIn() {
		quote := ""
		switch m.activeTab {
		case 0:
			quote = "please re-enter unique credentials to register"
		case 1:
			quote = "please re-enter your unique credentials to login"
		}
		rightContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(quote),
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

	quote := lipgloss.NewStyle().PaddingLeft(2).Foreground(cGray).Render("all new tabs opens with active left side of window")

	leftRows := []string{
		lblLeft, "", quote, "",
		sectionLbl.Render("app controls:"),
		renderShortcut("ALT+Q", "- quit application"),
		renderShortcut("ESC", "- go back / close error"),
		renderShortcut("ENTER", "- confirm / connect / send msg"),
		"",
		sectionLbl.Render("navigation:"),
		renderShortcut("TAB / RIGHT", "- next tab"),
		renderShortcut("SHIFT+TAB / LEFT", "- previous tab"),
		renderShortcut("ALT+RIGHT", "- right side"),
		renderShortcut("ALT+LEFT", "- left side"),
		renderShortcut("SHIFT+TAB / LEFT", "- previous tab"),
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
		renderShortcut("CTRL+P", "- paste smth"), "",
		sectionLbl.Render("friends contols:"),
		renderShortcut("ALT+X", "- deny friendship request"),
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
func horizLine(w int) string {
	if w <= 0 {
		return ""
	}
	return strings.Repeat("—", w-1) + "—"
}

func safeTruncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	cleanStr := ansi.Strip(s)
	runes := []rune(cleanStr)
	if len(runes) > maxW {
		return lipgloss.NewStyle().Foreground(cErr).Render(string(runes[:maxW-2]) + "..")
	}
	return s
}
