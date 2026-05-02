// package tui

// import (
// 	"aloh-tui/internal/tui/components/states"
// 	"aloh-tui/internal/tui/components/styles"
// 	"aloh-tui/internal/tui/components/titles"
// 	"aloh-tui/internal/tui/components/windows"
// 	"aloh-tui/internal/tui/tuierrs"
// 	"fmt"
// 	"strings"

// 	"github.com/charmbracelet/lipgloss"
// 	"github.com/charmbracelet/x/ansi"
// )

// func (m Model) View() string {
// 	if m.width == 0 || m.height == 0 {
// 		return "loading..."
// 	}

// 	var ui string

// 	borderStyle := styles.BorderStyle
// 	footerStyle := styles.FooterStyle
// 	frameChromeW := lipgloss.Width(borderStyle.Render("X")) - 1
// 	frameChromeH := lipgloss.Height(borderStyle.Render("X")) - 1

// 	usableW := m.width - frameChromeW
// 	usableH := m.height - frameChromeH

// 	title := styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.BIG_LOGO)

// 	if usableW < 69 || usableH < 21 {
// 		return "terminal too small"
// 	}
// 	if usableW < 69 || usableH < 30 {
// 		footerStyle = footerStyle.MarginTop(0)
// 		borderStyle = borderStyle.PaddingTop(0)
// 		title = styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.LITTLE_LOGO)
// 	}
// 	footerData := "alt+q (quit) | alt+h (help) | esc (back)"
// 	if m.curWindow == windows.DEF_WINDOW {
// 		footerData = "alt+q (quit) | alt+h (help) | tab (switch panel) | esc (back)"
// 	}
// 	footer := footerStyle.Width(usableW).MaxWidth(usableW).Render(footerData)
// 	titleH := lipgloss.Height(title)

// 	footerH := lipgloss.Height(footer)

// 	contentWrapperStyle := styles.ContentStyle
// 	contentChromeW := lipgloss.Width(contentWrapperStyle.Render("X")) - 1
// 	contentChromeH := lipgloss.Height(contentWrapperStyle.Render("X")) - 1

// 	centerW := usableW - contentChromeW
// 	centerH := usableH - titleH - footerH - contentChromeH

// 	if centerH < 10 || centerW < 10 {
// 		return "terminal too small"
// 	}

// 	baseBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("30"))
// 	boxChromeW := lipgloss.Width(baseBox.Render(""))
// 	boxChromeH := lipgloss.Height(baseBox.Render("X")) - 1

// 	centerBox := baseBox.Align(lipgloss.Center, lipgloss.Center)
// 	activeBox := centerBox.BorderForeground(lipgloss.Color("205"))
// 	activeBaseBox := baseBox.BorderForeground(lipgloss.Color("205"))

// 	leftTotalW := centerW / 3
// 	rightTotalW := centerW - leftTotalW
// 	topTotalH := centerH / 2
// 	botTotalH := centerH - topTotalH

// 	innerLeftW := leftTotalW - boxChromeW
// 	innerRightW := rightTotalW - boxChromeW
// 	innerTopH := topTotalH - boxChromeH
// 	innerBotH := botTotalH - boxChromeH
// 	innerRightH := centerH - boxChromeH

// 	if innerLeftW < 1 {
// 		innerLeftW = 1
// 	}
// 	if innerRightW < 1 {
// 		innerRightW = 1
// 	}
// 	if innerTopH < 1 {
// 		innerTopH = 1
// 	}
// 	if innerBotH < 1 {
// 		innerBotH = 1
// 	}
// 	if innerRightH < 1 {
// 		innerRightH = 1
// 	}

// 	switch m.curWindow {
// 	case windows.START_WINDOW:
// 		lbl := styles.HeaderStyle.Render("welcome")
// 		start := styles.ContentStyle.Render("press enter or tab to start")
// 		mainContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", start)

// 		pulseColors := []string{"#3a0088", "#5c00a6", "#9000c4", "#ce00cc", "#ff00aa", "#ce00cc", "#9000c4", "#5c00a6"}
// 		currentColor := pulseColors[m.animFrame%len(pulseColors)]

// 		clippedContent := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(centerW / 2).Render(mainContent)

// 		startBox := lipgloss.NewStyle().
// 			Border(lipgloss.RoundedBorder()).
// 			BorderForeground(lipgloss.Color(currentColor)).
// 			Align(lipgloss.Center, lipgloss.Center).
// 			Width(centerW / 2).
// 			Height(centerH - boxChromeH).
// 			Render(clippedContent)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, startBox)
// 		styledContent := contentWrapperStyle.Render(content)
// 		ui = m.zone.Mark("start", lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer))

// 	case windows.DEF_WINDOW:

// 		inputW := innerLeftW - 4
// 		if inputW < 1 {
// 			inputW = 1
// 		}
// 		for i := range m.regTextInputs {
// 			m.regTextInputs[i].Width = inputW
// 		}

// 		var topLeft string
// 		styleTopLeft := centerBox
// 		if m.state == states.REG_STATE || m.state == states.PROFILE_STATE {
// 			styleTopLeft = activeBox
// 		}

// 		var topContent string

// 		inputContainer := lipgloss.NewStyle().Width(inputW).Align(lipgloss.Left)
// 		if m.state == states.LOAD_STATE && m.prState == states.REG_STATE {
// 			lbl := styles.HeaderStyle.Render("please wait")
// 			loadingText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render("loading...")
// 			topContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", loadingText)
// 		} else if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
// 			lbl := styles.HeaderStyle.Render("profile")
// 			var spekaing = ""
// 			if m.user.Engines.AudioEngine.UserIsSpeaking() {
// 				spekaing = "🔊"
// 			}
// 			name := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Bold(true).Render(m.user.Data.Personal.Nickname + spekaing)
// 			topContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", name)
// 		} else {
// 			lbl := styles.HeaderStyle.Render("registration")
// 			topContent = lipgloss.JoinVertical(
// 				lipgloss.Center,
// 				lbl,
// 				"",
// 				inputContainer.Render(m.regTextInputs[0].View()),
// 				"",
// 				inputContainer.Render(m.regTextInputs[1].View()),
// 				"",
// 				inputContainer.Render(m.regTextInputs[2].View()),
// 			)
// 		}

// 		clippedTop := lipgloss.NewStyle().MaxWidth(innerLeftW).MaxHeight(innerTopH).Render(topContent)
// 		topLeft = styleTopLeft.Width(innerLeftW).Height(innerTopH).Render(clippedTop)

// 		topLeftZone := m.zone.Mark("top-left", topLeft)

// 		var botLeft string
// 		styleBotLeft := centerBox
// 		if m.state == states.CONN_STATE || m.state == states.LEAVE_STATE || m.state == states.LOGIN_STATE {
// 			styleBotLeft = activeBox
// 		}

// 		m.connTextInputs[0].Width = inputW
// 		m.logingInput[0].Width = inputW
// 		m.logingInput[1].Width = inputW

// 		var botContent string
// 		if m.state == states.LOAD_STATE && (m.prState == states.CONN_STATE || m.prState == states.LEAVE_STATE || m.prState == states.LOGIN_STATE) {
// 			lbl := styles.HeaderStyle.Render("please wait")
// 			loadingText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render("loading...")
// 			botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", loadingText)
// 		} else if m.user.Networking == nil {
// 			lbl := styles.HeaderStyle.Render("login")
// 			botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", inputContainer.Render(m.logingInput[0].View()), "", inputContainer.Render(m.logingInput[1].View()))
// 		} else {
// 			onlineStr := "zero users are online"
// 			if len(m.online) > 0 {
// 				onlineStr = strings.Join(m.online, "|")
// 			}
// 			onlineView := lipgloss.NewStyle().Width(innerLeftW).MaxWidth(innerLeftW).Height(1).Align(lipgloss.Center).Foreground(lipgloss.Color("240")).Render(onlineStr)

// 			if !m.connected {
// 				lbl := styles.HeaderStyle.Render("connect")
// 				botContent = lipgloss.JoinVertical(
// 					lipgloss.Center,
// 					lbl,
// 					onlineView,
// 					"",
// 					inputContainer.Render(m.connTextInputs[0].View()),
// 				)
// 			} else {
// 				lbl := styles.HeaderStyle.Render("leave")
// 				botContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "")
// 			}
// 		}
// 		clippedBot := lipgloss.NewStyle().MaxWidth(innerLeftW).MaxHeight(innerBotH).Render(botContent)
// 		botLeft = styleBotLeft.Width(innerLeftW).Height(innerBotH).Render(clippedBot)

// 		botLeftZone := m.zone.Mark("bot-left", botLeft)

// 		var right string
// 		styleRight := baseBox
// 		if m.state == states.CHAT_STATE {
// 			styleRight = activeBaseBox
// 		}

// 		var rightContent string
// 		if m.user.Networking != nil {
// 			chatTitle := styles.HeaderStyle.Render("chat")

// 			cons := make([]string, 0, len(m.connections))

// 			speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
// 			mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

// 			for _, c := range m.connections {
// 				var state = ""
// 				clearNickname := ansi.Strip(c)
// 				if _, ok := mutedUsers[clearNickname]; ok {
// 					state = "🔇"
// 				} else {
// 					if _, ok := speakingUsers[clearNickname]; ok {
// 						state = "🔊"
// 					}
// 				}

// 				cons = append(cons, c+" "+state)
// 			}

// 			strConn := strings.Join(cons, " | ")
// 			if strConn == "" {
// 				strConn = "no active connections"
// 			}

// 			activeConnectionsView := lipgloss.NewStyle().Bold(true).Width(innerRightW - 2).MaxWidth(innerRightW - 2).Align(lipgloss.Center).Render(strConn)

// 			m.chatTextInput.Width = innerRightW - 2
// 			chatInputContainer := lipgloss.NewStyle().Width(innerRightW - 2).Align(lipgloss.Left)
// 			inputView := chatInputContainer.Render(m.chatTextInput.View())

// 			usedSpaceH := lipgloss.Height(chatTitle) + lipgloss.Height(m.muteState) + lipgloss.Height(activeConnectionsView) + lipgloss.Height(inputView) + 1
// 			historyMaxH := innerRightH - usedSpaceH
// 			if historyMaxH < 0 {
// 				historyMaxH = 0
// 			}

// 			var historyBox string
// 			if historyMaxH > 0 {
// 				var allMsgsLines []string
// 				msgStyle := lipgloss.NewStyle().Width(innerRightW - 2).MaxWidth(innerRightW - 2)

// 				for _, msg := range m.messages {
// 					timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true).Render(msg.Time)
// 					renderedMsg := msgStyle.Render(timeStr + "> " + msg.Nickname + ": " + msg.Text)
// 					msgLine := strings.Split(renderedMsg, "\n")
// 					allMsgsLines = append(allMsgsLines, msgLine...)
// 				}

// 				lenAll := len(allMsgsLines)

// 				maxOffset := lenAll - historyMaxH
// 				if maxOffset < 0 {
// 					maxOffset = 0
// 				}

// 				if m.chatOffset > maxOffset {
// 					m.chatOffset = maxOffset
// 				}

// 				start := lenAll - historyMaxH - m.chatOffset
// 				if start < 0 {
// 					start = 0
// 				}

// 				end := lenAll - m.chatOffset
// 				if end < 0 {
// 					end = 0
// 				}

// 				visibleMsgs := allMsgsLines[start:end]

// 				historyView := strings.Join(visibleMsgs, "\n")
// 				historyBox = lipgloss.Place(innerRightW-2, historyMaxH, lipgloss.Left, lipgloss.Bottom, historyView)
// 			}

// 			rightContent = lipgloss.JoinVertical(lipgloss.Center, chatTitle, m.muteState, activeConnectionsView, historyBox, "", inputView)
// 		} else {
// 			rightContent = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("please register or login to start")
// 		}

// 		clippedRight := lipgloss.NewStyle().MaxWidth(innerRightW).MaxHeight(innerRightH).Render(rightContent)
// 		right = styleRight.Width(innerRightW).Height(innerRightH).Render(clippedRight)

// 		rightZone := m.zone.Mark("right", right)

// 		leftCol := lipgloss.JoinVertical(lipgloss.Left, topLeftZone, botLeftZone)
// 		mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightZone)

// 		styledContent := contentWrapperStyle.Render(mainContent)
// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

// 	case windows.SETTINGS_WINDOW:
// 		targetBoxWidth := (centerW) / 2

// 		innerW := targetBoxWidth - boxChromeW
// 		if innerW < 1 {
// 			innerW = 1
// 		}

// 		innerH := centerH - boxChromeH
// 		if innerH < 1 {
// 			innerH = 1
// 		}

// 		lblLeft := styles.HeaderStyle.Width(innerW).Align(lipgloss.Center).Render("microphones")

// 		listH := innerH - lipgloss.Height(lblLeft) - 1
// 		if listH < 1 {
// 			listH = 1
// 		}
// 		m.microphonesList.SetSize(innerW, listH)

// 		leftContent := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", m.microphonesList.View())

// 		micBox := baseBox.Width(innerW).
// 			Height(innerH).
// 			Render(leftContent)

// 		lblRight := styles.HeaderStyle.Width(innerW).Align(lipgloss.Center).Render("settings")

// 		denoiseText := fmt.Sprintf("denoise: %t", m.user.Data.Setup.Denoise)
// 		aecText := fmt.Sprintf("echo cancelling: %t", m.user.Data.Setup.AEC)
// 		filterText := fmt.Sprintf("filter: %t", m.user.Data.Setup.Filter)

// 		setupText := lipgloss.JoinVertical(lipgloss.Left, denoiseText, "", aecText, "", filterText)

// 		setupContent := styles.ContentStyle.
// 			Width(innerW).
// 			PaddingLeft(2).
// 			Render(setupText)

// 		rightContent := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", setupContent)

// 		setupBox := baseBox.Width(innerW).
// 			Height(innerH).
// 			Render(rightContent)

// 		boxesJoined := lipgloss.JoinHorizontal(
// 			lipgloss.Top,
// 			micBox,
// 			"",
// 			setupBox,
// 		)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, boxesJoined)

// 		styledContent := contentWrapperStyle.Render(content)

// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

// 	case windows.CONNECTIONS_WINDOW:
// 		targetBoxWidth := (centerW - boxChromeW) / 2
// 		lbl := styles.HeaderStyle.Width(targetBoxWidth).Align(lipgloss.Center).Render("connections")

// 		listContent := m.connectionsList.View()

// 		mainContent := lipgloss.JoinVertical(lipgloss.Left, lbl, "", listContent)

// 		clippedContent := lipgloss.NewStyle().
// 			MaxWidth(targetBoxWidth).
// 			MaxHeight(centerH - boxChromeH).
// 			Render(mainContent)

// 		connsBox := centerBox.
// 			Width(targetBoxWidth).
// 			Height(centerH - boxChromeH).
// 			Render(clippedContent)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, connsBox)
// 		styledContent := contentWrapperStyle.Render(content)

// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

// 	case windows.HELP_WINDOW:
// 		targetBoxWidth := int(float32(centerW) / 1.5)
// 		if targetBoxWidth < 40 {
// 			targetBoxWidth = 40
// 		}

// 		lbl := styles.HeaderStyle.Width(targetBoxWidth).Align(lipgloss.Center).Render("help & shortcuts")

// 		catTitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginTop(1)
// 		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true)

// 		navTitle := catTitleStyle.Render("navigation & ui:")
// 		navKeys := styles.ContentStyle.Render(fmt.Sprintf(
// 			"%s : switch panel\n%s : back/close window\n%s : confirm/send\n%s : open settings\n%s : open connections",
// 			keyStyle.Render("tab"), keyStyle.Render("esc"), keyStyle.Render("enter"), keyStyle.Render("alt+s"), keyStyle.Render("alt+c"),
// 		))

// 		audioTitle := catTitleStyle.Render("audio controls:")
// 		audioKeys := styles.ContentStyle.Render(fmt.Sprintf(
// 			"%s : mute/unmute microphone\n%s : deafen (mute all sounds)\n%s : switch denoise (in settings)\n%s : switch echo canceller (in settings)\n%s : switch filter (in settings)",
// 			keyStyle.Render("alt+v"), keyStyle.Render("alt+b"), keyStyle.Render("alt+d"), keyStyle.Render("alt+e"), keyStyle.Render("alt+f"),
// 		))

// 		usersTitle := catTitleStyle.Render("users (in connections):")
// 		usersKeys := styles.ContentStyle.Render(fmt.Sprintf(
// 			"%s : increase/decrease volume\n%s : mute/unmute user",
// 			keyStyle.Render("alt+up/down"), keyStyle.Render("alt+z"),
// 		))

// 		mainContent := lipgloss.JoinVertical(
// 			lipgloss.Center,
// 			lbl,
// 			navTitle, navKeys,
// 			audioTitle, audioKeys,
// 			usersTitle, usersKeys,
// 		)

// 		clippedHelp := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(targetBoxWidth).Render(mainContent)

// 		helpBox := centerBox.
// 			Width(targetBoxWidth).
// 			Height(centerH - boxChromeH).
// 			Render(clippedHelp)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, helpBox)
// 		styledContent := contentWrapperStyle.Render(content)

// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

// 	case windows.PROFILE_WINDOW:
// 		lbl := styles.HeaderStyle.Render("profile")
// 		nickname := styles.ContentStyle.Render("nickname: " + m.user.Data.Personal.Nickname)
// 		registerTime := styles.ContentStyle.Render("registered: " + m.user.Data.Personal.RegisterTime)

// 		mainContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", nickname, registerTime)

// 		clippedProfile := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(innerLeftW).Render(mainContent)
// 		profileBox := centerBox.Width(innerLeftW).Height(centerH - boxChromeH).Render(clippedProfile)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, profileBox)
// 		styledContent := contentWrapperStyle.Render(content)

// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)

// 	case windows.ERR_WINDOW:
// 		lbl := styles.HeaderStyle.Render("error")
// 		errText := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true).Render(tuierrs.CastError(m.err))
// 		errContent := lipgloss.JoinVertical(lipgloss.Center, lbl, "", errText)

// 		if m.user.Networking == nil && m.user.Data.Personal.Nickname != "" {
// 			errContent = lipgloss.JoinVertical(lipgloss.Center, lbl, "", errText, "", m.regTextInputs[1].View())
// 		}

// 		clippedErr := lipgloss.NewStyle().MaxHeight(centerH - boxChromeH).MaxWidth(innerLeftW).Render(errContent)
// 		errBox := centerBox.Width(innerLeftW).Height(centerH - boxChromeH).Render(clippedErr)

// 		content := lipgloss.Place(centerW, centerH, lipgloss.Center, lipgloss.Center, errBox)
// 		styledContent := contentWrapperStyle.Render(content)

// 		ui = lipgloss.JoinVertical(lipgloss.Center, title, styledContent, footer)
// 	}

// 	return m.zone.Scan(borderStyle.Render(ui))
// }
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
	cFocus    = lipgloss.Color("#A6E22E") // Неоновый зеленый
	cDim      = lipgloss.Color("#75715E") // Приглушенный серый
	cText     = lipgloss.Color("#F8F8F2") // Основной текст
	cAccent   = lipgloss.Color("#FD971F") // Оранжевый акцент
	cErr      = lipgloss.Color("#F92672") // Розово-красный
	cSubtext  = lipgloss.Color("#66D9EF") // Голубой

	headerActive   = lipgloss.NewStyle().Foreground(cFocus).Bold(true)
	headerInactive = lipgloss.NewStyle().Foreground(cDim).Bold(true)
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	padW, padH := 2, 1
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	if usableW < 60 || usableH < 18 {
		return "Terminal too small. Min size: 60x18."
	}

	// 1. Настройка футера (В стартовом меню только ALT+Q)
	var footerData string
	switch m.curWindow {
	case windows.START_WINDOW:
		footerData = "ALT+Q: Quit"
	case windows.DEF_WINDOW:
		footerData = "ALT+Q: Quit   |   ALT+H: Help   |   TAB: Switch   |   ESC: Back"
	default:
		footerData = "ALT+Q: Quit   |   ALT+H: Help   |   ESC: Back"
	}
	footer := lipgloss.NewStyle().Foreground(cDim).Width(usableW).Render(footerData)
	footerH := lipgloss.Height(footer)

	// 2. Глобальный логотип (скрываем в стартовом окне, там он будет по центру)
	var title string
	var titleH int
	if m.curWindow != windows.START_WINDOW {
		titleLogo := titles.BIG_LOGO
		if usableW < 80 || usableH < 25 {
			titleLogo = titles.LITTLE_LOGO
		}
		title = lipgloss.NewStyle().Foreground(cFocus).Render(titleLogo)
		titleH = lipgloss.Height(title)
	}

	// 3. Вычисление высоты Grid
	var gridH int
	if m.curWindow != windows.START_WINDOW {
		gridH = usableH - titleH - 2 - footerH
	} else {
		gridH = usableH - 1 - footerH
	}

	if gridH < 10 || usableW < 10 {
		return "Terminal too small."
	}

	var uiContent string
	switch m.curWindow {
	case windows.START_WINDOW:
		uiContent = m.renderStartView(usableW, gridH)
	case windows.DEF_WINDOW:
		uiContent = m.renderBorderlessDashboard(usableW, gridH)
	case windows.SETTINGS_WINDOW:
		uiContent = m.renderSettingsView(usableW, gridH)
	case windows.CONNECTIONS_WINDOW:
		uiContent = m.renderConnectionsView(usableW, gridH)
	case windows.HELP_WINDOW:
		uiContent = m.renderHelpView(usableW, gridH)
	case windows.PROFILE_WINDOW:
		uiContent = m.renderProfileView(usableW, gridH)
	case windows.ERR_WINDOW:
		uiContent = m.renderErrView(usableW, gridH)
	}

	// 4. Сборка всего окна
	var finalLayout string
	if m.curWindow != windows.START_WINDOW {
		finalLayout = lipgloss.JoinVertical(lipgloss.Left, title, "", uiContent, "", footer)
	} else {
		finalLayout = lipgloss.JoinVertical(lipgloss.Left, uiContent, "", footer)
	}
	
	screen := lipgloss.Place(usableW, usableH, lipgloss.Left, lipgloss.Top, finalLayout)

	return lipgloss.NewStyle().Padding(padH, padW).Render(m.zone.Scan(screen))
}

// =====================================================================
// ОСНОВНОЕ ОКНО
// =====================================================================

func (m Model) renderBorderlessDashboard(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2 

	topH := (h * 35) / 100
	botH := h - topH - 1

	activePrefix := "► "
	inactivePrefix := "  "
	contentPad := lipgloss.NewStyle().PaddingLeft(2).Width(leftW)

	// --- СЕКЦИЯ ЛИЧНОСТИ (Верх Лево) ---
	isTopActive := m.state == states.REG_STATE || m.state == states.PROFILE_STATE
	var topContent string

	for i := range m.regTextInputs {
		m.regTextInputs[i].Width = max(1, leftW-3)
	}

	lblTop := headerInactive.Render(inactivePrefix + "IDENTITY")
	if isTopActive { lblTop = headerActive.Render(activePrefix + "IDENTITY") }

	if m.state == states.LOAD_STATE && m.prState == states.REG_STATE {
		topContent = contentPad.Foreground(cDim).Render("Processing...")
	} else if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
		speaking := ""
		if m.user.Engines.AudioEngine.UserIsSpeaking() { speaking = " 🔊" }
		nick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Bold(true).Render("@" + m.user.Data.Personal.Nickname + speaking)
		topContent = contentPad.Render(nick)
	} else {
		topContent = lipgloss.JoinVertical(lipgloss.Left,
			contentPad.Render(m.regTextInputs[0].View()),
			contentPad.Render(m.regTextInputs[1].View()),
			contentPad.Render(m.regTextInputs[2].View()),
		)
	}
	topLeftPane := lipgloss.NewStyle().Width(leftW).Height(topH).Render(lipgloss.JoinVertical(lipgloss.Left, lblTop, topContent))

	// --- РАЗДЕЛИТЕЛЬ ---
	horizDivider := lipgloss.NewStyle().Foreground(cDim).Render(strings.Repeat("─", leftW))

	// --- СЕКЦИЯ СЕТИ (Низ Лево) ---
	isBotActive := m.state == states.CONN_STATE || m.state == states.LEAVE_STATE || m.state == states.LOGIN_STATE
	var botContent string

	m.connTextInputs[0].Width = max(1, leftW-3)
	m.logingInput[0].Width = max(1, leftW-3)
	m.logingInput[1].Width = max(1, leftW-3)

	lblBot := headerInactive.Render(inactivePrefix + "NETWORK")
	if isBotActive { lblBot = headerActive.Render(activePrefix + "NETWORK") }

	if m.state == states.LOAD_STATE && (m.prState == states.CONN_STATE || m.prState == states.LEAVE_STATE || m.prState == states.LOGIN_STATE) {
		botContent = contentPad.Foreground(cDim).Render("Connecting...")
	} else if m.user.Networking == nil {
		botContent = lipgloss.JoinVertical(lipgloss.Left,
			contentPad.Render(m.logingInput[0].View()),
			contentPad.Render(m.logingInput[1].View()),
		)
	} else {
		rawOnlineStr := "0 online"
		if len(m.online) > 0 {
			rawOnlineStr = fmt.Sprintf("%d online: %s", len(m.online), strings.Join(m.online, ", "))
		}
		
		onlineView := lipgloss.NewStyle().Width(leftW).PaddingLeft(2).Foreground(cDim).Render(rawOnlineStr)

		if !m.connected {
			botContent = lipgloss.JoinVertical(lipgloss.Left, onlineView, "", contentPad.Render(m.connTextInputs[0].View()))
		} else {
			leaveTxt := contentPad.Foreground(cErr).Render("[ ENTER to leave ]")
			botContent = lipgloss.JoinVertical(lipgloss.Left, onlineView, "", leaveTxt)
		}
	}
	
	botContentWrapped := lipgloss.NewStyle().MaxHeight(botH - 1).Render(botContent)
	botLeftPane := lipgloss.NewStyle().Width(leftW).Height(botH).Render(lipgloss.JoinVertical(lipgloss.Left, lblBot, botContentWrapped))

	// --- ВЕРТИКАЛЬНЫЙ РАЗДЕЛИТЕЛЬ ---
	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))

	// --- СЕКЦИЯ ЧАТА (Право) ---
	isChatActive := m.state == states.CHAT_STATE
	var rightContent string

	lblChat := headerInactive.Render(inactivePrefix + "COMMS")
	if isChatActive { lblChat = headerActive.Render(activePrefix + "COMMS") }

	if m.user.Networking != nil {
		cons := make([]string, 0, len(m.connections))
		speakingUsers := m.user.Engines.AudioEngine.FetchSpeakingUsers()
		mutedUsers := m.user.Engines.AudioEngine.FetchUsersMutes()

		for _, c := range m.connections {
			state := ""
			clearNick := ansi.Strip(c)
			if _, ok := mutedUsers[clearNick]; ok { state = " 🔇" } else if _, ok := speakingUsers[clearNick]; ok { state = " 🔊" }
			cons = append(cons, c+state)
		}

		rawConnStr := strings.Join(cons, "  ·  ")
		if rawConnStr == "" { rawConnStr = "No active connections" }
		
		safeConnStr := safeTruncate(rawConnStr, rightW-2)
		connsView := lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render(safeConnStr)

		m.chatTextInput.Width = max(1, rightW-3)
		inputView := lipgloss.NewStyle().PaddingLeft(2).Render(m.chatTextInput.View())

		var chatParts[]string
		chatParts = append(chatParts, lblChat)
		
		if m.muteState != "" {
			chatParts = append(chatParts, lipgloss.NewStyle().PaddingLeft(2).Render(m.muteState))
		}
		
		chatParts = append(chatParts, connsView, "") 
		
		usedH := 0
		for _, p := range chatParts {
			usedH += lipgloss.Height(p)
		}
		usedH += 1 
		usedH += lipgloss.Height(inputView)

		historyMaxH := max(0, h-usedH)
		var historyBox string

		if historyMaxH > 0 {
			var allMsgsLines[]string
			msgStyle := lipgloss.NewStyle().Width(rightW - 2).MaxWidth(rightW - 2).PaddingLeft(2)

			for _, msg := range m.messages {
				t := lipgloss.NewStyle().Foreground(cDim).Render(msg.Time)
				n := lipgloss.NewStyle().Foreground(cSubtext).Bold(true).Render(msg.Nickname + ":")
				txt := lipgloss.NewStyle().Foreground(cText).Render(msg.Text)
				
				renderedMsg := msgStyle.Render(fmt.Sprintf("%s %s %s", t, n, txt))
				allMsgsLines = append(allMsgsLines, strings.Split(renderedMsg, "\n")...)
			}

			lenAll := len(allMsgsLines)
			m.chatOffset = min(m.chatOffset, max(0, lenAll-historyMaxH))
			startIdx := max(0, lenAll-historyMaxH-m.chatOffset)
			endIdx := max(0, lenAll-m.chatOffset)

			visibleMsgs := allMsgsLines[startIdx:endIdx]
			historyBox = lipgloss.Place(rightW, historyMaxH, lipgloss.Left, lipgloss.Bottom, strings.Join(visibleMsgs, "\n"))
		}

		chatParts = append(chatParts, historyBox, "", inputView)
		rightContent = lipgloss.JoinVertical(lipgloss.Left, chatParts...)

	} else {
		rightContent = lipgloss.JoinVertical(lipgloss.Left, lblChat)
		emptyMsg := lipgloss.Place(rightW, h-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(cDim).Render("~ Offline ~"))
		rightContent = lipgloss.JoinVertical(lipgloss.Left, rightContent, emptyMsg)
	}

	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightContent)

	leftCol := lipgloss.JoinVertical(lipgloss.Left, m.zone.Mark("top-left", topLeftPane), horizDivider, m.zone.Mark("bot-left", botLeftPane))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", dividerPane, " ", m.zone.Mark("right", rightPane))
}

// =====================================================================
// ВТОРОСТЕПЕННЫЕ ОКНА
// =====================================================================

func (m Model) renderSettingsView(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2
	
	lblLeft := headerActive.Render("► MICROPHONES")
	m.microphonesList.SetSize(leftW-4, h-2)
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, lipgloss.NewStyle().PaddingLeft(2).Render(m.microphonesList.View()))
	leftPane := lipgloss.NewStyle().Width(leftW).Height(h).Render(leftBox)

	lblRight := headerActive.Render("► AUDIO FILTERS")
	st := lipgloss.NewStyle().Foreground(cFocus).Bold(true)
	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight, "",
		lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("Denoise: %s", st.Render(fmt.Sprintf("%t", m.user.Data.Setup.Denoise)))),
		lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("Echo:    %s", st.Render(fmt.Sprintf("%t", m.user.Data.Setup.AEC)))),
		lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("Filter:  %s", st.Render(fmt.Sprintf("%t", m.user.Data.Setup.Filter)))),
	)
	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightBox)

	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", dividerPane, " ", rightPane)
}

func (m Model) renderHelpView(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2

	keyStyle := lipgloss.NewStyle().Foreground(cFocus).Width(12)
	descStyle := lipgloss.NewStyle().Foreground(cText)
	row := func(k, d string) string { return lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(lipgloss.Top, keyStyle.Render(k), descStyle.Render(d))) }

	lblLeft := headerActive.Render("► NAVIGATION & USERS")
	leftBox := lipgloss.JoinVertical(lipgloss.Left,
		lblLeft, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cSubtext).Render("[ General ]"),
		row("TAB", "Switch Panel"),
		row("ESC", "Back"),
		row("ALT+S", "Settings"),
		row("ALT+C", "Connections"),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cSubtext).Render("[ Users ]"),
		row("ALT+UP/DN", "User Volume"),
		row("ALT+Z", "Mute User"),
	)
	leftPane := lipgloss.NewStyle().Width(leftW).Height(h).Render(leftBox)

	lblRight := headerActive.Render("► AUDIO CONTROLS")
	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cSubtext).Render("[ Microphone & Filters ]"),
		row("ALT+V", "Mute mic"),
		row("ALT+B", "Deafen all"),
		row("ALT+D", "Toggle Denoise"),
		row("ALT+E", "Toggle Echo Canceller"),
		row("ALT+F", "Toggle Audio Filter"),
	)
	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightBox)

	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", dividerPane, " ", rightPane)
}

func (m Model) renderConnectionsView(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2

	lblLeft := headerActive.Render("► ACTIVE CONNECTIONS")
	m.connectionsList.SetSize(leftW-4, h-2)
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, lipgloss.NewStyle().PaddingLeft(2).Render(m.connectionsList.View()))
	leftPane := lipgloss.NewStyle().Width(leftW).Height(h).Render(leftBox)

	// Добавили красивую панель с контроллерами и статусом
	lblRight := headerActive.Render("► DETAILS & CONTROLS")
	
	statusText := "Offline"
	statusColor := cDim
	if m.connected {
		statusText = "Connected"
		statusColor = cFocus
	}
	
	rightBox := lipgloss.JoinVertical(lipgloss.Left,
		lblRight, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Network Status: ") + lipgloss.NewStyle().Foreground(statusColor).Render(statusText),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("User Management:"),
		lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.NewStyle().Foreground(cFocus).Render("ALT+UP/DN") + lipgloss.NewStyle().Foreground(cText).Render(" - Adjust user volume")),
		lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.NewStyle().Foreground(cFocus).Render("ALT+Z")     + lipgloss.NewStyle().Foreground(cText).Render("     - Mute/Unmute user")),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Global Audio:"),
		lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.NewStyle().Foreground(cFocus).Render("ALT+B")     + lipgloss.NewStyle().Foreground(cText).Render("     - Deafen all sounds")),
	)
	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightBox)

	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", dividerPane, " ", rightPane)
}

func (m Model) renderProfileView(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2

	lblLeft := headerActive.Render("► PROFILE INFO")
	leftBox := lipgloss.JoinVertical(lipgloss.Left,
		lblLeft, "",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Nickname:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render("@"+m.user.Data.Personal.Nickname),
		"",
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Registered:"),
		lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(m.user.Data.Personal.RegisterTime),
	)
	leftPane := lipgloss.NewStyle().Width(leftW).Height(h).Render(leftBox)

	lblRight := headerActive.Render("► STATISTICS")
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("No statistics available."))
	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightBox)

	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", dividerPane, " ", rightPane)
}

func (m Model) renderStartView(w, h int) string {
	greenPulse :=[]string{"#004400", "#006600", "#008800", "#00AA00", "#00CC00", "#33FF33", "#00CC00", "#00AA00", "#008800", "#006600"}
	currentColor := greenPulse[m.animFrame%len(greenPulse)]
	
	titleLogo := titles.BIG_LOGO
	if w < 80 || h < 25 {
		titleLogo = titles.LITTLE_LOGO
	}
	
	logo := lipgloss.NewStyle().Foreground(cFocus).Width(w).Align(lipgloss.Center).Render(titleLogo)
	msg := lipgloss.NewStyle().Foreground(lipgloss.Color(currentColor)).Bold(true).Render("Press [ENTER] or [TAB] to initiate sequence.")
	msgCentered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(msg)
	
	content := lipgloss.JoinVertical(lipgloss.Center, logo, "", "", msgCentered)
	
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderErrView(w, h int) string {
	leftW := (w * 35) / 100
	rightW := w - leftW - 2

	lblLeft := lipgloss.NewStyle().Foreground(cErr).Bold(true).Render("► ERROR DETECTED")
	errText := lipgloss.NewStyle().PaddingLeft(2).Foreground(cText).Render(tuierrs.CastError(m.err))
	leftBox := lipgloss.JoinVertical(lipgloss.Left, lblLeft, "", errText)
	leftPane := lipgloss.NewStyle().Width(leftW).Height(h).Render(leftBox)

	lblRight := lipgloss.NewStyle().Foreground(cErr).Bold(true).Render("► RESOLUTION")
	var rightContent string
	if m.user.Networking == nil && m.user.Data.Personal.Nickname != "" {
		m.regTextInputs[1].Width = max(1, rightW-4)
		rightContent = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Please re-enter credentials:"), "",
			lipgloss.NewStyle().PaddingLeft(2).Render(m.regTextInputs[1].View()),
		)
	} else {
		rightContent = lipgloss.NewStyle().PaddingLeft(2).Foreground(cDim).Render("Check logs or configuration.")
	}
	
	rightBox := lipgloss.JoinVertical(lipgloss.Left, lblRight, "", rightContent)
	rightPane := lipgloss.NewStyle().Width(rightW).Height(h).Render(rightBox)

	dividerPane := lipgloss.NewStyle().Height(h).Render(vertLine(h))
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", dividerPane, " ", rightPane)
}

func vertLine(h int) string {
	if h <= 0 {
		return ""
	}
	return strings.Repeat("│\n", h-1) + "│"
}

func safeTruncate(s string, maxW int) string {
	if maxW <= 0 { return "" }
	cleanStr := ansi.Strip(s)
	runes :=[]rune(cleanStr)
	if len(runes) > maxW { return string(runes[:maxW-2]) + ".." }
	return s
}