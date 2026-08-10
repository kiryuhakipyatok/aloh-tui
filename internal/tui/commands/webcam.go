package commands

// import (
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/google/uuid"
// )

// type RawWebcamMsg struct {
// 	Id   uuid.UUID
// 	Data []byte
// }

// type WebcamMsg struct {
// 	Id    uuid.UUID
// 	Frame string
// }

// func WaitForRawWebcamMsgCmd(sub chan RawWebcamMsg) tea.Cmd {
// 	return func() tea.Msg {
// 		return <-sub
// 	}
// }

// func WaitForWebcamMsgCmd(sub chan WebcamMsg) tea.Cmd {
// 	return func() tea.Msg {
// 		return <-sub
// 	}
// }

// func RenderUsersFrameCmd(ve video.VideoEngine, id uuid.UUID, data []byte) tea.Cmd {
// 	return func() tea.Msg {
// 		msg := WebcamMsg{
// 			Id: id,
// 		}

// 		frame := ve.RenderUsersVideoTerminal(id, data)
// 		msg.Frame = frame
// 		return msg
// 	}
// }
