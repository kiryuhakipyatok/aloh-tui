package tui

import (
	"aloh-tui/internal/tui/components/lists"
	"cmp"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type List struct {
	lipList     list.Model
	lipDelegate lists.DynamicDelegate
}

func (m *Model) setupMicrohonesList() {
	if m.user != nil && m.user.Engines.AudioEngine != nil {
		ms := m.user.Engines.AudioEngine.FetchMicrophones()
		mics := make([]string, 0, len(ms))
		for mic := range ms {
			mics = append(mics, mic)
		}

		slices.SortFunc(mics, func(a, b string) int {
			return cmp.Compare(ms[a].Index, ms[b].Index)
		})
		curMic := m.user.Engines.AudioEngine.GetCurrentMicrophone().Name
		microphones := make([]list.Item, len(ms))
		for _, mic := range mics {
			v := ms[mic]
			mi := lists.MicItem{Name: v.Name, Channels: v.Channels, SampleRate: v.SampleRate, Format: v.Format}

			if mic == curMic {
				mi.Current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
			}

			microphones[v.Index] = mi
		}

		delegate := list.NewDefaultDelegate()
		delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
		delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

		delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
		delegate.SetSpacing(1)

		dd := lists.DynamicDelegate{DefaultDelegate: delegate}
		l := list.New(microphones, dd, m.width/2, m.height-4)

		l.DisableQuitKeybindings()
		l.Title = "select microphone"
		l.Select(0)
		l.SetShowStatusBar(false)
		l.SetShowTitle(false)
		l.SetFilteringEnabled(false)
		l.SetShowFilter(false)
		l.SetShowHelp(false)

		m.microphonesList = List{
			lipList:     l,
			lipDelegate: dd,
		}
	}
}

func (m *Model) setupConnestionsList() {
	conns := make([]list.Item, 0)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle()
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
	delegate.ShowDescription = true
	delegate.SetSpacing(1)

	dd := lists.DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(conns, dd, m.width/2, m.height-4)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	m.connectionsList = List{
		lipList:     l,
		lipDelegate: dd,
	}
}

func (m *Model) setupOnlineList() {
	online := make([]list.Item, len(m.online))

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
	delegate.SetSpacing(1)

	dd := lists.DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(online, dd, m.width/2, m.height-4)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	m.onlineList = List{
		lipList:     l,
		lipDelegate: dd,
	}
}

const (
	HARD_DENOISE = iota
	SOFT_DENOISE
	AEC
	EQUALIZER
	AUDIO_N
	DESKTOP_N
	APP_N
	TIME
	DATE
	ZONE
)

func (m *Model) setupSettingsList() {
	settings := []list.Item{
		lists.SettingsItem{
			Id:      HARD_DENOISE,
			Name:    "hard denoise",
			Desc:    "reduce noise hard",
			Enabled: m.user.Data.Setup.Audio.HardDenoise,
		},
		lists.SettingsItem{
			Id:      SOFT_DENOISE,
			Name:    "soft denoise",
			Desc:    "reduce noise soft",
			Enabled: m.user.Data.Setup.Audio.SoftDenoise,
		},
		lists.SettingsItem{
			Id:      AEC,
			Name:    "echocanceller",
			Desc:    "reduce echo",
			Enabled: m.user.Data.Setup.Audio.AEC,
		},
		lists.SettingsItem{
			Id:      EQUALIZER,
			Name:    "equalizer",
			Desc:    "reduce low freqs and increase high",
			Enabled: m.user.Data.Setup.Audio.Filter,
		},
		lists.SettingsItem{
			Id:      APP_N,
			Name:    "app notifications",
			Desc:    "notifications in app intreface",
			Enabled: m.user.Data.Setup.Notifications.AppNotifications,
		},
		lists.SettingsItem{
			Id:      AUDIO_N,
			Name:    "audio notifications",
			Desc:    "notifications with sound",
			Enabled: m.user.Data.Setup.Notifications.AudioNotifications,
		},
		lists.SettingsItem{
			Id:      DESKTOP_N,
			Name:    "desktop notifications",
			Desc:    "notifications with desktop notifying",
			Enabled: m.user.Data.Setup.Notifications.DesktopNotifications,
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
	delegate.SetSpacing(1)

	dd := lists.DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(settings, dd, m.width/2, m.height-4)

	l.Select(0)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	m.settingsList = List{
		lipList:     l,
		lipDelegate: dd,
	}
}

func (m *Model) setupFriendsReqsList() {
	fReqs := m.user.GetFriendsReqs()
	friendsReq := make([]list.Item, 0, len(fReqs))

	for _, f := range fReqs {
		friendsReq = append(friendsReq, lists.FriendReqItem{
			Nickname: f.Nickname,
			ReqTime:  f.ReqTime.Local().Format("2006-01-02"),
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
	delegate.SetSpacing(0)

	dd := lists.DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(friendsReq, dd, m.width/2, m.height-4)
	l.Select(0)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	m.friendsReqsList = List{
		lipList:     l,
		lipDelegate: dd,
	}
}

func (m *Model) setupApearenceList() {
	apearences := []list.Item{
		lists.ApearenceItem{
			Id:      TIME,
			Name:    "show time",
			Desc:    "current time in top-right corner",
			Enabled: m.user.Data.Setup.Appereance.ShowTime,
		},
		lists.ApearenceItem{
			Id:      DATE,
			Name:    "show date",
			Desc:    "current date in top-right corner",
			Enabled: m.user.Data.Setup.Appereance.ShowDate,
		},
		lists.ApearenceItem{
			Id:      ZONE,
			Name:    "show zone",
			Desc:    "current zone in top-right corner",
			Enabled: m.user.Data.Setup.Appereance.ShowZone,
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(cDim)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(cGray)
	delegate.SetSpacing(1)
	dd := lists.DynamicDelegate{DefaultDelegate: delegate}
	l := list.New(apearences, dd, m.width/2, m.height-4)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	//m.apearenceDelegate = delegate
	m.apearenceList = List{
		lipList:     l,
		lipDelegate: dd,
	}
	m.setListVisible(&m.apearenceList, false)
	//m.apearenceList.Select(-1)

}

func (m *Model) setListVisible(l *List, vis bool) {
	l.lipList.Select(0)
	if vis {
		l.lipDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(m.themeColor)
		l.lipDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(m.subThemeColor)
	} else {
		l.lipDelegate.Styles.SelectedTitle = l.lipDelegate.Styles.NormalTitle
		l.lipDelegate.Styles.SelectedDesc = l.lipDelegate.Styles.NormalDesc
	}

	l.lipList.SetDelegate(l.lipDelegate)
}

func (m *Model) updateFriendsReqList() tea.Cmd {
	fReqs := m.user.GetFriendsReqs()
	names := make([]string, 0, len(fReqs))
	for _, f := range fReqs {
		names = append(names, f.Nickname)
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))

	for _, name := range names {
		t := time.Now().Format("2006-01-02")
		newItems = append(newItems, lists.FriendReqItem{
			Nickname: name,
			ReqTime:  t,
		})
	}

	return m.friendsReqsList.lipList.SetItems(newItems)
}

func (m *Model) updateConnectionItemList(nickname string, volume float32, muted bool) tea.Cmd {
	var cmds []tea.Cmd
	items := m.connectionsList.lipList.Items()
	for i, v := range items {
		conn, ok := v.(lists.ConnectionItem)
		if !ok {
			continue
		}
		if conn.Nickname == nickname {
			conn.VolumeCoefficient = volume
			conn.Muted = muted
		}

		cmds = append(cmds, m.connectionsList.lipList.SetItem(i, conn))
	}
	return tea.Batch(cmds...)
}

func (m *Model) updateMicrophonesItemList(microphone string) tea.Cmd {
	var cmd tea.Cmd
	items := m.microphonesList.lipList.Items()
	for i, v := range items {
		mic, ok := v.(lists.MicItem)
		if !ok {
			continue
		}
		mic.Current = ""
		if mic.Name == microphone {
			mic.Current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
		}
		cmd = m.microphonesList.lipList.SetItem(i, mic)
	}
	return cmd
}

func (m *Model) updateSettingsItemList(settingId uint) tea.Cmd {
	var cmd tea.Cmd
	items := m.settingsList.lipList.Items()
	for i, v := range items {
		s, ok := v.(lists.SettingsItem)
		if !ok {
			continue
		}
		if s.Id == settingId {
			s.Enabled = !s.Enabled
			cmd = m.settingsList.lipList.SetItem(i, s)
			return cmd
		}
	}
	return cmd
}

func (m *Model) updateApearenceItemList(apearenceId uint) tea.Cmd {
	var cmd tea.Cmd
	items := m.apearenceList.lipList.Items()
	for i, v := range items {
		s, ok := v.(lists.ApearenceItem)
		if !ok {
			continue
		}
		if s.Id == apearenceId {
			s.Enabled = !s.Enabled
			cmd = m.apearenceList.lipList.SetItem(i, s)
			return cmd
		}
	}
	return cmd
}

func (m *Model) updateMicrophonesList() tea.Cmd {
	ms := m.user.Engines.AudioEngine.FetchMicrophones()
	mics := make([]string, 0, len(ms))
	for mic := range ms {
		mics = append(mics, mic)
	}

	slices.SortFunc(mics, func(a, b string) int {
		return cmp.Compare(ms[a].Index, ms[b].Index)
	})

	newItems := make([]list.Item, 0, len(mics))
	curAudioMic := m.user.Engines.AudioEngine.GetCurrentMicrophone().Name
	curUserMic := m.user.Data.Devices.Microphone
	var currFind bool
	for _, mic := range mics {
		v := ms[mic]
		mi := lists.MicItem{
			Name:       mic,
			Channels:   v.Channels,
			SampleRate: v.SampleRate,
			Format:     v.Format,
		}
		if mic == curAudioMic && mic == curUserMic {
			mi.Current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
			currFind = true
		}
		newItems = append(newItems, mi)
	}
	if !currFind {
		m.user.Data.Devices.Microphone = ""
	}
	return m.microphonesList.lipList.SetItems(newItems)
}

func (m *Model) updateOnlineList() tea.Cmd {
	names := make([]string, 0, len(m.online))
	for name := range m.online {
		names = append(names, name)
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))

	for _, name := range names {
		var bf string
		if m.user.Data.Statistics.BestFriend.Nickname == name {
			bf = m.user.Data.Setup.Appereance.BestFriendTag
		}
		newItems = append(newItems, lists.OnlineItem{
			Name:        name,
			Connections: m.online[name],
			BFTag:       bf,
		})
	}

	return m.onlineList.lipList.SetItems(newItems)
}

func (m *Model) updateConnectionsList() tea.Cmd {
	names := make([]string, 0, len(m.connections))
	for _, name := range m.connections {
		names = append(names, name)
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))
	for _, name := range names {
		var vc float32 = 1
		var muted bool
		clearName := ansi.Strip(name)
		us, ok := m.user.Data.Setup.Audio.UsersSetup[clearName]
		if ok {
			vc = us.VolumeCoefficient
			muted = us.Muted
		}
		var bf string
		if m.user.Data.Statistics.BestFriend.Nickname == clearName {
			bf = m.user.Data.Setup.Appereance.BestFriendTag
		}
		newItems = append(newItems, lists.ConnectionItem{
			Nickname:          name,
			VolumeCoefficient: vc,
			Muted:             muted,
			BFTag:             bf,
		})
	}

	return m.connectionsList.lipList.SetItems(newItems)
}
