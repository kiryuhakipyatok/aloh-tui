package users

func (u *User) GetUsersSetup(nickname string) *UsersSetup {
	u.mu.RLock()
	defer u.mu.RUnlock()
	us, ok := u.Data.Setup.Audio.UsersSetup[nickname]
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
