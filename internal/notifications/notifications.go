package notifications

import (
	"fmt"

	_ "embed"
	"github.com/gen2brain/beeep"
)

//go:embed notifications-files/notification.raw
var notificationBytes []byte

//go:embed notifications-files/image.png
var notificationImage []byte

func NotificationSoundBytes() []byte {
	return notificationBytes
}

func Notify(time, nickname, msg string) error {
	beeep.AppName = "aloh"

	if err := beeep.Notify("new message", fmt.Sprintf("%s> %s: %s", time, nickname, msg), notificationImage); err != nil {
		return err
	}
	return nil
}
