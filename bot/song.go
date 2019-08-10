package bot

import (
	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
)

// getSongOrCreate looks up a song in the database by its etterna ID, and retrieves it
// from the API if it doesn't exist
func getSongOrCreate(bot *eb.Bot, id int) (*model.Song, error) {
	song, err := bot.Songs.Get(id)

	if err != nil {
		return nil, err
	} else if song != nil {
		return song, nil
	}

	etternaSong, err := bot.API.GetSong(id)

	if err != nil {
		return nil, err
	}

	song = &model.Song{
		EtternaID:     etternaSong.ID,
		Artist:        etternaSong.Artist,
		Name:          etternaSong.Name,
		BackgroundURL: etternaSong.BackgroundURL,
	}

	if err := bot.Songs.Save(song); err != nil {
		return nil, err
	}

	return song, nil
}
