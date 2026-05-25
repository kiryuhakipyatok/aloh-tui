package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type OnlineMsg struct {
	Online map[string][]string
	Err    error
}

func FetchOnlineFriendsCmd(netw networking.Networking, nicknames []string) tea.Cmd {
	return func() tea.Msg {
		msg := OnlineMsg{}
		if len(nicknames) <= 0 {
			return msg
		}
		if netw != nil {
			online, err := netw.FetchOnlineFriends(nicknames)
			if err != nil {
				msg.Err = err
				return msg
			}

			msg.Online = online
		}

		return msg
	}
}

// func FetchOnlineCmd(netw networking.Networking, nickname string) tea.Cmd {
// 	return func() tea.Msg {
// 		msg := OnlineMsg{}
// 		online, err := netw.FetchCurrentOnline()
// 		if err != nil {
// 			msg.Err = err
// 			return msg
// 		}

// 		online = slices.DeleteFunc(online, func(v string) bool {
// 			return v == nickname
// 		})

// 		msg.Online = make(map[string][]string, len(online))
// 		var wg errgroup.Group
// 		wg.SetLimit(15)
// 		var mu sync.Mutex
// 		for _, v := range online {
// 			wg.Go(func() error {
// 				sessions, err := netw.FetchCurrentConnects(v)
// 				if err != nil {
// 					return err
// 				}
// 				mu.Lock()
// 				msg.Online[v] = sessions
// 				mu.Unlock()
// 				return nil
// 			})
// 		}

// 		if err := wg.Wait(); err != nil {
// 			msg.Err = err
// 		}

// 		return msg
// 	}
// }
