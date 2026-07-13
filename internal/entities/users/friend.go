package users

import "github.com/google/uuid"

type Friend struct {
	FriendPersonal   `json:"personal"`
	FriendDenoises   `json:"denoises"`
	FriendAppereance `json:"appereance"`
}

type FriendAppereance struct {
	Tagline string `json:"tagline"`
	//Color   string `json:"color"`
}

type FriendDenoises struct {
	HardDenoised bool `json:"hard-denoised"`
	SoftDenoised bool `json:"soft-denoised"`
}

type FriendPersonal struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}