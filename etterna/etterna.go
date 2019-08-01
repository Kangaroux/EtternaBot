package etterna

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

const (
	baseAPIURL = "https://api.etternaonline.com/v1"
	baseURL    = "https://etternaonline.com"
)

var (
	reUserID = regexp.MustCompile(`'userid': '(\d+)'`)
)

// EtternaAPIInterface is the interface for interacting with the etterna API.
type EtternaAPIInterface interface {
	GetByUsername(username string) (*User, error)
	GetUserID(username string) (int, error)
	GetScores(userID int, n int, sortColumn, sortAsc bool)
}

// New returns a ready-to-use instance of the etterna API
func New(apiKey string) EtternaAPI {
	return EtternaAPI{
		apiKey:     apiKey,
		baseAPIURL: baseAPIURL,
		baseURL:    baseURL,
	}
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

	if resp.StatusCode == http.StatusNotFound {
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(data))
		return nil, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     fmt.Sprintf("No user with username '%s' exists.", username),
		}
	}

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     fmt.Sprintf("Unexpected error trying to look up user."),
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     fmt.Sprintf("Unexpected error trying to look up user."),
		}
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     fmt.Sprintf("Unexpected error trying to look up user."),
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

	if resp.StatusCode == http.StatusNotFound {
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(data))
		return 0, &Error{
			Code:    ErrNotFound,
			Context: err,
			Msg:     fmt.Sprintf("No user with username '%s' exists.", username),
		}
	}

	if err != nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     fmt.Sprintf("Unexpected error trying to look up user ID."),
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: err,
			Msg:     fmt.Sprintf("Unexpected error trying to look up user ID."),
		}
	}

	match := reUserID.FindSubmatch(body)

	if match == nil {
		return 0, &Error{
			Code:    ErrUnexpected,
			Context: errors.New("failed to find userid submatch"),
			Msg:     fmt.Sprintf("Unexpected error trying to look up user ID."),
		}
	}

	id, _ := strconv.Atoi(string(match[1]))

	return id, nil
}
