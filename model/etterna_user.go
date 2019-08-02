package model

type UserServicer interface {
	// Gets the (cached) etterna user associated with a given discord ID in a given server
	GetDiscordID(serverID, discordID string) (*EtternaUser, error)

	// Gets the (cached) etterna user with a given username
	GetUsername(username string) (*EtternaUser, error)

	// Updates/creates the (cached) etterna user
	Save(user *EtternaUser) error
}

type EtternaUser struct {
	BaseModel
	Username      string  `db:"username"`
	EtternaID     int     `db:"etterna_id"`
	Avatar        string  `db:"avatar"`
	MSDOverall    float64 `db:"msd_overall"`
	MSDStream     float64 `db:"msd_stream"`
	MSDJumpstream float64 `db:"msd_jumpstream"`
	MSDHandstream float64 `db:"msd_handstream"`
	MSDStamina    float64 `db:"msd_stamina"`
	MSDJackSpeed  float64 `db:"msd_jackspeed"`
	MSDChordjack  float64 `db:"msd_chordjack"`
	MSDTechnical  float64 `db:"msd_technical"`
}
