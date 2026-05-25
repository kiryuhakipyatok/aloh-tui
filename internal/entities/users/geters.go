package users

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
	return u.Data.Setup.AppNotifications
}