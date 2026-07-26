package users

import (
	"aloh-tui/internal/entities/setups"
	"aloh-tui/pkg/errs"
	"encoding/json"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/google/uuid"
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

func (u *User) IncreaseAmountOfConnectionsByUser(iden Identity) (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	var newBf bool
	if !u.conatinsFriends(iden) {
		return newBf, nil
	}

	if uc, ok := u.Data.Setup.Audio.UsersSetup[iden.Nickname]; ok {
		uc.AmountOfConnections++
		u.Data.Setup.Audio.UsersSetup[iden.Nickname] = uc
		if iden == u.Data.Statistics.BestFriend.Identity {
			u.Data.Statistics.BestFriend.AmountOfConnections++
		} else if uc.AmountOfConnections > u.Data.Statistics.BestFriend.AmountOfConnections &&
			uc.AmountOfConnections > 10 {
			u.Data.Statistics.BestFriend.AmountOfConnections = uc.AmountOfConnections
			u.Data.Statistics.BestFriend.Identity = iden
			newBf = true
		}
		if err := u.UpdateUserJSON(); err != nil {
			return newBf, err
		}
	}
	return newBf, nil
}

func (u *User) CountMaxTimeInConnection(stop chan struct{}) error {
	defer func() error {
		u.mu.Lock()
		if u.Data.Statistics.MinutesInCurrentConnection > u.Data.Statistics.MaxTimeInConnetion {
			u.Data.Statistics.MaxTimeInConnetion = u.Data.Statistics.MinutesInCurrentConnection
		}
		u.Data.Statistics.MinutesInCurrentConnection = 0
		if err := u.UpdateUserJSON(); err != nil {
			u.mu.Unlock()
			return err
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
			u.Data.Statistics.MinutesInCurrentConnection++
			u.Data.Statistics.AmountOfMinutesInConnections++
			if err := u.UpdateUserJSON(); err != nil {
				u.mu.Unlock()
				return err
			}
			u.mu.Unlock()
		}
	}
}

func (u *User) OnOffAudioNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Notifications.AudioNotifications = !u.Data.Setup.Notifications.AudioNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffDesktopNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Notifications.DesktopNotifications = !u.Data.Setup.Notifications.DesktopNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffAppNotification() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Notifications.AppNotifications = !u.Data.Setup.Notifications.AppNotifications
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffAEC(aec bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Audio.AEC = aec
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffHardDenoice(denoise bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Audio.Denoises.HardDenoise = denoise
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffSoftDenoice(denoise bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Audio.Denoises.SoftDenoise = denoise
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

func (u *User) ChangeHeadphones(h string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Devices.Headphones = h
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffFilter(filter bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Audio.Filter = filter
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeThemeColor(color string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Appereance.ThemeColor = color
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeBFTag(tag string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Appereance.BestFriendTag = tag
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeBanTag(tag string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Appereance.BanTag = tag
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeTagline(tagline string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Account.Tagline = tagline
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeColor(color string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Account.Color = color
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) ChangeNotificationTag(tag string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Appereance.NotificaionTag = tag
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffShowTime() (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	res := !u.Data.Setup.Appereance.ShowTime
	u.Data.Setup.Appereance.ShowTime = res
	if err := u.UpdateUserJSON(); err != nil {
		return false, err
	}
	return res, nil
}

func (u *User) OnOffShowDate() (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	res := !u.Data.Setup.Appereance.ShowDate
	u.Data.Setup.Appereance.ShowDate = res
	if err := u.UpdateUserJSON(); err != nil {
		return false, err
	}
	return res, nil
}

func (u *User) OnOffShowZone() (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	res := !u.Data.Setup.Appereance.ShowZone
	u.Data.Setup.Appereance.ShowZone = res
	if err := u.UpdateUserJSON(); err != nil {
		return false, err
	}
	return res, nil
}

func (u *User) MuteUnmuteUser(nickname string, res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
	if !ok {
		us = setups.UsersSetup{
			VolumeCoefficient: 1,
		}
	}
	us.Muted = res
	u.Data.Setup.Audio.UsersSetup[nickname] = us
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffUsersHardDenoise(nickname string, res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
	if !ok {
		us = setups.UsersSetup{
			VolumeCoefficient: 1,
		}
	}
	us.HardDenoise = res
	u.Data.Setup.Audio.UsersSetup[nickname] = us
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) OnOffUsersSoftDenoise(nickname string, res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
	if !ok {
		us = setups.UsersSetup{
			VolumeCoefficient: 1,
		}
	}
	us.SoftDenoise = res
	u.Data.Setup.Audio.UsersSetup[nickname] = us
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) MuteUnmuteMic(res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if res {
		u.Data.Setup.Audio.Mutes.FullMute = false
	}
	u.Data.Setup.Audio.Mutes.MicMute = res
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) MuteUnmuteFull(res bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if res {
		u.Data.Setup.Audio.Mutes.MicMute = false
	}
	u.Data.Setup.Audio.Mutes.FullMute = res
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) SetUsersVolume(nickname string, vc float32) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
	if !ok {
		us = setups.UsersSetup{
			VolumeCoefficient: 1,
		}
	}

	us.VolumeCoefficient = vc

	u.Data.Setup.Audio.UsersSetup[nickname] = us

	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) NewFriendReq(frReq FriendReq) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.FriendsReqs = append(u.Data.Personal.FriendsReqs, frReq)
}

func (u *User) NewFriend(friend Friend) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.Friends[friend.ID] = friend
	u.Data.Personal.FriendsReqs = slices.DeleteFunc(u.Data.Personal.FriendsReqs, func(f FriendReq) bool {
		return f.Identity.ID == friend.ID
	})
}

func (u *User) DeleteFriendReq(id uuid.UUID) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.FriendsReqs = slices.DeleteFunc(u.Data.Personal.FriendsReqs, func(f FriendReq) bool {
		return f.Identity.ID == id
	})
}

func (u *User) DeleteFriend(iden Identity) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.Data.Personal.Friends, iden.ID)
	delete(u.Data.Setup.Audio.UsersSetup, iden.Nickname)
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) BlockUser(iden Identity) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.BlockedUsers = append(u.Data.Personal.BlockedUsers, iden)
	var (
		deletedFriend     bool
		deletedFriendNick string
	)
	maps.DeleteFunc(u.Data.Personal.Friends, func(id uuid.UUID, f Friend) bool {
		deletedFriend = id == iden.ID
		if deletedFriend {
			deletedFriendNick = f.Nickname
		}
		return deletedFriend
	})
	u.Data.Personal.FriendsReqs = slices.DeleteFunc(u.Data.Personal.FriendsReqs, func(f FriendReq) bool {
		return f.Identity == iden
	})

	if deletedFriend {
		delete(u.Data.Setup.Audio.UsersSetup, deletedFriendNick)
		bf := u.Data.Statistics.BestFriend
		if bf.Identity == iden {
			u.Data.Statistics.BestFriend = noBF()
		}
		if err := u.UpdateUserJSON(); err != nil {
			return err
		}
	}
	return nil
}

func (u *User) UnblockUser(iden Identity) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Personal.BlockedUsers = slices.DeleteFunc(u.Data.Personal.BlockedUsers, func(bIden Identity) bool {
		return iden == bIden
	})
}

func (u *User) NewUserSetup(nickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Setup.Audio.UsersSetup[nickname] = setups.UsersSetup{
		VolumeCoefficient: 1,
	}

	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) UpdateFriendTagline(id uuid.UUID, tagline string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, ok := u.Data.Personal.Friends[id]
	if !ok {
		return errs.ErrNotFriend()
	}

	f := u.Data.Personal.Friends[id]
	f.Tagline = tagline
	u.Data.Personal.Friends[id] = f

	return nil
}

func (u *User) NewNickname(newNickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Data.Identity.Nickname = newNickname
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) UpdateForeignNickname(iden Identity, newNickname string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	foreignId := iden.ID
	foreignNickname := iden.Nickname
	fr, ok := u.Data.Personal.Friends[foreignId]
	if ok {
		fr.Nickname = newNickname
		u.Data.Personal.Friends[iden.ID] = fr
		if u.Data.Statistics.BestFriend.Identity.ID == foreignId {
			u.Data.Statistics.BestFriend.Identity.Nickname = newNickname
		}
	} else {
		for i, b := range u.Data.Personal.BlockedUsers {
			if b.ID == foreignId {
				u.Data.Personal.BlockedUsers[i].Nickname = newNickname
				break
			}
		}
	}
	for i, fr := range u.Data.Personal.FriendsReqs {
		if fr.Identity.ID == foreignId {
			u.Data.Personal.FriendsReqs[i].Identity.Nickname = newNickname
			break
		}
	}
	us, ok := u.Data.Setup.Audio.UsersSetup[foreignNickname]
	if ok {
		delete(u.Data.Setup.Audio.UsersSetup, iden.Nickname)
		u.Data.Setup.Audio.UsersSetup[newNickname] = us
	}
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) UpdateFriendColor(iden Identity, newColor string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	foreignId := iden.ID
	fr, ok := u.Data.Personal.Friends[foreignId]
	if !ok {
		return errs.ErrNotFound()
	}
	fr.Color = newColor
	u.Data.Personal.Friends[iden.ID] = fr
	return nil
}
