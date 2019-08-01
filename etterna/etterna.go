package etterna

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	baseURL = "https://api.etternaonline.com/v1"
)

type EtternaAPIInterface interface {
	GetUserID(username string) (int, error)
	GetByUsername(username string) (*User, error)
	GetScores()
}

func New(apiKey string) EtternaAPI {
	return EtternaAPI{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

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

	resp, err := http.Get(api.baseURL + "/user_data" +
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
