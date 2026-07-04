package users

import "github.com/google/uuid"

type Friend struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
	Color    string    `json:"color"`
}
