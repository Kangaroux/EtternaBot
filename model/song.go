package model

type SongServicer interface {
	Get(etternaID int) (*Song, error)
	Save(song *Song) error
}

type Song struct {
	ID            int
	EtternaID     int    `db:"etterna_id"`
	BackgroundURL string `db:"background_url"`
	Artist        string
	Name          string
}
