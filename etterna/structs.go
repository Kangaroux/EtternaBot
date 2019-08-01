package etterna

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

type User struct {
	ID          int
	Username    string
	CountryCode string
	AvatarURL   string

	Overall    float64
	Stream     float64
	Jumpstream float64
	Handstream float64
	Stamina    float64
	JackSpeed  float64
	Chordjack  float64
	Technical  float64
}
