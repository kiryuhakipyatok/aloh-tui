package tui

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/notifications"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/logger"
	"math"
	"slices"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	MAX_VOLUME = 4.0
	MIN_VOLUME = 0.0
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress {
				if m.zone.Get("top-left").InBounds(msg) && m.user.Networking == nil {
					m.prState = m.state
					m.state = states.REG_STATE
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else if m.zone.Get("top-left").InBounds(msg) && m.user.Networking != nil {
					m.curWindow = windows.PROFILE_WINDOW
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else if m.zone.Get("bot-left").InBounds(msg) && m.user.Networking == nil {
					m.prState = m.state
					m.state = states.LOGIN_STATE
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else if m.zone.Get("bot-left").InBounds(msg) && m.user.Networking != nil && m.connected {
					m.prState = m.state
					m.state = states.LOAD_STATE
					m.cursor = 0
					cmds = append(cmds, commands.LeaveCmd(m.user.Networking, m.user.Engines.AudioEngine))
				} else if m.zone.Get("bot-left").InBounds(msg) && m.user.Networking != nil {
					m.prState = m.state
					m.state = states.CONN_STATE
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else if m.zone.Get("right").InBounds(msg) && m.user.Networking != nil && m.connected {
					m.prState = m.state
					m.state = states.CHAT_STATE
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else if m.zone.Get("start").InBounds(msg) {
					m.state = states.DEF_STATE
					m.curWindow = windows.DEF_WINDOW
					m.cursor = 0
					m.focusInputs()
					return m, textinput.Blink
				} else {
					if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
						m.state = states.DEF_STATE
						m.cursor = 0
						m.focusInputs()
						return m, textinput.Blink
					}
				}
			}

		case tea.MouseButtonWheelUp:
			if m.connected && m.state == states.CHAT_STATE && m.user.Networking != nil && m.zone.Get("right").InBounds(msg) {
				if m.chatOffset < m.getMaxChatOffset() {
					m.chatOffset++
				}
			}

		case tea.MouseButtonWheelDown:
			if m.connected && m.state == states.CHAT_STATE && m.user.Networking != nil && m.zone.Get("right").InBounds(msg) {
				m.chatOffset--
				if m.chatOffset < 0 {
					m.chatOffset = 0
				}
			}
		}

	case commands.UpdateTickMsg:
		if m.connected {
			return m, commands.UpdateTickCmd()
		} else {
			m.updateTick = false
			return m, nil
		}

	case commands.OnOffDenoiceMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}

	case commands.OnOffFilterMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}

	case commands.OnOffAECMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}

	case commands.UsersVolumeMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}

	case commands.MuteMsg:
		switch msg.Typee {
		case commands.MIC:
			if msg.Res {
				m.muteState = "🙊"
			} else {
				m.muteState = ""
			}
		case commands.FULL:
			if msg.Res {
				m.muteState = "🙊🙉"
			} else {
				m.muteState = ""
			}
		}

	case commands.MuteUnmuteUserMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}
		m.focusInputs()
		return m, textinput.Blink

	case commands.ConnectToUserMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		} else {
			m.state = states.CHAT_STATE
			m.connected = true
			m.focusInputs()
			m.user.Engines.AudioEngine.SetConnected()
			cmds = append(cmds, textinput.Blink, commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			return m, tea.Batch(cmds...)
		}
	case commands.SendInChatMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		} else {
			m.state = states.CHAT_STATE
			m.focusInputs()
			return m, textinput.Blink
		}
	case commands.LeaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		} else {
			m.state = states.CONN_STATE
			m.connected = false
			m.user.Engines.AudioEngine.SetDisconnected()
			m.messages = []commands.ChatMessage{}
			m.connections = []string{}
			m.user.Engines.AudioEngine.PlayNotification()
			m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "System", Text: "you disconnected!"})
			m.focusInputs()
			return m, textinput.Blink
		}

	case commands.AuthMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
			if msg.Typee != auth.DEFAULT {
				m.user.Data = entities.Data{}
			}
		} else {
			m.state = states.DEF_STATE
			if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
				m.setupMicrohonesList(m.user.Data.Devices.Microphone)

				m.user.Networking.ChatCallback(func(id string, data []byte) {
					t := time.Now().Format("15:04:05")
					msg := string(data)
					m.msgChan <- commands.ChatMessage{Time: t, Nickname: id, Text: msg}
					m.user.Engines.AudioEngine.PlayNotification()
					if err := notifications.Notify(t, id, msg); err != nil {
						m.log.Error("failed to notify", logger.Attr("userId", id), logger.Err(err))
					}
				})
				m.user.Networking.VoiceCallback(func(id string, data []byte) {
					m.user.Engines.AudioEngine.PlayUserVoice(id, data)
				})
			}

			cmds = append(cmds,
				commands.WaitForChatMessageCmd(m.msgChan),
				commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname),
				commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname),
				commands.TickCmd())

		}

	case commands.ChatMessage:
		msg.Nickname = m.coloredNickname(msg.Nickname)
		m.messages = append(m.messages, msg)
		return m, commands.WaitForChatMessageCmd(m.msgChan)

	case commands.OnlineMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		} else {
			m.online = msg.Online
		}

	case commands.ChangeMicrophoneMessage:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		}

	case commands.SessionsUpdateMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			m.curWindow = windows.ERR_WINDOW
		} else {

			joined := utils.Difference(msg.Sessions, m.connections)

			for _, v := range joined {
				color := lipgloss.Color(randomcolor.GetRandomColorInHex())
				nickname := lipgloss.NewStyle().Foreground(color).Render(v)
				m.connections = append(m.connections, nickname)
				m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "System", Text: nickname + " joined the chat!"})
				m.connected = true
				m.user.Engines.AudioEngine.SetConnected()
				cmds = append(cmds, commands.SetupUserVolumeCmd(m.user, v), commands.SetupUserMuteCmd(m.user, v))
				m.user.Engines.AudioEngine.PlayNotification()
			}

			left := utils.Difference(m.connections, msg.Sessions)

			for _, v := range left {
				m.connections = slices.DeleteFunc(m.connections, func(n string) bool {
					return n == v
				})
				m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "System", Text: v + " disconnected!"})
				m.user.Engines.AudioEngine.PlayNotification()
			}

			if len(left) > 0 && len(m.connections) == 0 {
				m.connected = false
				if m.user.Engines.AudioEngine != nil {
					m.user.Engines.AudioEngine.SetDisconnected()
				}

			}

			if len(joined) > 0 && !m.updateTick {
				cmds = append(cmds, commands.UpdateTickCmd())
				m.updateTick = true
			}

			cmds = append(cmds, m.updateConnectionsList())

			if len(joined) > 0 && (m.state == states.CONN_STATE || (m.state == states.LOAD_STATE && m.prState == states.CONN_STATE)) {
				m.state = states.CHAT_STATE
				m.focusInputs()
			}

		}

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
				cmds = append(cmds, commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			}

			if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
				cmds = append(cmds, commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			}
		}

		cmds = append(cmds, commands.TickCmd())

	case commands.AnimTickMsg:

		if m.curWindow == windows.START_WINDOW {
			m.animFrame++
			return m, commands.AnimTickCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+с", "alt+й", "alt+Й", "ctrl+С", "ctrl+C", "ctrl+c", "alt+q", "alt+Q":
			m.Clean()
			return m, tea.Quit

		case "alt+s", "alt+ы", "alt+S", "alt+Ы":
			if m.state != states.START_STATE && m.user.Networking != nil {
				m.curWindow = windows.SETTINGS_WINDOW
			}

		case "alt+d", "alt+в", "alt+D", "alt+В":
			if m.curWindow == windows.SETTINGS_WINDOW {
				cmds = append(cmds, commands.OnOffDenoiceCmd(m.user))
			}

		case "alt+e", "alt+у", "alt+E", "alt+У":
			if m.curWindow == windows.SETTINGS_WINDOW {
				cmds = append(cmds, commands.OnOffAECCmd(m.user))
			}

		case "alt+f", "alt+F", "alt+а", "alt+А":
			if m.curWindow == windows.SETTINGS_WINDOW {
				cmds = append(cmds, commands.OnOffFilterCmd(m.user))
			}

		case "alt+c", "alt+с", "alt+C", "alt+С":
			if m.state != states.START_STATE {
				m.curWindow = windows.CONNECTIONS_WINDOW
				m.connectionsList.SetSize(m.width/2, m.height-4)
			}

		case "alt+v", "alt+М", "alt+V", "alt+м":
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteMicCmd(m.user.Engines.AudioEngine))
			}

		case "alt+up":
			if m.curWindow == windows.CONNECTIONS_WINDOW && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.SelectedItem().(connectionItem); ok {
					vc := i.volumeCoefficient
					if vc >= MAX_VOLUME {
						return m, nil
					}

					vc = float32(math.Round(float64(vc+0.1)*10) / 10)

					if vc > MAX_VOLUME {
						vc = MAX_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.nickname), vc), m.updateConnectionItemList(i.nickname, vc, i.muted))
				}

			}

		case "alt+down":
			if m.curWindow == windows.CONNECTIONS_WINDOW && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.SelectedItem().(connectionItem); ok {
					vc := i.volumeCoefficient
					if vc <= MIN_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc-0.1)*10) / 10)
					if vc < MIN_VOLUME {
						vc = MIN_VOLUME
					} else if vc == 0 {
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.nickname), vc), m.updateConnectionItemList(i.nickname, vc, i.muted))
				}
			}

		case "alt+h", "alt+H", "alt+р", "alt+Р":
			if m.state != states.START_STATE {
				m.curWindow = windows.HELP_WINDOW
			}

		case "alt+b", "alt+и", "alt+B", "alt+И":
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteCmd(m.user.Engines.AudioEngine))
			}

		case "alt+z", "alt+Z", "alt+я", "alt+Я":
			if m.curWindow == windows.CONNECTIONS_WINDOW && m.connected {
				if i, ok := m.connectionsList.SelectedItem().(connectionItem); ok {
					cmds = append(cmds, commands.MuteUnmuteUserCmd(m.user, ansi.Strip(i.nickname)), m.updateConnectionItemList(i.nickname, i.volumeCoefficient, !i.muted))
				}
			}

		case "esc":
			if m.curWindow != windows.DEF_WINDOW {
				prevWindow := m.curWindow
				m.curWindow = windows.DEF_WINDOW

				if prevWindow == windows.ERR_WINDOW {
					m.err = nil
					if m.prState == states.LOAD_STATE {
						m.state = states.DEF_STATE
					} else {
						m.state = m.prState
					}
				}
			}

			m.cursor = 0
			m.focusInputs()
			return m, textinput.Blink

		case "tab":
			if m.curWindow == windows.DEF_WINDOW {
				switch m.state {
				case states.LOAD_STATE:
					return m, nil
				case states.REG_STATE:
					if m.user.Networking != nil {
						m.state = states.PROFILE_STATE
					} else {
						m.state = states.LOGIN_STATE
					}
				case states.CONN_STATE:
					m.state = states.CHAT_STATE

				case states.CHAT_STATE:
					if m.user.Networking != nil {
						m.state = states.PROFILE_STATE
					} else {
						m.state = states.REG_STATE
					}

				case states.LOGIN_STATE:
					if m.user.Networking != nil {
						m.state = states.PROFILE_STATE
					} else {
						m.state = states.REG_STATE
					}

				case states.PROFILE_STATE:

					if m.connected {
						m.state = states.LEAVE_STATE
					} else {
						m.state = states.CONN_STATE
					}

				case states.LEAVE_STATE:
					m.state = states.CHAT_STATE
				case states.START_STATE:
					m.state = states.LOAD_STATE
					m.curWindow = windows.DEF_WINDOW
					if m.user.Data.Personal.Nickname != "" && m.user.Data.Personal.RegisterTime != "" && m.user.Networking == nil {
						cmds = append(cmds, commands.AuthCmd(m.user, m.log, nil))
					} else {
						m.state = states.DEF_STATE
						m.focusInputs()
					}

				default:
					if m.user.Networking != nil {
						m.state = states.PROFILE_STATE
					} else {
						m.state = states.REG_STATE
					}

				}

				m.cursor = 0
				m.curWindow = windows.DEF_WINDOW
				m.focusInputs()
				return m, textinput.Blink
			}

		case "up":
			if m.cursor > 0 && m.state != states.CHAT_STATE {
				m.cursor--
				m.focusInputs()
				return m, textinput.Blink
			}

		case "down":
			switch m.state {
			case states.REG_STATE:
				if m.cursor < len(m.regTextInputs)-1 {
					m.cursor++
					m.focusInputs()
					return m, textinput.Blink
				}
			case states.LOGIN_STATE:
				if m.cursor < len(m.logingInput)-1 {
					m.cursor++
					m.focusInputs()
					return m, textinput.Blink
				}
			case states.CONN_STATE:
				if m.cursor < len(m.connTextInputs)-1 {
					m.cursor++
					m.focusInputs()
					return m, textinput.Blink
				}
			}

		case "enter":
			if m.curWindow == windows.SETTINGS_WINDOW {
				if i, ok := m.microphonesList.SelectedItem().(micItem); ok {
					cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.name))
				}
				break
			} else if m.curWindow != windows.DEF_WINDOW && m.curWindow != windows.START_WINDOW {
				break
			}
			switch m.state {
			case states.REG_STATE:
				m.prState = m.state
				m.state = states.LOAD_STATE
				for i := range m.regTextInputs {
					if m.regTextInputs[i].Value() == "" {
						break
					}
				}

				m.user.Data.Personal.Nickname = m.regTextInputs[0].Value()
				password := m.regTextInputs[1].Value()
				repPassword := m.regTextInputs[2].Value()
				m.user.Data.Personal.RegisterTime = time.Now().Format("2006-01-02")

				cmds = append(cmds, commands.RegisterCmd(m.user, m.log, []byte(password), []byte(repPassword)))

				for i := range m.regTextInputs {
					m.regTextInputs[i].Reset()
				}

			case states.LOGIN_STATE:
				m.prState = m.state
				m.state = states.LOAD_STATE
				for i := range m.logingInput {
					if m.logingInput[i].Value() == "" {
						break
					}
				}

				m.user.Data.Personal.Nickname = m.logingInput[0].Value()
				password := m.logingInput[1].Value()

				cmds = append(cmds, commands.LoginCmd(m.user, m.log, []byte(password)))

				for i := range m.logingInput {
					m.logingInput[i].Reset()
				}

			// case states.SETTINGS_STATE:
			// 	if m.curWindow == windows.SETTINGS_WINDOW {
			// 		if i, ok := m.microphonesList.SelectedItem().(micItem); ok {
			// 			cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.name))
			// 		}
			// 	}

			case states.CONN_STATE:
				m.prState = m.state
				m.state = states.LOAD_STATE
				for i := range m.connTextInputs {
					if m.connTextInputs[i].Value() == "" {
						break
					}
				}

				cmds = append(cmds, commands.ConnectToUserCmd(m.user.Networking, m.connTextInputs[0].Value()))

				for i := range m.connTextInputs {
					m.connTextInputs[i].Reset()
				}

			case states.START_STATE:
				m.state = states.LOAD_STATE
				m.curWindow = windows.DEF_WINDOW
				if m.user.Data.Personal.Nickname != "" && m.user.Data.Personal.RegisterTime != "" && m.user.Networking == nil {
					cmds = append(cmds, commands.AuthCmd(m.user, m.log, nil))
				} else {
					m.state = states.DEF_STATE
					m.focusInputs()
				}

			case states.LEAVE_STATE:
				m.prState = m.state
				m.state = states.LOAD_STATE
				cmds = append(cmds, commands.LeaveCmd(m.user.Networking, m.user.Engines.AudioEngine))

			case states.CHAT_STATE:
				m.prState = m.state
				val := m.chatTextInput.Value()
				if val != "" && m.user.Networking != nil {
					cmds = append(cmds, commands.SendInChatCmd(m.user.Networking, val))
					m.messages = append(m.messages, commands.ChatMessage{
						Time:     time.Now().Format("15:04:05"),
						Nickname: lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(m.user.Data.Personal.Nickname),
						Text:     val})
					m.chatTextInput.Reset()
				}
			case states.PROFILE_STATE:
				m.curWindow = windows.PROFILE_WINDOW
			}
		}

	}

	if m.curWindow == windows.DEF_WINDOW {
		switch m.state {
		case states.REG_STATE:
			for i := range m.regTextInputs {
				m.regTextInputs[i], cmd = m.regTextInputs[i].Update(msg)
				cmds = append(cmds, cmd)
			}
		case states.CONN_STATE:
			for i := range m.connTextInputs {
				m.connTextInputs[i], cmd = m.connTextInputs[i].Update(msg)
				cmds = append(cmds, cmd)
			}

		case states.CHAT_STATE:
			m.chatTextInput, cmd = m.chatTextInput.Update(msg)
			cmds = append(cmds, cmd)

		case states.LOGIN_STATE:
			for i := range m.logingInput {
				m.logingInput[i], cmd = m.logingInput[i].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	switch m.curWindow {
	case windows.SETTINGS_WINDOW:
		m.microphonesList, cmd = m.microphonesList.Update(msg)
		cmds = append(cmds, cmd)

	case windows.CONNECTIONS_WINDOW:
		m.connectionsList, cmd = m.connectionsList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
