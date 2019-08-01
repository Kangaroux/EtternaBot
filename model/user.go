package model

import "github.com/Kangaroux/etternabot/etterna"

type UserServicer interface {
	Get(username string) (*User, error)
	Save(user *User) error
}

type User struct {
	BaseModel
	Username string
	EtternaID int
	Avatar   string
	MSD      etterna.MSD
}
