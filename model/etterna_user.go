package model

import "database/sql"

type EtternaUserServicer interface {
	// Gets the id of the discord user associated with the given etterna user in a given server
	GetRegisteredDiscordUserID(serverID, username string) (string, error)

	// Gets the (cached) etterna user associated with a given discord ID in a given server
	GetRegisteredUser(serverID, discordID string) (*EtternaUser, error)

	// Gets the (cached) etterna user with a given username
	GetUsername(username string) (*EtternaUser, error)

	// Updates/creates the (cached) etterna user
	Save(user *EtternaUser) error

	// Registers a discord user with an etterna user for a particular discord server
	Register(username, serverID, discordID string) (bool, error)

	// Unregisters the discord user from any etterna users for a particular discord server
	Unregister(serverID, discordID string) (bool, error)
}

type EtternaUser struct {
	BaseModel
	Username           string         `db:"username"`
	EtternaID          int            `db:"etterna_id"`
	Avatar             string         `db:"avatar"`
	LastRecentScoreKey sql.NullString `db:"last_recent_score_key"`
	MSDOverall         float64        `db:"msd_overall"`
	MSDStream          float64        `db:"msd_stream"`
	MSDJumpstream      float64        `db:"msd_jumpstream"`
	MSDHandstream      float64        `db:"msd_handstream"`
	MSDStamina         float64        `db:"msd_stamina"`
	MSDJackSpeed       float64        `db:"msd_jackspeed"`
	MSDChordjack       float64        `db:"msd_chordjack"`
	MSDTechnical       float64        `db:"msd_technical"`
}
