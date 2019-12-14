package bot

import (
	"fmt"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/util"
)

// TrackAllRecentPlays gets recent plays for all registered etterna users and
// if there was a new recent play, prints the play in the scores channel of
// all servers that user is registered in
func TrackAllRecentPlays(bot *eb.Bot, minAcc float64) {
	serversToUpdate := make(map[uint]model.DiscordServer)
	users, err := bot.Users.GetRegisteredUsersForRecentPlays()

	if err != nil {
		fmt.Println("Failed to look up users for recent plays")
		return
	}

	for _, v := range users {
		s, err := getRecentPlay(bot, v.User.EtternaID)

		// Score isn't valid or we've already tracked this play. If you play the same song back-to-back
		// EO will sometimes overwrite the old score we also need to track when the score was recorded
		// so we know if it was overwritten or not
		if err != nil ||
			s == nil ||
			(v.User.LastRecentScoreKey.Valid && s.Key == v.User.LastRecentScoreKey.String &&
				v.User.LastRecentScoreDate != nil &&
				s.Date.Equal(*v.User.LastRecentScoreDate)) {
			continue
		}

		// Get the latest ratings of this user from the etterna API so we can compare with
		// the old rating we saved and see if the user gained rating from the play
		latestUser, err := bot.API.GetByUsername(v.User.Username)

		if err != nil {
			fmt.Println("Failed to look up recent user", v.User.Username, err)
			return
		}

		latestUser.Overall = util.RoundToPrecision(latestUser.Overall, 2)
		latestUser.Stream = util.RoundToPrecision(latestUser.Stream, 2)
		latestUser.Jumpstream = util.RoundToPrecision(latestUser.Jumpstream, 2)
		latestUser.Handstream = util.RoundToPrecision(latestUser.Handstream, 2)
		latestUser.Stamina = util.RoundToPrecision(latestUser.Stamina, 2)
		latestUser.JackSpeed = util.RoundToPrecision(latestUser.JackSpeed, 2)
		latestUser.Chordjack = util.RoundToPrecision(latestUser.Chordjack, 2)
		latestUser.Technical = util.RoundToPrecision(latestUser.Technical, 2)

		diffMSD := etterna.MSD{
			Overall:    latestUser.Overall - v.User.MSDOverall,
			Stream:     latestUser.Stream - v.User.MSDStream,
			Jumpstream: latestUser.Jumpstream - v.User.MSDJumpstream,
			Handstream: latestUser.Handstream - v.User.MSDHandstream,
			Stamina:    latestUser.Stamina - v.User.MSDStamina,
			JackSpeed:  latestUser.JackSpeed - v.User.MSDJackSpeed,
			Chordjack:  latestUser.Chordjack - v.User.MSDChordjack,
			Technical:  latestUser.Technical - v.User.MSDTechnical,
		}

		v.User.MSDOverall = latestUser.Overall
		v.User.MSDStream = latestUser.Stream
		v.User.MSDJumpstream = latestUser.Jumpstream
		v.User.MSDHandstream = latestUser.Handstream
		v.User.MSDStamina = latestUser.Stamina
		v.User.MSDJackSpeed = latestUser.JackSpeed
		v.User.MSDChordjack = latestUser.Chordjack
		v.User.MSDTechnical = latestUser.Technical
		v.User.RankOverall = latestUser.Rank.Overall
		v.User.RankStream = latestUser.Rank.Stream
		v.User.RankJumpstream = latestUser.Rank.Jumpstream
		v.User.RankHandstream = latestUser.Rank.Handstream
		v.User.RankStamina = latestUser.Rank.Stamina
		v.User.RankJackSpeed = latestUser.Rank.JackSpeed
		v.User.RankChordjack = latestUser.Rank.Chordjack
		v.User.RankTechnical = latestUser.Rank.Technical
		v.User.LastRecentScoreKey.String = s.Key
		v.User.LastRecentScoreKey.Valid = true
		v.User.LastRecentScoreDate = &s.Date

		bot.Users.Save(&v.User)

		gains := ""

		if diffMSD.Overall >= 0.01 {
			gains += fmt.Sprintf("➤ **Overall:** %.2f (+%.2f)\n", latestUser.Overall, diffMSD.Overall)
		}

		if diffMSD.Stream >= 0.01 {
			gains += fmt.Sprintf("➤ **Stream:** %.2f (+%.2f)\n", latestUser.Stream, diffMSD.Stream)
		}

		if diffMSD.Jumpstream >= 0.01 {
			gains += fmt.Sprintf("➤ **Jumpstream:** %.2f (+%.2f)\n", latestUser.Jumpstream, diffMSD.Jumpstream)
		}

		if diffMSD.Handstream >= 0.01 {
			gains += fmt.Sprintf("➤ **Handstream:** %.2f (+%.2f)\n", latestUser.Handstream, diffMSD.Handstream)
		}

		if diffMSD.Stamina >= 0.01 {
			gains += fmt.Sprintf("➤ **Stamina:** %.2f (+%.2f)\n", latestUser.Stamina, diffMSD.Stamina)
		}

		if diffMSD.JackSpeed >= 0.01 {
			gains += fmt.Sprintf("➤ **JackSpeed:** %.2f (+%.2f)\n", latestUser.JackSpeed, diffMSD.JackSpeed)
		}

		if diffMSD.Chordjack >= 0.01 {
			gains += fmt.Sprintf("➤ **Chordjack:** %.2f (+%.2f)\n", latestUser.Chordjack, diffMSD.Chordjack)
		}

		if diffMSD.Technical >= 0.01 {
			gains += fmt.Sprintf("➤ **Technical:** %.2f (+%.2f)\n", latestUser.Technical, diffMSD.Technical)
		}

		// Only display the song if the player got above a certain acc or if they gained pp
		if gains == "" && s.Accuracy < minAcc {
			continue
		}

		for _, server := range v.Servers {
			embed, err := getPlaySummaryAsDiscordEmbed(bot, s, &v.User)

			if err != nil {
				continue
			}

			if gains != "" {
				embed.Description += "\n\n" + gains
			}

			bot.Session.ChannelMessageSendEmbed(server.ScoreChannelID.String, embed)

			server.LastSongID.Int64 = int64(s.Song.ID)
			server.LastSongID.Valid = true

			serversToUpdate[server.ID] = server
		}
	}

	// Update all of the servers that we set the LastSongID field on
	for _, s := range serversToUpdate {
		bot.Servers.Save(&s)
	}
}
