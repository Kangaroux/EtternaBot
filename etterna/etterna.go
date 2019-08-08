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

// Payload received from the userScores endpoint
type scorePayload struct {
	SongName  string
	Rate      string `json:"user_chart_rate_rate"`
	Nerf      float64
	ScoreKey  string
	Date      string `json:"datetime"`
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

// APIInterface is the interface for interacting with the etterna API.
type APIInterface interface {
	GetByUsername(username string) (*User, error)
	GetUserID(username string) (int, error)
	GetScores(userID int, n uint, start uint, sortColumn string, sortAsc bool) ([]Score, error)
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

// GetScores returns a list of scores for a given user.
func (api *EtternaAPI) GetScores(userID int, n uint, start uint, sortColumn SortColumn, sortAsc bool) ([]Score, error) {
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

	resp, err := http.PostForm(api.baseURL+"/score/userScores", form)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retreive scores",
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
			Msg:     "Unexpected error trying to retreive scores",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retreive scores",
		}
	}

	scores := []Score{}

	for _, payload := range payload.Data {
		score, err := parseScorePayload(payload)

		if err != nil {
			return nil, &Error{
				Code:    ErrUnexpected,
				Context: err,
				Msg:     "Unexpected error trying to retreive scores",
			}
		}

		scores = append(scores, *score)
	}

	return scores, nil
}

// GetScoreDetail gets the full details of a song (except for the nerf rating, ty rop)
func (api *EtternaAPI) GetScoreDetail(scoreKey string) (*Score, error) {
	var payload []struct {
		scorePayload
		MaxCombo  string
		Valid     string
		Modifiers string
		DateTime  string
		MinesHit  string `json:"hitmine"`
	}

	reqURL := fmt.Sprintf(api.baseAPIURL+"/score?api_key=%s&key=%s", api.apiKey, scoreKey[:41])
	resp, err := http.PostForm(reqURL, url.Values{})

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retreive score details",
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
			Msg:     "Unexpected error trying to retreive score details",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retreive score details",
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
	score.MaxCombo, _ = strconv.Atoi(p.MaxCombo)
	score.Valid = p.Valid == "1"
	score.Date, _ = time.Parse("2006-01-02 15:04:05", p.DateTime)
	score.Mods = p.Modifiers
	score.MinesHit, _ = strconv.Atoi(p.MinesHit)

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
			Msg:     "Unexpected error trying to retreive song details",
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
			Msg:     "Unexpected error trying to retreive song details",
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     "Unexpected error trying to retreive song details",
		}
	}

	song := Song{
		Name:          payload[0].SongName,
		Author:        payload[0].Author,
		Artist:        payload[0].Artist,
		BackgroundURL: payload[0].Background,
		Key:           payload[0].SongKey,
	}

	song.ID, _ = strconv.Atoi(payload[0].ID)

	return &song, nil
}

func parseScorePayload(payload scorePayload) (*Score, error) {
	score := Score{}

	if err := parseWifeScore(payload.WifeScore, &score); err != nil {
		return nil, err
	}

	if err := parseSongNameAndID(payload.SongName, &score); err != nil {
		return nil, err
	}

	score.Nerfed = payload.Nerf
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
