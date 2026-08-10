package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/commands"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/styles"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	bz "github.com/lrstanley/bubblezone"
)

type userState struct {
	fullMute     bool
	micMute      bool
	hardDenoised bool
	softDenoised bool
	webcam       bool
}

type sizes struct {
	width  int
	height int
}

type tabs struct {
	defTabs []string
	regTabs []string

	tabsNotifications map[string]struct{}
}

type anims struct {
	logoAnim    []string
	notConnAnim []string
	aloneAnim   []string

	animFrame  int
	pulseFrame int
}

type modelStates struct {
	state     uint
	prState   uint
	sideState uint
}

type inputs struct {
	regTextInputs    []textinput.Model
	chatTextInput    textinput.Model
	logingInput      []textinput.Model
	friendsInputs    []textinput.Model
	appereanceInputs []textinput.Model
	nicknameInputs   []textinput.Model
	taglineInput     textinput.Model
	colorInput       textinput.Model
	passwordInputs   []textinput.Model
}

type chat struct {
	chatOffset int
	chatWidth  int
	chatHeight int

	messages []commands.ChatMessage
}

type modelLists struct {
	connectionsList lists.ConnectionsList

	friendsList lists.FriendsList

	settingsList      lists.SettingsList
	notificationsList lists.SwitcherList
	audioList         lists.SwitcherList
	microphonesList   lists.DeviceList
	headphonesList    lists.DeviceList
	apearenceList     lists.SwitcherList
	friendsReqsList   lists.FriendsReqsList

	accountList lists.SettingsList
}

type colors struct {
	themeColor    lipgloss.Color
	subThemeColor lipgloss.Color

	userColor   lipgloss.Color
	usersColors map[uuid.UUID]styles.UserColors
}

type chans struct {
	msgChan    chan commands.ChatMessage
	rawMsgChan chan commands.RawChatMessage

	peerConnectionsChan    chan commands.PeerConnectedMsg
	peerDisconnectionsChan chan commands.PeerDisconnectedMsg

	stopCountMinutesChan chan struct{}

	sshEventsChan  chan sshclient.Event
	netwEventsChan chan commands.NetworkEventMsg
}

type curs struct {
	activeTab int
	cursor    int

	curTime time.Time

	curWindow uint
}

type boolStates struct {
	ticked    bool
	connected bool
}

type other struct {
	spinner spinner.Model

	rightHeaderData []string

	imageBuffer []byte

	headerActiveStyle lipgloss.Style

	zone *bz.Manager

	//webcamUsersFrames map[uuid.UUID]userVideoFrame
	userWebcamFrame   string
}

type datas struct {
	connections       []users.Identity
	connecctionsNicks []string
	online            map[uuid.UUID][]users.Identity
	usersStates       map[uuid.UUID]*userState
}
