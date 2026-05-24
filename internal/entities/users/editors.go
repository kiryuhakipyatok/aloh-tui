package users

import (
	"encoding/json"
	"os"
	"slices"
	"time"
)

func (u *User) UpdateUserJSON() error {
	userData, err := json.Marshal(u.Data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(u.Paths.DataFilePath, userData, 0644); err != nil {
		return err
	}

	return nil
}

func (u *User) IncreaseAmountOfFriends(nickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.Friends = append(u.Data.Personal.Friends, nickname)
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) DecreaseAmountOfFriends(nickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.Friends = slices.DeleteFunc(u.Data.Personal.Friends, func(f string) bool {
		return f == nickname
	})

	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfMessages() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Statistics.AmountOfMessages++
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfConnections() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Statistics.AmountOfConnections++
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfConnectionsByUser(nickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if uc, ok := u.Data.Setup.UsersSetup[nickname]; ok {
		uc.AmountOfConnections++
		if uc.AmountOfConnections > u.Data.Statistics.BestFriend.AmountOfConnections {
			u.Data.Statistics.BestFriend.AmountOfConnections = uc.AmountOfConnections
			u.Data.Statistics.BestFriend.Nickname = nickname
		}
		if err := u.UpdateUserJSON(); err != nil {
			return err
		}
	}
	return nil
}

func (u *User) CountMaxTimeInConnection(stop chan struct{}) error {
	var currentTime uint
	defer func() error {
		u.mu.Lock()
		if currentTime > u.Data.Statistics.MaxTimeInConnetion {
			u.Data.Statistics.MaxTimeInConnetion = currentTime
			if err := u.UpdateUserJSON(); err != nil {
				return err
			}
		}
		u.mu.Unlock()
		return nil
	}()
	for {
		select {
		case <-stop:
			return nil
		case <-time.After(time.Minute * 1):
			u.mu.Lock()
			currentTime++
			u.Data.Statistics.AmountOfMinutesInConnections++
			if err := u.UpdateUserJSON(); err != nil {
				return err
			}
			u.mu.Unlock()
		}
	}
}

func (u *User) OnOffAudioNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.AudioNotifications = !u.Data.Setup.AudioNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffDesktopNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.DesktopNotifications = !u.Data.Setup.DesktopNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffAppNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.AppNotifications = !u.Data.Setup.AppNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffAEC(aec bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.AEC = aec
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffHardDenoice(denoice bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.HardDenoise = denoice
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffSoftDenoice(denoice bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.SoftDenoise = denoice
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeMicrophone(mic string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Devices.Microphone = mic
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffFilter(filter bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Filter = filter
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeThemeColor(color string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.ThemeColor = color
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeBFTag(tag string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.BestFriendTag = tag
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeNotificationSign(sign string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.NotificaionSign = sign
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) MuteUnmuteUser(nickname string, res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.UsersSetup[nickname]
	if !ok {
		us = &UsersSetup{}
	}
	us.Muted = res
	u.Data.Setup.UsersSetup[nickname] = us
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) SetUsersVolume(nickname string, vc float32) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.UsersSetup[nickname]
	if !ok {
		us = &UsersSetup{
			VolumeCoefficient: 1,
		}
	}

	us.VolumeCoefficient = vc

	u.Data.Setup.UsersSetup[nickname] = us

	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) GetUsersSetup(nickname string) *UsersSetup {
	u.mu.RLock()
	defer u.mu.RUnlock()
	us, ok := u.Data.Setup.UsersSetup[nickname]
	if ok {
		return us
	}
	return nil
}

func (u *User) GetFriendsReqs() []FriendReq {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Personal.FriendsReqs
}

func (u *User) GetFriends() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Personal.Friends
}

func (u *User) GetAppNotificationsState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.AppNotifications
}

func (u *User) NewFriendReq(nickname string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.FriendsReqs = append(u.Data.Personal.FriendsReqs, FriendReq{Nickname: nickname})
}

func (u *User) DeleteFriendReq(nickname string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.FriendsReqs = slices.DeleteFunc(u.Data.Personal.FriendsReqs, func(f FriendReq) bool {
		return f.Nickname == nickname
	})
}

func (u *User) IsFriend(nickname string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return slices.Contains(u.Data.Personal.Friends, nickname)
}