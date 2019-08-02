package service

import (
	"database/sql"
	"time"

	"github.com/Kangaroux/etternabot/model"
	"github.com/jmoiron/sqlx"
)

type UserService struct {
	db *sqlx.DB
}

// NewUserService returns a service for managing users in the database
func NewUserService(db *sqlx.DB) UserService {
	return UserService{db: db}
}

func (s UserService) Get(username string) (*model.User, error) {
	user := &model.User{}

	if err := s.db.Get(user, `SELECT * FROM "users" WHERE username=$1`, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		// log.Log.Errorf("Failed to lookup user username=%v err=%v", username, err)
		return nil, err
	}

	return user, nil
}

func (s UserService) Save(user *model.User) error {
	var err error

	now := time.Now().UTC()
	user.UpdatedAt = now

	if user.ID == 0 {
		user.CreatedAt = now
		q := `INSERT INTO "users" (
			created_at,
			updated_at,
			username,
			discord_id
			etterna_id,
			avatar,
			msd_overall,
			msd_stream,
			msd_jumpstream,
			msd_handstream,
			msd_stamina,
			msd_jackspeed,
			msd_chordjack,
			msd_technical
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`

		err = s.db.Get(&user.ID, q,
			user.CreatedAt,
			user.UpdatedAt,
			user.Username,
			user.DiscordID,
			user.EtternaID,
			user.Avatar,
			user.MSDOverall,
			user.MSDStream,
			user.MSDJumpstream,
			user.MSDHandstream,
			user.MSDStamina,
			user.MSDJackSpeed,
			user.MSDChordjack,
			user.MSDTechnical,
		)
	} else {
		q := `UPDATE "users" SET
			updated_at=$2,
			avatar=$3,
			discord_id=$4,
			msd_overall=$5,
			msd_stream=$6,
			msd_jumpstream=$7,
			msd_handstream=$8,
			msd_stamina=$9,
			msd_jackspeed=$10,
			msd_chordjack=$11,
			msd_technical=$12
		WHERE username=$1`

		_, err = s.db.Exec(q, user.CreatedAt,
			user.Username,
			user.UpdatedAt,
			user.Avatar,
			user.DiscordID,
			user.MSDOverall,
			user.MSDStream,
			user.MSDJumpstream,
			user.MSDHandstream,
			user.MSDStamina,
			user.MSDJackSpeed,
			user.MSDChordjack,
			user.MSDTechnical,
		)
	}

	if err != nil {
		// log.Log.Errorf("Failed to save user user=%v err=%v", user, err)
		return err
	}

	return nil
}
