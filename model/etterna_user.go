package model

import (
	"database/sql"
	"time"
)

type EtternaUserServicer interface {
	// Gets the id of the discord user associated with the given etterna user in a given server
	GetRegisteredDiscordUserID(serverID, username string) (string, error)

	// Gets the (cached) etterna user associated with a given discord ID in a given server
	GetRegisteredUser(serverID, discordID string) (*EtternaUser, error)

	// Gets the (cached) etterna user with a given username
	GetUsername(username string) (*EtternaUser, error)

	// Gets all etterna users that are registered as well as the discord server that
	// each user is registered in. Used for tracking recent plays
	GetRegisteredUsersForRecentPlays() ([]*RegisteredUserServers, error)

	// Updates/creates the (cached) etterna user
	Save(user *EtternaUser) error

	// Registers a discord user with an etterna user for a particular discord server
	Register(username, serverID, discordID string) (bool, error)

	// Unregisters the discord user from any etterna users for a particular discord server
	Unregister(serverID, discordID string) (bool, error)
}

type EtternaUser struct {
	BaseModel
	Username            string         `db:"username"`
	EtternaID           int            `db:"etterna_id"`
	Avatar              string         `db:"avatar"`
	LastRecentScoreKey  sql.NullString `db:"last_recent_score_key"`
	LastRecentScoreDate *time.Time     `db:"last_recent_score_date"`
	MSDOverall          float64        `db:"msd_overall"`
	MSDStream           float64        `db:"msd_stream"`
	MSDJumpstream       float64        `db:"msd_jumpstream"`
	MSDHandstream       float64        `db:"msd_handstream"`
	MSDStamina          float64        `db:"msd_stamina"`
	MSDJackSpeed        float64        `db:"msd_jackspeed"`
	MSDChordjack        float64        `db:"msd_chordjack"`
	MSDTechnical        float64        `db:"msd_technical"`
	RankOverall         int            `db:"rank_overall"`
	RankStream          int            `db:"rank_stream"`
	RankJumpstream      int            `db:"rank_jumpstream"`
	RankHandstream      int            `db:"rank_handstream"`
	RankStamina         int            `db:"rank_stamina"`
	RankJackSpeed       int            `db:"rank_jackspeed"`
	RankChordjack       int            `db:"rank_chordjack"`
	RankTechnical       int            `db:"rank_technical"`
}

type RegisteredUserServers struct {
	User    EtternaUser
	Servers []DiscordServer
}
