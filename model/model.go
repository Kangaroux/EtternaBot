package model

import "time"

type BaseModel struct {
	ID        uint      `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
