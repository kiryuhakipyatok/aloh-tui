package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/errs"
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

	"github.com/blacktop/go-termimg"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
)

const (
	MAX_VOLUME = 4.0
	MIN_VOLUME = 0.0
)

func (m Model) syncTabState() (Model, tea.Cmd) {
	cmds := []tea.Cmd{textinput.Blink}
	m.curWindow = windows.DEF_WINDOW
	m.prState = m.state
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
				lists.SetListVisible(&m.friendsList.DefList, true)
				m.unfocusInputs()
			case states.RIGHT_STATE:
				switch m.cursor {
				case 0:
					m.state = states.CONN_STATE
				default:
					m.state = states.FRIEND_STATE
				}
				m.focusInputs()
				lists.SetListVisible(&m.friendsList.DefList, false)
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
		case 3:
			m.state = states.WEBCAM_STATE
			delete(m.tabsNotifications, "webcam")
		case 4:
			m.state = states.PROFILE_STATE
			delete(m.tabsNotifications, "profile")
			switch m.sideState {
			case states.LEFT_STATE:
				if len(m.friendsReqsList.LipList.Items()) > 0 {
					lists.SetListVisible(&m.friendsReqsList.DefList, true)
					lists.SetListVisible(&m.apearenceList.DefList, false)
					m.unfocusInputs()
				} else {
					m.sideState = states.RIGHT_STATE
					return m.syncTabState()
				}

			case states.RIGHT_STATE:
				if m.cursor < len(m.appereanceInputs) {
					//m.apearenceList.Select(-1)
					lists.SetListVisible(&m.friendsReqsList.DefList, false)
					lists.SetListVisible(&m.apearenceList.DefList, false)
					m.focusInputs()
				} else if m.cursor == len(m.appereanceInputs) {
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
				cmds = append(cmds, commands.SecTickCmd())
			case states.ACCOUNT_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					lists.SetListVisible(&m.accountList.DefList, true)
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					lists.SetListVisible(&m.accountList.DefList, false)
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.NICKNAME_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					m.focusInputs()
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					m.unfocusInputs()
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.PASSWORD_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					m.focusInputs()
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					m.unfocusInputs()
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.TAGLINE_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					m.focusInputs()
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					m.unfocusInputs()
					lists.SetListVisible(&m.settingsList.DefList, true)
				}
			case states.COLOR_STATE:
				switch m.sideState {
				case states.RIGHT_STATE:
					m.focusInputs()
					lists.SetListVisible(&m.settingsList.DefList, false)
				case states.LEFT_STATE:
					m.unfocusInputs()
					lists.SetListVisible(&m.settingsList.DefList, true)
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
	return m, tea.Batch(cmds...)
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
					if m.user.Data.Identity.ID != uuid.Nil && m.user.Networking == nil {
						cmds = append(cmds, commands.AuthCmd(m.user, m.sshEventsChan, m.log))
					} else {
						m, cmd = m.syncTabState()
					}
					return m, cmd
				} else if m.curWindow == windows.DEF_WINDOW && m.state != states.LOAD_STATE {
					if !m.inCurrentWindow(msg) {
						m.cursor = 0
					}
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
					} else if m.zone.Get("webcamT").InBounds(msg) || m.zone.Get("webcamW").InBounds(msg) {
						m.activeTab = 3
						iden, ok := m.onUsersWebcam(msg)
						if ok {
							m.prState = m.state
							m.state = states.LOAD_STATE
							return m, commands.OnOffWindowWebcamCmd(m.user.Engines.VideoEngine, iden)
						}
					} else if m.zone.Get("profileT").InBounds(msg) || m.zone.Get("profileW").InBounds(msg) {
						m.activeTab = 4
					} else if m.zone.Get("settingsT").InBounds(msg) || m.zone.Get("settingsW").InBounds(msg) {
						m.activeTab = 5
					} else if m.activeTab == 3 {
						m.log.Info("click on webcam tab")
						iden, ok := m.onUsersWebcam(msg)
						if ok {
							m.prState = m.state
							m.state = states.LOAD_STATE
							return m, commands.OnOffWindowWebcamCmd(m.user.Engines.VideoEngine, iden)
						}
					} else {
						m.sideState = states.ZERO_STATE
						m, cmd = m.syncTabState()
						m.unfocusInputs()
						m.unfocusLists()
						return m, cmd
					}

					if m.zone.Get("leftSide").InBounds(msg) && m.sideState != states.LEFT_STATE {
						m.sideState = states.LEFT_STATE
					} else if m.zone.Get("rightSide").InBounds(msg) && m.sideState != states.RIGHT_STATE {
						m.sideState = states.RIGHT_STATE
					}

					m, cmd = m.syncTabState()
					return m, cmd
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

	case commands.NetworkEventMsg:
		id := msg.Id

		iden, err := getIdentityInConnections(m.connections, id)
		if err != nil {
			return m.Err(err)
		}
		nick := iden.Nickname

		us, ok := m.usersStates[id]
		if ok {
			switch msg.Event.Typee {
			case networking.MIC_MUTE:
				us.fullMute = false
				state, err := networking.DataToState(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.micMute = state
			case networking.FULL_MUTE:
				state, err := networking.DataToState(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.fullMute = state
				us.micMute = false
			case networking.HARD_DENOISE:
				state, err := networking.DataToState(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.hardDenoised = state
				iden, err := m.user.GetFriendIdentityById(id)
				if err != nil {
					return m.Err(err)
				}
				nick := iden.Nickname
				usE, err := m.user.GetUsersSetup(nick)
				if err != nil {
					return m.Err(err)
				}
				if us.hardDenoised && usE.HardDenoise {
					i, err := m.connectionsList.GetConnectionItem(iden)
					if err != nil {
						return m.Err(err)
					}
					seq := tea.Sequence(commands.OnOffUsersHardDenoise(m.user, iden, false),
						m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
							false, i.PersonalSoftDenoise))
					cmds = append(cmds, seq)
				}
			case networking.SOFT_DENOISE:
				state, err := networking.DataToState(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.softDenoised = state
				usE, err := m.user.GetUsersSetup(nick)
				if err != nil {
					return m.Err(err)
				}
				if us.softDenoised && usE.SoftDenoise {

					i, err := m.connectionsList.GetConnectionItem(iden)
					if err != nil {
						return m.Err(err)
					}
					seq := tea.Sequence(commands.OnOffUsersSoftDenoise(m.user, iden, false),
						m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
							i.PersonalHardDenoise, false))
					cmds = append(cmds, seq)
				}
			case networking.WEBCAM:
				state, err := networking.DataToState(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.webcam = state
				if !state {
					m.user.Engines.VideoEngine.RemoveUserFromUsersVideo(id)
				}
				m.log.Info("new webcam event", state, msg.Id)
				if m.activeTab != 3 {
					m.tabsNotifications["webcam"] = struct{}{}
				}

				// if state {
				// 	m.webcamUsersFrames[id] = userVideoFrame{
				// 		nickname: nick,
				// 	}
				// } else {
				// 	delete(m.webcamUsersFrames, id)
				// }

			case networking.GENERAL:
				generalData, err := networking.DataToGeneral(msg.Event.Data)
				if err != nil {
					return m.Err(err)
				}
				us.fullMute = generalData.FullMute
				us.micMute = generalData.MicMute
				us.softDenoised = generalData.SoftDenoise
				us.hardDenoised = generalData.HardDenoise
				usE, err := m.user.GetUsersSetup(nick)
				if err != nil {
					return m.Err(err)
				}

				if us.softDenoised && usE.SoftDenoise {
					i, err := m.connectionsList.GetConnectionItem(iden)
					if err != nil {
						return m.Err(err)
					}
					seq := tea.Sequence(commands.OnOffUsersSoftDenoise(m.user, iden, false),
						m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
							i.PersonalHardDenoise, false))
					cmds = append(cmds, seq)
				}
				if us.hardDenoised && usE.HardDenoise {
					i, err := m.connectionsList.GetConnectionItem(iden)
					if err != nil {
						return m.Err(err)
					}
					seq := tea.Sequence(commands.OnOffUsersHardDenoise(m.user, iden, false),
						m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
							false, i.PersonalSoftDenoise))
					cmds = append(cmds, seq)
				}
			}
		}
		return m, commands.WaitForNetworkEventMessageCmd(m.netwEventsChan)

	case sshclient.Event:
		var nickname string
		switch msg.Type {
		case sshclient.NEW_FRIEND_REQ:
			iden, err := sshclient.CastToIdentityData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			cmds = append(cmds, commands.NewFriendReqCmd(m.user, iden))
			//commands.NotifyCmd(nickname, "new friend request"))
		case sshclient.ACCEPT_FRIEND:
			iden, err := sshclient.CastToIdentityData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}

			cmds = append(cmds, commands.NewFriendCmd(m.user, iden))
			//commands.NotifyCmd(nickname, "your new friend"))
		case sshclient.DELETE_FRIEND:
			iden, err := sshclient.CastToIdentityData(msg.Data)
			if err != nil {
				m.log.Error("error delete friend when cast", err)
				return m.Err(err)
			}

			if m.activeTab != 4 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			delete(m.online, iden.ID)
			cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, iden, false))
			//commands.NotifyCmd(nickname, "no longer your friend"))
		case sshclient.BLOCK_USER:
			iden, err := sshclient.CastToIdentityData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			id := iden.ID
			if m.user.IsFriend(iden) {
				if m.activeTab != 4 {
					m.tabsNotifications["profile"] = struct{}{}
				}
				delete(m.online, id)
				cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, iden, false),
					commands.NotifyCmd(nickname, "blocked you"))
			}

		case sshclient.FRIEND_ONLINE:
			//if m.user.IsFriend(nickname) {
			fcd, err := sshclient.CastToFriendConnsData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			if m.activeTab != 0 {
				m.tabsNotifications["friends"] = struct{}{}
			}

			m.online[fcd.Identity.ID] = fcd.Connects
			cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)))

		case sshclient.FRIEND_OFFLINE:
			id, err := sshclient.CastToIdData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			delete(m.online, id)
			cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)))

		case sshclient.UPDATE_TAGLINE:
			td, err := sshclient.CastToTaglineData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			m.log.Info("event update tagline", td)
			cmds = append(cmds, commands.UpdateFriendsTaglineCmd(m.user, td.Identity, td.Tagline))
		case sshclient.UPDATE_COLOR:
			cd, err := sshclient.CastToColorData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			m.usersColors[cd.Identity.ID] = newUC(cd.Color)
			cmds = append(cmds, commands.UpdateFriendsColorCmd(m.user, cd.Identity, cd.Color))
		case sshclient.UPDATE_NICKNAME:
			nd, err := sshclient.CastToNickanameData(msg.Data)
			if err != nil {
				return m.Err(err)
			}
			m.log.Info("event update nickname", nd)
			cmds = append(cmds, commands.UpdateForeignNicknameCmd(m.user, nd.Identity, nd.Nickname))
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
			case commands.THEME_COLOR:
				m.setupColors()

				mics := m.microphonesList.LipList.Items()
				heads := m.headphonesList.LipList.Items()
				settings := m.settingsList.LipList.Items()
				conns := m.connectionsList.LipList.Items()
				friends := m.friendsList.LipList.Items()
				reqs := m.friendsReqsList.LipList.Items()
				apps := m.apearenceList.LipList.Items()
				notifs := m.notificationsList.LipList.Items()
				audio := m.audioList.LipList.Items()

				m.setupModelsLists()
				m.setupUsersLists()

				m.microphonesList.LipList.SetItems(mics)
				m.headphonesList.LipList.SetItems(heads)
				m.settingsList.LipList.SetItems(settings)
				m.connectionsList.LipList.SetItems(conns)
				m.friendsList.LipList.SetItems(friends)
				m.friendsReqsList.LipList.SetItems(reqs)
				m.apearenceList.LipList.SetItems(apps)
				m.notificationsList.LipList.SetItems(notifs)
				m.audioList.LipList.SetItems(audio)

				m, cmd = m.syncTabState()
				return m, cmd
			case commands.BFTAG:
				cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors),
					m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)))
				return m, cmd
			case commands.B_TAG:
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)))
				return m, cmd
			case commands.S_TIME:
				switch msg.Res {
				case true:
					m.rightHeaderData[1] = time.Now().Format("15:04:05")
					cmds = append(cmds, commands.TimeTickCmd())
				case false:
					m.rightHeaderData[1] = ""
				}
			case commands.S_DATE:
				switch msg.Res {
				case true:
					m.rightHeaderData[0] = m.curTime.Format("2006-01-02")
				case false:
					m.rightHeaderData[0] = ""
				}
			case commands.S_ZONE:
				switch msg.Res {
				case true:
					m.rightHeaderData[2] = m.curTime.Format("-07:00")
				case false:
					m.rightHeaderData[2] = ""
				}

			}
		}

	case commands.FriendsMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			switch msg.Typee {
			case sshclient.ACCEPT_FRIEND, sshclient.DENY_FRIEND, sshclient.DELETE_FRIEND,
				sshclient.BLOCK_USER, sshclient.UNBLOCK_USER, sshclient.NEW_FRIEND_REQ,
				sshclient.UPDATE_TAGLINE, sshclient.UPDATE_COLOR, sshclient.UPDATE_NICKNAME:
				if m.state == states.LOAD_STATE {
					m.state = m.prState
				}
			}
			iden := msg.Identity
			nickname := iden.Nickname
			id := iden.ID
			switch msg.Typee {
			case sshclient.ACCEPT_FRIEND:
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)),
					m.friendsReqsList.UpdateFriendsReqList(m.user),
					commands.NotifyCmd(nickname, "your new  friend"))

			case sshclient.DENY_FRIEND:
				cmds = append(cmds, m.friendsReqsList.UpdateFriendsReqList(m.user),
					commands.NotifyCmd(nickname, "friend request denied"))

			case sshclient.DELETE_FRIEND:
				delete(m.online, id)
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)),
					commands.NotifyCmd(nickname, "no longer your friend"))

			case sshclient.BLOCK_USER:
				m.log.Info("blocked", m.user.Data.Personal.BlockedUsers)
				delete(m.online, id)
				if isInConnections(m.connections, iden.ID) {
					cmds = append(cmds, commands.DisconnFromOne(m.user.Networking, iden))
				}
				// if m.user.IsFriend(iden){
				// 	cmds = append(cmds, m.c)
				// }
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)),
					m.friendsReqsList.UpdateFriendsReqList(m.user), commands.NotifyCmd(nickname, "blocked"))

			case sshclient.UNBLOCK_USER:
				m.log.Info("blocked", m.user.Data.Personal.BlockedUsers)
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)),
					commands.NotifyCmd(nickname, "unblocked"))

			case sshclient.NEW_FRIEND_REQ:
				cmds = append(cmds, m.friendsReqsList.UpdateFriendsReqList(m.user),
					commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
					commands.NotifyCmd(nickname, "new friend request"))

			case sshclient.UPDATE_TAGLINE:
				m.log.Info("friend msg update tagline", iden)
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)))
			case sshclient.UPDATE_COLOR:
				m.log.Info("friend msg update color", iden)
				if isInConnections(m.connections, id) {
					cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors))
				}
			case sshclient.UPDATE_NICKNAME:
				m.log.Info("friend msg update nickname", iden)
				cmds = append(cmds, m.friendsList.UpdateFriendsList(m.user, cloneMap(m.online)),
					m.friendsReqsList.UpdateFriendsReqList(m.user))
				if isInConnections(m.connections, id) {
					updateNicknameInConn(m.connections, id, iden.Nickname)
					cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors))
				}
			default:
				return m, nil
			}
			return m, tea.Batch(cmds...)
		}

	case commands.SoloDisconn:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else if m.connected {
			iden := msg.Identity
			nickname := iden.Nickname
			id := iden.ID
			m.connections = slices.DeleteFunc(m.connections, func(cIden users.Identity) bool {
				return iden == cIden
			})
			m.connecctionsNicks = slices.DeleteFunc(m.connecctionsNicks, func(n string) bool {
				return ansi.Strip(n) == nickname
			})

			if len(m.connections) == 0 && m.connected {
				m.connected = false
				m.connections = []users.Identity{}
				m.connecctionsNicks = []string{}
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

			if m.activeTab != 2 {
				m.tabsNotifications["voice"] = struct{}{}
			}
			delete(m.usersStates, id)

			cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors),
				commands.UpdateCurrentConnectsCmd(m.user, m.connections))
			t := time.Now().Format("15:04:05")

			m.messages = append(m.messages, commands.ChatMessage{Time: t,
				Identity: users.Identity{
					Nickname: "system",
				}, Text: styles.CErrStyle.Render(msg.Identity.Nickname) + styles.CErrStyle.Render(" banned!")})
			//delete(m.usersStates, msg.Nickname)
			cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors),
				commands.UpdateCurrentConnectsCmd(m.user, m.connections))

		}

	case commands.AccountMsg:
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
				account := m.user.GetAccount()
				m.userColor = setupUserColor(account.Color)
			}
		}

	case commands.StatiscticsMsg:
		err := msg.Err
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else if msg.Res {

			switch msg.Typee {
			case commands.BF:
				cmds = append(cmds, m.friendsList.UpdateBestFriend(m.user))
			default:

			}

		}

	case commands.OnOffDenoiceMsg, commands.OnOffFilterMsg, commands.OnOffAECMsg, commands.UsersVolumeMsg,
		commands.MuteUnmuteUserMsg, commands.NotificationMessage, commands.UserWebcam,
		commands.ChangeDeviceMessage, commands.UserInfoMsg, commands.UsersDenoiseMsg, commands.OnOffWindowWebcamMsg:
		var err error
		switch mes := msg.(type) {
		case commands.OnOffDenoiceMsg:
			err = mes.Err
		case commands.OnOffFilterMsg:
			err = mes.Err
		case commands.OnOffAECMsg:
			err = mes.Err
		case commands.UsersVolumeMsg:
			err = mes.Err
		case commands.MuteUnmuteUserMsg:
			err = mes.Err
		case commands.NotificationMessage:
			err = mes.Err
		case commands.ChangeDeviceMessage:
			err = mes.Err
		case commands.UserInfoMsg:
			err = mes.Err
		case commands.UsersDenoiseMsg:
			err = mes.Err
		case commands.UserWebcam:
			err = mes.Err
		case commands.OnOffWindowWebcamMsg:
			err = mes.Err
		}

		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
		} else {
			if m.state == states.LOAD_STATE {
				m.state = m.prState
			}
			m.focusInputs()
			return m, textinput.Blink
		}

	case commands.MuteMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
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
			m, cmd = m.syncTabState()
			cmds = append(cmds, cmd, commands.IncreaseAmountOfMessagesCmd(m.user))
			return m, tea.Batch(cmds...)
		}

	case commands.LeaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.state = states.ERR_STATE
		} else if m.connected {
			if m.user.Engines.AudioEngine != nil {
				if err := m.user.Engines.AudioEngine.SetDisconnected(); err != nil {
					return m.Err(err)
				}
			}

			if m.user.Engines.VideoEngine != nil {
				m.user.Engines.VideoEngine.SetDisconnected()
			}

			if m.activeTab == 2 && m.state == states.LOAD_STATE {
				m.state = m.prState
			}

			m.connected = false
			select {
			case m.stopCountMinutesChan <- struct{}{}:
			default:
			}
			m.messages = []commands.ChatMessage{}
			m.connections = []users.Identity{}
			m.connecctionsNicks = []string{}
			clear(m.usersStates)
			//clear(m.webcamUsersFrames)
			m.user.Engines.AudioEngine.PlayNotification()
			cmds = append(cmds, commands.UpdateCurrentConnectsCmd(m.user, m.connections))
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
			// if msg.Typee != sshclient.DEFAULT {
			// 	m.user.Data.Personal = users.Personal{}
			// }
		} else {
			switch msg.Typee {
			case sshclient.REGISTER:
				m.state = states.DEF_STATE
				m.activeTab = 0
			case sshclient.LOGIN:
				m.state = states.DEF_STATE
				m.activeTab = 0
			default:
				m.state = states.START_STATE
			}
			//m.activeTab = 0
			//m, cmd = m.syncTabState()

			m.setupUsersModel()
			m.setupUsersLists()
			m.setupFriendsColors()

			//if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
			cs := CallbacksSetup{
				User:            m.user,
				RawMsgChan:      m.rawMsgChan,
				PeerConnChan:    m.peerConnectionsChan,
				PeerDisconnChan: m.peerDisconnectionsChan,
				NetwEventChan:   m.netwEventsChan,
			}

			SetupCallbacks(cs)

			frReqs := m.user.GetFriendsReqs()
			if len(frReqs) > 0 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			//	}

			// ls := lists.ListSetup{
			// 	ThemeColor:       m.themeColor,
			// 	SubColor:         m.subThemeColor,
			// 	NormalDescColor:  styles.CGray,
			// 	NormalTitleColor: styles.CText,
			// }

			// if m.user.Engines.AudioEngine != nil {
			// 	m.microphonesList = lists.SetupDevicesList(m.user.Engines.AudioEngine, lists.MICROPHONE, ls)
			// 	m.headphonesList = lists.SetupDevicesList(m.user.Engines.AudioEngine, lists.HEADPHONES, ls)
			// }
			//friends := m.user.GetFriends()

			cmds = append(cmds,
				commands.WaitForChatMessageCmd(m.msgChan),
				commands.WaitForSSHEventMessageCmd(m.sshEventsChan),
				commands.WaitForNetworkEventMessageCmd(m.netwEventsChan),
				commands.WaitForRawChatMessageCmd(m.rawMsgChan),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
				commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan), textinput.Blink)
		}

	// case commands.RawWebcamMsg:
	// 	data := msg.Data
	// 	id := msg.Id
	// 	//webcamFrame := m.user.Engines.VideoEngine.RenderUsersVideoTerminal(id, data)

	// 	// select {
	// 	// case m.webcamMsgChan <- commands.WebcamMsg{Id: id, Frame: webcamFrame}:
	// 	// default:
	// 	// }
	// 	cmds = append(cmds, commands.RenderUsersFrameCmd(m.user.Engines.VideoEngine, id, data),
	// 		commands.WaitForRawWebcamMsgCmd(m.rawWebcamMsgChan))
	// 	return m, tea.Batch(cmds...)

	// case commands.WebcamMsg:
	// 	wuf, ok := m.webcamUsersFrames[msg.Id]
	// 	if ok {
	// 		wuf.frame = msg.Frame
	// 		m.webcamUsersFrames[msg.Id] = wuf
	// 	}

	// 	return m, commands.WaitForWebcamMsgCmd(m.webcamMsgChan)

	case commands.RawChatMessage:
		data := msg.Data
		var (
			textMsg                    string
			textForDesktopNotification string
		)
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
		} else {
			textMsg = string(msg.Data)
			textForDesktopNotification = textMsg
		}

		iden, err := getIdentityInConnections(m.connections, msg.Id)
		if err != nil {
			m.err = err
			m.state = states.ERR_STATE
			return m, commands.WaitForRawChatMessageCmd(m.rawMsgChan)
		}
		nickname := iden.Nickname

		m.msgChan <- commands.ChatMessage{Identity: iden, Time: msg.Time, Text: textMsg}

		userAudioN := m.user.GetAudioNotificationsState()
		userDesktopN := m.user.GetDesktopNotificationsState()
		if userAudioN {
			cmds = append(cmds, commands.PlayNotificationCmd(m.user.Engines.AudioEngine))
		}
		if userDesktopN {
			cmds = append(cmds, commands.NotifyCmd(nickname, textForDesktopNotification))
		}
		cmds = append(cmds, commands.WaitForRawChatMessageCmd(m.rawMsgChan))
		return m, tea.Batch(cmds...)

	case commands.ChatMessage:
		//msg.Nickname = m.coloredNickname(msg.Nickname)
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
	// 			cmd = m.friendsList.UpdateFriendsList(m.user, m.online)
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
		id := msg.Id
		var (
			iden users.Identity
			err  error
		)

		iden, err = m.user.GetFriendIdentityById(id)
		if err != nil {

			iden, err = m.getOnlineIdentity(cloneMap(m.online), id)
			if err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan)
			}
		}
		if m.user.IsBlocked(iden) {
			cmds = append(cmds, commands.DisconnFromOne(m.user.Networking, iden),
				commands.WaitForPeerConnectionCmd(m.peerConnectionsChan))
			return m, tea.Batch(cmds...)
		}

		nickname := iden.Nickname
		uc, ok := m.usersColors[id]
		if !ok {
			uc = newUC("")
			m.usersColors[id] = uc
		}

		coloredNickname := lipgloss.NewStyle().Foreground(uc.MainColor).Render(nickname)

		m.connections = append(m.connections, iden)
		m.connecctionsNicks = append(m.connecctionsNicks, coloredNickname)
		m.messages = append(m.messages, commands.ChatMessage{Time: msg.Time,
			Identity: users.Identity{
				Nickname: "system",
			}, Text: coloredNickname + " joined the chat!"})

		us, err := m.user.GetUsersSetup(nickname)
		if err != nil {
			if err := m.user.NewUserSetup(nickname); err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForPeerConnectionCmd(m.peerConnectionsChan)
			}
		}

		m.usersStates[id] = &userState{}
		if !m.connected {
			m.connected = true
			cmds = append(cmds, commands.CountMaxTimeInConnectionCmd(m.user, m.stopCountMinutesChan),
				commands.IncreaseAmountOfConnectionsCmd(m.user), commands.UpdateTickCmd())
			if m.user.Engines.AudioEngine != nil {
				m.user.Engines.AudioEngine.SetConnected()
			}
			if m.user.Engines.VideoEngine != nil {
				m.user.Engines.VideoEngine.SetConnected()
			}
		}
		if m.activeTab != 2 {
			m.tabsNotifications["voice"] = struct{}{}
		}

		seq := tea.Sequence(commands.IncreaseAmountOfConnectionsByUser(m.user, iden),
			m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors))

		cmds = append(cmds,
			commands.SetupUserVolumeCmd(m.user, iden),
			commands.OnOffUsersHardDenoise(m.user, iden, us.HardDenoise),
			commands.OnOffUsersSoftDenoise(m.user, iden, us.SoftDenoise),
			commands.SetupUserMuteCmd(m.user, iden),
			seq,
			commands.PlayNotificationCmd(m.user.Engines.AudioEngine),
			commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
			commands.UpdateCurrentConnectsCmd(m.user, m.connections), commands.SendUserInfo(m.user, id))

		// if m.activeTab == 0 {
		// 	m.activeTab = 1
		// 	m = m.syncTabState()
		// }

	case commands.PeerDisconnectedMsg:
		id := msg.Id
		var (
			iden users.Identity
			err  error
		)

		iden, err = m.user.GetFriendIdentityById(id)
		if err != nil {
			iden, err = m.getOnlineIdentity(cloneMap(m.online), id)
			if err != nil {
				m.err = err
				m.state = states.ERR_STATE
				return m, commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan)
			}
		}

		nickname := iden.Nickname
		m.connections = slices.DeleteFunc(m.connections, func(cIden users.Identity) bool {
			return iden == cIden
		})
		m.connecctionsNicks = slices.DeleteFunc(m.connecctionsNicks, func(n string) bool {
			return ansi.Strip(n) == nickname
		})

		if len(m.connections) == 0 && m.connected {

			m.connected = false
			m.connections = []users.Identity{}
			m.connecctionsNicks = []string{}
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
			if m.user.Engines.VideoEngine != nil {
				m.user.Engines.VideoEngine.SetDisconnected()
			}

		}

		cmds = append(cmds, commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan))

		//if !m.user.IsBlocked(iden) {

		if m.activeTab != 2 {
			m.tabsNotifications["voice"] = struct{}{}
		}
		delete(m.usersStates, id)
		//delete(m.webcamUsersFrames, id)
		//	id := iden.ID
		coloredNickname := lipgloss.NewStyle().Foreground(m.usersColors[id].MainColor).Render(nickname)

		if err := m.user.Engines.AudioEngine.RemoveFromUsersAudio(id); err != nil {
			m.err = err
			m.state = states.ERR_STATE
		}
		m.user.Engines.VideoEngine.RemoveUserFromUsersVideo(id)
		m.messages = append(m.messages, commands.ChatMessage{
			Identity: users.Identity{
				Nickname: "system",
			}, Time: msg.Time, Text: coloredNickname + " disconnected!"})
		cmds = append(cmds, m.connectionsList.UpdateConnectionsList(m.user, m.connections, m.usersColors),
			commands.PlayNotificationCmd(m.user.Engines.AudioEngine), commands.UpdateCurrentConnectsCmd(m.user, m.connections))
		//}

	case spinner.TickMsg:
		if (m.isLoggedIn() && m.state == states.LOAD_STATE &&
			(m.activeTab == 0 || m.activeTab == 3 || m.activeTab == 4 || m.activeTab == 5)) ||
			(!m.isLoggedIn() && m.state == states.LOAD_STATE) {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case commands.TickMsg:
		if m.state != states.LOAD_STATE {
			if m.user.Engines.AudioEngine != nil && m.activeTab == 5 && m.state == states.DEVICES_STATE {
				updatesDevices := tea.Sequence(commands.UpdateMicrophonesCmd(m.user), commands.UpdateHeadphonesCmd(m.user))
				cmds = append(cmds, updatesDevices, commands.SecTickCmd())
			}

		}

	case commands.UpdateTickMsg:
		if m.connected {
			return m, commands.UpdateTickCmd()
		}

	case commands.AnimTickMsg:
		if m.curWindow == windows.START_WINDOW {
			m.animFrame++
			return m, commands.AnimTickCmd()
		}

	case commands.PulseTickMsg:
		if m.curWindow == windows.START_WINDOW {
			m.pulseFrame++
			return m, commands.PulseTickCmd()
		}

	case time.Time:
		timeIsOn := m.user.GetShowTimeState()
		if timeIsOn {
			m.curTime = msg
			m.rightHeaderData[1] = m.curTime.Format("15:04:05")
			return m, commands.TimeTickCmd()
		}

	case tea.KeyMsg:
		if m.state == states.LOAD_STATE {
			switch msg.String() {
			case "alt+й", "alt+Й", "alt+q", "alt+Q":
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "alt+й", "alt+Й", "alt+q", "alt+Q":
			return m, tea.Quit

		case "alt+s", "alt+ы", "alt+S", "alt+Ы":
			if m.isLoggedIn() {
				m.activeTab = 5
				m, cmd = m.syncTabState()
			}
		case "alt+f", "alt+а", "alt+А", "alt+F":
			if m.isLoggedIn() {
				m.activeTab = 0
				m, cmd = m.syncTabState()
			}
		case "alt+g", "alt+G", "alt+п", "alt+П":
			if m.isLoggedIn() {
				m.activeTab = 2
				m, cmd = m.syncTabState()
			}

		case "alt+ц", "alt+Ц", "alt+w", "alt+W":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					iden := i.Identity
					nick := ansi.Strip(iden.Nickname)
					var sd bool
					id := i.Identity.ID
					usE, err := m.user.GetUsersSetup(nick)
					if err != nil {
						return m.Err(err)
					}
					iden.Nickname = nick

					sd = usE.SoftDenoise
					usT, ok := m.usersStates[id]

					if ok && ((!sd == true && !usT.softDenoised) || !sd == false) {
						seq := tea.Sequence(commands.OnOffUsersSoftDenoise(m.user, iden, !sd),
							m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
								i.PersonalHardDenoise, !sd))
						cmds = append(cmds, seq)
					}
				}
			}

		case "alt+r", "alt+R", "alt+к", "alt+К":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					iden := i.Identity
					nick := ansi.Strip(iden.Nickname)
					id := iden.ID
					var hd bool
					usE, err := m.user.GetUsersSetup(nick)
					if err != nil {
						return m.Err(err)
					}
					iden.Nickname = nick
					hd = usE.HardDenoise
					usT, ok := m.usersStates[id]
					if ok && ((!hd == true && !usT.hardDenoised) || !hd == false) {
						seq := tea.Sequence(commands.OnOffUsersHardDenoise(m.user, iden, !hd),
							m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient, i.Muted,
								!hd, i.PersonalSoftDenoise))
						cmds = append(cmds, seq)
					}
				}
			}

		case "alt+с", "alt+С", "alt+c", "alt+C":
			if m.isLoggedIn() {
				m.activeTab = 1
				m, cmd = m.syncTabState()
			}

		case "alt+d", "alt+D", "alt+в", "alt+В":
			if m.isLoggedIn() {
				m.activeTab = 3
				m, cmd = m.syncTabState()
			}

		case "alt+у", "alt+У", "alt+e", "alt+E":
			if m.isLoggedIn() {
				m.activeTab = 4
				m, cmd = m.syncTabState()
			}

		case "alt+v", "alt+М", "alt+V", "alt+м":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteMicCmd(m.user))
			}
		case "alt+b", "alt+и", "alt+B", "alt+И":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.user.Engines.AudioEngine != nil {
				cmds = append(cmds, commands.MuteUnmuteCmd(m.user))
			}

		case "alt+h", "alt+H", "alt+р", "alt+Р":
			if m.curWindow == windows.DEF_WINDOW {
				if m.state == states.HELP_STATE {
					m, cmd = m.syncTabState()
					return m, cmd
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
			if m.isLoggedIn() && m.activeTab == 2 && m.connected && m.user.Engines.AudioEngine != nil {
				if i, ok := m.connectionsList.LipList.SelectedItem().(lists.ConnectionItem); ok {
					iden := i.Identity
					nick := ansi.Strip(iden.Nickname)
					iden.Nickname = nick
					seq := tea.Sequence(commands.MuteUnmuteUserCmd(m.user, iden),
						m.connectionsList.UpdateConnectionItemList(iden, i.VolumeCoefficient,
							!i.Muted, i.PersonalHardDenoise, i.PersonalSoftDenoise))
					cmds = append(cmds, seq)
				}
			}

		case "alt+x", "alt+X", "alt+ч", "alt+Ч":
			if m.state == states.ERR_STATE {
				return m, nil
			}
			if m.isLoggedIn() && m.activeTab == 4 && m.friendsReqsList.LipList.Index() >= 0 {
				if i, ok := m.friendsReqsList.LipList.SelectedItem().(lists.FriendReqItem); ok {
					nick := ansi.Strip(i.Identity.Nickname)
					i.Identity.Nickname = nick
					cmds = append(cmds, commands.DenyFriendRequestCmd(m.user, i.Identity))
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
					iden := i.Identity
					nick := ansi.Strip(iden.Nickname)
					iden.Nickname = nick
					seq := tea.Sequence(commands.SetUserVolumeCmd(m.user, iden, vc),
						m.connectionsList.UpdateConnectionItemList(iden, vc, i.Muted, i.PersonalHardDenoise, i.PersonalSoftDenoise))
					cmds = append(cmds, seq)
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
					iden := i.Identity
					nick := ansi.Strip(iden.Nickname)
					iden.Nickname = nick
					seq := tea.Sequence(commands.SetUserVolumeCmd(m.user, iden, vc),
						m.connectionsList.UpdateConnectionItemList(iden, vc, i.Muted, i.PersonalHardDenoise, i.PersonalSoftDenoise))
					cmds = append(cmds, seq)
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
					m, cmd = m.syncTabState()
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
					m, cmd = m.syncTabState()
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
				m, cmd = m.syncTabState()
				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
				return m, cmd
			}
			if m.curWindow == windows.START_WINDOW {
				m.state = states.LOAD_STATE
				m.curWindow = windows.DEF_WINDOW
				m, cmd = m.syncTabState()
				m.focusInputs()
				return m, cmd
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

				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
				m, cmd = m.syncTabState()
				m.log.Info("states after right", m.state, m.prState)
				return m, cmd
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

				if m.sideState == states.ZERO_STATE {
					m.sideState = states.LEFT_STATE
				}
				m, cmd = m.syncTabState()
				m.log.Info("states after left", m.state, m.prState)
				return m, cmd
			}

		case "esc":
			switch m.state {
			case states.ERR_STATE:
				m.log.Info("states before esc", m.state, m.prState)
				m.err = nil
				m.curWindow = windows.DEF_WINDOW
				if m.prState == states.LOAD_STATE {
					m, cmd = m.syncTabState()
				} else {
					m.state = m.prState
				}
			case states.HELP_STATE:
				m.err = nil
				m.curWindow = windows.DEF_WINDOW
				if m.prState == states.LOAD_STATE {
					m, cmd = m.syncTabState()
				} else {
					m.state = m.prState
				}
			default:
				if m.activeTab != 5 {
					m.state = m.prState
				} else if m.prState == states.ACCOUNT_STATE {
					m.state = m.prState
					m.prState = states.SETTINGS_STATE
				} else {
					m.state = states.SETTINGS_STATE
				}
				m, cmd = m.syncTabState()
			}
			m.log.Info("states after esc", m.state, m.prState)
			return m, cmd

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
				case 5:
					if m.sideState == states.RIGHT_STATE {
						m.cursor--
						changed = true
					}
				case 4:
					if m.cursor == len(m.appereanceInputs) {
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
				m, cmd = m.syncTabState()
				return m, cmd
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
				case 5:
					var c bool
					if m.sideState == states.RIGHT_STATE {
						switch m.state {
						case states.NICKNAME_STATE:
							if m.cursor < len(m.nicknameInputs)-1 {
								c = true
							}
						case states.PASSWORD_STATE:
							if m.cursor < len(m.passwordInputs)-1 {
								c = true
							}
						}
					}
					if c {
						m.cursor++
						changed = true
					}
				case 4:
					if m.cursor < len(m.appereanceInputs) {
						// if m.cursor == len(m.appereanceInputs)-1 {
						// 	changed = true
						// 	break
						// }
						m.cursor++
						changed = true
					}
				}
			}
			if changed {
				m, cmd = m.syncTabState()
				return m, cmd
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
				m, cmd = m.syncTabState()
				m.focusInputs()
				return m, cmd
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

					for i := range m.regTextInputs {
						if _, em := isEmptyString(m.regTextInputs[i].Value()); em {
							break
						}
					}
					if err := m.regTextInputs[0].Err; err != nil {
						return m.Err(err)
					}
					m.user.Data.Identity.Nickname = m.regTextInputs[0].Value()

					if err := m.regTextInputs[1].Err; err != nil {
						return m.Err(err)
					}
					password := m.regTextInputs[1].Value()

					repPassword := m.regTextInputs[2].Value()
					m.user.Data.Personal.RegisterTime = time.Now().Format("2006-01-02")

					bP := []byte(password)

					if !slices.Equal([]byte(repPassword), bP) {
						return m.Err(errs.ErrPasswordsNotEqual())

					}

					m.prState = m.state
					m.state = states.LOAD_STATE

					cmds = append(cmds, commands.RegisterCmd(m.user, m.sshEventsChan, m.log, bP), m.spinner.Tick)
					for i := range m.regTextInputs {
						m.regTextInputs[i].Reset()
					}

				case 1:

					for i := range m.logingInput {
						if _, em := isEmptyString(m.logingInput[i].Value()); em {
							break
						}
					}
					m.user.Data.Identity.Nickname = m.logingInput[0].Value()
					password := m.logingInput[1].Value()
					m.prState = m.state
					m.state = states.LOAD_STATE
					cmds = append(cmds, commands.LoginCmd(m.user, m.sshEventsChan, m.log, []byte(password)),
						m.spinner.Tick)
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
						if i, ok := m.friendsList.LipList.SelectedItem().(lists.FriendItem); ok && i.IsOnline {
							conns := i.Connections
							iden := i.Identity
							if !m.connected {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToAllUsersCmd(m.user, iden), m.spinner.Tick)
							} else if m.connected && !isInConnections(conns, m.user.Data.Identity.ID) {
								m.prState = m.state
								m.state = states.LOAD_STATE

								sequence := tea.Sequence(commands.LeaveCmd(m.user.Networking),
									commands.ConnectToAllUsersCmd(m.user, iden))
								cmds = append(cmds, sequence, m.spinner.Tick)
							}
						}

					case states.RIGHT_STATE:
						defer m.friendsInputs[m.cursor].Reset()
						var ok bool
						switch m.cursor {
						case 0:
							nick, ok = isEmptyString(m.friendsInputs[m.cursor].Value())
							if ok {
								break
							}
							iden, err := m.user.GetFriendIdentityByNick(nick)
							if err != nil {
								return m.Err(err)
							}
							if !m.connected {
								m.prState = m.state
								m.state = states.LOAD_STATE
								cmds = append(cmds, commands.ConnectToAllUsersCmd(m.user, iden), m.spinner.Tick)
							} else if m.connected && !isInConnections(m.connections, iden.ID) {
								m.prState = m.state
								m.state = states.LOAD_STATE

								sequence := tea.Sequence(commands.LeaveCmd(m.user.Networking),
									commands.ConnectToAllUsersCmd(m.user, iden))
								cmds = append(cmds, sequence, m.spinner.Tick)
							}
						case 1:
							nick, ok = isEmptyString(m.friendsInputs[m.cursor].Value())
							if ok {
								break
							}
							iden, _ := m.user.GetFriendIdentityByNick(nick)
							m.log.Info("iden", iden)
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.SendFriendRequestCmd(m.user, iden), m.spinner.Tick)
						case 2:
							nick, ok = isEmptyString(m.friendsInputs[m.cursor].Value())
							if ok {
								break
							}
							iden, err := m.user.GetFriendIdentityByNick(nick)
							if err != nil {
								return m.Err(err)
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.DeleteFromFriendsCmd(m.user, iden, true), m.spinner.Tick)
						case 3:
							nick, ok = isEmptyString(m.friendsInputs[m.cursor].Value())
							if ok {
								break
							}
							iden, _ := m.user.GetFriendIdentityByNick(nick)

							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.BlockUserCmd(m.user, iden), m.spinner.Tick)
						case 4:
							nick, ok = isEmptyString(m.friendsInputs[m.cursor].Value())
							if ok {
								break
							}
							iden, _ := m.user.GetFriendIdentityByNick(nick)
							// if err != nil {
							// 	return m.Err(err)
							// }
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.UnblockUserCmd(m.user, iden), m.spinner.Tick)
						}

					default:
						return m, nil
					}

				case 1:
					val, ok := isEmptyString(m.chatTextInput.Value())
					if ok {
						break
					}
					toSend := []byte(val)
					if m.imageBuffer != nil {
						img, _, err := image.Decode(bytes.NewReader(m.imageBuffer))
						if err != nil {
							m.imageBuffer = nil
							return m.Err(err)
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
						imageWidget.SetProtocol(termimg.Halfblocks)

						toSend = utils.SetThreeFirstByte([]byte{'i', 'm', 'g'}, m.imageBuffer)

						imageWidget.SetSizeWithCorrection(int(float64(size.X)*1.5), int(float64(size.Y)*1.5))

						rendered, err := imageWidget.Render()
						if err != nil {
							m.imageBuffer = nil
							return m.Err(err)
						}

						val = fmt.Sprintf("\n%s", rendered)
						m.imageBuffer = nil
					}

					cmds = append(cmds, commands.SendInChatCmd(m.user.Networking, toSend))
					t := time.Now().Format("15:04:05")
					iden := m.user.GetUserIdentity()
					iden.Nickname = lipgloss.NewStyle().Foreground(lipgloss.Color(m.userColor)).Render(iden.Nickname)
					m.messages = append(m.messages, commands.ChatMessage{
						Identity: iden,
						Time:     t,
						Text:     val})
					m.chatTextInput.Reset()

				case 2:
					if m.connected {
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.LeaveCmd(m.user.Networking))
					}
				case 3:
					if m.connected {
						m.prState = m.state
						m.state = states.LOAD_STATE
						cmds = append(cmds, commands.OnOffWebcamCmd(m.user), m.spinner.Tick)
					}
				case 4:
					switch m.sideState {
					case states.LEFT_STATE:
						if i, ok := m.friendsReqsList.LipList.SelectedItem().(lists.FriendReqItem); ok {
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.AcceptFriendRequestCmd(m.user, i.Identity), m.spinner.Tick)
						}
					case states.RIGHT_STATE:
						switch m.cursor {
						case 0:
							newColor, ok := isEmptyString(m.appereanceInputs[0].Value())
							if ok {
								break
							}

							if newColor == "d" {
								newColor = users.DEF_TM
							}
							if !strings.HasPrefix(newColor, "#") {
								newColor = "#" + newColor
							}
							if len(newColor) != 7 {
								return m, nil
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeThemeColorCmd(m.user, newColor))
							m.appereanceInputs[0].Reset()
						case 1:
							newBFTag := strings.TrimSpace(m.appereanceInputs[1].Value())
							m.prState = m.state
							m.state = states.LOAD_STATE
							if newBFTag == "d" {
								newBFTag = users.DEF_BFTAG
							}
							cmds = append(cmds, commands.ChangeBFTagCmd(m.user, newBFTag))
							m.appereanceInputs[1].Reset()
						case 2:
							newNotifyTag := strings.TrimSpace(m.appereanceInputs[2].Value())

							if newNotifyTag == "d" {
								newNotifyTag = users.DEF_NOTTAG
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeNotificationTagCmd(m.user, newNotifyTag))
							m.appereanceInputs[2].Reset()
						case 3:
							newBanTag := strings.TrimSpace(m.appereanceInputs[3].Value())
							m.prState = m.state
							m.state = states.LOAD_STATE
							if newBanTag == "d" {
								newBanTag = users.DEF_BANTAG
							}
							cmds = append(cmds, commands.ChangeBanTagCmd(m.user, newBanTag))
							m.appereanceInputs[3].Reset()
						case 4:
							if i, ok := m.apearenceList.LipList.SelectedItem().(lists.SwitcherItem); ok {
								m.prState = m.state
								m.state = states.LOAD_STATE
								switch i.Id {
								case lists.TIME:
									cmds = append(cmds, commands.OnOffShowTime(m.user))
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
									cmds = append(cmds, commands.OnOffHardDenoiceCmd(m.user, m.connected))
								case lists.SOFT_DENOISE:
									cmds = append(cmds, commands.OnOffSoftDenoiceCmd(m.user, m.connected))
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
						case states.ACCOUNT_STATE:
							return m.selectAccountSetting()
						case states.NICKNAME_STATE:
							newNickname, ok := isEmptyString(m.nicknameInputs[0].Value())
							if ok {
								break
							}
							password, ok := isEmptyString(m.nicknameInputs[1].Value())
							if ok {
								break
							}
							m.prState = m.state
							m.state = states.LOAD_STATE

							cmds = append(cmds, commands.NewNicknameCmd(m.user, newNickname, []byte(password)), m.spinner.Tick)
							for i := range m.nicknameInputs {
								m.nicknameInputs[i].Reset()
							}

						case states.TAGLINE_STATE:
							newTagline := strings.TrimSpace(m.taglineInput.Value())
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeTaglineCmd(m.user, newTagline), m.spinner.Tick)
							m.taglineInput.Reset()
						case states.COLOR_STATE:

							newColor, ok := isEmptyString(m.colorInput.Value())
							if ok {
								break
							}
							if newColor == "r" {
								newColor = "#random"
							} else if !strings.HasPrefix(newColor, "#") {
								newColor = "#" + newColor
							}
							if len(newColor) != 7 {
								break
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.ChangeColorCmd(m.user, newColor), m.spinner.Tick)
							m.colorInput.Reset()
						case states.PASSWORD_STATE:

							oldPassword := []byte(strings.TrimSpace(m.passwordInputs[0].Value()))
							if err := m.passwordInputs[1].Err; err != nil {
								return m.Err(err)
							}
							newPassword := []byte(strings.TrimSpace(m.passwordInputs[1].Value()))
							if slices.Equal(newPassword, oldPassword) {
								return m.Err(errs.ErrOldAndNewPasswordEqual())
							}
							repPassword := []byte(strings.TrimSpace(m.passwordInputs[2].Value()))

							if !slices.Equal(newPassword, repPassword) {
								return m.Err(errs.ErrPasswordsNotEqual())
							}
							m.prState = m.state
							m.state = states.LOAD_STATE
							cmds = append(cmds, commands.NewPasswordCmd(m.user, oldPassword, newPassword), m.spinner.Tick)
							for i := range m.passwordInputs {
								m.passwordInputs[i].Reset()
							}
						}
					case states.LEFT_STATE:
						switch m.state {
						case states.SETTINGS_STATE:
							return m.selectSetting()
						case states.ACCOUNT_STATE:
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
					m.friendsList.LipList, cmd = m.friendsList.LipList.Update(msg)
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
					if m.cursor < len(m.appereanceInputs) {
						for i := range m.appereanceInputs {
							m.appereanceInputs[i], cmd = m.appereanceInputs[i].Update(msg)
							cmds = append(cmds, cmd)
						}
					} else if m.cursor == len(m.appereanceInputs) {
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
					case states.ACCOUNT_STATE:
						m.accountList.LipList, cmd = m.accountList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.NICKNAME_STATE:
						for i := range m.nicknameInputs {
							m.nicknameInputs[i], cmd = m.nicknameInputs[i].Update(msg)
							cmds = append(cmds, cmd)
						}
					case states.PASSWORD_STATE:
						for i := range m.passwordInputs {
							m.passwordInputs[i], cmd = m.passwordInputs[i].Update(msg)
							cmds = append(cmds, cmd)
						}
					case states.COLOR_STATE:
						m.colorInput, cmd = m.colorInput.Update(msg)
						cmds = append(cmds, cmd)
					case states.TAGLINE_STATE:
						m.taglineInput, cmd = m.taglineInput.Update(msg)
						cmds = append(cmds, cmd)
					}
				case states.LEFT_STATE:
					switch m.state {
					case states.SETTINGS_STATE:
						m.settingsList.LipList, cmd = m.settingsList.LipList.Update(msg)
						cmds = append(cmds, cmd)
					case states.ACCOUNT_STATE:
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
