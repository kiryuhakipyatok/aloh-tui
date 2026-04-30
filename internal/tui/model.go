package tui

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/notifications"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	bz "github.com/lrstanley/bubblezone"
)

type connectionItem struct {
	nickname          string
	volumeCoefficient float32
	muted             bool
}

func (ci connectionItem) Title() string { return ci.nickname }
func (ci connectionItem) Description() string {
	return fmt.Sprintf("volume: %.1f, muted: %t", ci.volumeCoefficient, ci.muted)
}
func (ci connectionItem) FilterValue() string { return ci.nickname }

type micItem struct {
	name string
}

func (mi micItem) Title() string       { return mi.name }
func (mi micItem) Description() string { return "" }
func (mi micItem) FilterValue() string { return mi.name }

type connData struct {
	nickname          string
	muted             bool
	volumeCoefficient float32
}

type Model struct {
	width  int
	height int

	state   uint
	prState uint

	curWindow uint

	regTextInputs  []textinput.Model
	connTextInputs []textinput.Model
	chatTextInput  textinput.Model
	logingInput    []textinput.Model

	speaking bool

	chatOffset int

	updateTick bool

	userColor string

	cursor int

	user *entities.User

	muteState string

	messages []commands.ChatMessage

	zone *bz.Manager

	usersColors map[string]lipgloss.Color
	connections []string
	online      []string

	microphonesList list.Model
	connectionsList list.Model

	connected bool

	msgChan chan commands.ChatMessage

	log *logger.Logger

	animFrame int

	err error
}

func NewModel(logFilePath, dataFilePath, keysPath string, appLogger *logger.Logger) (*Model, error) {
	op := "model.NewModel"
	log := appLogger.AddOp(op)
	log.Info("creating new model...")

	m := &Model{
		state:   states.START_STATE,
		prState: states.START_STATE,

		curWindow: windows.START_WINDOW,

		regTextInputs:  make([]textinput.Model, 3),
		connTextInputs: make([]textinput.Model, 1),
		logingInput:    make([]textinput.Model, 2),
		messages:       []commands.ChatMessage{},

		usersColors: make(map[string]lipgloss.Color),

		zone: bz.New(),

		connections: make([]string, 0),

		userColor: randomcolor.GetRandomColorInHex(),

		log: appLogger,

		msgChan: make(chan commands.ChatMessage, 100),
	}

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
			UsersSetup: make(map[string]entities.UsersSetup, 0),
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
			audioEngine, err := audio.NewAudioEngine(appLogger, user.Data.Devices.Microphone, user.Data.Setup.Denoise, user.Data.Setup.AEC)
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
				msg := string(data)
				m.msgChan <- commands.ChatMessage{Time: t, Nickname: id, Text: msg}
				m.user.Engines.AudioEngine.PlayNotification()
				if err := notifications.Notify(t, id, msg); err != nil {
					m.log.Error("failed to notify", logger.Attr("userId", id), logger.Err(err))
				}
			})
			networking.VoiceCallback(func(id string, data []byte) {
				audioEngine.PlayUserVoice(id, data)
			})

			user.Engines.AudioEngine = audioEngine
			user.Networking = networking

		}
	}

	m.user = &user

	m.setupMicrohonesList(m.user.Data.Devices.Microphone)
	m.setupConnestionsList()

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
	m.connTextInputs[0] = connTextInput

	chatInput := textinput.New()
	chatInput.Placeholder = "type a message..."
	m.chatTextInput = chatInput

	log.Info("model created successfully")

	return m, nil
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, commands.AnimTickCmd())
	if m.user.Networking != nil && m.user.Engines.AudioEngine != nil {
		cmds = append(cmds, commands.WaitForChatMessageCmd(m.msgChan),
			commands.FetchSessionsCmd(m.user.Networking, m.user.Data.Personal.Nickname),
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

func (m *Model) focusInputs() {
	for i := range m.regTextInputs {
		m.regTextInputs[i].Blur()
		m.regTextInputs[i].PromptStyle = lipgloss.NewStyle()
		m.regTextInputs[i].TextStyle = lipgloss.NewStyle()
	}
	for i := range m.connTextInputs {
		m.connTextInputs[i].Blur()
		m.connTextInputs[i].PromptStyle = lipgloss.NewStyle()
		m.connTextInputs[i].TextStyle = lipgloss.NewStyle()
	}

	for i := range m.logingInput {
		m.logingInput[i].Blur()
		m.logingInput[i].PromptStyle = lipgloss.NewStyle()
		m.logingInput[i].TextStyle = lipgloss.NewStyle()
	}
	m.chatTextInput.Blur()
	m.chatTextInput.PromptStyle = lipgloss.NewStyle()
	m.chatTextInput.TextStyle = lipgloss.NewStyle()

	switch m.state {
	case states.REG_STATE:
		m.regTextInputs[m.cursor].Focus()
		m.regTextInputs[m.cursor].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.regTextInputs[m.cursor].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	case states.CONN_STATE:
		m.connTextInputs[m.cursor].Focus()
		m.connTextInputs[m.cursor].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.connTextInputs[m.cursor].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	case states.CHAT_STATE:
		m.chatTextInput.Focus()
		m.chatTextInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.chatTextInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	case states.LOGIN_STATE:
		m.logingInput[m.cursor].Focus()
		m.logingInput[m.cursor].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.logingInput[m.cursor].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	}
}

func (m *Model) setupMicrohonesList(selected string) {
	if m.user != nil && m.user.Engines.AudioEngine != nil {
		ms := m.user.Engines.AudioEngine.FetchMicrophones()

		microphones := make([]list.Item, 0, len(ms))

		selectedIndex := 0

		for i, v := range ms {
			microphones = append(microphones, micItem{name: v.Name()})
			if v.Name() == selected {
				selectedIndex = i
			}
		}

		delegate := list.NewDefaultDelegate()
		delegate.ShowDescription = false
		delegate.SetSpacing(3)

		m.microphonesList = list.New(microphones, delegate, m.width/2, m.height-4)
		m.microphonesList.Title = "select microphone"
		m.microphonesList.Select(selectedIndex)
		m.microphonesList.SetShowStatusBar(false)
		m.microphonesList.SetShowTitle(true)
		m.microphonesList.SetFilteringEnabled(false)
		m.microphonesList.SetShowFilter(false)
		m.microphonesList.SetShowHelp(false)
	}
}

func (m *Model) setupConnestionsList() {
	conns := make([]list.Item, 0)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(3)

	m.connectionsList = list.New(conns, delegate, m.width/2, m.height-4)
	m.connectionsList.Title = "connections"
	m.connectionsList.SetShowStatusBar(false)
	m.connectionsList.SetShowTitle(false)
	m.connectionsList.SetFilteringEnabled(false)
	m.connectionsList.SetShowFilter(false)
	m.connectionsList.SetShowHelp(false)
}

func (m *Model) updateConnectionItemList(nickname string, volume float32, muted bool) tea.Cmd {
	var cmds []tea.Cmd
	items := m.connectionsList.Items()
	for i, v := range items {
		conn, ok := v.(connectionItem)
		if !ok {
			continue
		}
		if conn.nickname == nickname {
			conn.volumeCoefficient = volume
			conn.muted = muted
		}

		cmds = append(cmds, m.connectionsList.SetItem(i, conn))
	}
	return tea.Batch(cmds...)

}

func (m *Model) updateConnectionsList() tea.Cmd {

	var cmds []tea.Cmd
	items := m.connectionsList.Items()
	connsMap := make(map[string]struct{}, len(m.connections))
	for _, v := range m.connections {
		connsMap[v] = struct{}{}
	}

	itemsMap := make(map[string]struct{}, len(items))
	for _, v := range items {
		item, ok := v.(connectionItem)
		if !ok {
			continue
		}
		itemsMap[item.nickname] = struct{}{}
	}

	for i := len(items) - 1; i >= 0; i-- {
		item, ok := items[i].(connectionItem)
		if !ok {
			continue
		}
		_, ok = connsMap[item.nickname]
		if !ok {
			m.connectionsList.RemoveItem(i)
		}
	}

	for _, v := range m.connections {
		if _, ok := itemsMap[v]; !ok {
			itLenLen := len(m.connectionsList.Items())
			var vc float32 = 1
			var muted bool
			us, ok := m.user.Data.Setup.UsersSetup[ansi.Strip(v)]
			if ok {
				vc = us.VolumeCoefficient
				muted = us.Muted
			}
			cmds = append(cmds, m.connectionsList.InsertItem(itLenLen, connectionItem{nickname: v, muted: muted, volumeCoefficient: vc}))
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) coloredNickname(nickname string) string {
	for _, v := range m.connections {
		if nickname == ansi.Strip(v) {
			nickname = v
		}
	}
	return nickname
}

func (m Model) getMaxChatOffset() int {
	if m.width == 0 || m.height == 0 {
		return 0
	}

	borderStyle := styles.BorderStyle
	frameChromeW := lipgloss.Width(borderStyle.Render("X")) - 1
	frameChromeH := lipgloss.Height(borderStyle.Render("X")) - 1

	usableW := m.width - frameChromeW
	usableH := m.height - frameChromeH

	title := styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.BIG_LOGO)
	if usableW < 30 || usableH < 30 {
		borderStyle = styles.BorderStyle.PaddingTop(0)
		frameChromeH = lipgloss.Height(borderStyle.Render("X")) - 1
		title = styles.TitleStyle.Width(usableW).MaxWidth(usableW).Render(titles.LITTLE_LOGO)
	}

	titleH := lipgloss.Height(title)
	footerH := lipgloss.Height(styles.FooterStyle.Width(usableW).MaxWidth(usableW).Render("X"))

	contentChromeW := lipgloss.Width(styles.ContentStyle.Render("X")) - 1
	contentChromeH := lipgloss.Height(styles.ContentStyle.Render("X")) - 1

	centerW := usableW - contentChromeW
	centerH := usableH - titleH - footerH - contentChromeH

	baseBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	boxChromeW := lipgloss.Width(baseBox.Render(""))
	boxChromeH := lipgloss.Height(baseBox.Render("X")) - 1

	leftTotalW := centerW / 3
	innerRightW := (centerW - leftTotalW) - boxChromeW
	innerRightH := centerH - boxChromeH

	if innerRightW < 1 {
		innerRightW = 1
	}
	if innerRightH < 1 {
		innerRightH = 1
	}

	chatTitleH := lipgloss.Height(styles.HeaderStyle.Render("chat"))
	muteStateH := lipgloss.Height(m.muteState)

	activeConnectionsViewH := lipgloss.Height(lipgloss.NewStyle().Width(innerRightW - 2).MaxWidth(innerRightW - 2).Render("X"))
	inputViewH := lipgloss.Height(lipgloss.NewStyle().Width(innerRightW - 2).Render(m.chatTextInput.View()))

	usedSpaceH := chatTitleH + muteStateH + activeConnectionsViewH + inputViewH + 1
	historyMaxH := innerRightH - usedSpaceH
	if historyMaxH < 0 {
		historyMaxH = 0
	}

	var totalLines int
	msgStyle := lipgloss.NewStyle().Width(innerRightW - 2).MaxWidth(innerRightW - 2)

	for _, msg := range m.messages {
		timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Bold(true).Render(msg.Time)
		renderedMsg := msgStyle.Render(timeStr + "> " + msg.Nickname + ": " + msg.Text)
		totalLines += len(strings.Split(renderedMsg, "\n"))
	}

	maxOffset := totalLines - historyMaxH
	if maxOffset < 0 {
		maxOffset = 0
	}

	return maxOffset
}
