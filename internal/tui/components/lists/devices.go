package lists

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/audio"
	"cmp"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MICROPHONE = iota
	HEADPHONES
)

type DeviceItem struct {
	Name       string
	Channels   uint32
	SampleRate uint32
	Format     string
	Current    string
}

func (di DeviceItem) Title() string {
	return di.Name
}
func (di DeviceItem) Description() string {
	return di.DynamicDescription(false)
}
func (di DeviceItem) FilterValue() string {
	return di.Name
}

func (di DeviceItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("channels: %d, sample rate: %d, format: %s, ENTER to select %s", di.Channels, di.SampleRate, di.Format, di.Current)
	}

	return fmt.Sprintf("channels: %d, sample rate: %d, format: %s %s", di.Channels, di.SampleRate, di.Format, di.Current)
}

func SetupDevicesList(ae audio.AudioEngine, deviceType uint, ls ListSetup) DeviceList {
	if ae == nil {
		return DeviceList{}
	}

	var (
		ds        map[string]audio.DeviceInfo
		curDevice string
	)

	switch deviceType {
	case MICROPHONE:
		ds = ae.FetchMicrophones()
		curDevice = ae.GetCurrentMicrophone().Name
	case HEADPHONES:
		ds = ae.FetchHeadphones()
		curDevice = ae.GetCurrentHeadphones().Name
	default:
		return DeviceList{}
	}

	devicesNames := make([]string, 0, len(ds))
	for n := range ds {
		devicesNames = append(devicesNames, n)
	}

	slices.SortFunc(devicesNames, func(a, b string) int {
		return cmp.Compare(ds[a].Index, ds[b].Index)
	})

	devicesItems := make([]list.Item, len(ds))
	for _, d := range devicesNames {
		v := ds[d]
		di := DeviceItem{Name: v.Name, Channels: v.Channels, SampleRate: v.SampleRate, Format: v.Format}

		if d == curDevice {
			di.Current = lipgloss.NewStyle().Foreground(ls.ThemeColor).Render("CURRENT")
		}

		devicesItems[v.Index] = di
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}
	l := list.New(devicesItems, dd, 30, 30)

	l.DisableQuitKeybindings()
	l.Select(0)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return DeviceList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *DeviceList) UpdateDevicesItemList(device string) tea.Cmd {
	var cmd tea.Cmd
	items := l.LipList.Items()
	for i, v := range items {
		dev, ok := v.(DeviceItem)
		if !ok {
			continue
		}
		dev.Current = ""
		if dev.Name == device {
			dev.Current = lipgloss.NewStyle().Foreground(l.themeColor).Render("CURRENT")
		}
		cmd = l.LipList.SetItem(i, dev)
	}
	return cmd
}

func (l *DeviceList) UpdateDevicesList(user *users.User, deviceType uint) tea.Cmd {
	ae := user.Engines.AudioEngine
	if ae == nil {
		return nil
	}

	var (
		ds             map[string]audio.DeviceInfo
		curAudioDevice string
		curUserDevice  string
	)

	userDevices := user.GetDevices()

	switch deviceType {
	case MICROPHONE:
		ds = ae.FetchMicrophones()
		curAudioDevice = ae.GetCurrentMicrophone().Name
		curUserDevice = userDevices.Microphone
	case HEADPHONES:
		ds = ae.FetchHeadphones()
		curAudioDevice = ae.GetCurrentHeadphones().Name
		curUserDevice = userDevices.Headphones
	default:
		return nil
	}

	devicesNames := make([]string, 0, len(ds))
	for n := range ds {
		devicesNames = append(devicesNames, n)
	}

	slices.SortFunc(devicesNames, func(a, b string) int {
		return cmp.Compare(ds[a].Index, ds[b].Index)
	})

	newItems := make([]list.Item, 0, len(devicesNames))

	var currFind bool
	for _, d := range devicesNames {
		v := ds[d]
		mi := DeviceItem{
			Name:       d,
			Channels:   v.Channels,
			SampleRate: v.SampleRate,
			Format:     v.Format,
		}
		if d == curAudioDevice && d == curUserDevice {
			mi.Current = lipgloss.NewStyle().Foreground(l.themeColor).Render("CURRENT")
			currFind = true
		}
		newItems = append(newItems, mi)
	}
	if !currFind {
		switch deviceType {
		case MICROPHONE:
			user.Data.Devices.Microphone = ""
		case HEADPHONES:
			user.Data.Devices.Headphones = ""
		default:
			return nil
		}
	}
	return l.LipList.SetItems(newItems)
}
