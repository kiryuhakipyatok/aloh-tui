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

func (u *User) GetFriends() map[uuid.UUID]Friend {
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

func (u *User) IsFriend(iden Identity) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.conatinsFriends(iden)
}

func (u *User) IsBlocked(iden Identity) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return slices.ContainsFunc(u.Data.Personal.BlockedUsers, func(bIden Identity) bool {
		return iden.Nickname == bIden.Nickname || iden.ID == bIden.ID
	})
}

func (u *User) GetFriendIdentityById(id uuid.UUID) (Identity, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	iden := Identity{
		ID: id,
	}
	for _, f := range u.Data.Personal.Friends {
		i := f.Identity.ID
		if i == id {
			iden.Nickname = f.Identity.Nickname
			break
		}
	}
	if iden.ID == uuid.Nil {
		return iden, errs.ErrNotFound()
	}
	return iden, nil
}

func (u *User) GetFriendIdentityByNick(nickname string) (Identity, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	iden := Identity{
		Nickname: nickname,
	}
	for _, f := range u.Data.Personal.Friends {
		n := f.Nickname
		if n == nickname {
			iden.ID = f.Identity.ID
			break
		}
	}
	if iden.ID == uuid.Nil {
		return iden, errs.ErrNotFound()
	}
	return iden, nil
}

func (u *User) conatinsFriends(iden Identity) bool {
	if _, ok := u.Data.Personal.Friends[iden.ID]; ok {
		return true
	}
	for _, v := range u.Data.Personal.Friends {
		if v.Nickname == iden.Nickname {
			return true
		}
	}

	return false
}
