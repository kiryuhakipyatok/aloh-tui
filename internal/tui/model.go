package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"

	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/windows"
	"aloh-tui/pkg/logger"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	bz "github.com/lrstanley/bubblezone"
)

type Model struct {
	sizes
	tabs
	anims
	modelLists
	modelStates
	inputs
	chat
	colors
	chans
	curs
	boolStates
	datas
	other

	user *users.User
	log  *logger.Logger
	err  error
}

func NewModel(logFilePath, dataFilePath, keysPath string, appLogger *logger.Logger) (*Model, error) {
	op := "model.NewModel"
	log := appLogger.AddOp(op)
	log.Info("creating new model...")

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &Model{
		modelStates: modelStates{
			state:   states.START_STATE,
			prState: states.START_STATE,
		},

		curs: curs{
			curWindow: windows.START_WINDOW,
			curTime:   time.Now(),
		},

		tabs: tabs{
			defTabs:           []string{freindsTab, chatTab, voiceTab, videoTab, profileTab, settingsTab},
			regTabs:           []string{regTab, logTab},
			tabsNotifications: make(map[string]struct{}, 6),
		},

		anims: anims{
			logoAnim:    make([]string, 0, 4),
			notConnAnim: make([]string, 0, 4),
		},

		chans: chans{
			msgChan:    make(chan commands.ChatMessage, 100),
			rawMsgChan: make(chan commands.RawChatMessage, 100),

			peerConnectionsChan:    make(chan commands.PeerConnectedMsg, 100),
			peerDisconnectionsChan: make(chan commands.PeerDisconnectedMsg, 100),

			stopCountMinutesChan: make(chan struct{}, 1),

			sshEventsChan:  make(chan sshclient.Event, 50),
			netwEventsChan: make(chan commands.NetworkEventMsg, 50),
		},

		inputs: inputs{
			regTextInputs: make([]textinput.Model, 3),
			logingInput:   make([]textinput.Model, 2),

			appereanceInputs: make([]textinput.Model, 4),
			friendsInputs:    make([]textinput.Model, 5),

			passwordInputs: make([]textinput.Model, 3),
			nicknameInputs: make([]textinput.Model, 2),
		},

		chat: chat{
			messages: make([]commands.ChatMessage, 0, 20),
		},

		datas: datas{
			usersStates:       make(map[uuid.UUID]*userState, 5),
			online:            make(map[uuid.UUID][]users.Identity, 5),
			connections:       make([]users.Identity, 0, 3),
			connecctionsNicks: make([]string, 0, 3),
		},

		other: other{
			rightHeaderData: make([]string, 3),
			spinner:         sp,
			zone:            bz.New(),
		},

		colors: colors{
			usersColors: make(map[uuid.UUID]styles.UserColors, 5),
		},

		log: appLogger,
	}

	user := users.NewUser(logFilePath, keysPath, dataFilePath)

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

	m.user = user

	if m.user.Data.Identity.ID != uuid.Nil {
		m.prState = m.state
		m.state = states.LOAD_STATE
	}

	m.setupModel()
	//m.setup()

	log.Info("model created successfully")

	return m, nil
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, commands.AnimTickCmd(), commands.PulseTickCmd(), tea.EnableMouseCellMotion)
	if m.user.Data.Identity.ID != uuid.Nil {
		cmds = append(cmds, commands.AuthCmd(m.user, m.chans.sshEventsChan, m.log), m.spinner.Tick)
	}
	if m.user.GetShowTimeState() {
		cmds = append(cmds, commands.TimeTickCmd())
	}
	return tea.Batch(cmds...)
}

func (m *Model) Clean() {

	if m.user.Networking != nil {
		m.user.Networking.Close()
	}

	if m.user.Engines.AudioEngine != nil {
		m.user.Engines.AudioEngine.SetDisconnected()
		m.user.Engines.AudioEngine.Stop()
	}

	if m.user.SSHClient != nil {
		m.user.SSHClient.Close()
	}
}
