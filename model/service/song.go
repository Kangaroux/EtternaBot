package service

import (
	"database/sql"

	"github.com/Kangaroux/etternabot/model"
	"github.com/jmoiron/sqlx"
)

type SongService struct {
	db *sqlx.DB
}

func NewSongService(db *sqlx.DB) SongService {
	return SongService{db: db}
}

func (s SongService) Get(etternaID int) (*model.Song, error) {
	song := &model.Song{}

	if err := s.db.Get(song, `SELECT * FROM "songs" WHERE etterna_id=$1`, etternaID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return song, nil
}

func (s SongService) Save(song *model.Song) error {
	var err error

	if song.ID == 0 {
		q := `INSERT INTO "songs" (
			etterna_id,
			artist,
			name,
			background_url
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

		err = s.db.Get(&song.ID, q,
			song.EtternaID,
			song.Artist,
			song.Name,
			song.BackgroundURL,
		)
	} else {
		q := `UPDATE "songs" SET
			artist=$2,
			name=$3,
			background_url=$4
		WHERE id=$1`

		_, err = s.db.Exec(q,
			song.ID,
			song.Artist,
			song.Name,
			song.BackgroundURL,
		)
	}

	return err
}
