package service

import (
	"database/sql"
	"time"

	"github.com/Kangaroux/etternabot/model"
	"github.com/jmoiron/sqlx"
)

type DiscordServerService struct {
	db *sqlx.DB
}

func NewDiscordServerService(db *sqlx.DB) DiscordServerService {
	return DiscordServerService{db: db}
}

func (s DiscordServerService) Get(serverID string) (*model.DiscordServer, error) {
	server := &model.DiscordServer{}

	if err := s.db.Get(server, `SELECT * FROM "discord_servers" WHERE server_id=$1`, serverID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return server, nil
}

func (s DiscordServerService) Save(server *model.DiscordServer) error {
	var err error

	now := time.Now().UTC()
	server.UpdatedAt = now

	if server.ID == 0 {
		server.CreatedAt = now
		q := `INSERT INTO "discord_servers" (
			created_at,
			updated_at,
			command_prefix,
			server_id,
			score_channel_id,
			last_song_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

		err = s.db.Get(&server.ID, q,
			server.CreatedAt,
			server.UpdatedAt,
			server.CommandPrefix,
			server.ServerID,
			server.ScoreChannelID,
			server.LastSongID,
		)
	} else {
		q := `UPDATE "discord_servers" SET
			updated_at=$2,
			command_prefix=$3,
			score_channel_id=$4,
			last_song_id=$5
		WHERE id=$1`

		_, err = s.db.Exec(q,
			server.ID,
			server.UpdatedAt,
			server.CommandPrefix,
			server.ScoreChannelID,
			server.LastSongID,
		)
	}

	return err
}
