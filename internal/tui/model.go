package tui

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/internal/utils"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
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

	logoAnim []string

	state   uint
	prState uint

	spinner spinner.Model

	curWindow uint

	regTextInputs  []textinput.Model
	connTextInputs textinput.Model
	chatTextInput  textinput.Model
	logingInput    []textinput.Model

	defaultThemeColor string

	imageBuffer []byte

	speaking bool

	chatOffset int
	chatWidth  int
	chatHeight int

	ticked bool

	userColor string

	cursor int

	user *entities.User

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
	onlineDelegate      list.DefaultDelegate
	settingsList        list.Model
	settingsDelegate    list.DefaultDelegate

	connected bool

	msgChan    chan commands.ChatMessage
	rawMsgChan chan commands.RawChatMessage

	peerConnectionsChan    chan commands.PeerConnectedMsg
	peerDisconnectionsChan chan commands.PeerDisconnectedMsg

	log *logger.Logger

	animFrame int

	headerActiveStyle lipgloss.Style
	themeColor        lipgloss.Color
	subThemeColor     lipgloss.Color

	themeColorInput textinput.Model

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
		logoAnim:      make([]string, 0, 3),
		messages:      []commands.ChatMessage{},

		defaultThemeColor: "#A6E22E",

		spinner: sp,

		online: make(map[string][]string, 5),

		usersColors: make(map[string]userColors, 5),

		zone: bz.New(),

		connections: make([]string, 0),

		userColor: randomcolor.GetRandomColorInHex(),

		log: appLogger,

		msgChan:    make(chan commands.ChatMessage, 100),
		rawMsgChan: make(chan commands.RawChatMessage, 100),

		peerConnectionsChan:    make(chan commands.PeerConnectedMsg, 100),
		peerDisconnectionsChan: make(chan commands.PeerDisconnectedMsg, 100),
	}

	m.logoAnim = []string{titles.BIG_LOGO1, titles.BIG_LOGO2, titles.BIG_LOGO3, titles.BIG_LOGO2}

	user := entities.User{
		Paths: entities.Paths{
			LogFilePath:  logFilePath,
			KeysPath:     keysPath,
			DataFilePath: dataFilePath,
		},
	}

	userDataBytes, err := os.ReadFile(user.Paths.DataFilePath)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Error("failed to read userdata.json file", logger.Err(err))
			return nil, err
		}
	}

	userData := entities.Data{
		Setup: entities.Setup{
			UsersSetup:           make(map[string]entities.UsersSetup, 0),
			ThemeColor:           m.defaultThemeColor,
			AudioNotifications:   true,
			DesktopNotifications: true,
		},
	}

	if len(userDataBytes) > 1 {
		log.Info("userdata.json file is not empty")
		if err := json.Unmarshal(userDataBytes, &userData); err != nil {
			log.Error("failed to unmarshal userdata.json file", logger.Err(err))
			return nil, err
		}
	}

	user.Data = userData

	log.Info("userdata", user.Data)

	m.themeColor = lipgloss.Color(user.Data.Setup.ThemeColor)
	m.subThemeColor = lipgloss.Color(utils.DarkenHex(user.Data.Setup.ThemeColor, 0.7))
	m.headerActiveStyle = lipgloss.NewStyle().Foreground(m.themeColor).Bold(true)

	if userData.Personal.Nickname != "" && userData.Personal.RegisterTime != "" {

		logNickname := logger.Attr("nickname", userData.Personal.Nickname)

		log.Info("authorize user with existing user data", logNickname)
		if _, err := auth.Auth(userData.Personal.Nickname, keysPath, auth.DEFAULT, nil); err != nil {
			log.Error("err when auth", logger.Err(err))
			if !errors.Is(err, errs.ErrAuth) {
				log.Error("err auth", logger.Err(err))
				m.err = err
				m.state = states.ERR_STATE
				user.Data.Personal.Nickname = ""
				user.Data.Personal.RegisterTime = ""
			}
		} else {
			log.Info("user authorized successfully, networking setting...", logNickname)
			networking, err := networking.NewNetworking(userData.Personal.Nickname, user.Paths.LogFilePath)
			if err != nil {
				log.Error("failed to create networking", logger.Err(err), logNickname)
				return nil, err
			}
			audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
				Microphone:  user.Data.Devices.Microphone,
				Aec:         user.Data.Setup.AEC,
				HardDenoice: user.Data.Setup.HardDenoise,
				SoftDenoice: userData.Setup.SoftDenoise,
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
			t := time.Now().Format("15:04:05")
			networking.ChatCallback(func(id string, data []byte) {
				m.rawMsgChan <- commands.RawChatMessage{Time: t, Nickname: id, Data: data}

			})
			networking.VoiceCallback(func(id string, data []byte) {
				audioEngine.PlayUserVoice(id, data)
			})
			networking.PeerConnectedCallback(func(id string) {
				m.peerConnectionsChan <- commands.PeerConnectedMsg{Nickname: id, Time: t}
			})
			networking.PeerDisconnectedCallback(func(id string) {
				m.peerDisconnectionsChan <- commands.PeerDisconnectedMsg{Nickname: id, Time: t}
			})

			user.Engines.AudioEngine = audioEngine
			user.Networking = networking

		}
	}

	m.user = &user

	if m.user.Engines.AudioEngine != nil {
		m.setupMicrohonesList()
		m.setupSettingsList()
	}

	m.setupConnestionsList()
	m.setupOnlineList()

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

	connTextInput := textinput.New()
	connTextInput.Placeholder = "enter nickname"
	m.connTextInputs = connTextInput

	chatInput := textinput.New()
	chatInput.Placeholder = "type a message..."
	m.chatTextInput = chatInput

	themeColorInput := textinput.New()
	themeColorInput.Placeholder = "new color in hex, d to default"
	m.themeColorInput = themeColorInput

	log.Info("model created successfully")

	return m, nil
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, commands.AnimTickCmd(), m.spinner.Tick)
	if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
		cmds = append(cmds, commands.WaitForChatMessageCmd(m.msgChan),
			commands.WaitForRawChatMessageCmd(m.rawMsgChan),
			commands.WaitForPeerConnectionCmd(m.peerConnectionsChan),
			commands.WaitForPeerDisconnectionCmd(m.peerDisconnectionsChan),
			commands.FetchOnlineCmd(m.user.Networking, m.user.Data.Personal.Nickname),
			commands.TickCmd(), tea.EnableMouseCellMotion)
		return tea.Batch(cmds...)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Clean() {
	if m.user.Engines.AudioEngine != nil {
		m.user.Engines.AudioEngine.Stop()
	}

	if m.user.Networking != nil {
		m.user.Networking.Close()
	}
}
