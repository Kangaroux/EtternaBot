package model

type UserServicer interface {
	Get(username string) (*User, error)
	Save(user *User) error
}

type User struct {
	BaseModel
	Username      string  `db:"username"`
	DiscordID     string  `db:"discord_id"`
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
