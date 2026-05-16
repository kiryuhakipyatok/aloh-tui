package tui

import (
	"aloh-tui/internal/tui/components/lists"
	"cmp"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

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
	names := make([]string, 0, len(m.online))
	slices.Sort(names)
	online := make([]list.Item, len(m.online))
	for _, name := range names {
		online = append(online, lists.OnlineItem{
			Name:        name,
			Connections: m.online[name],
		})
	}

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

const (
	HARD_DENOISE = iota
	SOFT_DENOISE
	AEC
	EQUALIZER
	AUDIO_N
	DESKTOP_N
)

func (m *Model) setupSettingsList() {
	settings := []list.Item{
		lists.SettingsItem{
			Id:      HARD_DENOISE,
			Name:    "hard denoise",
			Desc:    "reduce noise hard",
			Enabled: m.user.Data.Setup.HardDenoise,
		},
		lists.SettingsItem{
			Id:      SOFT_DENOISE,
			Name:    "soft denoise",
			Desc:    "reduce noise soft",
			Enabled: m.user.Data.Setup.SoftDenoise,
		},
		lists.SettingsItem{
			Id:      AEC,
			Name:    "echocanceller",
			Desc:    "reduce echo",
			Enabled: m.user.Data.Setup.AEC,
		},
		lists.SettingsItem{
			Id:      EQUALIZER,
			Name:    "equalizer",
			Desc:    "reduce low freqs and increase high",
			Enabled: m.user.Data.Setup.Filter,
		},
		lists.SettingsItem{
			Id:      AUDIO_N,
			Name:    "audio notifications",
			Desc:    "notifications with sound",
			Enabled: m.user.Data.Setup.AudioNotifications,
		},
		lists.SettingsItem{
			Id:      DESKTOP_N,
			Name:    "desktop notifications",
			Desc:    "notifications with desktop notifying",
			Enabled: m.user.Data.Setup.DesktopNotifications,
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
		conn, ok := v.(lists.ConnectionItem)
		if !ok {
			continue
		}
		if conn.Nickname == nickname {
			conn.VolumeCoefficient = volume
			conn.Muted = muted
		}

		cmds = append(cmds, m.connectionsList.SetItem(i, conn))
	}
	return tea.Batch(cmds...)
}

func (m *Model) updateMicrophonesItemList(microphone string) tea.Cmd {
	var cmd tea.Cmd
	items := m.microphonesList.Items()
	for i, v := range items {
		mic, ok := v.(lists.MicItem)
		if !ok {
			continue
		}
		mic.Current = ""
		if mic.Name == microphone {
			mic.Current = lipgloss.NewStyle().Foreground(m.themeColor).Render("CURRENT")
		}
		cmd = m.microphonesList.SetItem(i, mic)
	}
	return cmd
}

func (m *Model) updateSettingsItemList(settingId uint) tea.Cmd {
	var cmd tea.Cmd
	items := m.settingsList.Items()
	for i, v := range items {
		s, ok := v.(lists.SettingsItem)
		if !ok {
			continue
		}
		if s.Id == settingId {
			s.Enabled = !s.Enabled
			cmd = m.settingsList.SetItem(i, s)
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
	return m.microphonesList.SetItems(newItems)
}

func (m *Model) updateOnlineList() tea.Cmd {
	names := make([]string, 0, len(m.online))
	for name := range m.online {
		names = append(names, name)
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))
	for _, name := range names {
		newItems = append(newItems, lists.OnlineItem{
			Name:        name,
			Connections: m.online[name],
		})
	}

	return m.onlineList.SetItems(newItems)
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
		us, ok := m.user.Data.Setup.UsersSetup[ansi.Strip(name)]
		if ok {
			vc = us.VolumeCoefficient
			muted = us.Muted
		}
		newItems = append(newItems, lists.ConnectionItem{
			Nickname:          name,
			VolumeCoefficient: vc,
			Muted:             muted,
		})
	}

	return m.connectionsList.SetItems(newItems)
}
