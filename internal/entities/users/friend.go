package users

type Friend struct {
	Identity         `json:"identity"`
	FriendDenoises   `json:"denoises"`
	FriendAppereance `json:"appereance"`
}

type FriendAppereance struct {
	Tagline string `json:"tagline"`
	Color   string `json:"color"`
}

type FriendDenoises struct {
	HardDenoised bool `json:"hard-denoised"`
	SoftDenoised bool `json:"soft-denoised"`
}
