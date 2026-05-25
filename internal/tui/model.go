package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bz "github.com/lrstanley/bubblezone"
)

type userColors struct {
	mainColor lipgloss.Color
	subColor  lipgloss.Color
}

type Model struct {
	width  int
	height int

	defTabs []string
	regTabs []string

	sideState uint

	activeTab int

	logoAnim    []string
	notConnAnim []string
	aloneAnim   []string

	state   uint
	prState uint

	spinner spinner.Model

	curWindow uint

	regTextInputs []textinput.Model
	chatTextInput textinput.Model
	logingInput   []textinput.Model
	friendsInputs []textinput.Model
	profileInputs []textinput.Model

	eventsChan chan sshclient.Event

	defaultThemeColor string

	imageBuffer []byte

	speaking bool

	chatOffset int
	chatWidth  int
	chatHeight int

	ticked bool

	userColor string

	cursor int

	user *users.User

	muteState string

	messages []commands.ChatMessage

	zone *bz.Manager

	usersColors map[string]userColors
	connections []string
	online      map[string][]string

	microphonesList     list.Model
	microphonesDelegate list.DefaultDelegate
	connectionsList     list.Model
	connectionsDelegate list.DefaultDelegate
	onlineList          list.Model
	onlineDelegate      lists.DynamicOnlineDelegate
	settingsList        list.Model
	settingsDelegate    list.DefaultDelegate
	friendsReqsList     list.Model
	friendsReqsDelegate lists.DynamicFriendsReqDelegate

	tabsNotifications map[string]struct{}

	connected     bool

	msgChan    chan commands.ChatMessage
	rawMsgChan chan commands.RawChatMessage

	peerConnectionsChan    chan commands.PeerConnectedMsg
	peerDisconnectionsChan chan commands.PeerDisconnectedMsg

	log *logger.Logger

	animFrame  int
	pulseFrame int

	headerActiveStyle lipgloss.Style
	themeColor        lipgloss.Color
	subThemeColor     lipgloss.Color

	defaultBFTag            string
	defaultNotificationSign string

	stopCountMinutesChan chan struct{}

	err error
}

func NewModel(logFilePath, dataFilePath, keysPath string, appLogger *logger.Logger) (*Model, error) {
	op := "model.NewModel"
	log := appLogger.AddOp(op)
	log.Info("creating new model...")

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &Model{
		state:   states.START_STATE,
		prState: states.START_STATE,

		curWindow: windows.START_WINDOW,

		defTabs: []string{"friends", "chat", "voice", "video", "profile", "settings"},
		regTabs: []string{"registration", "login"},

		regTextInputs: make([]textinput.Model, 3),
		logingInput:   make([]textinput.Model, 2),
		logoAnim:      make([]string, 0, 4),
		notConnAnim:   make([]string, 0, 4),
		profileInputs: make([]textinput.Model, 3),
		friendsInputs: make([]textinput.Model, 5),
		messages:      make([]commands.ChatMessage, 0, 20),

		eventsChan: make(chan sshclient.Event, 50),

		defaultThemeColor:       "#A6E22E",
		defaultBFTag:            "👑",
		defaultNotificationSign: "🔔",

		spinner: sp,

		online: make(map[string][]string, 5),

		stopCountMinutesChan: make(chan struct{}, 1),

		usersColors: make(map[string]userColors, 5),

		zone: bz.New(),

		connections: make([]string, 0, 3),

		userColor: randomcolor.GetRandomColorInHex(),

		log: appLogger,

		tabsNotifications: make(map[string]struct{}, 6),

		msgChan:    make(chan commands.ChatMessage, 100),
		rawMsgChan: make(chan commands.RawChatMessage, 100),

		peerConnectionsChan:    make(chan commands.PeerConnectedMsg, 100),
		peerDisconnectionsChan: make(chan commands.PeerDisconnectedMsg, 100),
	}

	m.logoAnim = []string{titles.BIG_LOGO1, titles.BIG_LOGO2, titles.BIG_LOGO3, titles.BIG_LOGO2}
	m.notConnAnim = []string{titles.NOT_CONN1, titles.NOT_CONN2, titles.NOT_CONN3, titles.NOT_CONN2}
	m.aloneAnim = []string{titles.ALONE1, titles.ALONE2, titles.ALONE3, titles.ALONE2}

	user := users.NewUser(logFilePath, keysPath, dataFilePath, m.defaultThemeColor)
	user.Data.Setup.BestFriendTag = m.defaultBFTag
	user.Data.Setup.NotificaionSign = m.defaultNotificationSign

	userDataBytes, err := os.ReadFile(user.Paths.DataFilePath)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Error("failed to read userdata.json file", logger.Err(err))
			return nil, err
		}
	}

	if len(userDataBytes) > 1 {
		log.Info("userdata.json file is not empty")
		if err := json.Unmarshal(userDataBytes, &user.Data); err != nil {
			log.Error("failed to unmarshal userdata.json file", logger.Err(err))
			return nil, err
		}
	}

	m.themeColor = lipgloss.Color(user.Data.Setup.ThemeColor)
	m.subThemeColor = lipgloss.Color(utils.DarkenHex(user.Data.Setup.ThemeColor, 0.7))
	m.headerActiveStyle = lipgloss.NewStyle().Foreground(m.themeColor).Bold(true)

	if user.Data.Personal.Nickname != "" && user.Data.Personal.RegisterTime != "" {

		logNickname := logger.Attr("nickname", user.Data.Personal.Nickname)

		log.Info("authorize user with existing user data", logNickname)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		client, personalData, err := sshclient.AuthSSHClient(ctx, appLogger, sshclient.SSHClientSetup{
			Nickname:   user.Data.Personal.Nickname,
			KeysPath:   user.Paths.KeysPath,
			Typee:      sshclient.DEFAULT,
			EventsChan: m.eventsChan,
			Password:   nil,
		})
		if err != nil {
			log.Error("err when sshclient", logger.Err(err))
			if !errors.Is(err, errs.ErrAuth) {
				log.Error("err sshclient", logger.Err(err))
				m.err = err
				m.state = states.ERR_STATE
				user.Data.Personal.Nickname = ""
				user.Data.Personal.RegisterTime = ""
			}
		} else {
			var pd struct {
				Nickname     string            `json:"nickname"`
				RegisterTime time.Time         `json:"registerTime"`
				FriendsReqs  []users.FriendReq `json:"friendsReqs"`
				Friends      []string          `json:"friends"`
			}

			if err := json.Unmarshal(personalData, &pd); err != nil {
				log.Error("failed to unmarshal users's personal data", logger.Err(err), logNickname)
				return nil, err
			}

			user.Data.Personal.Nickname = pd.Nickname
			user.Data.Personal.RegisterTime = pd.RegisterTime.Local().Format("2006-01-02")
			user.Data.Personal.FriendsReqs = pd.FriendsReqs
			if len(user.Data.Personal.FriendsReqs) > 0 {
				m.tabsNotifications["profile"] = struct{}{}
			}
			user.Data.Personal.Friends = pd.Friends
			user.SSHClient = client
			log.Info("user authorized successfully, networking setting...", logNickname)

			networking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
			if err != nil {
				log.Error("failed to create networking", logger.Err(err), logNickname)
				return nil, err
			}
			audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
				Microphone:  user.Data.Devices.Microphone,
				Aec:         user.Data.Setup.AEC,
				HardDenoice: user.Data.Setup.HardDenoise,
				SoftDenoice: user.Data.Setup.SoftDenoise,
				Filtered:    user.Data.Setup.Filter,
			})
			if err != nil {
				log.Error("failed to create audio engine", logger.Err(err))
				return nil, err
			}

			if err := audioEngine.SetNetworking(networking); err != nil {
				log.Error("failed to set newtorking to audio engine", logger.Err(err), logNickname)
				return nil, err
			}

			log.Info("setting netwoking callbacks...", logNickname)

			networking.ChatCallback(func(id string, data []byte) {
				t := time.Now().Format("15:04:05")
				m.rawMsgChan <- commands.RawChatMessage{Time: t, Nickname: id, Data: data}

			})
			networking.VoiceCallback(func(id string, data []byte) {

				audioEngine.PlayUserVoice(id, data)
			})
			networking.PeerConnectedCallback(func(id string) {
				t := time.Now().Format("15:04:05")
				m.peerConnectionsChan <- commands.PeerConnectedMsg{Nickname: id, Time: t}
			})
			networking.PeerDisconnectedCallback(func(id string) {
				t := time.Now().Format("15:04:05")
				m.peerDisconnectionsChan <- commands.PeerDisconnectedMsg{Nickname: id, Time: t}
			})
			user.SSHClient = client
			user.Engines.AudioEngine = audioEngine
			user.Networking = networking

		}
	}

	m.user = user

	if m.user.Engines.AudioEngine != nil {
		m.setupMicrohonesList()
		m.setupSettingsList()
	}

	m.setupConnestionsList()
	m.setupOnlineList()
	m.setupFriendsReqsList()

	for i := range m.regTextInputs {
		ti := textinput.New()
		ti.CharLimit = 32
		switch i {
		case 0:
			ti.Placeholder = "unique nickname"
		case 1:
			ti.Placeholder = "secret"
			ti.EchoMode = textinput.EchoPassword
		case 2:
			ti.Placeholder = "repeat secret"
			ti.EchoMode = textinput.EchoPassword
		}
		m.regTextInputs[i] = ti
	}

	for i := range m.logingInput {
		ti := textinput.New()
		ti.CharLimit = 32
		switch i {
		case 0:
			ti.Placeholder = "nickname"
		case 1:
			ti.Placeholder = "secret"
			ti.EchoMode = textinput.EchoPassword
		}
		m.logingInput[i] = ti
	}

	for i := range m.friendsInputs {
		ti := textinput.New()
		ti.CharLimit = 32
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

	for i := range m.profileInputs {
		ti := textinput.New()
		ti.CharLimit = 32
		switch i {
		case 0:
			ti.Placeholder = "new color in hex, d to default"
		case 1:
			ti.Placeholder = "new best friend tag, d to default"
		case 2:
			ti.Placeholder = "new notification sign, d to default"
		}
		m.profileInputs[i] = ti
	}

	log.Info("model created successfully")

	return m, nil
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, commands.AnimTickCmd(), commands.PulseTickCmd(), m.spinner.Tick)
	if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
		cmds = append(cmds, commands.FetchOnlineFriendsCmd(m.user.Networking, m.user.Data.Personal.Friends),
			commands.WaitForChatMessageCmd(m.msgChan),
			commands.WaitForEventMessageCmd(m.eventsChan),
			commands.WaitForRawChatMessageCmd(m.rawMsgChan),
			commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
			commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan),
			commands.TickCmd(), tea.EnableMouseCellMotion)
		return tea.Batch(cmds...)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Clean() {
	if m.user.Engines.AudioEngine != nil {
		m.user.Engines.AudioEngine.SetDisconnected()
		m.user.Engines.AudioEngine.Stop()
	}

	if m.user.Networking != nil {
		m.user.Networking.Close()
	}

	if m.user.SSHClient != nil {
		m.user.SSHClient.Close()
	}
}
