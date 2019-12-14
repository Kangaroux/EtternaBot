package etterna

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Kangaroux/etternabot/util"
	"github.com/Kangaroux/htmlquery"
)

const (
	baseAPIURL = "https://api.etternaonline.com/v1"
	baseURL    = "https://etternaonline.com"
)

type SortColumn int

const (
	SortSongName SortColumn = iota
	SortRate
	SortOverall
	SortNerf
	SortAccuracy
	SortDate
	SortStream
	SortJumpstream
	SortHandstream
	SortStamina
	SortJackSpeed
	SortChordjack
	SortTechnical
)

var (
	reUserID    = regexp.MustCompile(`'userid': '(\d+)'`)
	reJudgement = regexp.MustCompile(`([a-zA-Z]+):\s+(\d+)`)
)

// Payload received from score list endpoint
type scorePayload struct {
	Date      string      `json:"datetime"`
	Nerf      interface{} // This is a float64 except when the score is invalid then it's a string ???
	Rate      string      `json:"user_chart_rate_rate"`
	ScoreKey  string
	SongName  string
	WifeScore string

	Overall    string
	Stream     string
	Jumpstream string
	Handstream string
	Stamina    string
	JackSpeed  string
	Chordjack  string
	Technical  string
}

// Payload received from the score detail endpoint
type scoreDetailPayload struct {
	Date        string `json:"datetime"`
	MaxCombo    string
	MinesHit    string `json:"hitmine"`
	Modifiers   string
	Rate        string `json:"user_chart_rate_rate"`
	ScoreKey    string
	Valid       string
	WifeScore   string
	Username    string
	AvatarURL   string
	CountryCode string

	Overall    string
	Stream     string
	Jumpstream string
	Handstream string
	Stamina    string
	JackSpeed  string
	Chordjack  string
	Technical  string
}

// APIInterface is the interface for interacting with the etterna API.
type APIInterface interface {
	GetByUsername(username string) (*User, error)
	GetUserID(username string) (int, error)
	GetScores(userID int, search string, n uint, start uint, sortColumn string, sortAsc bool) ([]Score, error)
}

// New returns a ready-to-use instance of the etterna API
func New(apiKey string) EtternaAPI {
	return EtternaAPI{
		apiKey:     apiKey,
		baseAPIURL: baseAPIURL,
		baseURL:    baseURL,
	}
}

func (api *EtternaAPI) BaseAPIURL() string {
	return api.baseAPIURL
}

func (api *EtternaAPI) BaseURL() string {
	return api.baseURL
}

// GetByUsername returns the user data for a given username. If the user does not
// exist, the error code is ErrNotFound.
func (api *EtternaAPI) GetByUsername(username string) (*User, error) {
	var payload struct {
		Username    string
		CountryCode string
		Avatar      string
		Overall     string
		Stream      string
		Jumpstream  string
		Handstream  string
		Stamina     string
		JackSpeed   string
		Chordjack   string
		Technical   string
	}

	resp, err := http.Get(api.baseAPIURL + "/user_data" +
		fmt.Sprintf("?api_key=%s&username=%s", api.apiKey, username))

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &Error{
			Code: ErrNotFound,
			Msg:  "No user with that username exists.",
		}
	} else if resp.StatusCode == http.StatusForbidden {
		return nil, &Error{
			Code: ErrUnexpected,
			Msg:  "API access is denied due to insufficient permissions (bad API key?).",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	u := User{
		Username:    payload.Username,
		CountryCode: payload.CountryCode,
		AvatarURL:   payload.Avatar,
	}

	u.Overall, _ = strconv.ParseFloat(payload.Overall, 64)
	u.Stream, _ = strconv.ParseFloat(payload.Stream, 64)
	u.Jumpstream, _ = strconv.ParseFloat(payload.Jumpstream, 64)
	u.Handstream, _ = strconv.ParseFloat(payload.Handstream, 64)
	u.Stamina, _ = strconv.ParseFloat(payload.Stamina, 64)
	u.JackSpeed, _ = strconv.ParseFloat(payload.JackSpeed, 64)
	u.Chordjack, _ = strconv.ParseFloat(payload.Chordjack, 64)
	u.Technical, _ = strconv.ParseFloat(payload.Technical, 64)

	if err := api.getUserRanks(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUserID returns a user's ID. For some ungodly reason the user_data endpoint
// doesn't give you the ID, so we have to pull the HTML and look up the user ID
// by hand. Since grabbing the entire HTML is pretty expensive, user IDs should
// be cached whenever possible.
func (api *EtternaAPI) GetUserID(username string) (int, error) {
	resp, err := http.Get(api.baseURL + "/user/" + username)

	if err != nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user ID.",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return 0, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     "No user with that username exists.",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user ID.",
		}
	}

	match := reUserID.FindSubmatch(body)

	if match == nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: errors.New("failed to find userid submatch"),
			Msg:     "Unexpected error trying to look up user ID.",
		}
	}

	id, _ := strconv.Atoi(string(match[1]))

	return id, nil
}

// GetScores returns a list of valid scores for a given user.
func (api *EtternaAPI) GetScores(userID int, search string, n uint, start uint, sortColumn SortColumn, sortAsc bool) ([]Score, error) {
	var payload struct {
		Data []scorePayload
	}

	sortStr := ""

	if sortAsc {
		sortStr = "asc"
	} else {
		sortStr = "desc"
	}

	form := url.Values{}
	form.Set("start", strconv.Itoa(int(start)))
	form.Set("length", strconv.Itoa(int(n)))
	form.Set("userid", strconv.Itoa(userID))
	form.Set("order[0][column]", strconv.Itoa(int(sortColumn)))
	form.Set("order[0][dir]", sortStr)

	if search != "" {
		form.Set("search[value]", search)
	}

	resp, err := http.PostForm(api.baseURL+"/score/userScores", form)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve scores",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     "User does not exist.",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve scores",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve scores",
		}
	}

	scores := []Score{}

	for _, payload := range payload.Data {
		if payload.Nerf == "0" {
			continue
		}

		score, err := parseScorePayload(payload)

		if err != nil {
			return nil, &Error{
				Code:    ErrUnexpected,
				Context: err,
				Msg:     "Unexpected error trying to retrieve scores",
			}
		}

		scores = append(scores, *score)
	}

	return scores, nil
}

// GetScoreDetail gets the full details of a song (except for the nerf rating, ty rop)
func (api *EtternaAPI) GetScoreDetail(scoreKey string) (*Score, error) {
	var payload []scoreDetailPayload

	reqURL := fmt.Sprintf(api.baseAPIURL+"/score?api_key=%s&key=%s", api.apiKey, scoreKey[:41])
	resp, err := http.PostForm(reqURL, url.Values{})

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve score details",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     "Score does not exist.",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve score details",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve score details",
		}
	}

	p := payload[0]

	score := Score{}

	score.Overall, _ = strconv.ParseFloat(p.Overall, 64)
	score.Stream, _ = strconv.ParseFloat(p.Stream, 64)
	score.Jumpstream, _ = strconv.ParseFloat(p.Jumpstream, 64)
	score.Handstream, _ = strconv.ParseFloat(p.Handstream, 64)
	score.Stamina, _ = strconv.ParseFloat(p.Stamina, 64)
	score.JackSpeed, _ = strconv.ParseFloat(p.JackSpeed, 64)
	score.Chordjack, _ = strconv.ParseFloat(p.Chordjack, 64)
	score.Technical, _ = strconv.ParseFloat(p.Technical, 64)

	score.Overall = util.RoundToPrecision(score.Overall, 2)
	score.Stream = util.RoundToPrecision(score.Stream, 2)
	score.Jumpstream = util.RoundToPrecision(score.Jumpstream, 2)
	score.Handstream = util.RoundToPrecision(score.Handstream, 2)
	score.Stamina = util.RoundToPrecision(score.Stamina, 2)
	score.JackSpeed = util.RoundToPrecision(score.JackSpeed, 2)
	score.Chordjack = util.RoundToPrecision(score.Chordjack, 2)
	score.Technical = util.RoundToPrecision(score.Technical, 2)

	score.MaxCombo, _ = strconv.Atoi(p.MaxCombo)
	score.Valid = p.Valid == "1"
	score.Date, _ = time.Parse("2006-01-02 15:04:05", p.Date)
	score.Mods = p.Modifiers
	score.MinesHit, _ = strconv.Atoi(p.MinesHit)

	// The user ID is embedded at the end of the score key
	userID, _ := strconv.ParseInt(scoreKey[41:], 10, 32)

	score.User = User{
		ID:          int(userID),
		AvatarURL:   p.AvatarURL,
		CountryCode: p.CountryCode,
		Username:    p.Username,
	}

	return &score, nil
}

func (api *EtternaAPI) GetSong(id int) (*Song, error) {
	var payload []struct {
		SongKey    string
		ID         string
		SongName   string
		Author     string
		Artist     string
		Background string
	}

	reqURL := fmt.Sprintf(api.baseAPIURL+"/song?api_key=%s&key=%d", api.apiKey, id)
	resp, err := http.PostForm(reqURL, url.Values{})

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve song details",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     "Song does not exist.",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve song details",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retrieve song details",
		}
	}

	song := Song{
		ID:            id,
		Name:          payload[0].SongName,
		Artist:        payload[0].Artist,
		BackgroundURL: payload[0].Background,
		Key:           payload[0].SongKey,
	}

	return &song, nil
}

// getUserRanks gets the user's skillset rankings from the user_rank API
func (api *EtternaAPI) getUserRanks(user *User) error {
	var payload struct {
		Overall    string
		Stream     string
		Jumpstream string
		Handstream string
		Stamina    string
		JackSpeed  string
		Chordjack  string
		Technical  string
	}

	resp, err := http.Get(api.baseAPIURL + "/user_rank" +
		fmt.Sprintf("?api_key=%s&username=%s", api.apiKey, user.Username))

	if err != nil {
		return &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return &Error{
			Code: ErrNotFound,
			Msg:  "No user with that username exists.",
		}
	} else if resp.StatusCode == http.StatusForbidden {
		return &Error{
			Code: ErrUnexpected,
			Msg:  "API access is denied due to insufficient permissions (bad API key?).",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to look up user.",
		}
	}

	user.Rank.Overall, _ = strconv.Atoi(payload.Overall)
	user.Rank.Stream, _ = strconv.Atoi(payload.Stream)
	user.Rank.Jumpstream, _ = strconv.Atoi(payload.Jumpstream)
	user.Rank.Handstream, _ = strconv.Atoi(payload.Handstream)
	user.Rank.Stamina, _ = strconv.Atoi(payload.Stamina)
	user.Rank.JackSpeed, _ = strconv.Atoi(payload.JackSpeed)
	user.Rank.Chordjack, _ = strconv.Atoi(payload.Chordjack)
	user.Rank.Technical, _ = strconv.Atoi(payload.Technical)

	return nil
}

func parseScorePayload(payload scorePayload) (*Score, error) {
	score := Score{}

	if err := parseWifeScore(payload.WifeScore, &score); err != nil {
		return nil, err
	}

	if err := parseSongNameAndID(payload.SongName, &score); err != nil {
		return nil, err
	}

	// If the score is invalid the nerf value will be "0" otherwise it's a float, epic
	if val, ok := payload.Nerf.(float64); ok {
		score.Nerfed = val
	}

	score.Rate, _ = strconv.ParseFloat(payload.Rate, 64)
	score.Key = payload.ScoreKey
	score.Date, _ = time.Parse("2006-01-02", payload.Date)

	doc, err := htmlquery.Parse(strings.NewReader(payload.Overall))

	if err != nil {
		return nil, err
	}

	node := htmlquery.FindOne(doc, "//a")
	text := htmlquery.InnerText(node)

	score.Overall, _ = strconv.ParseFloat(text, 64)
	score.Stream, _ = strconv.ParseFloat(payload.Stream, 64)
	score.Jumpstream, _ = strconv.ParseFloat(payload.Jumpstream, 64)
	score.Handstream, _ = strconv.ParseFloat(payload.Handstream, 64)
	score.Stamina, _ = strconv.ParseFloat(payload.Stamina, 64)
	score.JackSpeed, _ = strconv.ParseFloat(payload.JackSpeed, 64)
	score.Chordjack, _ = strconv.ParseFloat(payload.Chordjack, 64)
	score.Technical, _ = strconv.ParseFloat(payload.Technical, 64)

	score.Overall = util.RoundToPrecision(score.Overall, 2)
	score.Stream = util.RoundToPrecision(score.Stream, 2)
	score.Jumpstream = util.RoundToPrecision(score.Jumpstream, 2)
	score.Handstream = util.RoundToPrecision(score.Handstream, 2)
	score.Stamina = util.RoundToPrecision(score.Stamina, 2)
	score.JackSpeed = util.RoundToPrecision(score.JackSpeed, 2)
	score.Chordjack = util.RoundToPrecision(score.Chordjack, 2)
	score.Technical = util.RoundToPrecision(score.Technical, 2)

	return &score, nil
}

func parseSongNameAndID(s string, score *Score) error {
	doc, err := htmlquery.Parse(strings.NewReader(s))

	if err != nil {
		return err
	}

	node := htmlquery.FindOne(doc, "//a")
	score.Song.Name = htmlquery.InnerText(node)
	urlParts := strings.Split(htmlquery.SelectAttr(node, "href"), "/")
	score.Song.ID, _ = strconv.Atoi(urlParts[len(urlParts)-1])

	return nil
}

// Parse the wifescore and judgements
func parseWifeScore(s string, score *Score) error {
	doc, err := htmlquery.Parse(strings.NewReader(s))

	if err != nil {
		return err
	}

	node := htmlquery.FindOne(doc, "//div")
	text := htmlquery.SelectAttr(node, "title")

	// Judgements are in the title attribute in the form "Marvelous: 1234<br/> Perfect: 1234<br/> "...
	matches := reJudgement.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		val, _ := strconv.Atoi(match[2])

		switch strings.ToLower(match[1]) {
		case "marvelous":
			score.Marvelous = val
		case "perfect":
			score.Perfect = val
		case "great":
			score.Great = val
		case "good":
			score.Good = val
		case "bad":
			score.Bad = val
		case "miss":
			score.Miss = val
		}
	}

	node = htmlquery.FindOne(doc, "//span")
	text = strings.TrimSuffix(htmlquery.InnerText(node), "%")
	score.Accuracy, _ = strconv.ParseFloat(text, 64)

	return nil
}
