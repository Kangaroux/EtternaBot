package bot

import (
	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
)

// getUserOrCreate returns the etterna user with the given username, inserting the user into the
// database automatically if they don't already exist
func getUserOrCreate(bot *eb.Bot, username string) (*model.EtternaUser, error) {
	user, err := bot.Users.GetUsername(username)

	if err != nil {
		return nil, err
	} else if user != nil {
		return user, nil
	}

	etternaUser, err := bot.API.GetByUsername(username)

	if err != nil {
		return nil, err
	}

	id, err := bot.API.GetUserID(username)

	if err != nil {
		return nil, err
	}

	user = &model.EtternaUser{
		Username:      etternaUser.Username,
		EtternaID:     id,
		Avatar:        etternaUser.AvatarURL,
		MSDOverall:    etternaUser.MSD.Overall,
		MSDStream:     etternaUser.MSD.Stream,
		MSDJumpstream: etternaUser.MSD.Jumpstream,
		MSDHandstream: etternaUser.MSD.Handstream,
		MSDStamina:    etternaUser.MSD.Stamina,
		MSDJackSpeed:  etternaUser.MSD.JackSpeed,
		MSDChordjack:  etternaUser.MSD.Chordjack,
		MSDTechnical:  etternaUser.MSD.Technical,
	}

	if err := bot.Users.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
