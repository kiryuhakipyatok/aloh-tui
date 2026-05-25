package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
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
	if !m.isLoggedIn() {
		//if m.state != states.LOAD_STATE {
		switch m.activeTab {
		case 0:
			m.state = states.REG_STATE
		case 1:
			m.state = states.LOGIN_STATE
		}
		//}
		m.focusInputs()
	} else {

		switch m.activeTab {
		case 0:
			// if m.connected {
			// 	m.state = states.DEF_STATE
			// } else {
			switch m.sideState {
			case states.LEFT_STATE:
				m.state = states.CONN_STATE
				m.onlineList.Select(0)
				m.unfocusInputs()
			case states.RIGHT_STATE:
				switch m.cursor {
				case 0:
					m.state = states.CONN_STATE
				case 1:
					m.state = states.FRIEND_STATE
				}
				m.focusInputs()
			}
			//}
			delete(m.tabsNotifications, "friends")

		case 1:
			m.state = states.CHAT_STATE
			delete(m.tabsNotifications, "chat")
			m.focusInputs()
		case 2:
			if m.connected {
				m.state = states.LEAVE_STATE
			} else {
				m.state = states.DEF_STATE
			}
			delete(m.tabsNotifications, "voice")
		case 4:
			m.state = states.PROFILE_STATE
			if m.cursor < len(m.profileInputs) {
				m.friendsReqsList.Select(-1)
				m.focusInputs()
			} else if m.cursor == len(m.profileInputs) {
				m.friendsReqsList.Select(0)
				m.unfocusInputs()
			}
			delete(m.tabsNotifications, "profile")

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
	cmds := make([]tea.Cmd, 0, 5)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		if m.state == states.LOAD_STATE {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionRelease {
				if m.curWindow == windows.START_WINDOW && m.zone.Get("start").InBounds(msg) {
					m.state = states.LOAD_STATE
					m.curWindow = windows.DEF_WINDOW
					if m.user.Data.Personal.Nickname != "" && m.user.Data.Personal.RegisterTime != "" && m.user.Networking == nil {
						cmds = append(cmds, commands.AuthCmd(m.user, m.eventsChan, m.log, nil))
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
						m.sideState = states.LEFT_STATE
						m = m.syncTabState()
						m.unfocusInputs()
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
	case sshclient.Event:
		nickname := msg.Data
		switch msg.Type {
		case sshclient.NEW_FRIEND:
			m.user.NewFriendReq(msg.Data)
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			cmds = append(cmds, m.updateFriendsReqList(),
				commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
				commands.NotifyCmd(nickname, "new friend request"))
		case sshclient.ACCEPT_FRIEND:
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			cmds = append(cmds, commands.IncreaseAmountOfFriendsCmd(m.user, nickname),
				commands.NotifyCmd(nickname, "your new friend"))
		case sshclient.DELETE_FRIEND:
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			delete(m.online, nickname)
			cmds = append(cmds, commands.DecreaseAmountOfFriendsCmd(m.user, nickname),
				m.updateOnlineList(), commands.NotifyCmd(nickname, "no longer your friend"))
		}
		cmds = append(cmds, commands.WaitForEventMessageCmd(m.eventsChan))

	case commands.ThemeColorMsg:
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

			m.friendsReqsDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
			m.friendsReqsDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

			m.microphonesList.SetDelegate(m.microphonesDelegate)
			m.settingsList.SetDelegate(m.settingsDelegate)
			m.connectionsList.SetDelegate(m.connectionsDelegate)
			m.onlineList.SetDelegate(m.onlineDelegate)
			m.friendsReqsList.SetDelegate(m.friendsReqsDelegate)

			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}

			cmd = m.updateMicrophonesItemList(m.user.Data.Devices.Microphone)
			return m, cmd
		}

	case commands.BFTagMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}
			cmds = append(cmds, m.updateConnectionsList(), m.updateOnlineList())
			return m, cmd
		}

	case commands.FriendsMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}
			nickname := msg.Nickname
			switch msg.Typee {
			case sshclient.ACCEPT_FRIEND:
				m.user.DeleteFriendReq(nickname)
				return m, m.updateFriendsReqList()
			case sshclient.DENY_FRIEND:
				m.user.DeleteFriendReq(nickname)
				return m, m.updateFriendsReqList()
			case sshclient.DELETE_FRIEND:
				delete(m.online, nickname)
				cmds = append(cmds, m.updateOnlineList())
				return m, tea.Batch(cmds...)
			case sshclient.BLOCK_USER:
				delete(m.online, nickname)
				m.user.DeleteFriendReq(nickname)
				cmds = append(cmds, m.updateOnlineList(), m.updateFriendsReqList())
				return m, tea.Batch(cmds...)
			case sshclient.NEW_FRIEND:
			}
		}

	case commands.OnOffDenoiceMsg, commands.OnOffFilterMsg, commands.OnOffAECMsg, commands.UsersVolumeMsg,
		commands.MuteUnmuteUserMsg, commands.StatiscticsMsg, commands.NotificaionSignMsg:
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
		case commands.StatiscticsMsg:
			err = m.Err
		case commands.NotificaionSignMsg:
			err = m.Err
		}

		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			if m.state == states.LOAD_STATE && (m.activeTab == 2 || m.activeTab == 4 || m.activeTab == 5) {
				m.state = m.prState
			}
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
			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}
		}
		// } else if !m.connected {
		// 	m.log.Info("state before conn cmd", logger.Attr("state", m.state), logger.Attr("prState", m.prState))
		// 	m.connected = true
		// 	if m.user.Engines.AudioEngine != nil {
		// 		m.user.Engines.AudioEngine.SetConnected()
		// 	}
		// 	m.log.Info("state after conn cmd", logger.Attr("state", m.state), logger.Attr("prState", m.prState))
		// 	cmds = append(cmds, commands.CountMaxTimeInConnectionCmd(m.user, m.stopCountMinutesChan), commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
		// 		commands.IncreaseAmountOfConnectionsCmd(m.user), textinput.Blink)
		// 	m.activeTab = 1
		// 	m = m.syncTabState()
		// }

	case commands.SendInChatMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			m.activeTab = 1
			m = m.syncTabState()
			cmds = append(cmds, commands.IncreaseAmountOfMessagesCmd(m.user), textinput.Blink)
			return m, tea.Batch(cmds...)
		}

	case commands.LeaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else if m.connected {
			if m.user.Engines.AudioEngine != nil {
				if err := m.user.Engines.AudioEngine.SetDisconnected(); err != nil {
					m.err = err
					m.state = states.ERR_STATE
					return m, nil
				}
			}
			if m.prState == states.CONN_STATE {
				m.state = states.LOAD_STATE
			}
			m.connected = false
			select {
			case m.stopCountMinutesChan <- struct{}{}:
			default:
			}
			m.messages = []commands.ChatMessage{}
			m.connections = []string{}
			m.user.Engines.AudioEngine.PlayNotification()
			m.activeTab = 0
			m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: "you disconnected!"})

		}

	case commands.AuthMsg:
		if msg.Err != nil {
			if m.user.Networking != nil {
				m.user.Networking.Close()
			}
			if m.user.Engines.AudioEngine != nil {
				m.user.Engines.AudioEngine.Stop()
			}
			m.err = msg.Err
			m.state = states.ERR_STATE
			if msg.Typee != sshclient.DEFAULT {
				m.user.Data.Personal = users.Personal{}
			}
		} else {
			m.activeTab = 0
			m = m.syncTabState()

			if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
				m.user.Networking.ChatCallback(func(id string, data []byte) {
					t := time.Now().Format("15:04:05")
					m.rawMsgChan <- commands.RawChatMessage{Time: t, Nickname: id, Data: data}

				})
				m.user.Networking.VoiceCallback(func(id string, data []byte) {
					m.user.Engines.AudioEngine.PlayUserVoice(id, data)
				})
				m.user.Networking.PeerConnectedCallback(func(id string) {
					t := time.Now().Format("15:04:05")
					m.peerConnectionsChan <- commands.PeerConnectedMsg{Nickname: id, Time: t}
				})
				m.user.Networking.PeerDisconnectedCallback(func(id string) {
					t := time.Now().Format("15:04:05")
					m.peerDisconnectionsChan <- commands.PeerDisconnectedMsg{Nickname: id, Time: t}
				})
			}

			m.setupMicrohonesList()
			m.setupSettingsList()

			cmds = append(cmds,
				commands.WaitForChatMessageCmd(m.msgChan), commands.WaitForRawChatMessageCmd(m.rawMsgChan),
				commands.WaitForEventMessageCmd(m.eventsChan),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
				commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan),
				commands.FetchOnlineFriendsCmd(m.user.Networking, m.user.Data.Personal.Friends),
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
			imageWidget.SetSizeWithCorrection(int(float64(size.X)*1.5), int(float64(size.Y)*1.5))
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
			cmds = append(cmds, commands.PlayNotificationCmd(m.user.Engines.AudioEngine))
		}
		if m.user.Data.Setup.DesktopNotifications {
			cmds = append(cmds, commands.NotifyCmd(m.user.Data.Personal.Nickname, textForDesktopNotification))
		}
		cmds = append(cmds, commands.WaitForRawChatMessageCmd(m.rawMsgChan))
		return m, tea.Batch(cmds...)

	case commands.ChatMessage:
		msg.Nickname = m.coloredNickname(msg.Nickname)
		m.messages = append(m.messages, msg)
		if m.activeTab != 1 {
			m.tabsNotifications["chat"] = struct{}{}
		}
		return m, commands.WaitForChatMessageCmd(m.msgChan)

	case commands.OnlineMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			if msg.Online != nil && !isEqualOnline(m.online, msg.Online) {
				m.online = msg.Online
				if m.activeTab != 0 {
					m.tabsNotifications["friends"] = struct{}{}
				}
				cmd = m.updateOnlineList()
				return m, cmd
			}
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
			cmd = m.updateMicrophonesList()
			return m, cmd
		}

	case commands.PeerConnectedMsg:
		hex := randomcolor.GetRandomColorInHex()
		color := lipgloss.Color(hex)
		nickname := lipgloss.NewStyle().Foreground(color).Render(msg.Nickname)
		m.usersColors[msg.Nickname] = userColors{
			mainColor: color,
			subColor:  lipgloss.Color(utils.DarkenHex(hex, 0.7)),
		}
		m.connections = append(m.connections, nickname)
		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: nickname + " joined the chat!"})

		if _, ok := m.user.Data.Setup.UsersSetup[msg.Nickname]; !ok {
			m.user.Data.Setup.UsersSetup[msg.Nickname] = &users.UsersSetup{
				VolumeCoefficient: 1,
			}

			if err := m.user.UpdateUserJSON(); err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, nil
			}

		}

		cmds = append(cmds, commands.IncreaseAmountOfConnectionsByUser(m.user, msg.Nickname), commands.SetupUserVolumeCmd(m.user, msg.Nickname),
			commands.SetupUserMuteCmd(m.user, msg.Nickname), m.updateConnectionsList(),
			commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
			commands.WaitForPeerConnectionCmd(m.peerConnectionsChan))
		if !m.connected {
			m.connected = true
			cmds = append(cmds, commands.CountMaxTimeInConnectionCmd(m.user, m.stopCountMinutesChan),
				commands.IncreaseAmountOfConnectionsCmd(m.user))
			if m.user.Engines.AudioEngine != nil {
				m.user.Engines.AudioEngine.SetConnected()
			}
		}
		if m.activeTab != 2 {
			m.tabsNotifications["voice"] = struct{}{}
		}
		if m.activeTab == 0 {
			m.activeTab = 1
			m = m.syncTabState()
		}

	case commands.PeerDisconnectedMsg:
		var colored string
		m.connections = slices.DeleteFunc(m.connections, func(n string) bool {
			if ansi.Strip(n) == msg.Nickname {
				colored = n
				return true
			}
			return false
		})
		if m.activeTab != 2 {
			m.tabsNotifications["voice"] = struct{}{}
		}
		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: colored + " disconnected!"})
		cmds = append(cmds, m.updateConnectionsList(),
			commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
			commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan))
		if len(m.connections) == 0 && m.connected {
			m.connected = false
			select {
			case m.stopCountMinutesChan <- struct{}{}:
			default:
			}

			if m.user.Engines.AudioEngine != nil {
				if err := m.user.Engines.AudioEngine.SetDisconnected(); err != nil {
					m.err = err
					m.state = states.ERR_STATE
				}
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" && len(m.user.Data.Personal.Friends) > 0 {
				cmds = append(cmds, commands.FetchOnlineFriendsCmd(m.user.Networking, m.user.Data.Personal.Friends))
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
		//if m.curWindow == windows.START_WINDOW || (m.activeTab == 2 && m.isLoggedIn()) {
		m.animFrame++
		return m, commands.AnimTickCmd()
	//}
	case commands.PulseTickMsg:
		//if m.curWindow == windows.START_WINDOW || (m.activeTab == 2 && m.isLoggedIn()) {
		m.pulseFrame++
		return m, commands.PulseTickCmd()
		//}

	case tea.KeyMsg:
		if m.state == states.LOAD_STATE {
			switch msg.String() {
			case "alt+q", "alt+Q", "ctrl+c":
				return m, tea.Quit
			default:
				return m, nil
			}
		}
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
				if i, ok := m.connectionsList.SelectedItem().(lists.ConnectionItem); ok {
					cmds = append(cmds, commands.MuteUnmuteUserCmd(m.user, ansi.Strip(i.Nickname)),
						m.updateConnectionItemList(i.Nickname, i.VolumeCoefficient, !i.Muted))
				}
			}
		case "alt+x", "alt+X", "alt+ч", "alt+Ч":
			if m.isLoggedIn() && m.activeTab == 4 && m.friendsReqsList.Index() >= 0 {
				if i, ok := m.friendsReqsList.SelectedItem().(lists.FriendReqItem); ok {
					cmds = append(cmds, commands.DenyFriendRequestCmd(m.user, i.Nickname))
				}
			}

		case "alt+up":
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.SelectedItem().(lists.ConnectionItem); ok {
					vc := i.VolumeCoefficient
					if vc >= MAX_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc+0.1)*10) / 10)
					if vc > MAX_VOLUME {
						vc = MAX_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.Nickname), vc),
						m.updateConnectionItemList(i.Nickname, vc, i.Muted))
				}
			}

		case "alt+down":
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.SelectedItem().(lists.ConnectionItem); ok {
					vc := i.VolumeCoefficient
					if vc <= MIN_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc-0.1)*10) / 10)
					if vc < MIN_VOLUME {
						vc = MIN_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.Nickname), vc),
						m.updateConnectionItemList(i.Nickname, vc, i.Muted))
				}
			}
		case "alt+left":
			if m.isLoggedIn() && (m.activeTab == 0 || m.activeTab == 5) {
				if m.sideState == states.RIGHT_STATE {
					m.sideState = states.LEFT_STATE
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
				if m.sideState == states.LEFT_STATE {
					m.sideState = states.RIGHT_STATE

					switch m.activeTab {
					case 0:
						m.onlineList.Select(-1)
						m.friendsReqsList.Select(-1)
						m.focusInputs()
					case 5:
						m.microphonesList.Select(0)
						m.settingsList.Select(-1)
					}
					return m, textinput.Blink
				}
			}

		case "tab":
			if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
				m.cursor = 0
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
				m.sideState = states.LEFT_STATE
				return m, textinput.Blink
			}
			if m.curWindow == windows.START_WINDOW {
				m.state = states.LOAD_STATE
				m.curWindow = windows.DEF_WINDOW
				m = m.syncTabState()
				m.focusInputs()
				return m, textinput.Blink
			}

		case "right":
			if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
				m.cursor = 0
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
				m.sideState = states.LEFT_STATE
				return m, textinput.Blink
			}

		case "shift+tab", "left":
			if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
				m.cursor = 0
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
				m.sideState = states.LEFT_STATE
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
			var changed bool
			if !m.isLoggedIn() && m.cursor > 0 {
				if m.cursor > 0 {
					m.cursor--
					changed = true
				}

			} else if m.cursor > 0 {
				switch m.activeTab {
				case 0:
					if m.sideState == states.RIGHT_STATE {
						m.cursor--
						changed = true
					}
				case 4:
					if m.cursor == len(m.profileInputs) {
						if m.friendsReqsList.Index() <= 0 {
							m.cursor--
							changed = true
						}
					} else {
						m.cursor--
						changed = true
					}
				}
			}
			if changed {
				m = m.syncTabState()
				return m, textinput.Blink
			}

		case "down":
			var changed bool
			if !m.isLoggedIn() {
				switch m.activeTab {
				case 0:
					if m.cursor < len(m.regTextInputs)-1 {
						m.cursor++
						changed = true
					}
				case 1:
					if m.cursor < len(m.logingInput)-1 {
						m.cursor++
						changed = true
					}
				}

			} else {
				switch m.activeTab {
				case 0:
					if m.sideState == states.RIGHT_STATE && m.cursor < len(m.friendsInputs)-1 {
						m.cursor++
						changed = true
					}
				case 4:
					if m.cursor < len(m.profileInputs) {
						if len(m.friendsReqsList.Items()) <= 0 && m.cursor == len(m.profileInputs)-1 {
							changed = true
							break
						}
						m.cursor++
						changed = true
					}
				}
			}
			if changed {
				m = m.syncTabState()
				return m, textinput.Blink
			}
			// m.focusInputs()
			// return m, textinput.Blink

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
				m.err = nil
				m.curWindow = windows.DEF_WINDOW
				if m.prState == states.LOAD_STATE {
					m = m.syncTabState()
				} else {
					m.state = m.prState
				}
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

					cmds = append(cmds, commands.RegisterCmd(m.user, m.eventsChan, m.log, []byte(password), []byte(repPassword)))
					for i := range m.regTextInputs {
						m.regTextInputs[i].Reset()
					}

				case 1:
					m.prState = m.state
					m.state = states.LOAD_STATE
					m.user.Data.Personal.Nickname = m.logingInput[0].Value()
					password := m.logingInput[1].Value()

					cmds = append(cmds, commands.LoginCmd(m.user, m.eventsChan, m.log, []byte(password)))
					for i := range m.logingInput {
						m.logingInput[i].Reset()
					}
				}
			} else {
				switch m.activeTab {
				case 0:

					var nick string
					switch m.sideState {
					case states.LEFT_STATE:
						if i, ok := m.onlineList.SelectedItem().(lists.OnlineItem); ok {
							nick := i.Name
							conns := i.Connections
							if !m.connected {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToUserCmd(m.user.Networking, nick))
							} else if m.connected && !slices.Contains(conns, m.user.Data.Personal.Nickname) {
								m.prState = m.state
								m.state = states.LOAD_STATE

								sequence := tea.Sequence(commands.LeaveCmd(m.user.Networking), commands.ConnectToUserCmd(m.user.Networking, nick))
								cmds = append(cmds, sequence)
							}
						}

					case states.RIGHT_STATE:
						switch m.cursor {
						case 0:
							if !m.connected {
								if m.friendsInputs[m.cursor].Value() != "" {
									nick = m.friendsInputs[m.cursor].Value()
									m.friendsInputs[m.cursor].Reset()
								} else {
									return m, nil
								}
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToUserCmd(m.user.Networking, nick))
							}
						case 1:
							if m.friendsInputs[m.cursor].Value() != "" {
								nick = m.friendsInputs[m.cursor].Value()
								m.friendsInputs[m.cursor].Reset()
							} else {
								return m, nil
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.SendFriendRequestCmd(m.user, nick))
						case 2:
							if m.friendsInputs[m.cursor].Value() != "" {
								nick = m.friendsInputs[m.cursor].Value()
								m.friendsInputs[m.cursor].Reset()
							} else {
								return m, nil
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, nick))
						case 3:
							if m.friendsInputs[m.cursor].Value() != "" {
								nick = m.friendsInputs[m.cursor].Value()
								m.friendsInputs[m.cursor].Reset()
							} else {
								return m, nil
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.BlockUserCmd(m.user, nick))
						case 4:
							if m.friendsInputs[m.cursor].Value() != "" {
								nick = m.friendsInputs[m.cursor].Value()
								m.friendsInputs[m.cursor].Reset()
							} else {
								return m, nil
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.UnblockUserCmd(m.user, nick))
						}

						m.friendsInputs[m.cursor].Reset()
					default:
						return m, nil
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

						imageWidget.SetSizeWithCorrection(int(float64(size.X)*1.5), int(float64(size.Y)*1.5))

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
						t := time.Now().Format("15:04:05")
						m.messages = append(m.messages, commands.ChatMessage{
							Time:     t,
							Nickname: lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(m.user.Data.Personal.Nickname),
							Text:     val})
						m.chatTextInput.Reset()
					}
				case 2:
					if m.connected {
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.LeaveCmd(m.user.Networking))
					}
				case 4:
					switch m.cursor {
					case 0:
						newColor := strings.TrimSpace(m.profileInputs[0].Value())
						if newColor == "" {

							return m, nil
						}
						if newColor == "d" {
							newColor = m.defaultThemeColor
						} else if !strings.HasPrefix(newColor, "#") {
							newColor = "#" + newColor
						}
						if len(newColor) != 7 {
							return m, nil
						}
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.ChangeThemeColorCmd(m.user, newColor))
						m.profileInputs[0].Reset()
					case 1:
						newBFTag := strings.TrimSpace(m.profileInputs[1].Value())
						if newBFTag == "d" {
							newBFTag = m.defaultBFTag
						}
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.ChangeBFTagCmd(m.user, newBFTag))
						m.profileInputs[1].Reset()
					case 2:
						newNotifySign := strings.TrimSpace(m.profileInputs[2].Value())
						if newNotifySign == "d" {
							newNotifySign = m.defaultNotificationSign
						}
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.ChangeNotificationSignCmd(m.user, newNotifySign))
						m.profileInputs[2].Reset()
					case 3:
						if i, ok := m.friendsReqsList.SelectedItem().(lists.FriendReqItem); ok {
							m.prState = m.state
							nick := i.Nickname
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.AcceptFriendRequestCmd(m.user, nick))
						}
					}

				case 5:
					switch m.sideState {
					case states.RIGHT_STATE:
						if i, ok := m.microphonesList.SelectedItem().(lists.MicItem); ok {
							cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.Name), m.updateMicrophonesItemList(i.Name))
						}
					case states.LEFT_STATE:
						if i, ok := m.settingsList.SelectedItem().(lists.SettingsItem); ok {
							switch i.Id {
							case HARD_DENOISE:
								cmds = append(cmds, commands.OnOffHardDenoiceCmd(m.user))
							case SOFT_DENOISE:
								cmds = append(cmds, commands.OnOffSoftDenoiceCmd(m.user))
							case AEC:
								cmds = append(cmds, commands.OnOffAECCmd(m.user))
							case EQUALIZER:
								cmds = append(cmds, commands.OnOffFilterCmd(m.user))
							case APP_N:
								cmds = append(cmds, commands.OnOffAppNotifications(m.user))
							case AUDIO_N:
								cmds = append(cmds, commands.OnOffAudioNotifications(m.user))
							case DESKTOP_N:
								cmds = append(cmds, commands.OnOffDesktopNotifications(m.user))
							default:
								return m, nil
							}

							cmds = append(cmds, m.updateSettingsItemList(i.Id))
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
				if m.sideState == states.RIGHT_STATE {
					for i := range m.friendsInputs {
						m.friendsInputs[i], cmd = m.friendsInputs[i].Update(msg)
						cmds = append(cmds, cmd)
					}
				}
				if m.sideState == states.LEFT_STATE {
					m.onlineList, cmd = m.onlineList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 1:
				m.chatTextInput, cmd = m.chatTextInput.Update(msg)
				cmds = append(cmds, cmd)
			case 2:
				if m.sideState == states.LEFT_STATE {
					m.connectionsList, cmd = m.connectionsList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 4:
				if m.cursor < len(m.profileInputs) {
					for i := range m.profileInputs {
						m.profileInputs[i], cmd = m.profileInputs[i].Update(msg)
						cmds = append(cmds, cmd)
					}
				} else if m.cursor == len(m.profileInputs) {
					m.friendsReqsList, cmd = m.friendsReqsList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 5:
				if m.sideState == states.RIGHT_STATE {
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