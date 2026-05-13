package tui

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/notifications"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"bytes"
	"fmt"
	"image"
	"math"
	"slices"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"golang.design/x/clipboard"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/blacktop/go-termimg"
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
	m.cursor = 0
	if !m.isLoggedIn() {
		switch m.activeTab {
		case 0:
			m.state = states.REG_STATE
		case 1:
			m.state = states.LOGIN_STATE
		}
		m.focusInputs()
	} else {
		switch m.activeTab {
		case 0:
			if m.connected {
				m.state = states.DEF_STATE
			} else {
				m.state = states.CONN_STATE
			}
			m.onlineList.Select(0)
			m.unfocusInputs()
		case 1:
			m.state = states.CHAT_STATE
			m.focusInputs()
		case 2:
			if m.connected {
				m.state = states.LEAVE_STATE
			} else {
				m.state = states.DEF_STATE
			}
		case 4:
			m.state = states.PROFILE_STATE
			m.focusInputs()
		case 5:
			m.microphonesList.Select(-1)
			m.settingsList.Select(0)
		default:
			m.state = states.DEF_STATE
		}

	}

	//m.focusInputs()
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
				} else if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
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
						m.sideState = 0
						m = m.syncTabState()
						return m, nil
					}
					m = m.syncTabState()
					return m, textinput.Blink
				}
			}

		case tea.MouseButtonWheelUp:
			if m.isLoggedIn() && m.activeTab == 1 && m.connected {
				maxOffset, _, _ := m.getChatSizes()
				if m.chatOffset < maxOffset {
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

	// case commands.UpdateTickMsg:
	// 	if m.connected {
	// 		return m, commands.UpdateTickCmd()
	// 	} else {
	// 		//m.updateTick = false
	// 		return m, nil
	// 	}

	case commands.ChangeThemeMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			m.themeColor = lipgloss.Color(m.user.Data.Setup.ThemeColor)
			m.subThemeColor = lipgloss.Color(utils.DarkenHex(m.user.Data.Setup.ThemeColor, 0.7))
			m.headerActiveStyle = lipgloss.NewStyle().Foreground(m.themeColor).Bold(true)

			m.microphonesDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
			m.microphonesDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

			m.settingsDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
			m.settingsDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

			m.connectionsDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

			m.onlineDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
			m.onlineDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

			m.microphonesList.SetDelegate(m.microphonesDelegate)
			m.settingsList.SetDelegate(m.settingsDelegate)
			m.connectionsList.SetDelegate(m.connectionsDelegate)
			m.onlineList.SetDelegate(m.onlineDelegate)

			m.state = m.prState
			m.unfocusInputs()
			return m, m.updateMicrophonesItemList(m.user.Data.Devices.Microphone)
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
			return m, textinput.Blink
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
				m.user.Data.Personal = entities.Personal{}
			}
		} else {
			m.activeTab = 0
			m = m.syncTabState()

			if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
				m.setupMicrohonesList()
				t := time.Now().Format("15:04:05")
				m.user.Networking.ChatCallback(func(id string, data []byte) {

					m.rawMsgChan <- commands.RawChatMessage{Time: t, Nickname: id, Data: data}

				})
				m.user.Networking.VoiceCallback(func(id string, data []byte) {
					m.user.Engines.AudioEngine.PlayUserVoice(id, data)
				})
				m.user.Networking.PeerConnectedCallback(func(id string) {
					m.peerConnectionsChan <- commands.PeerConnectedMsg{Nickname: id, Time: t}
				})
				m.user.Networking.PeerDisconnectedCallback(func(id string) {
					m.peerDisconnectionsChan <- commands.PeerDisconnectedMsg{Nickname: id, Time: t}
				})
			}

			cmds = append(cmds,
				commands.WaitForChatMessageCmd(m.msgChan), commands.WaitForRawChatMessageCmd(m.rawMsgChan),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
				commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan),
				commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname),
				commands.TickCmd(), textinput.Blink)
		}

	case commands.RawChatMessage:
		data := msg.Data
		textMsg := string(msg.Data)
		textForDesktopNotification := textMsg
		dataLen := len(data)
		if dataLen > 3 && slices.Equal(data[:3], []byte{'i', 'm', 'g'}) {
			img, _, err := image.Decode(bytes.NewReader(data[3:]))
			if err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForRawChatMessageCmd(m.rawMsgChan)
			}
			size := img.Bounds().Size()
			_, cw, ch := m.getChatSizes()
			if ch < 1 {
				ch = 1
			}

			if size.X > cw {
				size.X = cw
				size.Y = ch
			} else if size.Y > ch {
				size.X = cw
				size.Y = ch
			}

			imageWidget := termimg.NewImageWidgetFromImage(img)
			imageWidget.SetProtocol(termimg.Auto)
			imageWidget.SetSizeWithCorrection(size.X, size.Y)
			textMsg, err = imageWidget.Render()
			if err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForRawChatMessageCmd(m.rawMsgChan)
			}
			textForDesktopNotification = fmt.Sprintf("image with len: %d", dataLen)
			textMsg = fmt.Sprintf("\n%s", textMsg)
		}

		m.msgChan <- commands.ChatMessage{Nickname: msg.Nickname, Time: msg.Time, Text: textMsg}

		if m.user.Data.Setup.AudioNotifications {
			m.user.Engines.AudioEngine.PlayNotification()
		}
		if m.user.Data.Setup.DesktopNotifications {
			if err := notifications.Notify(msg.Time, msg.Nickname, textForDesktopNotification); err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForRawChatMessageCmd(m.rawMsgChan)
			}
		}
		return m, commands.WaitForRawChatMessageCmd(m.rawMsgChan)

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
			return m, m.updateOnlineList()
		}

	case commands.ChangeMicrophoneMessage:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		}

	case commands.UpdateMicrophonesMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			return m, m.updateMicrophonesList()
		}

	case commands.PeerConnectedMsg:
		color := lipgloss.Color(randomcolor.GetRandomColorInHex())
		nickname := lipgloss.NewStyle().Foreground(color).Render(msg.Nickname)
		m.connections = append(m.connections, nickname)
		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: nickname + " joined the chat!"})
		if !m.connected {
			m.connected = true
			if m.user.Engines.AudioEngine != nil {
				m.user.Engines.AudioEngine.SetConnected()
			}
		}

		cmds = append(cmds, commands.SetupUserVolumeCmd(m.user, msg.Nickname),
			commands.SetupUserMuteCmd(m.user, msg.Nickname), m.updateConnectionsList(), commands.WaitForPeerConnectionCmd(m.peerConnectionsChan))
		if m.activeTab == 0 || m.prState == states.CONN_STATE {
			m.activeTab = 1
			m = m.syncTabState()
		}
		m.user.Engines.AudioEngine.PlayNotification()

	case commands.PeerDisconnectedMsg:
		var colored string
		m.connections = slices.DeleteFunc(m.connections, func(n string) bool {
			if ansi.Strip(n) == msg.Nickname {
				colored = n
				return true
			}
			return false
		})
		if len(m.connections) == 0 {
			m.connected = false
			if m.user.Engines.AudioEngine != nil {
				m.user.Engines.AudioEngine.SetDisconnected()
			}
		}

		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: colored + " disconnected!"})
		cmds = append(cmds, m.updateConnectionsList(), commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan))
		m.user.Engines.AudioEngine.PlayNotification()

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" {
				cmds = append(cmds, commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			}
			if m.user.Engines.AudioEngine != nil && m.activeTab == 5 {
				cmds = append(cmds, commands.UpdateMicrophonesCmd(m.user))
			}

			// if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" && m.activeTab == 0 {
			// 	cmds = append(cmds, commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname))
			// }
			// if m.user.Engines.AudioEngine != nil && m.activeTab == 5 {
			// 	cmds = append(cmds, commands.UpdateMicrophonesCmd(m.user))
			// }
		}
		cmds = append(cmds, commands.TickCmd())

	case commands.AnimTickMsg:
		if m.curWindow == windows.START_WINDOW {
			m.animFrame++
			return m, commands.AnimTickCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "alt+й", "alt+Й", "alt+q", "alt+Q":
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

		case "ctrl+p", "ctrl+P", "ctrl+З", "ctrl+з":
			if m.connected && m.activeTab == 1 {
				m.log.Info("ctrl+p")
				textData := clipboard.Read(clipboard.FmtText)
				if len(textData) > 0 {
					m.chatTextInput.SetValue(string(textData))
					return m, nil
				}

				imgData := clipboard.Read(clipboard.FmtImage)

				lid := len(imgData)

				if lid > 0 {
					m.chatTextInput.SetValue(fmt.Sprintf("image with len: %d", lid))
					m.imageBuffer = imgData
					return m, nil
				}
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
		case "alt+left":
			if m.isLoggedIn() && (m.activeTab == 0 || m.activeTab == 5) {
				if m.sideState == 1 {
					m.sideState = 0
					m.unfocusInputs()

					switch m.activeTab {
					case 0:
						m.onlineList.Select(0)
					case 5:
						m.microphonesList.Select(-1)
						m.settingsList.Select(0)
					}
				}
			}

		case "alt+right":
			if m.isLoggedIn() && (m.activeTab == 0 || m.activeTab == 5) {
				if m.sideState == 0 {
					m.sideState = 1

					switch m.activeTab {
					case 0:
						m.onlineList.Select(-1)
						m.focusInputs()
					case 5:
						m.microphonesList.Select(0)
						m.settingsList.Select(-1)
					}
					return m, textinput.Blink
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
				m.sideState = 0
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
				m.sideState = 0
				return m, textinput.Blink
			}

		case "esc":
			switch m.state {
			case states.ERR_STATE:
				m.err = nil
				m.curWindow = windows.DEF_WINDOW
				if m.prState == states.LOAD_STATE {
					m = m.syncTabState()
				} else {
					m.state = m.prState
				}
			case states.HELP_STATE:
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
			if (!m.isLoggedIn() || (m.activeTab != 0 && m.activeTab != 2 && m.activeTab != 5)) && m.cursor > 0 {
				m.cursor--
				m.focusInputs()
				return m, textinput.Blink
			}

		case "down":
			if !m.isLoggedIn() || (m.activeTab != 0 && m.activeTab != 2 && m.activeTab != 5) {
				if m.activeTab == 0 && m.cursor < len(m.regTextInputs)-1 {
					m.cursor++
				} else if m.activeTab == 1 && m.cursor < len(m.logingInput)-1 {
					m.cursor++
				}

				m.focusInputs()
				return m, textinput.Blink
			}

		case "enter":
			if m.state == states.LOAD_STATE {
				return m, nil
			}
			if m.curWindow == windows.START_WINDOW {
				m.state = states.LOAD_STATE
				m.curWindow = windows.DEF_WINDOW
				m = m.syncTabState()
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
					if !m.connected {
						var nick string
						switch m.sideState {
						case 0:
							if i, ok := m.onlineList.SelectedItem().(onlineItem); ok {
								nick = i.name
							} else {
								return m, nil
							}
						case 1:
							if m.connTextInputs.Value() != "" {
								nick = m.connTextInputs.Value()
								m.connTextInputs.Reset()
							} else {
								return m, nil
							}
						default:
							return m, nil
						}
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.ConnectToUserCmd(m.user.Networking, nick))
					}

				case 1:
					val := m.chatTextInput.Value()
					toSend := []byte(val)
					if m.imageBuffer != nil {
						img, _, err := image.Decode(bytes.NewReader(m.imageBuffer))
						if err != nil {
							m.err = err
							m.state = states.ERR_STATE
							m.imageBuffer = nil
							return m, nil
						}
						size := img.Bounds().Size()
						_, cw, ch := m.getChatSizes()
						if ch < 1 {
							ch = 1
						}

						if size.X > cw {
							size.X = cw
							size.Y = ch
						} else if size.Y > ch {
							size.X = cw
							size.Y = ch
						}
						imageWidget := termimg.NewImageWidgetFromImage(img)
						imageWidget.SetProtocol(termimg.Auto)

						toSend = utils.SetThreeFirstByte([]byte{'i', 'm', 'g'}, m.imageBuffer)

						imageWidget.SetSizeWithCorrection(size.X, size.Y)

						rendered, err := imageWidget.Render()
						if err != nil {
							m.err = err
							m.state = states.ERR_STATE
							m.imageBuffer = nil
							return m, nil
						}

						val = fmt.Sprintf("\n%s", rendered)
						m.imageBuffer = nil
					}
					if val != "" && m.user.Networking != nil {
						cmds = append(cmds, commands.SendInChatCmd(m.user.Networking, toSend))
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
				case 4:
					m.prState = m.state
					m.state = states.LOAD_STATE
					newColor := m.themeColorInput.Value()
					if newColor == "d" {
						newColor = m.defaultThemeColor
					} else if !strings.HasPrefix(newColor, "#") {
						newColor = "#" + newColor
					}
					if len(newColor) != 7 {
						break
					}

					cmds = append(cmds, commands.ChangeThemeCmd(m.user, newColor))
					m.themeColorInput.Reset()
					m.unfocusInputs()
				case 5:
					switch m.sideState {
					case 1:
						if i, ok := m.microphonesList.SelectedItem().(micItem); ok {
							cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.name), m.updateMicrophonesItemList(i.name))
						}
					case 0:
						if i, ok := m.settingsList.SelectedItem().(settingsItem); ok {
							switch i.id {
							case DENOISE:
								cmds = append(cmds, commands.OnOffDenoiceCmd(m.user))
							case AEC:
								cmds = append(cmds, commands.OnOffAECCmd(m.user))
							case EQUALIZER:
								cmds = append(cmds, commands.OnOffFilterCmd(m.user))
							case AUDIO_N:
								cmds = append(cmds, commands.OnOffAudioNotifications(m.user))
							case DESKTOP_N:
								cmds = append(cmds, commands.OnOffDesktopNotifications(m.user))
							default:
								return m, nil
							}

							cmds = append(cmds, m.updateSettingsItemList(i.name))
						}
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
				if m.sideState == 1 {
					m.connTextInputs, cmd = m.connTextInputs.Update(msg)
					cmds = append(cmds, cmd)
				}
				if m.sideState == 0 {
					m.onlineList, cmd = m.onlineList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 1:
				m.chatTextInput, cmd = m.chatTextInput.Update(msg)
				cmds = append(cmds, cmd)
			case 2:
				if m.sideState == 0 {
					m.connectionsList, cmd = m.connectionsList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 4:
				m.themeColorInput, cmd = m.themeColorInput.Update(msg)
				cmds = append(cmds, cmd)
			case 5:
				if m.sideState == 1 {
					m.microphonesList, cmd = m.microphonesList.Update(msg)
					cmds = append(cmds, cmd)
				} else {
					m.settingsList, cmd = m.settingsList.Update(msg)
					cmds = append(cmds, cmd)
				}

			}
		}
	}

	return m, tea.Batch(cmds...)
}
