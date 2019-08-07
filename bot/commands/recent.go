package commands

import (
	"fmt"
	"time"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/util"
	"github.com/bwmarrin/discordgo"
)

func RecentPlay(bot *eb.Bot, m *discordgo.MessageCreate) {
	user, err := bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	} else if user == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered with an Etterna user. "+
			"Please register using the `setuser` command, or specify a user: recent <username>")
	}
}

func getPlay(bot *eb.Bot, scoreID string) {
	// score, err := bot.API.GetScoreDetail(scoreID)

	// if err != nil {
	// 	fmt.Println("Failed to look up score", scoreID, err)
	// 	return
	// }

	// s := scores[0]

	// // We've already seen this score
	// if v.User.LastRecentScoreKey.Valid && s.Key == v.User.LastRecentScoreKey.String {
	// 	fmt.Println("No new scores", s.Key)
	// 	continue
	// }

	// if err := bot.API.GetScoreDetail(&s); err != nil {
	// 	fmt.Println("Failed to get score details", s.Key, err)
	// 	return
	// }
}

func TrackRecentPlays(bot *eb.Bot, minAcc float64) {
	users, err := bot.Users.GetRegisteredUsersForRecentPlays()

	if err != nil {
		fmt.Println("Failed to look up users for recent plays")
		return
	}

	for _, v := range users {
		scores, err := bot.API.GetScores(v.User.EtternaID, 1, 0, etterna.SortDate, false)

		if err != nil {
			fmt.Println("Failed to look up recent scores", v.User.Username, err)
			return
		}

		s := scores[0]

		// We've already seen this score
		if v.User.LastRecentScoreKey.Valid && s.Key == v.User.LastRecentScoreKey.String {
			fmt.Println("No new scores", s.Key)
			continue
		}

		if err := bot.API.GetScoreDetail(&s); err != nil {
			fmt.Println("Failed to get score details", s.Key, err)
			return
		}

		latestUser, err := bot.API.GetByUsername(v.User.Username)

		if err != nil {
			fmt.Println("Failed to look up recent user", v.User.Username, err)
			return
		}

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

		v.User.MSDOverall = util.TruncateFloat(latestUser.Overall, 2)
		v.User.MSDStream = util.TruncateFloat(latestUser.Stream, 2)
		v.User.MSDJumpstream = util.TruncateFloat(latestUser.Jumpstream, 2)
		v.User.MSDHandstream = util.TruncateFloat(latestUser.Handstream, 2)
		v.User.MSDStamina = util.TruncateFloat(latestUser.Stamina, 2)
		v.User.MSDJackSpeed = util.TruncateFloat(latestUser.JackSpeed, 2)
		v.User.MSDChordjack = util.TruncateFloat(latestUser.Chordjack, 2)
		v.User.MSDTechnical = util.TruncateFloat(latestUser.Technical, 2)
		v.User.LastRecentScoreKey.String = s.Key
		v.User.LastRecentScoreKey.Valid = true

		bot.Users.Save(&v.User)

		// If the score is invalid don't post it
		if !s.Valid {
			fmt.Println("Score isn't valid")
			continue
		}

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

		song, err := bot.API.GetSong(s.Song.ID)

		if err != nil {
			fmt.Println("Failed to get song details", song.ID, err)
			return
		}

		s.Song = *song
		rateStr := fmt.Sprintf("%.2f", s.Rate)
		length := len(rateStr)

		// Remove a trailing zero if it exists (0.80 -> 0.8, 1.00 -> 1.0)
		if rateStr[length-1] == '0' {
			rateStr = rateStr[:length-1]
		}

		scoreURL := fmt.Sprintf("https://etternaonline.com/score/view/%s%d", s.Key, v.User.EtternaID)
		description := fmt.Sprintf(
			"**[%s (%sx)](%s)**\n\n"+
				"➤ **Acc:** %.2f%% @ %sx\n"+
				"➤ **Score:** %.2f | **Nerfed:** %.2f\n"+
				"➤ **Hits:** %d/%d/%d/%d/%d/%d\n"+
				"➤ **Max combo:** x%d",
			s.Song.Name,
			rateStr,
			scoreURL,
			s.Accuracy,
			rateStr,
			s.MSD.Overall,
			s.Nerfed,
			s.Marvelous,
			s.Perfect,
			s.Great,
			s.Good,
			s.Bad,
			s.Miss,
			s.MaxCombo)

		if gains != "" {
			description += "\n\n" + gains
		}

		msg := &discordgo.MessageEmbed{
			URL: scoreURL,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "Recent play by " + v.User.Username,
				IconURL: "https://etternaonline.com/avatars/" + v.User.Avatar,
			},
			Color:       8519899,
			Description: description,
			Timestamp:   s.Date.Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				IconURL: "https://i.imgur.com/HwIkGCk.png",
				Text:    v.User.Username,
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://etternaonline.com/song_images/bg/" + s.Song.BackgroundURL,
			},
		}

		for _, server := range v.Servers {
			bot.Session.ChannelMessageSendEmbed(server.ScoreChannelID.String, msg)
		}
	}
}
