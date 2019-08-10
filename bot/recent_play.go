package bot

import (
	"fmt"
	"time"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/bwmarrin/discordgo"
)

const (
	recentPlayLookupCount = 10
)

// getRecentPlay looks up the most recent, valid play for a user.
func getRecentPlay(bot *eb.Bot, etternaID int) (*etterna.Score, error) {
	scores, err := bot.API.GetScores(etternaID, recentPlayLookupCount, 0, etterna.SortDate, false)

	if err != nil {
		fmt.Println("Failed to look up recent scores", err)
		return nil, err
	}

	// User has no recent, valid scores
	if len(scores) == 0 {
		return nil, nil
	}

	s := scores[0]
	details, err := bot.API.GetScoreDetail(s.Key)

	if err != nil {
		fmt.Println("Failed to look up score", s.Key, err)
		return nil, err
	}

	s.MaxCombo = details.MaxCombo
	s.MinesHit = details.MinesHit
	s.Mods = details.Mods

	return &s, nil
}

// getPlaySummaryAsDiscordEmbed returns a discord embed object for displaying the score
func getPlaySummaryAsDiscordEmbed(bot *eb.Bot, score *etterna.Score, user *model.EtternaUser) (*discordgo.MessageEmbed, error) {
	song, err := getSongOrCreate(bot, score.Song.ID)

	if err != nil {
		fmt.Println("Failed to get song details", score.Song.ID, err)
		return nil, err
	}

	score.Song.Name = song.Name
	score.Song.Artist = song.Artist
	score.Song.BackgroundURL = song.BackgroundURL
	rateStr := fmt.Sprintf("%.2f", score.Rate)
	length := len(rateStr)

	// Remove a trailing zero if it exists (0.80 -> 0.8, 1.00 -> 1.0)
	if rateStr[length-1] == '0' {
		rateStr = rateStr[:length-1]
	}

	scoreURL := fmt.Sprintf(bot.API.BaseURL()+"/score/view/%s%d", score.Key, user.EtternaID)
	description := fmt.Sprintf(
		"**[%s (%sx)](%s)**\n\n"+
			"➤ **Acc:** %.2f%% @ %sx\n"+
			"➤ **Score:** %.2f | **Nerfed:** %.2f\n"+
			"➤ **Hits:** %d/%d/%d/%d/%d/%d\n"+
			"➤ **Max combo:** x%d",
		score.Song.Name,
		rateStr,
		scoreURL,
		score.Accuracy,
		rateStr,
		score.MSD.Overall,
		score.Nerfed,
		score.Marvelous,
		score.Perfect,
		score.Great,
		score.Good,
		score.Bad,
		score.Miss,
		score.MaxCombo)

	msg := &discordgo.MessageEmbed{
		URL: scoreURL,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Recent play by " + user.Username,
			IconURL: bot.API.BaseURL() + "/avatars/" + user.Avatar,
		},
		Color:       embedColor,
		Description: description,
		Timestamp:   score.Date.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://i.imgur.com/HwIkGCk.png",
			Text:    user.Username,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: bot.API.BaseURL() + "/song_images/bg/" + score.Song.BackgroundURL,
		},
	}

	return msg, nil
}
