package service

import (
	"database/sql"

	"github.com/Kangaroux/etternabot/model"
)

type UserService struct {
	db *sql.DB
}

func (us *UserService) Get(username string) (*model.User, error) {
	return nil, nil
}

func (us *UserService) Save(user *model.User) error {
	return nil
}
