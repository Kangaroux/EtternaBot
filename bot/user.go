package bot

import (
	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/util"
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

// getLatestUserInfo gets the latest MSD, ranks, and avatar for the user
func getLatestUserInfo(bot *eb.Bot, user *model.EtternaUser) error {
	etternaUser, err := bot.API.GetByUsername(user.Username)

	if err != nil {
		return err
	}

	user.Avatar = etternaUser.AvatarURL
	user.MSDOverall = util.TruncateFloat(etternaUser.Overall, 2)
	user.MSDStream = util.TruncateFloat(etternaUser.Stream, 2)
	user.MSDJumpstream = util.TruncateFloat(etternaUser.Jumpstream, 2)
	user.MSDHandstream = util.TruncateFloat(etternaUser.Handstream, 2)
	user.MSDStamina = util.TruncateFloat(etternaUser.Stamina, 2)
	user.MSDJackSpeed = util.TruncateFloat(etternaUser.JackSpeed, 2)
	user.MSDChordjack = util.TruncateFloat(etternaUser.Chordjack, 2)
	user.MSDTechnical = util.TruncateFloat(etternaUser.Technical, 2)
	user.RankOverall = etternaUser.Rank.Overall
	user.RankStream = etternaUser.Rank.Stream
	user.RankJumpstream = etternaUser.Rank.Jumpstream
	user.RankHandstream = etternaUser.Rank.Handstream
	user.RankStamina = etternaUser.Rank.Stamina
	user.RankJackSpeed = etternaUser.Rank.JackSpeed
	user.RankChordjack = etternaUser.Rank.Chordjack
	user.RankTechnical = etternaUser.Rank.Technical

	return nil
}
