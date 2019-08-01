package etterna

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	api := New("test")

	require.Equal(t, "test", api.apiKey)
	require.Equal(t, baseURL, api.baseURL)
}

func TestGetUsername(t *testing.T) {
	t.Run("should visit the correct URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1/user_data?api_key=testkey&username=jesse", r.URL.RequestURI())
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL + "/v1"

		api.GetByUsername("jesse")
	})

	t.Run("should error when not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL + "/v1"

		user, err := api.GetByUsername("jesse")

		require.Error(t, err)
		require.Equal(t, ErrNotFound, err.(*Error).Code)
		require.Nil(t, user)
	})

	t.Run("should return user on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, _ := json.Marshal(map[string]string{
				"username":    "jesse",
				"countrycode": "US",
				"avatar":      "avatar.jpg",
				"Overall":     "1.10",
				"Stream":      "2.20",
				"Jumpstream":  "3.30",
				"Handstream":  "4.40",
				"Stamina":     "5.50",
				"JackSpeed":   "6.60",
				"Chordjack":   "7.70",
				"Technical":   "8.80",
			})

			w.Write(data)
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL + "/v1"

		user, err := api.GetByUsername("jesse")

		require.NoError(t, err)
		require.Equal(t, User{
			Username:    "jesse",
			CountryCode: "US",
			AvatarURL:   "avatar.jpg",
			Overall:     1.10,
			Stream:      2.20,
			Jumpstream:  3.30,
			Handstream:  4.40,
			Stamina:     5.50,
			JackSpeed:   6.60,
			Chordjack:   7.70,
			Technical:   8.80,
		}, *user)
	})
}
