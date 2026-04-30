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
	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	var ui string

	borderStyle := styles.BorderStyle
	footerStyle := styles.FooterStyle
	frameChromeW := lipgloss.Width(borderStyle.Render("X")) - 1
	frameChromeH := lipgloss.Height(borderStyle.Render("X")) - 1

	usableW := m.width - frameChromeW
	usableH := m.height - frameChromeH

	title := styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.BIG_LOGO)

	if usableW < 69 || usableH < 21 {
		return "terminal too small"
	}
	if usableW < 69 || usableH < 30 {
		footerStyle = footerStyle.MarginTop(0)
		borderStyle = borderStyle.PaddingTop(0)
		title = styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.LITTLE_LOGO)
	}
	footerData := "alt+q (quit) | alt+h (help) | esc (back)"
	if m.curWindow == windows.DEF_WINDOW {
		footerData = "alt+q (quit) | alt+h (help) | tab (switch panel) | esc (back)"
	}
	footer := footerStyle.Width(usableW).MaxWidth(usableW).Render(footerData)
	titleH := lipgloss.Height(title)

	footerH := lipgloss.Height(footer)

	contentWrapperStyle := styles.ContentStyle
	contentChromeW := lipgloss.Width(contentWrapperStyle.Render("X")) - 1
	contentChromeH := lipgloss.Height(contentWrapperStyle.Render("X")) - 1

	centerW := usableW - contentChromeW
	centerH := usableH - titleH - footerH - contentChromeH

	if centerH < 10 || centerW < 10 {
		return "terminal too small"
	}

	baseBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("30"))
	boxChromeW := lipgloss.Width(baseBox.Render(""))
	boxChromeH := lipgloss.Height(baseBox.Render("X")) - 1

	centerBox := baseBox.Align(lipgloss.Center, lipgloss.Center)
	activeBox := centerBox.BorderForeground(lipgloss.Color("205"))
	activeBaseBox := baseBox.BorderForeground(lipgloss.Color("205"))

	leftTotalW := centerW / 3
	rightTotalW := centerW - leftTotalW
	topTotalH := centerH / 2
	botTotalH := centerH - topTotalH

	innerLeftW := leftTotalW - boxChromeW
	innerRightW := rightTotalW - boxChromeW
	innerTopH := topTotalH - boxChromeH
	innerBotH := botTotalH - boxChromeH
	innerRightH := centerH - boxChromeH

	if innerLeftW < 1 {
		innerLeftW = 1
	}
	if innerRightW < 1 {
		innerRightW = 1
	}
	if innerTopH < 1 {
		innerTopH = 1
	}
	if innerBotH < 1 {
		innerBotH = 1
	}
	if innerRightH < 1 {
		innerRightH = 1
	}

	switch m.curWindow {
	case windows.START_WINDOW:
		lbl := styles.HeaderStyle.Render("welcome")
		start := styles.ContentStyle.Render("press enter or tab to start")
		mainContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", start)

		pulseColors := []string{"#3a0088", "#5c00a6", "#9000c4", "#ce00cc", "#ff00aa", "#ce00cc", "#9000c4", "#5c00a6"}
		currentColor := pulseColors[m.animFrame%len(pulseColors)]

		clippedContent := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(centerW / 2).Render(mainContent)

		startBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(currentColor)).
			Align(lipgloss.Center, lipgloss.Center).
			Width(centerW / 2).
			Height(centerH - boxChromeH).
			Render(clippedContent)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, startBox)
		styledContent := contentWrapperStyle.Render(content)
		ui = m.zone.Mark("start", lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer))

	case windows.DEF_WINDOW:

		inputW := innerLeftW - 4
		if inputW < 1 {
			inputW = 1
		}
		for i := range m.regTextInputs {
			m.regTextInputs[i].Width = inputW
		}

		var topLeft string
		styleTopLeft := centerBox
		if m.state == states.REG_STATE || m.state == states.PROFILE_STATE {
			styleTopLeft = activeBox
		}

		var topContent string

		inputContainer := lipgloss.NewStyle().Width(inputW).Align(lipgloss.Left)
		if m.state == states.LOAD_STATE && m.prState == states.REG_STATE {
			lbl := styles.HeaderStyle.Render("please wait")
			loadingText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render("loading...")
			topContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", loadingText)
		} else if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
			lbl := styles.HeaderStyle.Render("profile")
			var spekaing = ""
			if m.user.Engines.AudioEngine.UserIsSpeaking() {
				spekaing = "🔊"
			}
			name := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Bold(true).Render(m.user.Data.Personal.Nickname + spekaing)
			topContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", name)
		} else {
			lbl := styles.HeaderStyle.Render("registration")
			topContent = lipgloss.JoinVertical(
				lipgloss.Center,
				lbl,
				"",
				inputContainer.Render(m.regTextInputs[0].View()),
				"",
				inputContainer.Render(m.regTextInputs[1].View()),
				"",
				inputContainer.Render(m.regTextInputs[2].View()),
			)
		}

		clippedTop := lipgloss.NewStyle().MaxWidth(innerLeftW).MaxHeight(innerTopH).Render(topContent)
		topLeft = styleTopLeft.Width(innerLeftW).Height(innerTopH).Render(clippedTop)

		topLeftZone := m.zone.Mark("top-left", topLeft)

		var botLeft string
		styleBotLeft := centerBox
		if m.state == states.CONN_STATE || m.state == states.LEAVE_STATE || m.state == states.LOGIN_STATE {
			styleBotLeft = activeBox
		}

		m.connTextInputs[0].Width = inputW
		m.logingInput[0].Width = inputW
		m.logingInput[1].Width = inputW

		var botContent string
		if m.state == states.LOAD_STATE && (m.prState == states.CONN_STATE || m.prState == states.LEAVE_STATE || m.prState == states.LOGIN_STATE) {
			lbl := styles.HeaderStyle.Render("please wait")
			loadingText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render("loading...")
			botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", loadingText)
		} else if m.user.Networking == nil {
			lbl := styles.HeaderStyle.Render("login")
			botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", inputContainer.Render(m.logingInput[0].View()), "", inputContainer.Render(m.logingInput[1].View()))
		} else {
			onlineStr := "zero users are online"
			if len(m.online) > 0 {
				onlineStr = strings.Join(m.online, "|")
			}
			onlineView := lipgloss.NewStyle().Width(innerLeftW).MaxWidth(innerLeftW).Height(1).Align(lipgloss.Center).Foreground(lipgloss.Color("240")).Render(onlineStr)

			if !m.connected {
				lbl := styles.HeaderStyle.Render("connect")
				botContent = lipgloss.JoinVertical(
					lipgloss.Center,
					lbl,
					onlineView,
					"",
					inputContainer.Render(m.connTextInputs[0].View()),
				)
			} else {
				lbl := styles.HeaderStyle.Render("leave")
				botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "")
			}
		}
		clippedBot := lipgloss.NewStyle().MaxWidth(innerLeftW).MaxHeight(innerBotH).Render(botContent)
		botLeft = styleBotLeft.Width(innerLeftW).Height(innerBotH).Render(clippedBot)

		botLeftZone := m.zone.Mark("bot-left", botLeft)

		var right string
		styleRight := baseBox
		if m.state == states.CHAT_STATE {
			styleRight = activeBaseBox
		}

		var rightContent string
		if m.user.Networking != nil {
			chatTitle := styles.HeaderStyle.Render("chat")

			cons := make([]string, 0, len(m.connections))

			speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
			mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

			for _, c := range m.connections {
				var state = ""
				clearNickname := ansi.Strip(c)
				if _, ok := mutedUsers[clearNickname]; ok {
					state = "🔇"
				} else {
					if _, ok := speakingUsers[clearNickname]; ok {
						state = "🔊"
					}
				}

				cons = append(cons, c+" "+state)
			}

			strConn := strings.Join(cons, " | ")
			if strConn == "" {
				strConn = "no active connections"
			}

			activeConnectionsView := lipgloss.NewStyle().Bold(true).Width(innerRightW - 2).MaxWidth(innerRightW - 2).Align(lipgloss.Center).Render(strConn)

			m.chatTextInput.Width = innerRightW - 2
			chatInputContainer := lipgloss.NewStyle().Width(innerRightW - 2).Align(lipgloss.Left)
			inputView := chatInputContainer.Render(m.chatTextInput.View())

			usedSpaceH := lipgloss.Height(chatTitle) + lipgloss.Height(m.muteState) + lipgloss.Height(activeConnectionsView) + lipgloss.Height(inputView) + 1
			historyMaxH := innerRightH - usedSpaceH
			if historyMaxH < 0 {
				historyMaxH = 0
			}

			var historyBox string
			if historyMaxH > 0 {
				var allMsgsLines []string
				msgStyle := lipgloss.NewStyle().Width(innerRightW - 2).MaxWidth(innerRightW - 2)

				for _, msg := range m.messages {
					timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true).Render(msg.Time)
					renderedMsg := msgStyle.Render(timeStr + "> " + msg.Nickname + ": " + msg.Text)
					msgLine := strings.Split(renderedMsg, "\n")
					allMsgsLines = append(allMsgsLines, msgLine...)
				}

				lenAll := len(allMsgsLines)

				maxOffset := lenAll - historyMaxH
				if maxOffset < 0 {
					maxOffset = 0
				}

				if m.chatOffset > maxOffset {
					m.chatOffset = maxOffset
				}

				start := lenAll - historyMaxH - m.chatOffset
				if start < 0 {
					start = 0
				}

				end := lenAll - m.chatOffset
				if end < 0 {
					end = 0
				}

				visibleMsgs := allMsgsLines[start:end]

				historyView := strings.Join(visibleMsgs, "\n")
				historyBox = lipgloss.Place(innerRightW-2, historyMaxH, lipgloss.Left, lipgloss.Bottom, historyView)
			}

			rightContent = lipgloss.JoinVertical(lipgloss.Center, chatTitle, m.muteState, activeConnectionsView, historyBox, "", inputView)
		} else {
			rightContent = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("please register or login to start")
		}

		clippedRight := lipgloss.NewStyle().MaxWidth(innerRightW).MaxHeight(innerRightH).Render(rightContent)
		right = styleRight.Width(innerRightW).Height(innerRightH).Render(clippedRight)

		rightZone := m.zone.Mark("right", right)

		leftCol := lipgloss.JoinVertical(lipgloss.Left, topLeftZone, botLeftZone)
		mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightZone)

		styledContent := contentWrapperStyle.Render(mainContent)
		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

	case windows.SETTINGS_WINDOW:
		targetBoxWidth := (centerW) / 2

		innerW := targetBoxWidth - boxChromeW
		if innerW < 1 {
			innerW = 1
		}

		innerH := centerH - boxChromeH
		if innerH < 1 {
			innerH = 1
		}

		lblLeft := styles.HeaderStyle.Width(innerW).Align(lipgloss.Center).Render("microphones")

		listH := innerH - lipgloss.Height(lblLeft) - 1
		if listH < 1 {
			listH = 1
		}
		m.microphonesList.SetSize(innerW, listH)

		leftContent := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", m.microphonesList.View())

		micBox := baseBox.Width(innerW).
			Height(innerH).
			Render(leftContent)

		lblRight := styles.HeaderStyle.Width(innerW).Align(lipgloss.Center).Render("settings")

		denoiseText := fmt.Sprintf("denoise: %t", m.user.Data.Setup.Denoise)
		aecText := fmt.Sprintf("echo cancelling: %t", m.user.Data.Setup.AEC)

		setupText := lipgloss.JoinVertical(lipgloss.Left, denoiseText, "", aecText)

		setupContent := styles.ContentStyle.
			Width(innerW).
			PaddingLeft(2).
			Render(setupText)

		rightContent := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", setupContent)

		setupBox := baseBox.Width(innerW).
			Height(innerH).
			Render(rightContent)

		boxesJoined := lipgloss.JoinHorizontal(
			lipgloss.Top,
			micBox,
			"",
			setupBox,
		)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, boxesJoined)

		styledContent := contentWrapperStyle.Render(content)

		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

	case windows.CONNECTIONS_WINDOW:
		targetBoxWidth := (centerW - boxChromeW) / 2
		lbl := styles.HeaderStyle.Width(targetBoxWidth).Align(lipgloss.Center).Render("connections")

		listContent := m.connectionsList.View()

		mainContent := lipgloss.JoinVertical(lipgloss.Left, lbl, "", listContent)

		clippedContent := lipgloss.NewStyle().
			MaxWidth(targetBoxWidth).
			MaxHeight(centerH - boxChromeH).
			Render(mainContent)

		connsBox := centerBox.
			Width(targetBoxWidth).
			Height(centerH - boxChromeH).
			Render(clippedContent)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, connsBox)
		styledContent := contentWrapperStyle.Render(content)

		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

	case windows.HELP_WINDOW:
		targetBoxWidth := int(float32(centerW) / 1.5)
		if targetBoxWidth < 40 {
			targetBoxWidth = 40
		}

		lbl := styles.HeaderStyle.Width(targetBoxWidth).Align(lipgloss.Center).Render("help & shortcuts")

		catTitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginTop(1)
		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true)

		navTitle := catTitleStyle.Render("navigation & ui:")
		navKeys := styles.ContentStyle.Render(fmt.Sprintf(
			"%s : switch panel\n%s : back/close window\n%s : confirm/send\n%s : open settings\n%s : open connections",
			keyStyle.Render("tab"), keyStyle.Render("esc"), keyStyle.Render("enter"), keyStyle.Render("alt+s"), keyStyle.Render("alt+c"),
		))

		audioTitle := catTitleStyle.Render("audio controls:")
		audioKeys := styles.ContentStyle.Render(fmt.Sprintf(
			"%s : mute/unmute microphone\n%s : deafen (mute all sounds)\n%s : switch denoise (in settings)\n%s : switch echo canceller (in settings)",
			keyStyle.Render("alt+v"), keyStyle.Render("alt+b"), keyStyle.Render("alt+d"), keyStyle.Render("alt+e"),
		))

		usersTitle := catTitleStyle.Render("users (in connections):")
		usersKeys := styles.ContentStyle.Render(fmt.Sprintf(
			"%s : increase/decrease volume\n%s : mute/unmute user",
			keyStyle.Render("alt+up/down"), keyStyle.Render("alt+f"),
		))

		mainContent := lipgloss.JoinVertical(
			lipgloss.Center,
			lbl,
			navTitle, navKeys,
			audioTitle, audioKeys,
			usersTitle, usersKeys,
		)

		clippedHelp := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(targetBoxWidth).Render(mainContent)

		helpBox := centerBox.
			Width(targetBoxWidth).
			Height(centerH - boxChromeH).
			Render(clippedHelp)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, helpBox)
		styledContent := contentWrapperStyle.Render(content)

		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

	case windows.PROFILE_WINDOW:
		lbl := styles.HeaderStyle.Render("profile")
		nickname := styles.ContentStyle.Render("nickname: " + m.user.Data.Personal.Nickname)
		registerTime := styles.ContentStyle.Render("registered: " + m.user.Data.Personal.RegisterTime)

		mainContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", nickname, registerTime)

		clippedProfile := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(innerLeftW).Render(mainContent)
		profileBox := centerBox.Width(innerLeftW).Height(centerH - boxChromeH).Render(clippedProfile)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, profileBox)
		styledContent := contentWrapperStyle.Render(content)

		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

	case windows.ERR_WINDOW:
		lbl := styles.HeaderStyle.Render("error")
		errText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true).Render(tuierrs.CastError(m.err))
		errContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", errText)

		if m.user.Networking == nil && m.user.Data.Personal.Nickname != "" {
			errContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", errText, "", m.regTextInputs[1].View())
		}

		clippedErr := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(innerLeftW).Render(errContent)
		errBox := centerBox.Width(innerLeftW).Height(centerH - boxChromeH).Render(clippedErr)

		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, errBox)
		styledContent := contentWrapperStyle.Render(content)

		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)
	}

	return m.zone.Scan(borderStyle.Render(ui))
}
