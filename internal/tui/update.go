package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"bytes"
	"fmt"
	"image"
	"math"
	"runtime/debug"
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
				lists.SetListVisible(&m.onlineList.DefList, true)
				m.unfocusInputs()
			case states.RIGHT_STATE:
				switch m.cursor {
				case 0:
					m.state = states.CONN_STATE
				case 1:
					m.state = states.FRIEND_STATE
				}
				m.focusInputs()
				lists.SetListVisible(&m.onlineList.DefList, false)
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
			delete(m.tabsNotifications, "profile")
			switch m.sideState {
			case states.LEFT_STATE:
				lists.SetListVisible(&m.friendsReqsList.DefList, true)
				lists.SetListVisible(&m.apearenceList.DefList, false)
				m.unfocusInputs()
			case states.RIGHT_STATE:
				if m.cursor < len(m.profileInputs) {
					//m.apearenceList.Select(-1)
					lists.SetListVisible(&m.friendsReqsList.DefList, false)
					lists.SetListVisible(&m.apearenceList.DefList, false)
					m.focusInputs()
				} else if m.cursor == len(m.profileInputs) {
					lists.SetListVisible(&m.friendsReqsList.DefList, false)
					lists.SetListVisible(&m.apearenceList.DefList, true)
					//m.apearenceList.Select(0)
					m.unfocusInputs()
				}
			}

		case 5:
			switch m.state {
			case states.AUDIO_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					lists.SetListVisible(&m.audioList.DefList, true)
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					lists.SetListVisible(&m.audioList.DefList, false)
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.NOTIFICATIONS_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					lists.SetListVisible(&m.notificationsList.DefList, true)
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					lists.SetListVisible(&m.notificationsList.DefList, false)
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.DEVICES_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					lists.SetListVisible(&m.microphonesList.DefList, true)
					lists.SetListVisible(&m.headphonesList.DefList, false)
				case states.LEFT_STATE:
					lists.SetListVisible(&m.microphonesList.DefList, false)
					lists.SetListVisible(&m.headphonesList.DefList, true)
				}
			case states.BINDS_STATE:
			default:
				m.state = states.SETTINGS_STATE
				m.sideState = states.LEFT_STATE
				lists.SetListVisible(&m.settingsList.DefList, true)
			}
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
						cmds = append(cmds, commands.AuthCmd(m.user, m.sshEventsChan, m.log, nil))
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
						m.sideState = states.ZERO_STATE
						m = m.syncTabState()
						m.unfocusInputs()
						m.unfocusLists()
						return m, nil
					}

					if m.zone.Get("leftSide").InBounds(msg) && m.sideState != states.LEFT_STATE {
						m.sideState = states.LEFT_STATE
					} else if m.zone.Get("rightSide").InBounds(msg) && m.sideState != states.RIGHT_STATE {
						m.sideState = states.RIGHT_STATE
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

	// case commands.NetworkEventMsg:
	// 	nickname := msg.Nickname
	// 	us, ok := m.usersStates[nickname]
	// 	if ok {
	// 		switch msg.Event.Typee {
	// 		case alohnetwork.MIC_MUTE:
	// 			us.fullMute = false
	// 			us.micMute = msg.Event.State
	// 		case alohnetwork.FULL_MUTE:
	// 			us.fullMute = msg.Event.State
	// 			us.micMute = false
	// 		}
	// 	}

	case sshclient.Event:
		nickname := msg.Data
		switch msg.Type {
		case sshclient.NEW_FRIEND_REQ:
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			cmds = append(cmds, commands.NewFriendReqCmd(m.user, nickname))
		case sshclient.ACCEPT_FRIEND:
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			cmds = append(cmds, commands.NewFriendCmd(m.user, nickname))
		case sshclient.DELETE_FRIEND:
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			delete(m.online, nickname)
			cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, nickname, false),
				m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)), commands.NotifyCmd(nickname, "no longer your friend"))
		case sshclient.BLOCK_USER:
			if m.user.IsFriend(nickname) {
				if m.activeTab != 4 {
					m.tabsNotifications["profile"] = struct{}{}
				}
				delete(m.online, nickname)
				cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, nickname, false),
					m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)), commands.NotifyCmd(nickname, "blocked you"))
			}

		case sshclient.FRIEND_ONLINE:
			//if m.user.IsFriend(nickname) {
			if m.activeTab != 0 {
				m.tabsNotifications["friends"] = struct{}{}
			}

			m.online[nickname] = make([]string, 0)
			cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)))
		//}
		case sshclient.FRIEND_OFFLINE:
			m.log.Info("friend offline")
			delete(m.online, nickname)
			cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)))

		}
		cmds = append(cmds, commands.WaitForSSHEventMessageCmd(m.sshEventsChan))

	case commands.AppereanceMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}
			switch msg.Typee {
			case commands.COLOR:
				m.themeColor = lipgloss.Color(m.user.Data.Setup.Appereance.ThemeColor)
				m.subThemeColor = lipgloss.Color(utils.DarkenHex(m.user.Data.Setup.Appereance.ThemeColor, 0.7))
				m.headerActiveStyle = lipgloss.NewStyle().Foreground(m.themeColor).Bold(true)

				m.microphonesList.LipDelegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.microphonesList.LipDelegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.microphonesList.LipDelegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.settingsList.LipDelegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.settingsList.LipDelegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.connectionsList.LipDelegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.onlineList.LipDelegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.onlineList.LipDelegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.friendsReqsList.LipDelegate.DefaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.friendsReqsList.LipDelegate.DefaultDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.apearenceList.LipDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
				m.apearenceList.LipDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

				m.microphonesList.LipList.SetDelegate(m.microphonesList.LipDelegate)
				m.settingsList.LipList.SetDelegate(m.settingsList.LipDelegate)
				m.connectionsList.LipList.SetDelegate(m.connectionsList.LipDelegate)
				m.onlineList.LipList.SetDelegate(m.onlineList.LipDelegate)
				m.friendsReqsList.LipList.SetDelegate(m.friendsReqsList.LipDelegate)
				m.apearenceList.LipList.SetDelegate(m.apearenceList.LipDelegate)

				m, cmd = m.syncTabState(), m.microphonesList.UpdateDevicesList(m.user, lists.MICROPHONE)
				return m, cmd
			case commands.BFTAG:
				cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections),
					m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)))
				return m, cmd
			case commands.B_TAG:
				cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)))
				return m, cmd
			}
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
				commands.NotifyCmd(nickname, "your new friend")
				return m, m.friendsReqsList.UpdateFriendsReqList(m.user)
			case sshclient.DENY_FRIEND:
				return m, m.friendsReqsList.UpdateFriendsReqList(m.user)
			case sshclient.DELETE_FRIEND:
				delete(m.online, nickname)
				cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)))
				return m, tea.Batch(cmds...)
			case sshclient.BLOCK_USER:
				delete(m.online, nickname)
				if isInConnections(m.connections, nickname) {
					cmds = append(cmds, commands.DisconnFromOne(m.user.Networking, nickname))
				}

				cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)), m.friendsReqsList.UpdateFriendsReqList(m.user), commands.NotifyCmd(nickname, "blocked"))
				return m, tea.Batch(cmds...)
			case sshclient.UNBLOCK_USER:
				cmds = append(cmds, m.onlineList.UpdateOnlineList(m.user, cloneMap(m.online)), commands.NotifyCmd(nickname, "unblocked"))
				return m, tea.Batch(cmds...)
			case sshclient.NEW_FRIEND_REQ:
				cmds = append(cmds, m.friendsReqsList.UpdateFriendsReqList(m.user),
					commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
					commands.NotifyCmd(nickname, "new friend request"))
				return m, tea.Batch(cmds...)
			}
		}

	case commands.SoloDisconn:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else if m.connected {
			t := time.Now().Format("15:04:05")

			m.messages = append(m.messages, commands.ChatMessage{Time: t, Nickname: "system", Text: styles.CErrStyle.Render(msg.Nickname) + styles.CErrStyle.Render(" banned!")})
			//delete(m.usersStates, msg.Nickname)
			cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections))
			if len(m.connections) == 0 {
				m.connected = false
				m.connections = []string{}
				m.messages = []commands.ChatMessage{}
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

		}

	case commands.OnOffDenoiceMsg, commands.OnOffFilterMsg, commands.OnOffAECMsg, commands.UsersVolumeMsg,
		commands.MuteUnmuteUserMsg, commands.StatiscticsMsg, commands.NotificationMessage, commands.ChangeDeviceMessage:
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
		case commands.NotificationMessage:
			err = m.Err
		case commands.ChangeDeviceMessage:
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
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
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
		}

	case commands.ConnectMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else if m.state == states.LOAD_STATE {
			m.state = m.prState
		}

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
				defer func() {
					if r := recover(); r != nil {
						m.log.Error("panic:", r, string(debug.Stack()))
					}
				}()
				if err := m.user.Engines.AudioEngine.SetDisconnected(); err != nil {
					m.err = err
					m.state = states.ERR_STATE
					return m, nil
				}
			}

			if m.prState == states.CONN_STATE {
				m.state = states.LOAD_STATE
			} else if m.state == states.LOAD_STATE {
				m.state = m.prState
			}

			m.connected = false
			select {
			case m.stopCountMinutesChan <- struct{}{}:
			default:
			}
			m.messages = []commands.ChatMessage{}
			m.connections = []string{}
			//clear(m.usersStates)
			m.user.Engines.AudioEngine.PlayNotification()
			//m.activeTab = 0
			// m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: "you disconnected!"})

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

			ls := lists.ListSetup{
				ThemeColor:       m.themeColor,
				SubColor:         m.subThemeColor,
				NormalDescColor:  styles.CGray,
				NormalTitleColor: styles.CText,
			}

			if m.user.Engines.AudioEngine != nil {
				m.microphonesList = lists.SetupDevicesList(m.user.Engines.AudioEngine, lists.MICROPHONE, ls)
				m.headphonesList = lists.SetupDevicesList(m.user.Engines.AudioEngine, lists.HEADPHONES, ls)
			}
			//friends := m.user.GetFriends()
			cmds = append(cmds,
				commands.WaitForChatMessageCmd(m.msgChan), commands.WaitForRawChatMessageCmd(m.rawMsgChan),
				commands.WaitForSSHEventMessageCmd(m.sshEventsChan),
				commands.WaitForNetworkEventMessageCmd(m.netwEventsChan),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
				commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan),
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

		userAudioN := m.user.GetAudioNotificationsState()
		userDesktopN := m.user.GetDesktopNotificationsState()
		if userAudioN {
			cmds = append(cmds, commands.PlayNotificationCmd(m.user.Engines.AudioEngine))
		}
		if userDesktopN {
			cmds = append(cmds, commands.NotifyCmd(msg.Nickname, textForDesktopNotification))
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

	// case commands.OnlineMsg:
	// 	if msg.Err != nil {
	// 		m.err = msg.Err
	// 		m.state = states.ERR_STATE
	// 	} else {
	// 		if msg.Online != nil && !isEqualOnline(msg.Online, m.online) {
	// 			if m.activeTab != 0 && isNewInOnline(msg.Online, m.online) {
	// 				m.tabsNotifications["friends"] = struct{}{}
	// 			}
	// 			m.online = msg.Online
	// 			cmd = m.onlineList.UpdateOnlineList(m.user, m.online)
	// 			return m, cmd
	// 		}
	// 	}

	case commands.UpdateDevicesMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else {
			switch msg.Typee {
			case lists.HEADPHONES:
				cmd = m.headphonesList.UpdateDevicesList(m.user, lists.HEADPHONES)
			case lists.MICROPHONE:
				cmd = m.microphonesList.UpdateDevicesList(m.user, lists.MICROPHONE)
			default:
				return m, nil
			}
			return m, cmd
		}

	case commands.PeerConnectedMsg:
		m.log.Info("new connect", msg.Nickname)
		nick := msg.Nickname
		if m.user.IsBlocked(nick) {
			m.log.Info("disconn from blocked", nick)
			cmds = append(cmds, commands.DisconnFromOne(m.user.Networking, nick),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan))
			return m, tea.Batch(cmds...)
		}
		hex := randomcolor.GetRandomColorInHex()
		color := lipgloss.Color(hex)
		nickname := lipgloss.NewStyle().Foreground(color).Render(nick)
		m.usersColors[msg.Nickname] = userColors{
			mainColor: color,
			subColor:  lipgloss.Color(utils.DarkenHex(hex, 0.7)),
		}
		m.connections = append(m.connections, nickname)
		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: nickname + " joined the chat!"})

		us := m.user.GetUsersSetup(msg.Nickname)
		if us == nil {
			if err := m.user.NewUserSetup(nickname); err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, nil
			}
		}

		//m.usersStates[nick] = &userState{}

		cmds = append(cmds, commands.IncreaseAmountOfConnectionsByUser(m.user, msg.Nickname), commands.SetupUserVolumeCmd(m.user, msg.Nickname),
			commands.SetupUserMuteCmd(m.user, msg.Nickname), m.connectionsList.UpdateConnectionsList(m.user, m.connections),
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
		// if m.activeTab == 0 {
		// 	m.activeTab = 1
		// 	m = m.syncTabState()
		// }

	case commands.PeerDisconnectedMsg:
		m.log.Info("peer disconnecting")
		nick := msg.Nickname
		cmds = append(cmds, commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan))
		defer func() {
			if r := recover(); r != nil {
				m.log.Error("panic:", r, string(debug.Stack()))
			}
		}()
		if !m.user.IsBlocked(nick) {
			var colored string
			m.connections = slices.DeleteFunc(m.connections, func(n string) bool {
				if ansi.Strip(n) == nick {
					colored = n
					return true
				}
				return false
			})
			m.log.Info("deketed from connections")
			if m.activeTab != 2 {
				m.tabsNotifications["voice"] = struct{}{}
			}
			//delete(m.usersStates, nick)
			m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time, Nickname: "system", Text: colored + " disconnected!"})
			cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections),
				commands.PlayNotificationCmd(m.user.Engines.AudioEngine))
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

				m.log.Info("peer disconnected you are solo")
			}
		}

	case spinner.TickMsg:
		if (m.isLoggedIn() && m.state == states.LOAD_STATE && (m.activeTab == 0 || m.activeTab == 4)) || (!m.isLoggedIn() && m.state == states.LOAD_STATE) {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			// friends := m.user.GetFriends()
			// if m.user.Networking != nil && m.user.Data.Personal.Nickname != "" && len(friends) > 0 {
			// 	cmds = append(cmds, commands.FetchOnlineFriendsCmd(m.user.Networking, friends))
			// }
			if m.user.Engines.AudioEngine != nil && m.activeTab == 5 && m.state == states.DEVICES_STATE {
				updatesDevices := tea.Sequence(commands.UpdateMicrophonesCmd(m.user), commands.UpdateHeadphonesCmd(m.user))
				cmds = append(cmds, updatesDevices)
			}

		}
		cmds = append(cmds, commands.TickCmd())

	case commands.AnimTickMsg:
		if m.curWindow == windows.START_WINDOW {
			m.animFrame++
			return m, commands.AnimTickCmd()
		}
	//}
	case commands.PulseTickMsg:
		if m.curWindow == windows.START_WINDOW {
			m.pulseFrame++
			return m, commands.PulseTickCmd()
		}

	case time.Time:
		timeIsOn := m.user.GetShowTimeState()
		if timeIsOn {
			m.curTime = msg
			return m, commands.TimeTickCmd()
		}

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
		case "alt+f", "alt+а", "alt+А", "alt+F":
			if m.isLoggedIn() {
				m.activeTab = 0
				m = m.syncTabState()
			}
		case "alt+g", "alt+G", "alt+п", "alt+П":
			if m.isLoggedIn() {
				m.activeTab = 2
				m = m.syncTabState()
			}

		case "alt+с", "alt+С", "alt+c", "alt+C":
			if m.isLoggedIn() {
				m.activeTab = 1
				m = m.syncTabState()
			}

		case "alt+d", "alt+D", "alt+в", "alt+В":
			if m.isLoggedIn() {
				m.activeTab = 3
				m = m.syncTabState()
			}

		case "alt+у", "alt+У", "alt+e", "alt+E":
			if m.isLoggedIn() {
				m.activeTab = 4
				m = m.syncTabState()
			}

		case "alt+v", "alt+М", "alt+V", "alt+м":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteMicCmd(m.user.Engines.AudioEngine, m.user.Networking))
			}
		case "alt+b", "alt+и", "alt+B", "alt+И":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteCmd(m.user.Engines.AudioEngine, m.user.Networking))
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
			if m.state == states.ERR_STATE {
				return m, nil
			}
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
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 2 && m.connected {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					cmds = append(cmds, commands.MuteUnmuteUserCmd(m.user, ansi.Strip(i.Nickname)),
						m.connectionsList.UpdateConnectionItemList(i.Nickname, i.VolumeCoefficient, !i.Muted))
				}
			}
		case "alt+x", "alt+X", "alt+ч", "alt+Ч":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 4 && m.friendsReqsList.LipList.Index() >= 0 {
				if i, ok := m.friendsReqsList.LipList.SelectedItem().(lists.FriendReqItem); ok {
					cmds = append(cmds, commands.DenyFriendRequestCmd(m.user, i.Nickname))
				}
			}

		case "alt+up":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					vc := i.VolumeCoefficient
					if vc >= MAX_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc+0.1)*10) / 10)
					if vc > MAX_VOLUME {
						vc = MAX_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.Nickname), vc),
						m.connectionsList.UpdateConnectionItemList(i.Nickname, i.VolumeCoefficient, !i.Muted))
				}
			}

		case "alt+down":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					vc := i.VolumeCoefficient
					if vc <= MIN_VOLUME {
						return m, nil
					}
					vc = float32(math.Round(float64(vc-0.1)*10) / 10)
					if vc < MIN_VOLUME {
						vc = MIN_VOLUME
					}
					cmds = append(cmds, commands.SetUserVolumeCmd(m.user, ansi.Strip(i.Nickname), vc),
						m.connectionsList.UpdateConnectionItemList(i.Nickname, i.VolumeCoefficient, !i.Muted))
				}
			}
		case "alt+left":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && (m.activeTab == 0 || m.activeTab == 4 || m.activeTab == 5) {

				if m.sideState == states.RIGHT_STATE {
					m.cursor = 0
					m.sideState = states.LEFT_STATE
					m = m.syncTabState()
				}

			}

		case "alt+right":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && (m.activeTab == 0 || m.activeTab == 4 || m.activeTab == 5) {
				if m.sideState == states.LEFT_STATE && m.state != states.SETTINGS_STATE {
					m.cursor = 0
					m.sideState = states.RIGHT_STATE
					m = m.syncTabState()
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
				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
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
				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
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
				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
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
			default:
				if m.activeTab != 5 {
					m.state = m.prState
				} else {
					m.state = states.SETTINGS_STATE
				}
				m = m.syncTabState()
			}
			return m, textinput.Blink

		case "up":
			if m.state == states.ERR_STATE {
				return m, nil
			}
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
						if m.apearenceList.LipList.Index() <= 0 {
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
			if m.state == states.ERR_STATE {
				return m, nil
			}
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
						// if m.cursor == len(m.profileInputs)-1 {
						// 	changed = true
						// 	break
						// }
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
			if m.state == states.ERR_STATE {
				return m, nil
			}
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

			// if m.state == states.ERR_STATE {
			// 	m.err = nil
			// 	m.curWindow = windows.DEF_WINDOW
			// 	if m.prState == states.LOAD_STATE {
			// 		m = m.syncTabState()
			// 	} else {
			// 		m.state = m.prState
			// 	}
			// }

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

					cmds = append(cmds, commands.RegisterCmd(m.user, m.sshEventsChan, m.log, []byte(password), []byte(repPassword)), m.spinner.Tick)
					for i := range m.regTextInputs {
						m.regTextInputs[i].Reset()
					}

				case 1:
					m.prState = m.state
					m.state = states.LOAD_STATE
					m.user.Data.Personal.Nickname = m.logingInput[0].Value()
					password := m.logingInput[1].Value()

					cmds = append(cmds, commands.LoginCmd(m.user, m.sshEventsChan, m.log, []byte(password)), m.spinner.Tick)
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
						if i, ok := m.onlineList.LipList.SelectedItem().(lists.OnlineItem); ok {
							nick := i.Name
							conns := i.Connections
							if !m.connected {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToAllUsersCmd(m.user, nick), m.spinner.Tick)
							} else if m.connected && !isInConnections(conns, m.user.Data.Personal.Nickname) {
								m.prState = m.state
								m.state = states.LOAD_STATE

								sequence := tea.Sequence(commands.LeaveCmd(m.user.Networking), commands.ConnectToAllUsersCmd(m.user, nick), m.spinner.Tick)
								cmds = append(cmds, sequence)
							}
						}

					case states.RIGHT_STATE:
						switch m.cursor {
						case 0:
							nick = strings.TrimSpace(m.friendsInputs[m.cursor].Value())
							if nick == "" {
								return m, nil
							}
							if !m.connected {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToAllUsersCmd(m.user, nick))
							} else if m.connected && !isInConnections(m.connections, nick) {
								m.prState = m.state
								m.state = states.LOAD_STATE

								sequence := tea.Sequence(commands.LeaveCmd(m.user.Networking), commands.ConnectToAllUsersCmd(m.user, nick))
								cmds = append(cmds, sequence)
							}
						case 1:
							nick = strings.TrimSpace(m.friendsInputs[m.cursor].Value())
							if nick == "" {
								return m, nil
							}

							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.SendFriendRequestCmd(m.user, nick), m.spinner.Tick)
						case 2:
							nick = strings.TrimSpace(m.friendsInputs[m.cursor].Value())
							if nick == "" {
								return m, nil
							}

							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, nick, true), m.spinner.Tick)
						case 3:
							nick = strings.TrimSpace(m.friendsInputs[m.cursor].Value())
							if nick == "" {
								return m, nil
							}

							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.BlockUserCmd(m.user, nick), m.spinner.Tick)
						case 4:
							nick = strings.TrimSpace(m.friendsInputs[m.cursor].Value())
							if nick == "" {
								return m, nil
							}

							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.UnblockUserCmd(m.user, nick), m.spinner.Tick)
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
					switch m.sideState {
					case states.LEFT_STATE:
						if i, ok := m.friendsReqsList.LipList.SelectedItem().(lists.FriendReqItem); ok {
							nick := i.Nickname
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.AcceptFriendRequestCmd(m.user, nick), m.spinner.Tick)
						}
					case states.RIGHT_STATE:
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
							newNotifyTag := strings.TrimSpace(m.profileInputs[2].Value())
							if newNotifyTag == "d" {
								newNotifyTag = m.defaultNotificationTag
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeNotificationTagCmd(m.user, newNotifyTag))
							m.profileInputs[2].Reset()
						case 3:
							newBanTag := strings.TrimSpace(m.profileInputs[3].Value())
							if newBanTag == "d" {
								newBanTag = m.defaultBanTag
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeBanTagCmd(m.user, newBanTag))
							m.profileInputs[3].Reset()
						case 4:
							if i, ok := m.apearenceList.LipList.SelectedItem().(lists.SwitcherItem); ok {
								m.prState = m.state
								m.state = states.LOAD_STATE
								switch i.Id {
								case lists.TIME:
									timeSeq := tea.Sequence(commands.OnOffShowTime(m.user), commands.TimeTickCmd())
									cmds = append(cmds, timeSeq)
								case lists.DATE:
									cmds = append(cmds, commands.OnOffShowDate(m.user))
								case lists.ZONE:
									cmds = append(cmds, commands.OnOffShowZone(m.user))
								default:
									return m, nil
								}

								cmds = append(cmds, m.apearenceList.UpdateSwitcherItemList(i.Id))
							}
						}
					}

				case 5:
					switch m.sideState {
					case states.RIGHT_STATE:
						switch m.state {
						case states.DEVICES_STATE:
							m.prState = m.state
							m.state = states.LOAD_STATE
							if i, ok := m.microphonesList.LipList.SelectedItem().(lists.DeviceItem); ok {
								cmds = append(cmds, commands.ChangeMicrophoneCmd(m.user, i.Name),
									m.microphonesList.UpdateDevicesItemList(i.Name))
							}
						case states.AUDIO_STATE:
							if i, ok := m.audioList.LipList.SelectedItem().(lists.SwitcherItem); ok {
								m.prState = m.state
								m.state = states.LOAD_STATE
								switch i.Id {
								case lists.AEC:
									cmds = append(cmds, commands.OnOffAECCmd(m.user))
								case lists.HARD_DENOISE:
									cmds = append(cmds, commands.OnOffHardDenoiceCmd(m.user))
								case lists.SOFT_DENOISE:
									cmds = append(cmds, commands.OnOffSoftDenoiceCmd(m.user))
								case lists.EQUALIZER:
									cmds = append(cmds, commands.OnOffFilterCmd(m.user))
								}
								cmds = append(cmds, m.audioList.UpdateSwitcherItemList(i.Id))
							}
						case states.NOTIFICATIONS_STATE:
							if i, ok := m.notificationsList.LipList.SelectedItem().(lists.SwitcherItem); ok {
								m.prState = m.state
								m.state = states.LOAD_STATE
								switch i.Id {
								case lists.APP_N:
									cmds = append(cmds, commands.OnOffAppNotifications(m.user))
								case lists.AUDIO_N:
									cmds = append(cmds, commands.OnOffAudioNotifications(m.user))
								case lists.DESKTOP_N:
									cmds = append(cmds, commands.OnOffDesktopNotifications(m.user))
								}
								cmds = append(cmds, m.notificationsList.UpdateSwitcherItemList(i.Id))
							}

						}
					case states.LEFT_STATE:
						switch m.state {
						case states.SETTINGS_STATE:
							return m.selectSetting()
						case states.AUDIO_STATE:
							return m.selectSetting()
						case states.NOTIFICATIONS_STATE:
							return m.selectSetting()
						case states.DEVICES_STATE:
							if i, ok := m.headphonesList.LipList.SelectedItem().(lists.DeviceItem); ok {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ChangeHeadphonesCmd(m.user, i.Name),
									m.headphonesList.UpdateDevicesItemList(i.Name))
							}
						}
					}

				}
			}
		}
	}

	if m.curWindow == windows.DEF_WINDOW && m.sideState != states.ZERO_STATE {
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
				switch m.sideState {
				case states.RIGHT_STATE:
					for i := range m.friendsInputs {
						m.friendsInputs[i], cmd = m.friendsInputs[i].Update(msg)
						cmds = append(cmds, cmd)
					}
				case states.LEFT_STATE:
					m.onlineList.LipList, cmd = m.onlineList.LipList.Update(msg)
					cmds = append(cmds, cmd)
				}

			case 1:
				if m.connected {
					m.chatTextInput, cmd = m.chatTextInput.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 2:
				if m.sideState == states.LEFT_STATE {
					m.connectionsList.LipList, cmd = m.connectionsList.LipList.Update(msg)
					cmds = append(cmds, cmd)
				}
			case 4:
				switch m.sideState {
				case states.RIGHT_STATE:
					if m.cursor < len(m.profileInputs) {
						for i := range m.profileInputs {
							m.profileInputs[i], cmd = m.profileInputs[i].Update(msg)
							cmds = append(cmds, cmd)
						}
					} else if m.cursor == len(m.profileInputs) {
						m.apearenceList.LipList, cmd = m.apearenceList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					}

				case states.LEFT_STATE:
					m.friendsReqsList.LipList, cmd = m.friendsReqsList.LipList.Update(msg)
					cmds = append(cmds, cmd)
				}

			case 5:
				switch m.sideState {
				case states.RIGHT_STATE:
					switch m.state {
					case states.AUDIO_STATE:
						m.audioList.LipList, cmd = m.audioList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.NOTIFICATIONS_STATE:
						m.notificationsList.LipList, cmd = m.notificationsList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.DEVICES_STATE:
						m.microphonesList.LipList, cmd = m.microphonesList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					}
				case states.LEFT_STATE:
					switch m.state {
					case states.SETTINGS_STATE:
						m.settingsList.LipList, cmd = m.settingsList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.AUDIO_STATE:
						m.settingsList.LipList, cmd = m.settingsList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.NOTIFICATIONS_STATE:
						m.settingsList.LipList, cmd = m.settingsList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.DEVICES_STATE:
						m.headphonesList.LipList, cmd = m.headphonesList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					}
				}

			}
		}
	}

	return m, tea.Batch(cmds...)
}
