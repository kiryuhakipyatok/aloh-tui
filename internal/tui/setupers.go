package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/utils"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

type CallbacksSetup struct {
	User            *users.User
	RawMsgChan      chan commands.RawChatMessage
	PeerConnChan    chan commands.PeerConnectedMsg
	PeerDisconnChan chan commands.PeerDisconnectedMsg
	NetwEventChan   chan commands.NetworkEventMsg
}

func SetupCallbacks(cs CallbacksSetup) {
	cs.User.Networking.ChatCallback(func(id uuid.UUID, data []byte) {
		t := time.Now().Format("15:04:05")
		cs.RawMsgChan <- commands.RawChatMessage{Time: t, Id: id, Data: data}
	})
	cs.User.Networking.VoiceCallback(func(id uuid.UUID, data []byte) {
		cs.User.Engines.AudioEngine.PlayUserVoice(id, data)
	})
	cs.User.Networking.WebcamCallback(func(id uuid.UUID, data []byte) {
		cs.User.Engines.VideoEngine.RenderUsersWebcam(id, data)
	})
	cs.User.Networking.ScreenCallback(func(id uuid.UUID, data []byte) {
		cs.User.Engines.VideoEngine.RenderUsersScreen(id, data)
	})
	cs.User.Networking.PeerConnectedCallback(func(id uuid.UUID) {
		t := time.Now().Format("15:04:05")
		cs.PeerConnChan <- commands.PeerConnectedMsg{Id: id, Time: t}
	})
	cs.User.Networking.PeerDisconnectedCallback(func(id uuid.UUID) {
		t := time.Now().Format("15:04:05")
		cs.PeerDisconnChan <- commands.PeerDisconnectedMsg{Id: id, Time: t}
	})

	cs.User.Networking.EventCallback(func(id uuid.UUID, e alohnetwork.Event) {
		cs.NetwEventChan <- commands.NetworkEventMsg{Id: id, Event: e}
	})
}

func (m *Model) setupModel() {
	m.setupColors()
	m.setupInputs()
	m.setupModelsLists()
	m.setupAnims()
}

func (m *Model) setupUsersModel() {
	m.setupColors()
	m.setupDTZ()
}

func (m *Model) setupInputs() {
	for i := range m.regTextInputs {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		switch i {
		case 0:
			ti.CharLimit = 24
			ti.Placeholder = "unique nickname"
		case 1:
			ti.Placeholder = "secret"
			ti.EchoMode = textinput.EchoPassword
			ti.Validate = validatePassword
		case 2:
			ti.Placeholder = "repeat secret"
			ti.EchoMode = textinput.EchoPassword
			ti.Validate = validatePassword
		}
		m.regTextInputs[i] = ti
	}

	for i := range m.logingInput {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		ti.CharLimit = 24
		switch i {
		case 0:
			ti.Placeholder = "nickname"
		case 1:
			ti.Placeholder = "secret"
			ti.EchoMode = textinput.EchoPassword
		}
		m.logingInput[i] = ti
	}

	for i := range m.nicknameInputs {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		switch i {
		case 0:
			ti.CharLimit = 24
			ti.Placeholder = "nickname"
		case 1:
			ti.Placeholder = "password"
			ti.EchoMode = textinput.EchoPassword
		}
		m.nicknameInputs[i] = ti
	}

	for i := range m.passwordInputs {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		ti.EchoMode = textinput.EchoPassword
		switch i {
		case 0:
			ti.Placeholder = "old password"
		case 1:
			ti.Placeholder = "new password"
		case 2:
			ti.Placeholder = "repeat new password"
		}
		m.passwordInputs[i] = ti
	}

	for i := range m.friendsInputs {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		ti.CharLimit = 24
		switch i {
		case 0:
			ti.Placeholder = "connect to friend"
		case 1:
			ti.Placeholder = "send friend request"
		case 2:
			ti.Placeholder = "delete from friends"
		case 3:
			ti.Placeholder = "block user"
		case 4:
			ti.Placeholder = "unblock user"
		}
		m.friendsInputs[i] = ti
	}

	chatInput := textinput.New()
	chatInput.Placeholder = "type a message..."
	m.chatTextInput = chatInput

	colorInput := textinput.New()
	colorInput.Placeholder = "new color in hex, r to tandom"
	m.colorInput = colorInput

	taglineInput := textinput.New()
	taglineInput.Placeholder = "new tagline"
	taglineInput.CharLimit = 28
	m.taglineInput = taglineInput

	for i := range m.appereanceInputs {
		ti := textinput.New()
		ti.PlaceholderStyle = styles.CGrayStyle
		ti.CharLimit = 7
		switch i {
		case 0:
			ti.Placeholder = "new color in hex, d to default"
		case 1:
			ti.Placeholder = "new best friend tag, d to default"
		case 2:
			ti.CharLimit = 1
			ti.Placeholder = "new notification tag, d to default"
		case 3:
			ti.CharLimit = 1
			ti.Placeholder = "new ban tag, d to default"
		}
		m.appereanceInputs[i] = ti
	}

	// for i := range m.accountInputs {
	// 	ti := textinput.New()
	// 	ti.PlaceholderStyle = styles.CGrayStyle
	// 	switch i {
	// 	case 0:
	// 		ti.CharLimit = 24
	// 		ti.Placeholder = "new nickname"
	// 	case 2:
	// 		ti.CharLimit = 28
	// 		ti.Placeholder = "new tagline"
	// 	}
	// 	m.appereanceInputs[i] = ti
	// }
}

func (m *Model) setupUsersLists() {
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

	m.friendsList = lists.SetupFriendsList(m.user, ls)

	m.friendsReqsList = lists.SetupFriendsReqsList(m.user, ls)

	m.apearenceList = lists.SetupSwitcherList(m.user, lists.APEREANCE, lists.ListSetup{
		ThemeColor:       m.themeColor,
		SubColor:         m.subThemeColor,
		NormalDescColor:  styles.CGray,
		NormalTitleColor: lipgloss.AdaptiveColor{Light: styles.Black, Dark: styles.White},
	})

	m.notificationsList = lists.SetupSwitcherList(m.user, lists.NOTIFICATIONS, ls)

	m.audioList = lists.SetupSwitcherList(m.user, lists.AUDIO, ls)
	m.accountList = lists.SetupAccountList(m.user, ls)
}

func (m *Model) setupFriendsColors() {
	friends := m.user.GetFriends()
	for _, f := range friends {
		m.usersColors[f.ID] = newUC(f.Color)
	}
}

func newUC(color string) styles.UserColors {
	var uc styles.UserColors
	if color != users.DEF_COLOR {
		uc = styles.UserColors{
			MainColor: lipgloss.Color(color),
			SubColor:  lipgloss.Color(utils.DarkenHex(color, 0.7)),
		}
	} else {
		hex := randomcolor.GetRandomColorInHex()
		uc = styles.UserColors{
			MainColor: lipgloss.Color(hex),
			SubColor:  lipgloss.Color(utils.DarkenHex(hex, 0.7)),
		}
	}
	return uc
}

func (m *Model) setupModelsLists() {
	ls := lists.ListSetup{
		ThemeColor:       m.themeColor,
		SubColor:         m.subThemeColor,
		NormalDescColor:  styles.CGray,
		NormalTitleColor: styles.CText,
	}

	m.connectionsList = lists.SetupConnestionsList(ls)

	m.settingsList = lists.SetupSettingsList(ls)
}

func (m *Model) setupDTZ() {
	if m.user.GetShowDateState() {
		curDate := m.curTime.Format("2006-01-02")
		m.rightHeaderData[0] = curDate
	}

	if m.user.GetShowTimeState() {
		curTime := m.curTime.Format("15:04:05")
		m.rightHeaderData[1] = curTime
	}

	if m.user.GetShowZoneState() {
		curZone := m.curTime.Format("-07:00")
		m.rightHeaderData[2] = curZone
	}
}

func (m *Model) setupAnims() {
	m.logoAnim = []string{titles.BIG_LOGO1, titles.BIG_LOGO2, titles.BIG_LOGO3, titles.BIG_LOGO2}
	m.notConnAnim = []string{titles.NOT_CONN1, titles.NOT_CONN2, titles.NOT_CONN3, titles.NOT_CONN2}
	m.aloneAnim = []string{titles.ALONE1, titles.ALONE2, titles.ALONE3, titles.ALONE2}
}

func setupUserColor(color string) lipgloss.Color {
	var c string
	if color == "#random" {
		c = randomcolor.GetRandomColorInHex()
	} else {
		c = color
	}
	return lipgloss.Color(c)
}

func (m *Model) setupColors() {
	m.userColor = setupUserColor(m.user.Data.Account.Color)
	m.themeColor = lipgloss.Color(m.user.Data.Setup.Appereance.ThemeColor)
	m.subThemeColor = lipgloss.Color(utils.DarkenHex(m.user.Data.Setup.Appereance.ThemeColor, 0.7))
	m.headerActiveStyle = lipgloss.NewStyle().Foreground(m.themeColor).Bold(true)
}
