package tui

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/tui/components/lists"
	"aloh-tui/internal/tui/components/states"
	"aloh-tui/internal/tui/components/styles"
	"aloh-tui/internal/tui/components/titles"
	"aloh-tui/pkg/errs"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

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
	usedH += lipgloss.Height(inputView)

	historyMaxH := max(0, h-usedH)

	var maxOffset int
	if historyMaxH > 0 {
		var allMsgsLines []string
		msgStyle := lipgloss.NewStyle().Width(w - 2).MaxWidth(w - 2).PaddingLeft(2)

		for _, msg := range m.messages {
			t := lipgloss.NewStyle().Render(msg.Time)
			n := lipgloss.NewStyle().Bold(true).Render(msg.Identity.Nickname + ":")
			txt := lipgloss.NewStyle().Render(msg.Text)
			renderedMsg := msgStyle.Render(fmt.Sprintf("%s %s %s", t, n, txt))
			allMsgsLines = append(allMsgsLines, strings.Split(renderedMsg, "\n")...)
		}

		lenAll := len(allMsgsLines)

		maxOffset = lenAll - historyMaxH - 1
	}

	return maxOffset, w, historyMaxH
}

// func (m Model) coloredNickname(nickname string) string {
// 	for _, v := range m.connections {
// 		if nickname == ansi.Strip(v) {
// 			nickname = v
// 		}
// 	}
// 	return nickname
// }

func (m Model) getOfflineUsers() []uuid.UUID {
	onlineMap := make(map[uuid.UUID]struct{})
	for o := range m.online {
		onlineMap[o] = struct{}{}
	}

	friends := m.user.GetFriends()

	offline := make([]uuid.UUID, 0, len(friends))

	for _, f := range friends {
		if _, ok := onlineMap[f.ID]; !ok {
			offline = append(offline, f.ID)
		}
	}

	return offline
}

func isNewInOnline(newOnline, oldOnline map[string][]string) bool {
	for n := range newOnline {
		if _, ok := oldOnline[n]; !ok {
			return true
		}
	}
	return false
}

func isEqualOnline(newOnline, oldOnline map[string][]string) bool {
	return maps.EqualFunc(newOnline, oldOnline, func(arr1 []string, arr2 []string) bool {
		slices.Sort(arr1)
		slices.Sort(arr2)
		return slices.Equal(arr1, arr2)
	})
}

func isInConnections(conns []users.Identity, id uuid.UUID) bool {
	return slices.ContainsFunc(conns, func(connIden users.Identity) bool {
		return id == connIden.ID
	})
}

func getIdentityInConnections(conns []users.Identity, id uuid.UUID) (users.Identity, error) {
	var iden users.Identity
	for _, c := range conns {
		if c.ID == id {
			iden = c
		}
	}
	if iden.ID == uuid.Nil {
		return iden, errs.ErrNotFound()
	}
	return iden, nil
}

func (m *Model) unfocusLists() {
	lists.SetListVisible(&m.apearenceList.DefList, false)
	lists.SetListVisible(&m.connectionsList.DefList, false)
	lists.SetListVisible(&m.friendsReqsList.DefList, false)
	lists.SetListVisible(&m.settingsList.DefList, false)
	lists.SetListVisible(&m.friendsList.DefList, false)
	lists.SetListVisible(&m.microphonesList.DefList, false)
	lists.SetListVisible(&m.headphonesList.DefList, false)
}

func (m Model) selectSetting() (Model, tea.Cmd) {
	if i, ok := m.settingsList.LipList.SelectedItem().(lists.SettingsItem); ok {
		m.prState = m.state
		m.state = states.LOAD_STATE
		switch i.Id {
		case lists.DEVICES_SETTINGS:
			m.state = states.DEVICES_STATE
		case lists.AUDIO_SETTINGS:
			m.state = states.AUDIO_STATE
			m.sideState = states.RIGHT_STATE
		case lists.NOTIFICATIONS_SETTINGS:
			m.state = states.NOTIFICATIONS_STATE
			m.sideState = states.RIGHT_STATE
		case lists.BINDS_SETTINGS:
			m.state = states.BINDS_STATE
		case lists.ACCOUNT_SETTINGS:
			m.state = states.ACCOUNT_STATE
		default:
			return m, nil
		}
		return m.syncTabState()
	}

	return m, nil
}

func (m Model) selectAccountSetting() (Model, tea.Cmd) {
	if i, ok := m.accountList.LipList.SelectedItem().(lists.AccountItem); ok {
		m.prState = m.state
		m.state = states.LOAD_STATE
		switch i.Id {
		case lists.NICKNAME_SETTINGS:
			m.state = states.NICKNAME_STATE
		case lists.PASSWORD_SETTINGS:
			m.state = states.PASSWORD_STATE
		case lists.TAGLINE_SETTINGS:
			m.state = states.TAGLINE_STATE
		case lists.COLOR_SETTINGS:
			m.state = states.COLOR_STATE
		default:
			return m, nil
		}
		m.log.Info("states", m.prState, m.state)
		return m.syncTabState()
	}
	return m, nil
}

func (m Model) selectDevicesSetting() (Model, tea.Cmd) {
	if i, ok := m.devicesList.LipList.SelectedItem().(lists.DeviceTypeItem); ok {
		m.prState = m.state
		m.state = states.LOAD_STATE
		switch i.Id {
		case lists.HEADPHONES_SETTINGS:
			m.state = states.HEADPHONES_SET_STATE
		case lists.WEBCAM_SETTINGS:
			m.state = states.WEBCAM_SET_STATE
		case lists.MICROPHONE_SETTINGS:
			m.state = states.MICROPHONE_SET_STATE
		default:
			return m, nil
		}
		m.log.Info("states", m.prState, m.state)
		return m.syncTabState()
	}
	return m, nil
}

func cloneMap(original map[uuid.UUID][]users.Identity) map[uuid.UUID][]users.Identity {
	cp := make(map[uuid.UUID][]users.Identity, len(original))
	for k, v := range original {
		cp[k] = v
	}
	return cp
}

func (m Model) inCurrentWindow(msg tea.MouseMsg) bool {
	return m.zone.Get("registerW").InBounds(msg) || m.zone.Get("videoW").InBounds(msg) ||
		m.zone.Get("friendsW").InBounds(msg) || m.zone.Get("chatW").InBounds(msg) ||
		m.zone.Get("voiceW").InBounds(msg) || m.zone.Get("videoW").InBounds(msg) ||
		m.zone.Get("profileW").InBounds(msg) || m.zone.Get("settingsW").InBounds(msg)
}

func (m Model) Err(err error) (Model, tea.Cmd) {
	m.err = err
	m.prState = m.state
	m.state = states.ERR_STATE
	return m, nil
}

func vertLine(h int) string {
	if h <= 0 {
		return ""
	}
	return strings.Repeat("│\n", h-1) + "│"
}
func horizLine(w int) string {
	if w <= 0 {
		return ""
	}
	return strings.Repeat("—", w-1) + "—"
}

func safeTruncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	cleanStr := ansi.Strip(s)
	runes := []rune(cleanStr)
	if len(runes) > maxW {
		return styles.CErrStyle.Render(string(runes[:maxW-2]) + "..")
	}
	return s
}

func (m *Model) getOnlineIdentity(online map[uuid.UUID][]users.Identity, id uuid.UUID) (users.Identity, error) {
	var (
		iden users.Identity
		wg   sync.WaitGroup
	)

	stop := make(chan struct{}, 1)

	for _, conns := range online {
		wg.Go(func() {
			m.log.Info("conns", conns)
			for _, c := range conns {
				select {
				case <-stop:
					return
				default:
					m.log.Info("c", c)
					if c.ID == id {
						iden = c
						stop <- struct{}{}
						return
					}
				}
			}
		})
	}

	wg.Wait()

	if iden.ID == uuid.Nil {
		return iden, errs.ErrNotFound()
	}
	m.log.Info("iden in get", iden)
	return iden, nil

}

func updateNicknameInConn(conns []users.Identity, id uuid.UUID, newNickname string) {
	for i, c := range conns {
		if c.ID == id {
			conns[i].Nickname = newNickname
			break
		}
	}
}

func isEmptyString(s string) (string, bool) {
	trimmed := strings.TrimSpace(s)
	if len(strings.TrimSpace(s)) == 0 {
		return "", true
	}
	return trimmed, false
}

func validatePassword(s string) error {
	entropy := passwordvalidator.GetEntropy(s)
	if err := passwordvalidator.Validate(s, 60); err != nil {
		return errs.ErrInvalidPassword(entropy)
	}
	return nil
}

func (m Model) onUsersWebcam(msg tea.MouseMsg) (users.Identity, bool) {
	var (
		id     uuid.UUID
		zoneId string
	)

	userIden := m.user.GetUserIdentity()
	zoneId = fmt.Sprintf("webcam-%s", userIden.ID.String())
	if m.zone.Get(zoneId).InBounds(msg) {
		return userIden, true
	}

	userFrames := m.user.Engines.VideoEngine.GetUsersWebcamFramesTerminal()

	for i := range userFrames {

		zoneId = fmt.Sprintf("webcam-%s", i.String())

		if m.zone.Get(zoneId).InBounds(msg) {
			id = i
		}

	}

	iden, err := m.user.GetFriendIdentityById(id)
	if err != nil {
		return iden, false
	}

	return iden, true
}

func (m Model) onUsersScreen(msg tea.MouseMsg) (users.Identity, bool) {
	var (
		id     uuid.UUID
		zoneId string
	)

	userIden := m.user.GetUserIdentity()
	zoneId = fmt.Sprintf("screen-%s", userIden.ID.String())
	if m.zone.Get(zoneId).InBounds(msg) {
		return userIden, true
	}

	userFrames := m.user.Engines.VideoEngine.GetUsersScreenFramesTerminal()

	for i := range userFrames {

		zoneId = fmt.Sprintf("screen-%s", i.String())

		if m.zone.Get(zoneId).InBounds(msg) {
			id = i
		}

	}

	iden, err := m.user.GetFriendIdentityById(id)
	if err != nil {
		return iden, false
	}

	return iden, true
}

func (m Model) renderVideoFrames(frames []string) string {
	lenF := len(frames)

	if lenF <= 0 {
		return ""
	}

	div := lenF / 2

	if !((lenF-div)%2 == 0) {
		div--
	}

	upFrames := make([]string, 0, lenF/2)
	downFrames := make([]string, 0, (lenF/2)+1)

	if div < 2 {
		return lipgloss.JoinHorizontal(lipgloss.Center, frames...)
	} else {
		for i, f := range frames {
			if i < div {
				upFrames = append(upFrames, f)
				continue
			}
			if i > div {
				downFrames = append(downFrames, f)
			}

		}
	}

	upJoined := lipgloss.JoinHorizontal(lipgloss.Center, upFrames...)
	downJoined := lipgloss.JoinHorizontal(lipgloss.Center, downFrames...)

	full := lipgloss.JoinVertical(lipgloss.Center, upJoined, downJoined)

	return full
}
