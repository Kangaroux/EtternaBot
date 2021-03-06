package etterna

import "time"

type Error struct {
	Code    int
	Context error
	Msg     string
}

type EtternaAPI struct {
	apiKey     string
	baseAPIURL string
	baseURL    string
}

type Judgements struct {
	Marvelous int
	Perfect   int
	Great     int
	Good      int
	Bad       int
	Miss      int
}

type MSD struct {
	Overall    float64
	Stream     float64
	Jumpstream float64
	Handstream float64
	Stamina    float64
	JackSpeed  float64
	Chordjack  float64
	Technical  float64
}

type Rank struct {
	Overall    int
	Stream     int
	Jumpstream int
	Handstream int
	Stamina    int
	JackSpeed  int
	Chordjack  int
	Technical  int
}

type Song struct {
	ID            int
	Name          string
	Artist        string
	BackgroundURL string
	Key           string
}

type Score struct {
	Accuracy float64
	Date     time.Time
	Key      string
	Rate     float64
	Nerfed   float64
	MaxCombo int
	Mods     string
	Valid    bool
	MinesHit int
	Song     Song
	User     User

	Judgements
	MSD
}

type User struct {
	ID          int
	AvatarURL   string
	CountryCode string
	Username    string

	MSD
	Rank Rank
}
