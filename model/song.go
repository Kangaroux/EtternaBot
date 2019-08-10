package model

type SongServicer interface {
	Get(id int) (*Song, error)
	Save(song *Song) error
}

type Song struct {
	ID            int
	Name          string
	Author        string
	Artist        string
	BackgroundURL string
	Key           string
}
