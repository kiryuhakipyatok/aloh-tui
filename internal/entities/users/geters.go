package users

import (
	"aloh-tui/pkg/errs"
	"slices"

	"github.com/google/uuid"
)

func (u *User) GetUsersSetup(nickname string) (UsersSetup, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
	if !ok {
		return us, errs.ErrNotFound()
	}
	return us, nil
}

func (u *User) GetFriendsReqs() []FriendReq {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Personal.FriendsReqs
}

func (u *User) GetFriends() map[uuid.UUID]*Friend {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Personal.Friends
}

func (u *User) GetAmountOfFriends() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.Data.Personal.Friends)
}

func (u *User) GetAmountOfBlocked() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.Data.Personal.BlockedUsers)
}

func (u *User) GetMinutesInCurrentConenction() uint {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.MinutesInCurrentConnection
}

func (u *User) GetMaxTimeInConnetion() uint {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.MaxTimeInConnetion
}

func (u *User) GetAmountOfMessages() uint {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.AmountOfMessages
}

func (u *User) GetAmountOfMinutesInConnections() uint {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.AmountOfMinutesInConnections
}

func (u *User) GetAmountOfConnections() uint {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.AmountOfConnections
}

func (u *User) GetBestFriend() BestFriend {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Statistics.BestFriend
}

func (u *User) GetAppNotificationsState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Notifications.AppNotifications
}

func (u *User) GetDesktopNotificationsState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Notifications.DesktopNotifications
}

func (u *User) GetAudioNotificationsState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Notifications.AudioNotifications
}

func (u *User) GetShowTimeState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.ShowTime
}

func (u *User) GetShowDateState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.ShowDate
}

func (u *User) GetShowZoneState() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.ShowZone
}

func (u *User) GetNotificationTag() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.NotificaionTag
}

func (u *User) GetBanTag() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.BanTag
}

func (u *User) GetBFTag() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.BestFriendTag
}

func (u *User) GetTagline() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance.Tagline
}

func (u *User) GetBinds() Binds {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Binds
}

func (u *User) GetAudio() Audio {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Audio
}

func (u *User) GetDenoises() Denoises {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Audio.Denoises
}

func (u *User) GetMutes() Mutes {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Audio.Mutes
}

func (u *User) GetNotifications() Notifications {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Notifications
}

func (u *User) GetDevices() Devices {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Devices
}

func (u *User) GetAppereance() Appereance {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Data.Setup.Appereance
}

func (u *User) IsFriend(id uuid.UUID) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.conatinsFriends(id)
}

func (u *User) IsBlocked(nickname string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return slices.Contains(u.Data.Personal.BlockedUsers, nickname)
}

func (u *User) GetFriendNickname(id uuid.UUID) string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	f, ok := u.Data.Personal.Friends[id]
	if !ok {
		return ""
	}
	return f.Nickname
}

func (u *User) GetFriendId(nickname string) uuid.UUID {
	u.mu.RLock()
	defer u.mu.RUnlock()
	var id uuid.UUID
	for _, f := range u.Data.Personal.Friends {
		n := f.Nickname
		if n == nickname {
			id = f.ID
			break
		}
	}
	return id
}

func (u *User) conatinsFriends(id uuid.UUID) bool {
	_, ok := u.Data.Personal.Friends[id]
	return ok
}
