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
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

func (ci connectionItem) Title() string {
	return ci.nickname
}
func (ci connectionItem) Description() string {
	return fmt.Sprintf("volume: %.1f, muted: %t", ci.volumeCoefficient, ci.muted)
}
func (ci connectionItem) FilterValue() string {
	return ci.nickname
}

type micItem struct {
	name       string
	channels   uint32
	sampleRate uint32
	current    string
}

func (mi micItem) Title() string {
	return mi.name
}
func (mi micItem) Description() string {
	return fmt.Sprintf("channels: %d, sample rate: %d, %s", mi.channels, mi.sampleRate, mi.current)
}
func (mi micItem) FilterValue() string {
	return mi.name
}

type onlineItem struct {
	name string
}

func (oi onlineItem) Title() string {
	return oi.name
}
func (oi onlineItem) Description() string {
	return "solo"
}
func (oi onlineItem) FilterValue() string {
	return oi.name
}

const (
	HARD_DENOISE = iota
	SOFT_DENOISE
	AEC
	EQUALIZER
	AUDIO_N
	DESKTOP_N
)

type settingsItem struct {
	id          uint
	name        string
	description string
	enabled     bool
}

func (si settingsItem) Title() string {
	return si.name
}
func (si settingsItem) Description() string {
	return fmt.Sprintf("%s, state: %t", si.description, si.enabled)
}
func (si settingsItem) FilterValue() string {
	return si.name
}

type connData struct {
	nickname          string
	muted             bool
	volumeCoefficient float32
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

	usersColors map[string]lipgloss.Color
	connections []string
	online      []string

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

		usersColors: make(map[string]lipgloss.Color),

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

	m.setupMicrohonesList()
	m.setupConnestionsList()
	m.setupOnlineList()
	m.setupSettingsList()

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

func (m *Model) focusInputs() {
	m.unfocusInputs()
	ps := lipgloss.NewStyle().Foreground(lipgloss.Color(m.themeColor))
	switch m.state {
	case states.REG_STATE:
		m.regTextInputs[m.cursor].Focus()
		m.regTextInputs[m.cursor].PromptStyle = ps
	case states.CONN_STATE:
		if m.sideState == 1 {
			m.connTextInputs.Focus()
			m.connTextInputs.PromptStyle = ps
		}
	case states.CHAT_STATE:
		m.chatTextInput.Focus()
		m.chatTextInput.PromptStyle = ps
	case states.LOGIN_STATE:
		m.logingInput[m.cursor].Focus()
		m.logingInput[m.cursor].PromptStyle = ps
	case states.PROFILE_STATE:
		m.themeColorInput.Focus()
		m.themeColorInput.PromptStyle = ps
	}
}

func (m *Model) unfocusInputs() {
	for i := range m.regTextInputs {
		m.regTextInputs[i].Blur()
		m.regTextInputs[i].PromptStyle = lipgloss.NewStyle()
		m.regTextInputs[i].TextStyle = lipgloss.NewStyle()
	}

	m.connTextInputs.Blur()
	m.connTextInputs.PromptStyle = lipgloss.NewStyle()
	m.connTextInputs.TextStyle = lipgloss.NewStyle()

	for i := range m.logingInput {
		m.logingInput[i].Blur()
		m.logingInput[i].PromptStyle = lipgloss.NewStyle()
		m.logingInput[i].TextStyle = lipgloss.NewStyle()
	}

	m.chatTextInput.Blur()
	m.chatTextInput.PromptStyle = lipgloss.NewStyle()
	m.chatTextInput.TextStyle = lipgloss.NewStyle()

	m.themeColorInput.Blur()
	m.themeColorInput.PromptStyle = lipgloss.NewStyle()
	m.themeColorInput.TextStyle = lipgloss.NewStyle()
}

func (m *Model) setupMicrohonesList() {
	if m.user != nil && m.user.Engines.AudioEngine != nil {
		ms := m.user.Engines.AudioEngine.FetchMicrophones()

		microphones := make([]list.Item, len(ms))

		for i, v := range ms {
			mi := micItem{name: v.Name, channels: v.Channels, sampleRate: v.SampleRate, current: ""}

			if i == m.user.Engines.AudioEngine.GetCurrentMicrophone().Name {
				mi.current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
			}

			microphones[v.Index] = mi
		}

		delegate := list.NewDefaultDelegate()
		delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
		delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

		delegate.Styles.DimmedTitle = lipgloss.NewStyle().Foreground(cText)
		delegate.SetSpacing(1)
		m.microphonesDelegate = delegate
		m.microphonesList = list.New(microphones, delegate, m.width/2, m.height-4)
		m.microphonesList.DisableQuitKeybindings()
		m.microphonesList.Title = "select microphone"
		m.microphonesList.Select(-1)
		m.microphonesList.SetShowStatusBar(false)
		m.microphonesList.SetShowTitle(false)
		m.microphonesList.SetFilteringEnabled(false)
		m.microphonesList.SetShowFilter(false)
		m.microphonesList.SetShowHelp(false)
	}
}

func (m *Model) setupConnestionsList() {
	conns := make([]list.Item, 0)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	m.connectionsDelegate = delegate
	m.connectionsList = list.New(conns, delegate, m.width/2, m.height-4)
	m.connectionsList.DisableQuitKeybindings()
	m.connectionsList.Select(0)
	m.connectionsList.SetShowStatusBar(false)
	m.connectionsList.SetShowTitle(false)
	m.connectionsList.SetFilteringEnabled(false)
	m.connectionsList.SetShowFilter(false)
	m.connectionsList.SetShowHelp(false)
}

func (m *Model) setupOnlineList() {
	online := make([]list.Item, 0)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Foreground(cText)
	delegate.SetSpacing(0)
	m.onlineDelegate = delegate
	m.onlineList = list.New(online, delegate, m.width/2, m.height-4)
	m.onlineList.DisableQuitKeybindings()
	m.onlineList.SetShowStatusBar(false)
	m.onlineList.SetShowTitle(false)
	m.onlineList.SetFilteringEnabled(false)
	m.onlineList.SetShowFilter(false)
	m.onlineList.SetShowHelp(false)
}

func (m *Model) setupSettingsList() {
	settings := []list.Item{
		settingsItem{
			id:          HARD_DENOISE,
			name:        "hard denoise",
			description: "reduce noise hard",
			enabled:     m.user.Data.Setup.HardDenoise,
		},
		settingsItem{
			id:          SOFT_DENOISE,
			name:        "soft denoise",
			description: "reduce noise soft",
			enabled:     m.user.Data.Setup.SoftDenoise,
		},
		settingsItem{
			id:          AEC,
			name:        "echocanceller",
			description: "reduce echo",
			enabled:     m.user.Data.Setup.AEC,
		},
		settingsItem{
			id:          EQUALIZER,
			name:        "equalizer",
			description: "reduce low freqs and increase high",
			enabled:     m.user.Data.Setup.Filter,
		},
		settingsItem{
			id:          AUDIO_N,
			name:        "audio notifications",
			description: "puck puck",
			enabled:     m.user.Data.Setup.AudioNotifications,
		},
		settingsItem{
			id:          DESKTOP_N,
			name:        "desktop notification",
			description: "kvadrat zadral",
			enabled:     m.user.Data.Setup.DesktopNotifications,
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Foreground(cText)
	delegate.SetSpacing(1)
	m.settingsDelegate = delegate
	m.settingsList = list.New(settings, delegate, m.width/2, m.height-4)
	m.settingsList.Select(0)
	m.settingsList.DisableQuitKeybindings()
	m.settingsList.SetShowStatusBar(false)
	m.settingsList.SetShowTitle(false)
	m.settingsList.SetFilteringEnabled(false)
	m.settingsList.SetShowFilter(false)
	m.settingsList.SetShowHelp(false)
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

func (m *Model) updateMicrophonesItemList(microphone string) tea.Cmd {
	var cmd tea.Cmd
	items := m.microphonesList.Items()
	for i, v := range items {
		mic, ok := v.(micItem)
		if !ok {
			continue
		}
		mic.current = ""
		if mic.name == microphone {
			mic.current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
		}
		cmd = m.microphonesList.SetItem(i, mic)
	}
	return cmd
}

func (m *Model) updateSettingsItemList(settingId uint) tea.Cmd {
	var cmd tea.Cmd
	items := m.settingsList.Items()
	for i, v := range items {
		s, ok := v.(settingsItem)
		if !ok {
			continue
		}
		if s.id == settingId {
			s.enabled = !s.enabled
			cmd = m.settingsList.SetItem(i, s)
			return cmd
		}
	}
	return cmd
}

func (m *Model) updateMicrophonesList() tea.Cmd {
	var cmds []tea.Cmd
	items := m.microphonesList.Items()
	ms := m.user.Engines.AudioEngine.FetchMicrophones()
	micsMap := make(map[string]struct{}, len(ms))
	for i := range ms {
		micsMap[i] = struct{}{}
	}

	itemsMap := make(map[string]struct{}, len(items))
	for _, v := range items {
		item, ok := v.(micItem)
		if !ok {
			continue
		}
		itemsMap[item.name] = struct{}{}
	}

	for i := len(items) - 1; i >= 0; i-- {
		item, ok := items[i].(micItem)
		if !ok {
			continue
		}
		_, ok = micsMap[item.name]
		if !ok {
			m.microphonesList.RemoveItem(i)
		}
	}

	for i, v := range ms {
		if _, ok := itemsMap[i]; !ok {
			itLenLen := len(m.microphonesList.Items())
			cmds = append(cmds, m.microphonesList.InsertItem(itLenLen, micItem{name: i, sampleRate: v.SampleRate, channels: v.Channels}))
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) updateOnlineList() tea.Cmd {
	var cmds []tea.Cmd
	items := m.onlineList.Items()
	onlineMap := make(map[string]struct{}, len(m.online))
	for _, v := range m.online {
		onlineMap[v] = struct{}{}
	}

	itemsMap := make(map[string]struct{}, len(items))
	for _, v := range items {
		item, ok := v.(onlineItem)
		if !ok {
			continue
		}
		itemsMap[item.name] = struct{}{}
	}

	for i := len(items) - 1; i >= 0; i-- {
		item, ok := items[i].(onlineItem)
		if !ok {
			continue
		}
		_, ok = onlineMap[item.name]
		if !ok {
			m.onlineList.RemoveItem(i)
		}
	}

	for _, v := range m.online {
		if _, ok := itemsMap[v]; !ok {
			itLenLen := len(m.onlineList.Items())
			cmds = append(cmds, m.onlineList.InsertItem(itLenLen, onlineItem{name: v}))
		}
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

func (m Model) getChatSizes() (int, int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0, 0
	}

	padW, padH := 2, 1
	usableW := m.width - (padW * 2)
	usableH := m.height - (padH * 2)

	footerH := lipgloss.Height("X")
	logo := lipgloss.NewStyle().Render(titles.A)
	logoH := lipgloss.Height(logo)

	gridH := usableH - footerH - logoH
	activeBorder := tabBorderWithBottom("┘", " ", "└")

	activeTabStyle := lipgloss.NewStyle().
		Border(activeBorder, true).
		Bold(true).
		Align(lipgloss.Center)

	tab := m.defTabs[0]
	style := activeTabStyle
	border, _, _, _, _ := style.GetBorder()
	border.BottomLeft = "│"
	border.BottomRight = "│"
	styledT := style.Border(border).Render(tab)
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, styledT)
	w := usableW - 4
	h := gridH - lipgloss.Height(tabsRow) - 2

	m.chatTextInput.Width = max(1, w-4)
	inputView := lipgloss.NewStyle().PaddingLeft(2).Render(m.chatTextInput.View())

	rawConnStr := "X"
	connsView := lipgloss.NewStyle().PaddingLeft(2).Render(safeTruncate(rawConnStr, w-2))

	usersAudioState := "X"

	var chatParts []string

	chatParts = append(chatParts, usersAudioState, connsView, "")

	usedH := 0
	for _, p := range chatParts {
		usedH += lipgloss.Height(p)
	}
	usedH += lipgloss.Height(inputView) + 1

	historyMaxH := max(0, h-usedH)

	var maxOffset int
	if historyMaxH > 0 {
		var allMsgsLines []string
		msgStyle := lipgloss.NewStyle().Width(w - 2).MaxWidth(w - 2).PaddingLeft(2)

		for _, msg := range m.messages {
			t := lipgloss.NewStyle().Render(msg.Time)
			n := lipgloss.NewStyle().Bold(true).Render(msg.Nickname + ":")
			txt := lipgloss.NewStyle().Render(msg.Text)
			renderedMsg := msgStyle.Render(fmt.Sprintf("%s %s %s", t, n, txt))
			allMsgsLines = append(allMsgsLines, strings.Split(renderedMsg, "\n")...)
		}

		lenAll := len(allMsgsLines)

		maxOffset = lenAll - historyMaxH
	}

	return maxOffset, w, historyMaxH
}
