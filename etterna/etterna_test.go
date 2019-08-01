package etterna

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	api := New("test")

	require.Equal(t, "test", api.apiKey)
	require.Equal(t, baseAPIURL, api.baseAPIURL)
	require.Equal(t, baseURL, api.baseURL)
}

func TestGetByUsername(t *testing.T) {
	t.Run("should visit the correct URL", func(t *testing.T) {
		ok := make(chan bool, 1)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1/user_data?api_key=testkey&username=jesse", r.URL.RequestURI())
			ok <- true
		}))

		defer server.Close()

		api := New("testkey")
		api.baseAPIURL = server.URL + "/v1"

		api.GetByUsername("jesse")

		select {
		case <-ok:
		case <-time.After(time.Second):
			t.FailNow()
		}
	})

	t.Run("should error when not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer server.Close()

		api := New("testkey")
		api.baseAPIURL = server.URL + "/v1"

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
		api.baseAPIURL = server.URL + "/v1"

		user, err := api.GetByUsername("jesse")

		require.NoError(t, err)
		require.Equal(t, User{
			Username:    "jesse",
			CountryCode: "US",
			AvatarURL:   "avatar.jpg",
			MSD: MSD{
				Overall:    1.10,
				Stream:     2.20,
				Jumpstream: 3.30,
				Handstream: 4.40,
				Stamina:    5.50,
				JackSpeed:  6.60,
				Chordjack:  7.70,
				Technical:  8.80,
			},
		}, *user)
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("should visit the correct URL", func(t *testing.T) {
		ok := make(chan bool, 1)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/user/jesse", r.URL.RequestURI())
			ok <- true
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		api.GetUserID("jesse")

		select {
		case <-ok:
		case <-time.After(time.Second):
			t.FailNow()
		}
	})

	t.Run("should error when not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		_, err := api.GetUserID("jesse")

		require.Error(t, err)
		require.Equal(t, ErrNotFound, err.(*Error).Code)
	})

	t.Run("should return user ID on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`
				<!DOCTYPE html>
				<head></head>
				<body>
					<script>
						var blah = {
							'userid': '123'
						};
					</script>
				</body>
			`))
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		id, err := api.GetUserID("jesse")

		require.NoError(t, err)
		require.Equal(t, 123, id)
	})
}

func TestGetScores(t *testing.T) {
	t.Run("should send the right request", func(t *testing.T) {
		start := 10
		userID := 1234
		length := 25
		sortColumn := SortDate
		sortAsc := false

		ok := make(chan bool, 1)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/score/userScores", r.URL.RequestURI())

			r.ParseForm()

			require.Equal(t, strconv.Itoa(int(start)), r.PostForm.Get("start"))
			require.Equal(t, strconv.Itoa(int(length)), r.PostForm.Get("length"))
			require.Equal(t, strconv.Itoa(int(userID)), r.PostForm.Get("userid"))
			require.Equal(t, strconv.Itoa(int(sortColumn)), r.PostForm.Get("order[0][column]"))
			require.Equal(t, "desc", r.PostForm.Get("order[0][dir]"))

			ok <- true
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		api.GetScores(userID, uint(length), uint(start), sortColumn, sortAsc)

		select {
		case <-ok:
		case <-time.After(time.Second):
			t.FailNow()
		}
	})

	t.Run("should error when not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		_, err := api.GetScores(123, 25, 0, SortAccuracy, true)

		require.Error(t, err)
		require.Equal(t, ErrNotFound, err.(*Error).Code)
	})

	t.Run("should return scores on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"draw":0,"recordsTotal":2304,"recordsFiltered":2304,"data":[{"addDeleteScore":"\n                        <div class='add-showcase'><i class='fas fa-plus-circle add-score'><\/i><\/div>\n                    ","songname":"<a href=\"https:\/\/etternaonline.com\/song\/view\/2254\">ETERNAL DRAIN<\/a>","user_chart_rate_rate":"0.80","Overall":"<a href=\"https:\/\/etternaonline.com\/score\/view\/S03d9aef6758d50f60dedc2fc6b855dd22aa6b1e24118\">19.84<\/a>","Nerf":19.84,"wifescore":"<div data-toggle='tooltip' data-html='true' data-placement='right' \n                                                title='Marvelous: 1489<br\/>\n                                                Perfect: 509<br\/>\n                                                Great: 162<br\/>\n                                                Good: 21<br\/>\n                                                Bad: 12<br\/>\n                                                Miss: 16<br\/>\n                                                '\n                                                ><span class='a'>89.89%<\/span><\/div>","datetime":"2019-07-31","stream":"17.06","jumpstream":"19.84","handstream":"14.6","stamina":"18.56","jackspeed":"16.67","chordjack":"11.09","technical":"19.1","nocc":"Off","scorekey":"S03d9aef6758d50f60dedc2fc6b855dd22aa6b1e2"},{"addDeleteScore":"\n                        <div class='add-showcase'><i class='fas fa-plus-circle add-score'><\/i><\/div>\n                    ","songname":"<a href=\"https:\/\/etternaonline.com\/song\/view\/65980\">Bagpipe<\/a>","user_chart_rate_rate":"1.00","Overall":"<a href=\"https:\/\/etternaonline.com\/score\/view\/Sd52bb9428e7551ba527418d787e8906dc0d33c6a4118\">21.01<\/a>","Nerf":21.01,"wifescore":"<div data-toggle='tooltip' data-html='true' data-placement='right' \n                                                title='Marvelous: 880<br\/>\n                                                Perfect: 283<br\/>\n                                                Great: 75<br\/>\n                                                Good: 10<br\/>\n                                                Bad: 11<br\/>\n                                                Miss: 9<br\/>\n                                                '\n                                                ><span class='a'>89.22%<\/span><\/div>","datetime":"2019-07-31","stream":"21.01","jumpstream":"14.94","handstream":"12.62","stamina":"17.52","jackspeed":"15.99","chordjack":"10.63","technical":"18.4","nocc":"Off","scorekey":"Sd52bb9428e7551ba527418d787e8906dc0d33c6a"},{"addDeleteScore":"\n                        <div class='add-showcase'><i class='fas fa-plus-circle add-score'><\/i><\/div>\n                    ","songname":"<a href=\"https:\/\/etternaonline.com\/song\/view\/9835\">115<\/a>","user_chart_rate_rate":"0.90","Overall":"<a href=\"https:\/\/etternaonline.com\/score\/view\/S8006c8f77cafc768d10b98486988063cbbd9d1434118\">18.67<\/a>","Nerf":18.48,"wifescore":"<div data-toggle='tooltip' data-html='true' data-placement='right' \n                                                title='Marvelous: 2112<br\/>\n                                                Perfect: 419<br\/>\n                                                Great: 105<br\/>\n                                                Good: 18<br\/>\n                                                Bad: 7<br\/>\n                                                Miss: 9<br\/>\n                                                '\n                                                ><span class='aa'>94.51%<\/span><\/div>","datetime":"2019-07-31","stream":"16.87","jumpstream":"18.67","handstream":"15.56","stamina":"17.79","jackspeed":"14.29","chordjack":"9.57","technical":"18.64","nocc":"Off","scorekey":"S8006c8f77cafc768d10b98486988063cbbd9d143"},{"addDeleteScore":"\n                        <div class='add-showcase'><i class='fas fa-plus-circle add-score'><\/i><\/div>\n                    ","songname":"<a href=\"https:\/\/etternaonline.com\/song\/view\/9863\">The Gears<\/a>","user_chart_rate_rate":"1.00","Overall":"<a href=\"https:\/\/etternaonline.com\/score\/view\/Sd34b00b7c0372b225b1f49e90e2c0f630b67cd4b4118\">19.72<\/a>","Nerf":19.72,"wifescore":"<div data-toggle='tooltip' data-html='true' data-placement='right' \n                                                title='Marvelous: 2362<br\/>\n                                                Perfect: 776<br\/>\n                                                Great: 258<br\/>\n                                                Good: 33<br\/>\n                                                Bad: 13<br\/>\n                                                Miss: 42<br\/>\n                                                '\n                                                ><span class='a'>87.81%<\/span><\/div>","datetime":"2019-07-31","stream":"18.2","jumpstream":"17.44","handstream":"16.73","stamina":"19.72","jackspeed":"16.41","chordjack":"10.89","technical":"16.93","nocc":"Off","scorekey":"Sd34b00b7c0372b225b1f49e90e2c0f630b67cd4b"},{"addDeleteScore":"\n                        <div class='add-showcase'><i class='fas fa-plus-circle add-score'><\/i><\/div>\n                    ","songname":"<a href=\"https:\/\/etternaonline.com\/song\/view\/12957\">300<\/a>","user_chart_rate_rate":"1.00","Overall":"<a href=\"https:\/\/etternaonline.com\/score\/view\/S631326cd374dd997a2dd5dcf1027457e43aaceaa4118\">17.51<\/a>","Nerf":17.51,"wifescore":"<div data-toggle='tooltip' data-html='true' data-placement='right' \n                                                title='Marvelous: 1278<br\/>\n                                                Perfect: 336<br\/>\n                                                Great: 113<br\/>\n                                                Good: 22<br\/>\n                                                Bad: 6<br\/>\n                                                Miss: 17<br\/>\n                                                '\n                                                ><span class='a'>88.92%<\/span><\/div>","datetime":"2019-07-31","stream":"17.51","jumpstream":"16.48","handstream":"12.35","stamina":"16.69","jackspeed":"15.17","chordjack":"10.07","technical":"17.5","nocc":"Off","scorekey":"S631326cd374dd997a2dd5dcf1027457e43aaceaa"}]}`))
		}))

		defer server.Close()

		api := New("testkey")
		api.baseURL = server.URL

		// Stubbed data so the args don't matter
		scores, err := api.GetScores(0, 0, 0, 0, false)

		require.NoError(t, err)
		require.Equal(t, 5, len(scores))

		// Testing 1 of the 5 because I'm too lazy to do this for two scores
		require.Equal(t, "S03d9aef6758d50f60dedc2fc6b855dd22aa6b1e2", scores[0].Key)
		require.Equal(t, "ETERNAL DRAIN", scores[0].Song.Name)
		require.Equal(t, 2254, scores[0].Song.ID)
		require.Equal(t, "2019-07-31", scores[0].Date.Format("2006-01-02"))
		require.Equal(t, 89.89, scores[0].Accuracy)
		require.Equal(t, 0.80, scores[0].Rate)
		require.Equal(t, 1489, scores[0].Marvelous)
		require.Equal(t, 509, scores[0].Perfect)
		require.Equal(t, 162, scores[0].Great)
		require.Equal(t, 21, scores[0].Good)
		require.Equal(t, 12, scores[0].Bad)
		require.Equal(t, 16, scores[0].Miss)
		require.Equal(t, 19.84, scores[0].Nerfed)
		require.Equal(t, 19.84, scores[0].Overall)
		require.Equal(t, 17.06, scores[0].Stream)
		require.Equal(t, 19.84, scores[0].Jumpstream)
		require.Equal(t, 14.60, scores[0].Handstream)
		require.Equal(t, 18.56, scores[0].Stamina)
		require.Equal(t, 16.67, scores[0].JackSpeed)
		require.Equal(t, 11.09, scores[0].Chordjack)
		require.Equal(t, 19.10, scores[0].Technical)
	})
}
