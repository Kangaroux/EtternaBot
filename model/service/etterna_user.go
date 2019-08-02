package service

import (
	"database/sql"
	"time"

	"github.com/Kangaroux/etternabot/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type EtternaUserService struct {
	db *sqlx.DB
}

// NewUserService returns a service for managing users in the database
func NewUserService(db *sqlx.DB) EtternaUserService {
	return EtternaUserService{db: db}
}

// GetRegisteredDiscordUserID looks up the discord ID of the user who is registered
// to the given etterna user in the given discord server. If the etterna user is
// not currently registered in the server, returns an empty string as the discord ID
func (s EtternaUserService) GetRegisteredDiscordUserID(serverID, username string) (string, error) {
	var discordID string

	query := `
		SELECT discord_user_id FROM "users_discord_servers"
		WHERE server_id=$1 AND username=$2
	`

	if err := s.db.Get(&discordID, query, serverID, username); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}

		return "", err
	}

	return discordID, nil
}

// GetRegisteredUser looks up the etterna user that the given discord ID is registered
// with in the given discord server.
func (s EtternaUserService) GetRegisteredUser(serverID, discordID string) (*model.EtternaUser, error) {
	user := &model.EtternaUser{}
	query := `
		SELECT u.* FROM "users_discord_servers" uds
		INNER JOIN "etterna_users" u ON u.username=uds.username
		WHERE uds.discord_user_id=$1 AND uds.server_id=$2
	`

	if err := s.db.Get(user, query, discordID, serverID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		// log.Log.Errorf("Failed to lookup user discord_id=%v err=%v", id, err)
		return nil, err
	}

	return user, nil
}

// GetUsername returns the matching cached etterna user. If the user is not
// found, that doesn't mean the user doesn't exist, but that we likely haven't
// cached it yet
func (s EtternaUserService) GetUsername(username string) (*model.EtternaUser, error) {
	user := &model.EtternaUser{}

	if err := s.db.Get(user, `SELECT * FROM "etterna_users" WHERE username=$1`, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		// log.Log.Errorf("Failed to lookup user username=%v err=%v", username, err)
		return nil, err
	}

	return user, nil
}

// Save updates an etterna user, creating a new record if necessary
func (s EtternaUserService) Save(user *model.EtternaUser) error {
	var err error

	now := time.Now().UTC()
	user.UpdatedAt = now

	if user.ID == 0 {
		user.CreatedAt = now
		q := `INSERT INTO "etterna_users" (
			created_at,
			updated_at,
			username,
			etterna_id,
			avatar,
			last_recent_score_key,
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
			user.EtternaID,
			user.Avatar,
			user.LastRecentScoreKey,
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
		q := `UPDATE "etterna_users" SET
			updated_at=$2,
			avatar=$3,
			last_recent_score_key=$4,
			msd_overall=$5,
			msd_stream=$6,
			msd_jumpstream=$7,
			msd_handstream=$8,
			msd_stamina=$9,
			msd_jackspeed=$10,
			msd_chordjack=$11,
			msd_technical=$12
		WHERE username=$1`

		_, err = s.db.Exec(q,
			user.Username,
			user.UpdatedAt,
			user.Avatar,
			user.LastRecentScoreKey,
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

// Register associates an etterna user with a discord user for a particular server.
// For a given server, there needs to be a One-to-One relationship between etterna
// users and discord users. If the constraint is violated, returns false, nil
func (s EtternaUserService) Register(username, serverID, discordID string) (bool, error) {
	query := `
		INSERT INTO "users_discord_servers" (
			username,
			discord_user_id,
			server_id
		)
		VALUES ($1, $2, $3)
	`

	if _, err := s.db.Exec(query, username, discordID, serverID); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Unregister deletes any association to etterna users for the given discord user
// in the given server. If the user was registered, returns true, nil
func (s EtternaUserService) Unregister(serverID, discordID string) (bool, error) {
	var err error
	var result sql.Result

	query := `
		DELETE FROM "users_discord_servers"
		WHERE server_id=$1 AND discord_user_id=$2
	`

	if result, err = s.db.Exec(query, serverID, discordID); err != nil {
		return false, err
	}

	affected, _ := result.RowsAffected()
	return affected > 0, nil
}
