package service

import (
	"database/sql"

	"github.com/Kangaroux/etternabot/model"
	"github.com/jmoiron/sqlx"
)

type SongService struct {
	db *sqlx.DB
}

func NewSOngService(db *sqlx.DB) SongService {
	return SongService{db: db}
}

func (s SongService) Get(id int) (*model.Song, error) {
	song := &model.Song{}

	if err := s.db.Get(song, `SELECT * FROM "songs" WHERE id=$1`, id); err != nil {
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
			artist,
			name,
			background_url,
			key
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

		err = s.db.Get(&song.ID, q,
			song.Artist,
			song.Name,
			song.BackgroundURL,
			song.Key,
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
