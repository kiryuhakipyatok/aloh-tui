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
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	MAX_VOLUME = 4.0
	MIN_VOLUME = 0.0
)

func (m Model) syncTabState() Model {
	m.curWindow = windows.DEF_WINDOW
	if !m.isLoggedIn() {
		switch m.activeTab {
		case 0:
			m.state = states.REG_STATE
		case 1:
			m.state = states.LOGIN_STATE
		}
	} else {
		switch m.activeTab {
		case 0:
			if m.connected {
				m.state = states.DEF_STATE
			} else {
				m.state = states.CONN_STATE
			}
		case 1:
			m.state = states.CHAT_STATE
		case 2:
			if m.connected {
				m.state = states.LEAVE_STATE
			} else {
				m.state = states.DEF_STATE
			}
		default:
			m.state = states.DEF_STATE
		}
	}
	m.cursor = 0
	m.focusInputs()
	return m
}

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
			if msg.Action == tea.MouseActionRelease {
				if m.curWindow == windows.START_WINDOW && m.zone.Get("start").InBounds(msg) {
					m.state = states.LOAD_STATE
					m.curWindow = windows.DEF_WINDOW
					if m.user.Data.Personal.Nickname != "" && m.user.Data.Personal.RegisterTime != "" && m.user.Networking == nil {
						cmds = append(cmds, commands.AuthCmd(m.user, m.log, nil))
					} else {
						m = m.syncTabState()
					}
					return m, textinput.Blink
				} else if m.curWindow == windows.DEF_WINDOW {
					if m.zone.Get("registerT").InBounds(msg) || m.zone.Get("registerW").InBounds(msg) {
						m.activeTab = 0
					} else if m.zone.Get("loginT").InBounds(msg) || m.zone.Get("loginW").InBounds(msg) {
						m.activeTab = 1
					} else if m.zone.Get("friendsT").InBounds(msg) || m.zone.Get("friendsW").InBounds(msg) {
						m.activeTab = 0
					} else if m.zone.Get("chatT").InBounds(msg) || m.zone.Get("chatW").InBounds(msg) {
						m.activeTab = 1
					} else if m.zone.Get("voiceT").InBounds(msg) || m.zone.Get("voiceW").InBounds(msg) {
						m.activeTab = 2
					} else if m.zone.Get("videoT").InBounds(msg) || m.zone.Get("videoW").InBounds(msg) {
						m.activeTab = 3
					} else if m.zone.Get("profileT").InBounds(msg) || m.zone.Get("profileW").InBounds(msg) {
						m.activeTab = 4
					} else if m.zone.Get("settingsT").InBounds(msg) || m.zone.Get("settingsW").InBounds(msg) {
						m.activeTab = 5
					} else {
						m.unfocusInputs()
						return m, nil
					}
					m = m.syncTabState()
					return m, textinput.Blink
				}
			}

		case tea.MouseButtonWheelUp:
			if m.isLoggedIn() && m.activeTab == 1 && m.connected {
				if m.chatOffset < m.getMaxChatOffset() {
					m.chatOffset++
				}
			}

		case tea.MouseButtonWheelDown:
			if m.isLoggedIn() && m.activeTab == 1 && m.connected {
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

	case commands.OnOffDenoiceMsg, commands.OnOffFilterMsg, commands.OnOffAECMsg, commands.UsersVolumeMsg, commands.MuteUnmuteUserMsg:
		var err error
		switch m := msg.(type) {
		case commands.OnOffDenoiceMsg:
			err = m.Err
		case commands.OnOffFilterMsg:
			err = m.Err
		case commands.OnOffAECMsg:
			err = m.Err
		case commands.UsersVolumeMsg:
			err = m.Err
		case commands.MuteUnmuteUserMsg:
			err = m.Err
		}

		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			m.focusInputs()
			return m, textinput.Blink
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

	case commands.ConnectToUserMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			m.connected = true
			m.user.Engines.AudioEngine.SetConnected()
			m.activeTab = 1
			m = m.syncTabState()
			cmds = append(cmds, textinput.Blink, commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			return m, tea.Batch(cmds...)
		}

	case commands.SendInChatMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			m.activeTab = 1
			m = m.syncTabState()
			return m, textinput.Blink
		}

	case commands.LeaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			m.connected = false
			m.user.Engines.AudioEngine.SetDisconnected()
			m.messages = []commands.ChatMessage{}
			m.connections = []string{}
			m.user.Engines.AudioEngine.PlayNotification()
			m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: "you disconnected!"})

			m.activeTab = 0
			m = m.syncTabState()
			return m, textinput.Blink
		}

	case commands.AuthMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
			if msg.Typee != auth.DEFAULT {
				m.user.Data = entities.Data{}
			}
		} else {
			m.activeTab = 0
			m = m.syncTabState()

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
		} else {
			m.online = msg.Online
		}

	case commands.ChangeMicrophoneMessage:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		}

	case commands.SessionsUpdateMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			joined := utils.Difference(msg.Sessions, m.connections)

			for _, v := range joined {
				color := lipgloss.Color(randomcolor.GetRandomColorInHex())
				nickname := lipgloss.NewStyle().Foreground(color).Render(v)
				m.connections = append(m.connections, nickname)
				m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: nickname + " joined the chat!"})
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
				m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: v + " disconnected!"})
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

			if len(joined) > 0 && (m.activeTab == 0 || m.prState == states.CONN_STATE) {
				m.activeTab = 1
				m = m.syncTabState()
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
				cmds = append(cmds, commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname), commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname))
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
		case "ctrl+c", "alt+й", "alt+Й", "ctrl+С", "ctrl+C", "alt+q", "alt+Q":
			m.Clean()
			return m, tea.Quit

		case "alt+s", "alt+ы", "alt+S", "alt+Ы":
			if m.isLoggedIn() {
				m.activeTab = 5
				m = m.syncTabState()
			}
		case "alt+c", "alt+с", "alt+C", "alt+С":
			if m.isLoggedIn() {
				m.activeTab = 0
				m = m.syncTabState()
			}
		case "alt+a", "alt+A", "alt+ф", "alt+Ф":
			if m.isLoggedIn() {
				m.activeTab = 2
				m = m.syncTabState()
			}

		case "alt+d", "alt+в", "alt+D", "alt+В":
			if m.isLoggedIn() && m.activeTab == 5 {
				cmds = append(cmds, commands.OnOffDenoiceCmd(m.user))
			}
		case "alt+e", "alt+у", "alt+E", "alt+У":
			if m.isLoggedIn() && m.activeTab == 5 {
				cmds = append(cmds, commands.OnOffAECCmd(m.user))
			}
		case "alt+f", "alt+F", "alt+а", "alt+А":
			if m.isLoggedIn() && m.activeTab == 5 {
				cmds = append(cmds, commands.OnOffFilterCmd(m.user))
			}

		case "alt+v", "alt+М", "alt+V", "alt+м":
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteMicCmd(m.user.Engines.AudioEngine))
			}
		case "alt+b", "alt+и", "alt+B", "alt+И":
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteCmd(m.user.Engines.AudioEngine))
			}

		case "alt+h", "alt+H", "alt+р", "alt+Р":
			if m.curWindow == windows.DEF_WINDOW {
				if m.state == states.HELP_STATE {
					m = m.syncTabState()
					return m, textinput.Blink
				}
				m.state = states.HELP_STATE
			}

		case "alt+z", "alt+Z", "alt+я", "alt+Я":
			if m.isLoggedIn() && m.activeTab == 2 && m.connected {
				if i, ok := m.connectionsList.SelectedItem().(connectionItem); ok {
					cmds = append(cmds, commands.MuteUnmuteUserCmd(m.user, ansi.Strip(i.nickname)), m.updateConnectionItemList(i.nickname, i.volumeCoefficient, !i.muted))
				}
			}

		case "alt+up":
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
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
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.SelectedItem().(connectionItem); ok {
					vc := i.volumeCoefficient
					if vc <= MIN_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc-0.1)*10) / 10)
					if vc < MIN_VOLUME {
						vc = MIN_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.nickname), vc), m.updateConnectionItemList(i.nickname, vc, i.muted))
				}
			}

		case "tab", "right":
			if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
				maxTabs := 2
				if m.isLoggedIn() {
					maxTabs = 6
				}
				if m.activeTab+1 >= maxTabs {
					m.activeTab = 0
				} else {
					m.activeTab++
				}
				m = m.syncTabState()
				return m, textinput.Blink
			}

		case "shift+tab", "left":
			if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
				maxTabs := 2
				if m.isLoggedIn() {
					maxTabs = 6
				}
				if m.activeTab-1 < 0 {
					m.activeTab = maxTabs - 1
				} else {
					m.activeTab--
				}
				m = m.syncTabState()
				return m, textinput.Blink
			}

		case "esc":
			if m.state == states.ERR_STATE {
				m.err = nil
				m.curWindow = windows.DEF_WINDOW
				if m.prState == states.LOAD_STATE {
					m = m.syncTabState()
				} else {
					m.state = m.prState
				}
			}
			return m, textinput.Blink

		case "up":
			if m.activeTab != 2 && m.activeTab != 5 && m.cursor > 0 {
				m.cursor--
				m.focusInputs()
				return m, textinput.Blink
			}

		case "down":
			if m.activeTab != 2 && m.activeTab != 5 {
				if !m.isLoggedIn() {
					if m.activeTab == 0 && m.cursor < len(m.regTextInputs)-1 {
						m.cursor++
					} else if m.activeTab == 1 && m.cursor < len(m.logingInput)-1 {
						m.cursor++
					}
				} else {
					if m.activeTab == 0 && m.cursor < len(m.connTextInputs)-1 {
						m.cursor++
					}
				}
				m.focusInputs()
				return m, textinput.Blink
			}

		case "enter":

			if m.curWindow == windows.START_WINDOW {
				m.state = states.LOAD_STATE
				m.curWindow = windows.DEF_WINDOW
				if m.user.Data.Personal.Nickname != "" && m.user.Data.Personal.RegisterTime != "" && m.user.Networking == nil {
					cmds = append(cmds, commands.AuthCmd(m.user, m.log, nil))
				} else {
					m = m.syncTabState()
				}
				m.focusInputs()
				return m, textinput.Blink
			}

			if m.state == states.ERR_STATE {
				m.curWindow = windows.DEF_WINDOW
				m = m.syncTabState()
				m.focusInputs()
				return m, textinput.Blink
			}

			if !m.isLoggedIn() {
				switch m.activeTab {
				case 0:
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

				case 1:
					m.prState = m.state
					m.state = states.LOAD_STATE
					m.user.Data.Personal.Nickname = m.logingInput[0].Value()
					password := m.logingInput[1].Value()

					cmds = append(cmds, commands.LoginCmd(m.user, m.log, []byte(password)))
					for i := range m.logingInput {
						m.logingInput[i].Reset()
					}
				}
			} else {
				switch m.activeTab {
				case 0:
					if !m.connected && m.connTextInputs[0].Value() != "" {
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.ConnectToUserCmd(m.user.Networking, m.connTextInputs[0].Value()))
						for i := range m.connTextInputs {
							m.connTextInputs[i].Reset()
						}
					}
				case 1:
					val := m.chatTextInput.Value()
					if val != "" && m.user.Networking != nil {
						cmds = append(cmds, commands.SendInChatCmd(m.user.Networking, val))
						m.messages = append(m.messages, commands.ChatMessage{
							Time:     time.Now().Format("15:04:05"),
							Nickname: lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(m.user.Data.Personal.Nickname),
							Text:     val})
						m.chatTextInput.Reset()
					}
				case 2:
					if m.connected {
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.LeaveCmd(m.user.Networking, m.user.Engines.AudioEngine))
					}
				case 5:
					if i, ok := m.microphonesList.SelectedItem().(micItem); ok {
						cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.name))
					}
				}
			}
		}
	}

	if m.curWindow == windows.DEF_WINDOW {
		if !m.isLoggedIn() {
			switch m.activeTab {
			case 0:
				for i := range m.regTextInputs {
					m.regTextInputs[i], cmd = m.regTextInputs[i].Update(msg)
					cmds = append(cmds, cmd)
				}
			case 1:
				for i := range m.logingInput {
					m.logingInput[i], cmd = m.logingInput[i].Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		} else {
			switch m.activeTab {
			case 0:
				for i := range m.connTextInputs {
					m.connTextInputs[i], cmd = m.connTextInputs[i].Update(msg)
					cmds = append(cmds, cmd)
				}
			case 1:
				m.chatTextInput, cmd = m.chatTextInput.Update(msg)
				cmds = append(cmds, cmd)
			case 2:
				m.connectionsList, cmd = m.connectionsList.Update(msg)
				cmds = append(cmds, cmd)
			case 5:
				m.microphonesList, cmd = m.microphonesList.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}
