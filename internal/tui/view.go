package tui

import (
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/tui/tuierrs"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m Model) isLoggedIn() bool {
	return m.user != nil && m.user.Networking != nil && m.user.Data.Identity.ID != uuid.Nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 || (!m.isLoggedIn() && m.state == states.LOAD_STATE) {
		return lipgloss.NewStyle().Foreground(m.themeColor).
			Render(fmt.Sprintf("%s%s", m.spinner.View(), "loading..."))
	}

	padW, padH := 2, 0
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	if usableW < 76 || usableH < 24 {
		return "terminal too small"
	}

	if m.curWindow == windows.START_WINDOW {
		footer := styles.CDimStyle.Width(usableW).Render("ALT+Q: quit")
		uiContent := m.renderStartView(usableW, usableH-lipgloss.Height(footer)-1)
		finalLayout := lipgloss.JoinVertical(lipgloss.Left, uiContent, "", footer)
		return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)))
	}

	rightFooterPart := lipgloss.NewStyle().Foreground(m.themeColor).Render(titles.LL)

	rightFooterW := lipgloss.Width(rightFooterPart)

	footerData := "ALT+Q: quit | TAB/ARRS: switch tabs | ESC: back | ALT+H: help"

	leftFooterPart := styles.CDimStyle.Width(usableW - rightFooterW).
		Align(lipgloss.Left).
		Render(footerData)

	footer := lipgloss.JoinHorizontal(lipgloss.Bottom, leftFooterPart, rightFooterPart)
	footerH := lipgloss.Height(footer)

	userIden := m.user.GetUserIdentity()

	rightHeaderPart := styles.CDimStyle.Padding(1, 1, 0, 0).
		Render(strings.TrimSpace(strings.Join(m.rightHeaderData, " ")) + " " + userIden.Nickname)
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

	inactiveTabStyle := styles.InactiveTabStyle.Border(inactiveBorder, true).Align(lipgloss.Center)

	activeTabStyle := styles.ActiveTabStyle.Border(activeBorder, true).Foreground(m.themeColor)

	var renderedTabs []string
	baseWidth := usableW / len(tabs)
	remainder := usableW % len(tabs)

	for i, t := range tabs {
		if _, ok := m.tabsNotifications[t]; ok && m.user.GetAppNotificationsState() {
			t += fmt.Sprintf(" %s", m.user.GetNotificationTag())
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
		uiContent = m.renderErrView(windowInnerW, windowInnerH)
	case states.HELP_STATE:
		uiContent = m.renderHelpView(windowInnerW, windowInnerH)
	default:
		if !m.isLoggedIn() && m.state != states.LOAD_STATE {
			switch m.activeTab {
			case 0:
				uiContent = m.renderRegistrationTab(windowInnerW, windowInnerH)
			case 1:
				uiContent = m.renderLoginTab(windowInnerW, windowInnerH)
			}
		} else if m.isLoggedIn() {
			switch m.activeTab {
			case 0:
				uiContent = m.renderFriendsTab(windowInnerW, windowInnerH)
			case 1:
				uiContent = m.renderChatTab(windowInnerW, windowInnerH, styles.CText)
			case 2:
				uiContent = m.renderVoiceTab(windowInnerW, windowInnerH)
			case 3:
				uiContent = m.renderVideoTab(windowInnerW, windowInnerH)
			case 4:
				uiContent = m.renderProfileView(windowInnerW, windowInnerH, styles.CText)
			case 5:
				uiContent = m.renderSettingsView(windowInnerW, windowInnerH)
			}
		}
	}

	contentBox := styles.WindowStyle.Width(usableW - 2).Height(windowInnerH).Render(lipgloss.NewStyle().Padding(0, 1).Render(uiContent))

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
		userColors := m.usersColors[c.ID]
		mainStyle := lipgloss.NewStyle().Foreground(userColors.MainColor)
		subStyle := lipgloss.NewStyle().Foreground(userColors.SubColor)
		bar := lipgloss.NewStyle().Foreground(userColors.SubColor).Render(fmt.Sprintf("voice power %.1f: ", 0.0))
		if _, ok := mutedUsers[c.ID]; ok {
			state = " 🔇"
		} else if rms, ok := speakingUsers[c.ID]; ok {
			state = " 🔊"
			bar = mainStyle.Render(fmt.Sprintf("voice power %.1f:", rms)) +
				subStyle.Render(strings.Repeat("▐", int(rms)/100))
		} else if us, ok := m.usersStates[c.ID]; ok {
			if us.fullMute {
				state = " 🙊🙉"
			} else if us.micMute {
				state = " 🙊"
			}
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
		textStyle := styles.CGrayBold.Render(titles.ALONE1)
		leftMiddleSect = lipgloss.Place(
			leftW,
			leftMiddleHeight-1,
			lipgloss.Center,
			lipgloss.Center,
			textStyle,
		)
	} else {
		rawStatesW := lipgloss.Width(rawStatesStr)
		m.connectionsList.DefList.LipList.SetSize(leftW-rawStatesW-4, h-2)
		textStyle := styles.PaddingLeftStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, m.connectionsList.DefList.LipList.View(), rawStatesStr))
		leftMiddleSect = lipgloss.Place(
			leftW,
			leftMiddleHeight,
			lipgloss.Left,
			lipgloss.Top,
			textStyle,
		)
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", leftMiddleSect)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := m.headerActiveStyle.Render("► details")
	voicePowerLbl := ""
	statusText := "not connected"
	statusColor := styles.CDim
	disc := ""
	if m.connected {
		voicePowerLbl = styles.PaddingLeftCDimStyle.Width(rightW).Align(lipgloss.Center).Render("voice powers")
		statusText = "connected"
		statusColor = m.themeColor
		disc = styles.CErrPaddingStyle.Render("ENTER to disconnect")
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
		textStyle := styles.CGrayBold.Render(titles.NOT_CONN1)
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

	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)
	rightPane = m.zone.Mark("rightSide", rightPane)
	leftPane = m.zone.Mark("leftSide", leftPane)
	dividerPanel := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))
	return m.zone.Mark("voiceW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPanel, rightPane))
}

func (m Model) renderVideoTab(w, h int) string {
	return m.zone.Mark("videoW", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, safeTruncate(styles.CDimStyle.Render("~ video functionality coming soon ~"), w)))
}

func (m Model) renderFriendsTab(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblTopLeft := m.headerActiveStyle.Render("► friends states")

	m.friendsList.DefList.LipList.SetSize(leftW-2, h-4)

	leftTopContent := lipgloss.NewStyle().PaddingLeft(2).Render(m.friendsList.DefList.LipList.View())

	on := len(m.online)
	of := len(m.user.Data.Personal.Friends) - on
	leftBotContent := styles.CDimStyle.Render(fmt.Sprintf("online: %d, offline: %d", on, of))

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblTopLeft, "", leftTopContent, "", leftBotContent)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblTopRight := m.headerActiveStyle.Render("► connect to friend")

	rightTopH := (h - 1) / 2
	rightBotH := h - rightTopH - 1
	m.friendsInputs[0].Width = max(1, rightW-4)
	var rightTopContent string
	if m.state == states.LOAD_STATE && m.prState == states.CONN_STATE {
		rightTopContent = styles.PaddingLeftCDimStyle.Render(m.spinner.View() + "connecting to friend...")
	} else if m.connected {
		alr := lipgloss.NewStyle().Foreground(m.themeColor).Render("you're ")
		curMins := styles.CTextStyle.Render(fmt.Sprintf("%d min", m.user.GetMinutesInCurrentConenction()))
		withWho := lipgloss.NewStyle().Foreground(m.themeColor).
			Render(fmt.Sprintf(" in connection with: %s", strings.Join(m.connecctionsNicks, " · ")))
		rightTopContent = lipgloss.JoinVertical(lipgloss.Left,
			styles.PaddingLeftCDimStyle.Render("enter nickname to:"),
			"",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.friendsInputs[0].View()),
			"",
			safeTruncate(lipgloss.NewStyle().PaddingLeft(2).
				Render(alr+curMins+withWho), rightW),
		)
	} else {
		rightTopContent = lipgloss.JoinVertical(lipgloss.Left,
			styles.PaddingLeftCDimStyle.Render("enter nickname to:"),
			"",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.friendsInputs[0].View()),
		)
	}

	lblBottomRight := m.headerActiveStyle.Render("► manage friends")
	var rightBottomContent string
	if m.state == states.LOAD_STATE && m.prState == states.FRIEND_STATE {
		switch m.cursor {
		case 1:
			rightBottomContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"sending friend request..."), rightW)
		case 2:
			rightBottomContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"deleting user from friends..."), rightW)
		case 3:
			rightBottomContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"blocking user..."), rightW)
		case 4:
			rightBottomContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"unblocking user..."), rightW)
		}
	} else {
		inputWidth := max(1, (rightW-4)/2)
		for i := range m.friendsInputs {
			m.friendsInputs[i].Width = inputWidth
		}
		frReq := safeTruncate(styles.PaddingLeftStyle.Width(inputWidth).Render(m.friendsInputs[1].View()), inputWidth)
		delFr := safeTruncate(styles.PaddingLeftStyle.Width(inputWidth).Render(m.friendsInputs[2].View()), inputWidth)
		bl := safeTruncate(styles.PaddingLeftStyle.Width(inputWidth).Render(m.friendsInputs[3].View()), inputWidth)
		unb := safeTruncate(styles.PaddingLeftStyle.Width(inputWidth).Render(m.friendsInputs[4].View()), inputWidth)
		if rightBotH > 10 {
			rightBottomContent = lipgloss.JoinVertical(lipgloss.Left,
				safeTruncate(styles.PaddingLeftCDimStyle.Render("enter nickname to:"), rightW),
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
				safeTruncate(styles.PaddingLeftCDimStyle.Render("enter nickname to:"), rightW),
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
	dividerHorizPane := lipgloss.Place(rightW, 1, lipgloss.Center, lipgloss.Center, styles.CDimStyle.Render(horizLine(rightW)))
	rightBox := lipgloss.JoinVertical(lipgloss.Left, rightTop, dividerHorizPane, rightBot)

	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerVertPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))
	rightPane = m.zone.Mark("rightSide", rightPane)
	leftPane = m.zone.Mark("leftSide", leftPane)
	return m.zone.Mark("friendsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerVertPane, rightPane))
}

func (m Model) renderSettingsView(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	var (
		lblLeft     string
		leftContent string

		lblRight     string
		rightContent string
	)

	renderSetup := func(keys, desc string, maxW int) string {
		k := styles.CDimStyle.Width(24).Render(keys)
		d := styles.CTextStyle.Render(desc)
		return safeTruncate(lipgloss.NewStyle().PaddingLeft(2).Render(k+d), maxW)
	}

	effState := m.state
	if effState == states.LOAD_STATE {
		effState = m.prState
	}

	switch effState {
	case states.SETTINGS_STATE:
		lblRight = m.headerActiveStyle.Render("► your settings setup")

		colW := (rightW - 2) / 2

		binds := m.user.GetBinds()
		audio := m.user.GetAudio()
		notifications := m.user.GetNotifications()
		devices := m.user.GetDevices()

		leftRight := lipgloss.JoinVertical(lipgloss.Left, styles.PaddingLeftCGrayStyle.Render("binds"),
			renderSetup("friends tab", binds.FriendsTab, colW),
			renderSetup("chat tab", binds.ChatTab, colW),
			renderSetup("voice tab", binds.VoiceTab, colW),
			renderSetup("video tab", binds.VideoTab, colW),
			renderSetup("profile tab", binds.ProfileTab, colW),
			renderSetup("settings tab", binds.SettingsTab, colW),
			renderSetup("mute mic", binds.MicMute, colW),
			renderSetup("full mute", binds.FullMute, colW),
			renderSetup("user's mute", binds.UserMute, colW),
		)

		leftLeft := lipgloss.JoinVertical(lipgloss.Left, styles.PaddingLeftCGrayStyle.Render("audio"),
			renderSetup("hard denoise", fmt.Sprintf("%t", audio.Denoises.HardDenoise), colW),
			renderSetup("soft denoise", fmt.Sprintf("%t", audio.Denoises.SoftDenoise), colW),
			renderSetup("echocanceller", fmt.Sprintf("%t", audio.AEC), colW),
			renderSetup("equalizer", fmt.Sprintf("%t", audio.Filter), colW),
			"", styles.PaddingLeftCGrayStyle.Render("notifications"),
			renderSetup("app notifications", fmt.Sprintf("%t", notifications.AppNotifications), colW),
			renderSetup("desktop notifications", fmt.Sprintf("%t", notifications.DesktopNotifications), colW),
			renderSetup("audio notifications", fmt.Sprintf("%t", notifications.AudioNotifications), colW),
		)

		leftTop := lipgloss.JoinVertical(lipgloss.Left, styles.PaddingLeftCGrayStyle.Render("devices"),
			renderSetup("headphones", devices.Headphones, rightW),
			renderSetup("microphone", devices.Microphone, rightW),
		)

		styledLeftLeft := lipgloss.NewStyle().Width(colW).Render(leftLeft)
		styledLeftRight := lipgloss.NewStyle().Width(rightW - colW).Render(leftRight)
		middle := lipgloss.JoinHorizontal(lipgloss.Top, styledLeftLeft, styledLeftRight)

		rightContent = lipgloss.JoinVertical(lipgloss.Left, leftTop, "", middle)

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())

	case states.AUDIO_STATE:
		lblRight = m.headerActiveStyle.Render("► audio settings")
		m.audioList.LipList.SetSize(rightW-2, h-2)
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.audioList.LipList.View())

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())

	case states.NOTIFICATIONS_STATE:
		lblRight = m.headerActiveStyle.Render("► notification settings")
		m.notificationsList.LipList.SetSize(rightW-2, h-2)
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.notificationsList.LipList.View())

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())

	case states.DEVICES_STATE:
		lblLeft = m.headerActiveStyle.Render("► headphones")
		m.headphonesList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.headphonesList.LipList.View())

		lblRight = m.headerActiveStyle.Render("► microphones")
		m.microphonesList.LipList.SetSize(rightW-2, h-2)
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.microphonesList.LipList.View())

	case states.BINDS_STATE:
		lblRight = m.headerActiveStyle.Render("► audio binds")
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Render("audio binds")

		lblLeft = m.headerActiveStyle.Render("► quick jumps binds")
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render("quick jumps binds")

	case states.ACCOUNT_STATE:
		lblRight = m.headerActiveStyle.Render("► account settings")
		m.accountList.LipList.SetSize(rightW-2, h-2)
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.accountList.LipList.View())

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())
	case states.NICKNAME_STATE:
		lblRight = m.headerActiveStyle.Render("► account settings")
		if m.state == states.LOAD_STATE {
			rightContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"changing nickname..."), rightW)
		} else {
			iden := m.user.GetUserIdentity()
			for i := range m.nicknameInputs {
				m.nicknameInputs[i].Width = max(1, rightW-6)
			}

			nickSettings := lipgloss.JoinVertical(lipgloss.Left,
				styles.PaddingLeftCDimStyle.Render("current nickname: "+lipgloss.NewStyle().
					Foreground(m.themeColor).Render(iden.Nickname)),
				"",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.nicknameInputs[0].View()),
				"",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.nicknameInputs[1].View()),
			)

			rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(nickSettings)
		}

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())
	case states.COLOR_STATE:
		lblRight = m.headerActiveStyle.Render("► account settings")
		if m.state == states.LOAD_STATE {
			rightContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"changing color..."), rightW)
		} else {
			account := m.user.GetAccount()
			m.colorInput.Width = max(1, rightW-6)
			var ac string
			if account.Color != "" {
				ac = lipgloss.NewStyle().Foreground(lipgloss.Color(account.Color)).Render(account.Color)
			} else {
				ac = "random"
			}
			colorSettings := lipgloss.JoinVertical(lipgloss.Left,
				styles.PaddingLeftCDimStyle.Render("current color: "+ac), "",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.colorInput.View()))

			rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(colorSettings)
		}

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())
	case states.TAGLINE_STATE:
		lblRight = m.headerActiveStyle.Render("► account settings")
		if m.state == states.LOAD_STATE {
			rightContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"changing tagline..."), rightW)
		} else {
			account := m.user.GetAccount()

			m.taglineInput.Width = max(1, rightW-6)

			taglineSettings := lipgloss.JoinVertical(lipgloss.Left,
				styles.PaddingLeftCDimStyle.Render("current tagline: "+lipgloss.NewStyle().
					Foreground(m.themeColor).Render(account.Tagline)), "",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.taglineInput.View()),
			)

			rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(taglineSettings)
		}

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())
	case states.PASSWORD_STATE:
		lblRight = m.headerActiveStyle.Render("► account settings")
		if m.state == states.LOAD_STATE {
			rightContent = safeTruncate(styles.PaddingLeftCDimStyle.Render(m.spinner.View()+"changing password..."), rightW)
		} else {
			for i := range m.passwordInputs {
				m.passwordInputs[i].Width = max(1, rightW-6)
			}

			passwordSettings := lipgloss.JoinVertical(lipgloss.Left,
				styles.PaddingLeftCDimStyle.Render("current password: ************"), "",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.passwordInputs[0].View()), "",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.passwordInputs[1].View()), "",
				lipgloss.NewStyle().PaddingLeft(2).Render(m.passwordInputs[2].View()),
			)

			rightContent = lipgloss.NewStyle().PaddingLeft(2).Render(passwordSettings)
		}

		lblLeft = m.headerActiveStyle.Render("► app settings")
		m.settingsList.LipList.SetSize(leftW-2, h-2)
		leftContent = lipgloss.NewStyle().PaddingLeft(2).Render(m.settingsList.LipList.View())
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", leftContent)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", rightContent)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))

	rightPane = m.zone.Mark("rightSide", rightPane)
	leftPane = m.zone.Mark("leftSide", leftPane)

	return m.zone.Mark("settingsW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderProfileView(w, h int, cText lipgloss.AdaptiveColor) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	leftTopH := (2 * (h - 1)) / 3
	lblTopLeft := m.headerActiveStyle.Render("► profile info")

	leftLTopBoxWidth := 24

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

		k := styles.CDimStyle.Width(kw).Render(kTrunc)
		d := lipgloss.NewStyle().Foreground(cText).Render(data)
		return safeTruncate(lipgloss.NewStyle().PaddingRight(2).Render(k+d), rightStatsW)
	}
	bf := m.user.GetBestFriend()
	blocked := m.user.GetAmountOfBlocked()
	friends := m.user.GetAmountOfFriends()
	leftRTopBox := lipgloss.JoinVertical(lipgloss.Left, "", "",
		renderStatisctic("amount of friends: ", fmt.Sprintf("%d", friends)),
		"",
		renderStatisctic("max time in connection: ", fmt.Sprintf("%d minutes", m.user.GetMaxTimeInConnetion())),
		"",
		renderStatisctic("amount of messages: ", fmt.Sprintf("%d", m.user.GetAmountOfMessages())),
		"",
		renderStatisctic("minutes in connections: ", fmt.Sprintf("%d", m.user.GetAmountOfMinutesInConnections())),
		"",
		renderStatisctic("amount of connections: ", fmt.Sprintf("%d", m.user.GetAmountOfConnections())),
		"",
		renderStatisctic("amount of blocked: ", fmt.Sprintf("%d", blocked)))

	leftRTopBoxH := lipgloss.Height(leftRTopBox)
	leftRTopBox = lipgloss.Place(rightStatsW, leftRTopBoxH, lipgloss.Right, lipgloss.Top, leftRTopBox)

	nick := lipgloss.JoinVertical(lipgloss.Left,
		styles.PaddingLeftCDimStyle.Width(leftLTopBoxWidth).Render("nickname:"),
		styles.PaddingLeftCTextStyle.Render(m.user.Data.Identity.Nickname))
	reg := lipgloss.JoinVertical(lipgloss.Left,
		styles.PaddingLeftCDimStyle.Width(leftLTopBoxWidth).Render("registered:"),
		styles.PaddingLeftCTextStyle.Render(m.user.Data.Personal.RegisterTime))
	bfView := lipgloss.JoinVertical(lipgloss.Left,
		styles.PaddingLeftCDimStyle.Width(leftLTopBoxWidth).Render("best friend:"),
		styles.PaddingLeftCTextStyle.Render(fmt.Sprintf("%s\n(%d conns)", bf.Identity.Nickname, bf.AmountOfConnections)))
	personalInfo := lipgloss.JoinVertical(lipgloss.Left, nick, "", reg, "", bfView)

	leftLTopBox := lipgloss.JoinVertical(lipgloss.Left, lblTopLeft, "", personalInfo)

	leftTopBox := lipgloss.JoinHorizontal(lipgloss.Left, leftLTopBox, leftRTopBox)

	actualLeftTopH := lipgloss.Height(leftTopBox)
	if leftTopH < actualLeftTopH {
		leftTopH = actualLeftTopH
	}

	leftTopBox = lipgloss.NewStyle().MaxHeight(leftTopH).Render(leftTopBox)

	lblBotLeft := m.headerActiveStyle.Render("► friend requests")

	leftBotH := max(0, h-leftTopH-1)
	var leftBotContent string

	if m.state == states.LOAD_STATE && m.prState == states.PROFILE_STATE && m.sideState == states.LEFT_STATE {
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).
			Render(m.spinner.View() + "accepting friend request...")
	} else if len(m.friendsReqsList.DefList.LipList.Items()) <= 0 {
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).Render("zero friend request")
	} else {
		listH := max(0, leftBotH-2)
		listW := max(1, leftW-2)
		m.friendsReqsList.DefList.LipList.SetSize(listW, listH)
		leftBotContent = lipgloss.NewStyle().MaxHeight(leftBotH).PaddingLeft(2).Render(m.friendsReqsList.DefList.LipList.View())
	}

	leftBotBox := lipgloss.JoinVertical(lipgloss.Left, lblBotLeft, "", leftBotContent)
	leftBotBox = lipgloss.NewStyle().MaxHeight(leftBotH).Render(leftBotBox)
	leftBot := lipgloss.Place(leftW, leftBotH, lipgloss.Left, lipgloss.Top, leftBotBox)
	leftTop := lipgloss.Place(leftW, leftTopH, lipgloss.Left, lipgloss.Top, leftTopBox)
	dividerHorizPane := lipgloss.Place(leftW, 1, lipgloss.Center, lipgloss.Center, styles.CDimStyle.Render(horizLine(leftW)))

	leftBox := lipgloss.JoinVertical(lipgloss.Left, leftTop, dividerHorizPane, leftBot)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, lipgloss.NewStyle().MaxHeight(h).Render(leftBox))

	lblRight := m.headerActiveStyle.Render("► appereance")

	for i := range m.appereanceInputs {
		m.appereanceInputs[i].Width = max(1, rightW-6)
	}
	tags := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("tags")
	datetimes := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("datetimes")

	apereance := m.user.GetAppereance()

	rightTopBox := lipgloss.JoinVertical(lipgloss.Left,
		styles.PaddingLeftCDimStyle.Render("theme color: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(apereance.ThemeColor)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[0].View()),
		"", tags, "",
		styles.PaddingLeftCDimStyle.Render("best friend tag: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(apereance.BestFriendTag)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[1].View()),
		"",
		styles.PaddingLeftCDimStyle.Render("notification tag: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(apereance.NotificaionTag)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[2].View()),
		"",
		styles.PaddingLeftCDimStyle.Render("ban tag: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(apereance.BanTag)),
		lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[3].View()),
	//	"",
	// styles.PaddingLeftCDimStyle.Render("tagline: "+lipgloss.NewStyle().Foreground(m.themeColor).Render(apereance.Tagline)),
	// lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[4].View()),
	// "",
	// styles.PaddingLeftCDimStyle.Render("nickname: "+styles.CTextStyle.Render(iden.Nickname)),
	// lipgloss.NewStyle().PaddingLeft(2).Render(m.appereanceInputs[5].View()),
	)

	rightTopH := lipgloss.Height(rightTopBox)

	staticRightH := rightTopH + 5
	rightBotH := max(0, h-staticRightH)

	listRightW := max(1, rightW-2)
	m.apearenceList.DefList.LipList.SetSize(listRightW, rightBotH)

	rightBotBox := lipgloss.NewStyle().PaddingLeft(2).MaxHeight(rightBotH).Render(m.apearenceList.DefList.LipList.View())

	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight, "",
		rightTopBox, "",
		datetimes, "",
		rightBotBox,
	)

	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, lipgloss.NewStyle().MaxHeight(h).Render(rightBox))

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))
	rightPane = m.zone.Mark("rightSide", rightPane)
	leftPane = m.zone.Mark("leftSide", leftPane)
	return m.zone.Mark("profileW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}

func (m Model) renderRegistrationTab(w, h int) string {
	inputPad := lipgloss.NewStyle().PaddingLeft(2).Width(w)
	var content string
	if m.state == states.LOAD_STATE && m.prState == states.REG_STATE {
		content = styles.CDimStyle.Render(m.spinner.View() + "processing registration...")
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
		content = styles.CDimStyle.Render(m.spinner.View() + "login proccesing...")
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

	var userAudioState string

	chatW := w - 4
	m.chatTextInput.Width = max(1, chatW)

	var bottomSect string
	middleH := h - lipgloss.Height(userAudioState) - 2
	if middleH < 0 {
		middleH = 0
	}

	var middleSect string

	if !m.connected {

		textStyle := styles.CGrayBold.Render(titles.ALONE1)

		middleSect = lipgloss.Place(
			w,
			middleH,
			lipgloss.Center,
			lipgloss.Center,
			textStyle,
		)
	} else {
		iden := m.user.GetUserIdentity()
		coloredUserNickname := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).
			Render(iden.Nickname)
		userAudioState = coloredUserNickname

		if m.user.Engines.AudioEngine.CheckUserIsSpeaking() {
			userAudioState = coloredUserNickname + " 🔊"
		}

		mutes := m.user.GetMutes()

		if mutes.FullMute {
			userAudioState += " 🙊🙉"
		} else if mutes.MicMute {
			userAudioState += " 🙊"
		}

		cons := make([]string, 0, len(m.connections))
		speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
		mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

		for _, c := range m.connections {

			var state, rel string
			id := c.ID
			if _, ok := mutedUsers[id]; ok {
				state = " 🔇"
			} else if _, ok := speakingUsers[id]; ok {
				state = " 🔊"
			} else if us, ok := m.usersStates[id]; ok {
				if us.fullMute {
					state = " 🙊🙉"
				} else if us.micMute {
					state = " 🙊"
				}
			}

			bfTag := m.user.GetBFTag()
			bf := m.user.GetBestFriend()
			if c == bf.Identity {
				rel = bfTag + " "
			}

			usersColor := m.usersColors[c.ID]

			coloredNick := lipgloss.NewStyle().Foreground(usersColor.MainColor).Render(c.Nickname)

			cons = append(cons, rel+coloredNick+state)
		}

		rawConnStr := styles.CDimStyle.Render("with: ") + strings.Join(cons, " · ")
		connView := safeTruncate(styles.CDimStyle.Render(rawConnStr), w-2)

		historyMaxH := middleH - lipgloss.Height(connView) - 1
		if historyMaxH < 0 {
			historyMaxH = 0
		}

		var historyBox string
		if historyMaxH > 0 {
			msgStyle := lipgloss.NewStyle().Width(w - 2).MaxWidth(w - 2).PaddingLeft(2)
			fgColorStyle := lipgloss.NewStyle().Foreground(c)

			var gatheredLines []string
			linesNeeded := historyMaxH + m.chatOffset

			for i := len(m.messages) - 1; i >= 0; i-- {
				msg := m.messages[i]
				t := styles.CDimStyle.Render(msg.Time)
				var coloredNick string
				usersColor, ok := m.usersColors[msg.Identity.ID]
				if ok {
					coloredNick = lipgloss.NewStyle().Foreground(usersColor.MainColor).Render(msg.Identity.Nickname)
				} else {
					coloredNick = msg.Identity.Nickname
				}

				var renderedMsg string
				if strings.HasPrefix(msg.Text, "\n") {
					renderedMsg = styles.PaddingLeftStyle.Render(fmt.Sprintf("%s %s %s", t, coloredNick, msg.Text))
				} else {
					txt := fgColorStyle.Render(msg.Text)
					renderedMsg = msgStyle.Render(fmt.Sprintf("%s %s %s", t, coloredNick, txt))
				}

				lines := strings.Split(renderedMsg, "\n")

				for j := len(lines) - 1; j >= 0; j-- {
					gatheredLines = append(gatheredLines, lines[j])
				}

				if len(gatheredLines) >= linesNeeded {
					break
				}
			}

			start := m.chatOffset
			if start > len(gatheredLines) {
				start = len(gatheredLines)
			}
			end := start + historyMaxH
			if end > len(gatheredLines) {
				end = len(gatheredLines)
			}
			viewportLinesReversed := gatheredLines[start:end]

			viewportLines := make([]string, len(viewportLinesReversed))
			for i, line := range viewportLinesReversed {
				viewportLines[len(viewportLinesReversed)-1-i] = line
			}

			historyBox = lipgloss.Place(w, historyMaxH, lipgloss.Left, lipgloss.Bottom, strings.Join(viewportLines, "\n"))
		}

		middleContent := lipgloss.JoinVertical(lipgloss.Left, connView, "", historyBox)
		middleSect = lipgloss.Place(w, middleH, lipgloss.Left, lipgloss.Top, middleContent)
		bottomSect = styles.PaddingLeftStyle.MaxWidth(chatW).Render(m.chatTextInput.View())
	}

	finalView := lipgloss.JoinVertical(lipgloss.Left, userAudioState, middleSect, "", bottomSect)
	return m.zone.Mark("chatW", finalView)
}

func (m Model) renderStartView(w, h int) string {
	pulse := styles.WhitePulse
	if !lipgloss.HasDarkBackground() {
		pulse = styles.BlackPulse
	}

	currentColor := pulse[m.pulseFrame%len(pulse)]
	titleLogo := m.logoAnim[m.animFrame%len(m.logoAnim)]

	logo := lipgloss.NewStyle().Foreground(m.themeColor).Render(titleLogo)
	msg := lipgloss.NewStyle().Foreground(lipgloss.Color(currentColor)).Bold(true).Render(titles.START)
	msgCentered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(msg)
	content := lipgloss.JoinVertical(lipgloss.Center, logo, "", "", msgCentered)
	return m.zone.Mark("start", lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content))
}

func (m Model) renderErrView(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := styles.CErrBoldStyle.Render("► error detected")
	errText := styles.PaddingLeftCTextStyle.Render(tuierrs.CastError(m.err))
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", errText)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	lblRight := styles.CErrBoldStyle.Render("► resolution")
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
			styles.CDimStyle.Render(safeTruncate(quote, rightW)),
		)
	} else {
		rightContent = styles.CDimStyle.Render("check logs")
	}
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", rightContent)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane)
}

func (m Model) renderHelpView(w, h int) string {
	leftW := (w - 3) / 2
	rightW := w - leftW - 3

	lblLeft := m.headerActiveStyle.Render("► general & navigation")
	lblRight := m.headerActiveStyle.Render("► audio & voice controls")

	renderShortcut := func(keys, desc string) string {
		k := lipgloss.NewStyle().Foreground(m.themeColor).Width(22).Render(keys)
		d := styles.CTextStyle.Render(desc)
		return safeTruncate(styles.PaddingLeftStyle.Render(k+d), leftW)
	}

	quote := styles.PaddingLeftCGrayStyle.
		Render("hint - phd: personal hard denoise, psd: personal soft denoise")

	leftRows := []string{
		lblLeft, "",
		styles.PaddingLeftCGrayStyle.Render("app controls:"),
		renderShortcut("ALT+Q", "- quit application"),
		renderShortcut("ESC", "- go back / close error"),
		renderShortcut("ENTER", "- confirm / connect / send msg"),
		"",
		styles.PaddingLeftCGrayStyle.Render("navigation:"),
		renderShortcut("TAB / RIGHT", "- next tab"),
		renderShortcut("SHIFT+TAB / LEFT", "- previous tab"),
		renderShortcut("ALT+RIGHT", "- right side"),
		renderShortcut("ALT+LEFT", "- left side"),
		renderShortcut("SHIFT+TAB / LEFT", "- previous tab"),
		renderShortcut("UP / DN", "- move cursor in lists/inputs"),
		"",
		styles.PaddingLeftCGrayStyle.Render("quick jump:"),
		renderShortcut("ALT+F", "- go to friends tab"),
		renderShortcut("ALT+C", "- go to chat tab"),
		renderShortcut("ALT+G", "- go to voice tab"),
		renderShortcut("ALT+D", "- go to video tab"),
		renderShortcut("ALT+E", "- go to profile tab"),
		renderShortcut("ALT+S", "- go to settings tab"),
		renderShortcut("ALT+H", "- go to / close help menu"),
	}

	leftBox := lipgloss.JoinVertical(lipgloss.Left, leftRows...)
	leftPane := lipgloss.Place(leftW, h, lipgloss.Left, lipgloss.Top, leftBox)

	rightRows := []string{
		lblRight, "", quote, "",
		styles.PaddingLeftCGrayStyle.Render("voice controls:"),
		renderShortcut("ALT+V", "- toggle mic mute"),
		renderShortcut("ALT+B", "- toggle full mute"),
		renderShortcut("ALT+Z", "- toggle user's mute"),
		renderShortcut("ALT+W", "- toggle personal soft denoise"),
		renderShortcut("ALT+R", "- toggle personal hard denoise"),
		renderShortcut("ALT+UP/DN", "- increase / decrease user's volume"), "",
		styles.PaddingLeftCGrayStyle.Render("chat controls:"),
		renderShortcut("CTRL+P", "- paste smth"), "",
		styles.PaddingLeftCGrayStyle.Render("friends contols:"),
		renderShortcut("ALT+X", "- deny friendship request"),
		renderShortcut("ENTER", "- accept friendship request"),
	}

	rightBox := lipgloss.JoinVertical(lipgloss.Left, rightRows...)
	rightPane := lipgloss.Place(rightW, h, lipgloss.Left, lipgloss.Top, rightBox)

	dividerPane := lipgloss.Place(3, h, lipgloss.Center, lipgloss.Top, styles.CDimStyle.Render(vertLine(h)))
	return m.zone.Mark("profileW", lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane))
}
