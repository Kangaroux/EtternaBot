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

type Song struct {
	ID            int
	Name          string
	Author        string
	Artist        string
	BackgroundURL string
	SongKey       string
}

type Score struct {
	Accuracy float64
	Date     time.Time
	Key      string
	Rate     float64
	SongName string
	SongID   int
	Nerfed   float64

	Judgements
	MSD
}

type User struct {
	ID          int
	AvatarURL   string
	CountryCode string
	Username    string

	MSD
}
