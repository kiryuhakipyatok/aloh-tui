package lists

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/media/video"
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
	WEBCAMS

	AUDIO_TYPE
	VIDEO_TYPE
)

type DeviceItem struct {
	Name    string
	typee   uint
	props   props
	Current string
}

type props any

type audioProps struct {
	Channels   uint32
	SampleRate uint32
	Format     string
}

type videProps struct {
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
	var desc string
	switch di.typee {
	case AUDIO_TYPE:
		adi, ok := di.props.(audioProps)
		if ok {
			desc = fmt.Sprintf("channels: %d, sample rate: %d, format: %s", adi.Channels, adi.SampleRate, adi.Format)
		}
	case VIDEO_TYPE:
		_, ok := di.props.(videProps)
		if ok {
			desc = "webcam"
		}
	}
	if isSelected {
		return fmt.Sprintf("%s, ENTER to select %s", desc, di.Current)
	}

	return fmt.Sprintf("%s %s", desc, di.Current)
}

func SetupDevicesList(engines users.Engines, deviceType uint, ls ListSetup) DeviceList {
	ae := engines.AudioEngine
	ve := engines.VideoEngine
	if ae == nil {
		return DeviceList{}
	}
	if ve == nil {
		return DeviceList{}
	}

	var (
		devices   map[string]media.Device
		curDevice string
	)

	switch deviceType {
	case MICROPHONE:
		devices = ae.FetchMicrophones()
		di, ok := ae.GetCurrentMicrophone().(audio.DeviceInfo)
		if !ok {
			return DeviceList{}
		}
		curDevice = di.Name
	case HEADPHONES:
		devices = ae.FetchHeadphones()
		di, ok := ae.GetCurrentHeadphones().(audio.DeviceInfo)
		if !ok {
			return DeviceList{}
		}
		curDevice = di.Name
	case WEBCAMS:
		devices = ve.FetchWebcams()
		di, ok := ve.GetCurrentWebcam().(video.DeviceInfo)
		if !ok {
			return DeviceList{}
		}
		curDevice = di.Name
	default:
		return DeviceList{}
	}

	devicesNames := make([]string, 0, len(devices))
	for n := range devices {
		devicesNames = append(devicesNames, n)
	}

	switch deviceType {
	case MICROPHONE, HEADPHONES:
		slices.SortFunc(devicesNames, func(a, b string) int {
			dia, ok := devices[a].(audio.DeviceInfo)
			if !ok {
				return -1
			}
			dib, ok := devices[a].(audio.DeviceInfo)
			if !ok {
				return -1
			}
			return cmp.Compare(dia.Index, dib.Index)
		})
	case WEBCAMS:
		slices.SortFunc(devicesNames, func(a, b string) int {
			dia, ok := devices[a].(video.DeviceInfo)
			if !ok {
				return -1
			}
			dib, ok := devices[a].(video.DeviceInfo)
			if !ok {
				return -1
			}
			return cmp.Compare(dia.Index, dib.Index)
		})
	}

	devicesItems := make([]list.Item, len(devices))

	switch deviceType {
	case HEADPHONES, MICROPHONE:
		for _, d := range devicesNames {
			v := devices[d]
			adi, ok := v.(audio.DeviceInfo)
			if !ok {
				continue
			}
			di := DeviceItem{
				Name:  adi.Name,
				typee: AUDIO_TYPE,
				props: audioProps{
					Channels:   adi.Channels,
					Format:     adi.Format,
					SampleRate: adi.SampleRate,
				},
			}

			if d == curDevice {
				di.Current = lipgloss.NewStyle().Foreground(ls.ThemeColor).Render("CURRENT")
			}

			devicesItems[adi.Index] = di
		}
	case WEBCAMS:

		for _, d := range devicesNames {
			v := devices[d]
			vdi, ok := v.(video.DeviceInfo)
			if !ok {
				continue
			}
			di := DeviceItem{
				Name:  vdi.Name,
				typee: VIDEO_TYPE,
				props: videProps{},
			}

			if d == curDevice {
				di.Current = lipgloss.NewStyle().Foreground(ls.ThemeColor).Render("CURRENT")
			}

			devicesItems[vdi.Index] = di
		}

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
	ve := user.Engines.VideoEngine
	if ae == nil {
		return nil
	}

	var (
		devices   map[string]media.Device
		curDevice string
		//curUserDevice string
	)

	//userDevices := user.GetDevices()

	switch deviceType {
	case MICROPHONE:
		devices = ae.FetchMicrophones()
		di, ok := ae.GetCurrentMicrophone().(audio.DeviceInfo)
		if !ok {
			return nil
		}
		curDevice = di.Name
		//curUserDevice = userDevices.Microphone
	case HEADPHONES:
		devices = ae.FetchHeadphones()
		di, ok := ae.GetCurrentHeadphones().(audio.DeviceInfo)
		if !ok {
			return nil
		}
		curDevice = di.Name
		//curUserDevice = userDevices.Headphones
	case WEBCAMS:
		devices = ve.FetchWebcams()
		di, ok := ae.GetCurrentHeadphones().(audio.DeviceInfo)
		if !ok {
			return nil
		}
		curDevice = di.Name
		//curUserDevice = userDevices.Headphones
	default:
		return nil
	}

	devicesNames := make([]string, 0, len(devices))
	for n := range devices {
		devicesNames = append(devicesNames, n)
	}

	switch deviceType {
	case MICROPHONE, HEADPHONES:
		slices.SortFunc(devicesNames, func(a, b string) int {
			dia, ok := devices[a].(audio.DeviceInfo)
			if !ok {
				return -1
			}
			dib, ok := devices[a].(audio.DeviceInfo)
			if !ok {
				return -1
			}
			return cmp.Compare(dia.Index, dib.Index)
		})
	case WEBCAMS:
		slices.SortFunc(devicesNames, func(a, b string) int {
			dia, ok := devices[a].(video.DeviceInfo)
			if !ok {
				return -1
			}
			dib, ok := devices[a].(video.DeviceInfo)
			if !ok {
				return -1
			}
			return cmp.Compare(dia.Index, dib.Index)
		})
	}

	newItems := make([]list.Item, len(devicesNames))

	var currFind bool
	switch deviceType {
	case HEADPHONES, MICROPHONE:
		for _, d := range devicesNames {
			v := devices[d]
			adi, ok := v.(audio.DeviceInfo)
			if !ok {
				continue
			}
			di := DeviceItem{
				Name:  adi.Name,
				typee: AUDIO_TYPE,
				props: audioProps{
					Channels:   adi.Channels,
					Format:     adi.Format,
					SampleRate: adi.SampleRate,
				},
			}

			if d == curDevice {
				di.Current = lipgloss.NewStyle().Foreground(l.themeColor).Render("CURRENT")
			}

			newItems[adi.Index] = di
		}
	case WEBCAMS:

		for _, d := range devicesNames {
			v := devices[d]
			vdi, ok := v.(video.DeviceInfo)
			if !ok {
				continue
			}
			di := DeviceItem{
				Name:  vdi.Name,
				typee: VIDEO_TYPE,
				props: videProps{},
			}

			if d == curDevice {
				di.Current = lipgloss.NewStyle().Foreground(l.themeColor).Render("CURRENT")
			}

			newItems[vdi.Index] = di
		}

	}
	if !currFind {
		switch deviceType {
		case MICROPHONE:
			user.Data.Devices.Microphone = ""
		case HEADPHONES:
			user.Data.Devices.Headphones = ""
		case WEBCAMS:
			user.Data.Devices.Webcam = ""
		default:
			return nil
		}
	}
	return l.LipList.SetItems(newItems)
}
