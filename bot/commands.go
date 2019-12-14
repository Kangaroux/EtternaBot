package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/util"
	"github.com/bwmarrin/discordgo"
)

var (
	reCompareRate = regexp.MustCompile(`compare@(\d*\.?\d*)`)
)

// CmdCompare gets the user's best score for the last song posted in the server
func CmdCompare(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate, args []string) {
	var err error
	var user *model.EtternaUser

	if !server.LastSongID.Valid {
		bot.Session.ChannelMessageSend(m.ChannelID, "No scores to compare to.")
		return
	}

	if len(args) == 1 {
		user, err = bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)
	} else {
		user, err = getUserOrCreate(bot, args[1], false)
	}

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	} else if user == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered with an Etterna user. "+
			"Please register using the `setuser` command, or specify a user: recent <username>")
		return
	}

	bot.Session.ChannelTyping(m.ChannelID)
	song, err := getSongOrCreate(bot, int(server.LastSongID.Int64))

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// Get the top nerf score for this song by this user
	scores, err := bot.API.GetScores(user.EtternaID, song.Name, 50, 0, etterna.SortNerf, false)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	var score *etterna.Score

	// Find the first matching score
	for i, s := range scores {
		if s.Song.ID == int(server.LastSongID.Int64) {
			score = &scores[i]
			break
		}
	}

	if len(scores) == 0 || score == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has no scores on '%s'", user.Username, song.Name))
		return
	}

	details, err := bot.API.GetScoreDetail(score.Key)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	score.MaxCombo = details.MaxCombo
	score.Valid = details.Valid
	score.Date = details.Date
	score.Mods = details.Mods
	score.MinesHit = details.MinesHit

	embed, err := getPlaySummaryAsDiscordEmbed(bot, score, user)
	embed.Author.Name = "Played by " + user.Username

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// CmdCompareRate gets the user's best score for the last song posted in the server
// at a specific rate
func CmdCompareRate(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate, args []string) {
	var err error
	var user *model.EtternaUser

	match := reCompareRate.FindStringSubmatch(args[0])

	fmt.Println(args)

	if match == nil || len(match[1]) == 0 {
		bot.Session.ChannelMessageSend(m.ChannelID, "Usage: compare@<rate> [user]")
		return
	}

	rate, _ := strconv.ParseFloat(match[1], 64)

	if rate < 0.7 || rate > 3.0 {
		bot.Session.ChannelMessageSend(m.ChannelID, "Rate must be between 0.7 and 3.0.")
		return
	}

	split := strings.Split(args[0], ".")

	// Has a decimal part
	if len(split) == 2 {
		// Verify the decimal part is not longer than 2 digits and that if it is 2 digits the
		// leading number is a 0 or a 5.
		if len(split[1]) > 2 || (len(split[1]) == 2 && split[1][1] != '0' && split[1][1] != '5') {
			bot.Session.ChannelMessageSend(m.ChannelID, "Rate must be in 0.05 increments.")
			return
		}
	}

	if !server.LastSongID.Valid {
		bot.Session.ChannelMessageSend(m.ChannelID, "No scores to compare to.")
		return
	}

	if len(args) == 1 {
		user, err = bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)
	} else {
		user, err = getUserOrCreate(bot, args[1], false)
	}

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	} else if user == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered with an Etterna user. "+
			"Please register using the `setuser` command, or specify a user: recent <username>")
		return
	}

	bot.Session.ChannelTyping(m.ChannelID)
	song, err := getSongOrCreate(bot, int(server.LastSongID.Int64))

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// Get the top nerf score for this song by this user
	scores, err := bot.API.GetScores(user.EtternaID, song.Name, 100, 0, etterna.SortNerf, false)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	var score *etterna.Score
	hasAnyScore := false

	// Find the first matching score at this rate
	for i, s := range scores {
		if s.Song.ID == int(server.LastSongID.Int64) {
			hasAnyScore = true

			if s.Rate == rate {
				score = &scores[i]
				break
			}
		}
	}

	rateStr := fmt.Sprintf("%.2f", rate)
	length := len(rateStr)

	// Remove a trailing zero if it exists (0.80 -> 0.8, 1.00 -> 1.0)
	if rateStr[length-1] == '0' {
		rateStr = rateStr[:length-1]
	}

	if len(scores) == 0 || !hasAnyScore {
		bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has no scores on '%s'", user.Username, song.Name))
		return
	} else if hasAnyScore && score == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has no scores on '%s' at %s", user.Username, song.Name, rateStr))
		return
	}

	details, err := bot.API.GetScoreDetail(score.Key)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	score.MaxCombo = details.MaxCombo
	score.Valid = details.Valid
	score.Date = details.Date
	score.Mods = details.Mods
	score.MinesHit = details.MinesHit

	embed, err := getPlaySummaryAsDiscordEmbed(bot, score, user)
	embed.Author.Name = "Played by " + user.Username

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// CmdHelp prints some help text for the user
func CmdHelp(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate) {

	prefix := server.CommandPrefix

	embed := &discordgo.MessageEmbed{
		Title: "EtternaBot Help",
		Description: "I'm a bot for tracking Etterna Online plays. https://etternaonline.com\nFor commands, " +
			"use this prefix: `" + prefix + "`\n\nI can also post score summaries if you send a link to a score.",

		Fields: []*discordgo.MessageEmbedField{

			&discordgo.MessageEmbedField{
				Name:   "**help**",
				Value:  "Shows this help text. Cool.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**setuser** <username>",
				Value:  "Links an Etterna Online user to you. This will cause your recent plays to be tracked automatically.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**unset**",
				Value:  "Unlinks you from any Etterna Online users. Your recent plays will no longer be tracked.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**compare** [username]",
				Value:  "Compares you or someone else's best score on the last posted song.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**compare**@<rate> [username]",
				Value:  "Compares you or someone else's best score on the last posted song at a specific rate. The rate must be a number between 0.7 and 3.0, and it must be in 0.05 increments.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**profile**",
				Value:  "Gets a summary of your current ranks and ratings.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**recent** [username]",
				Value:  "Gets a summary of your latest play, or the play of whichever player you specify.",
				Inline: false,
			},

			&discordgo.MessageEmbedField{
				Name:   "**vs** <username> [username]",
				Value:  "Compares two user's profiles. If you only specify one username, that user's profile will be compared to yours.",
				Inline: false,
			},
		},

		Color: embedColor,
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// CmdProfile displays a user's current rank and ratings
func CmdProfile(bot *eb.Bot, m *discordgo.MessageCreate, args []string) {
	var err error
	var user *model.EtternaUser

	if len(args) == 1 {
		user, err = bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)
	} else if len(args) > 1 {
		user, err = getUserOrCreate(bot, args[1], true)
	}

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	} else if user == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered with an Etterna user. "+
			"Please register using the `setuser` command, or specify a user: recent <username>")
		return
	}

	// The profile command should always show the latest info for the user, however we
	// don't want to cache it since the nerf calc can cause these to change, and the recent
	// plays shows what the calc adjusted
	getLatestUserInfo(bot, user)

	var description string

	description += fmt.Sprintf("➤ **Overall:** %.2f (#%d)\n", user.MSDOverall, user.RankOverall)
	description += fmt.Sprintf("➤ **Stream:** %.2f (#%d)\n", user.MSDStream, user.RankStream)
	description += fmt.Sprintf("➤ **Jumpstream:** %.2f (#%d)\n", user.MSDJumpstream, user.RankJumpstream)
	description += fmt.Sprintf("➤ **Handstream:** %.2f (#%d)\n", user.MSDHandstream, user.RankHandstream)
	description += fmt.Sprintf("➤ **Stamina:** %.2f (#%d)\n", user.MSDStamina, user.RankStamina)
	description += fmt.Sprintf("➤ **JackSpeed:** %.2f (#%d)\n", user.MSDJackSpeed, user.RankJackSpeed)
	description += fmt.Sprintf("➤ **Chordjack:** %.2f (#%d)\n", user.MSDChordjack, user.RankChordjack)
	description += fmt.Sprintf("➤ **Technical:** %.2f (#%d)\n", user.MSDTechnical, user.RankTechnical)

	profileURL := bot.API.BaseURL() + "/user/" + user.Username

	embed := &discordgo.MessageEmbed{
		Description: description,
		Color:       embedColor,
		Title:       "View profile",
		URL:         profileURL,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: "https://i.imgur.com/HwIkGCk.png",
			Name:    "EtternaOnline: " + user.Username,
			URL:     profileURL,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: bot.API.BaseURL() + "/avatars/" + user.Avatar,
		},
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// CmdRecentPlay gets a user's most recent valid play and prints it in the discord channel
func CmdRecentPlay(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate, args []string) {
	var err error
	var user *model.EtternaUser

	if len(args) == 1 {
		user, err = bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)
	} else if len(args) > 1 {
		user, err = getUserOrCreate(bot, args[1], false)
	}

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	} else if user == nil {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered with an Etterna user. "+
			"Please register using the `setuser` command, or specify a user: recent <username>")
		return
	}

	bot.Session.ChannelTyping(m.ChannelID)
	score, err := getRecentPlay(bot, user.EtternaID)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	embed, err := getPlaySummaryAsDiscordEmbed(bot, score, user)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)

	server.LastSongID.Int64 = int64(score.Song.ID)
	server.LastSongID.Valid = true

	bot.Servers.Save(server)
}

// CmdSetScoresChannel sets which discord channel the bot should post scores in
// when tracking recent plays
func CmdSetScoresChannel(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate) {
	server.ScoreChannelID.String = m.ChannelID
	server.ScoreChannelID.Valid = true

	if err := bot.Servers.Save(server); err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
}

// CmdSetUser links a discord user with an etterna user. Only one discord user
// can be linked to a given etterna user at a time in a server. Likewise, discord
// users can only be linked to one etterna user at a time in a server.
func CmdSetUser(bot *eb.Bot, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		bot.Session.ChannelMessageSend(m.ChannelID,
			"Usage: setuser <username>")
		return
	}

	username := strings.TrimSpace(args[1])
	discordID, err := bot.Users.GetRegisteredDiscordUserID(m.GuildID, username)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if discordID == m.Author.ID {
		bot.Session.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("You are already registered as '%s'.", username))
		return
	} else if discordID != "" {
		bot.Session.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Another user is already registered as '%s'.", username))
		return
	}

	// Get the etterna user registered with this discord user
	user, err := bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if user != nil {
		bot.Session.ChannelMessageSend(m.ChannelID,
			"You are already registered as another user. Use the 'unset' "+
				"command first and try again.")
		return
	}

	// The discord user is not associated with any etterna users, look up
	// the etterna user with that username
	user, err = getUserOrCreate(bot, username, false)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	ok, err := bot.Users.Register(user.Username, m.GuildID, m.Author.ID)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if !ok {
		// This can only happen due to a data race, still worth checking though
		bot.Session.ChannelMessageSend(m.ChannelID,
			"You are currently registered as another user. Use the 'unset' command "+
				"first and try again.")
	} else {
		bot.Session.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Success! You are now registered as '%s'.", user.Username))
	}
}

// CmdUnsetUser unregisters the given discord user from an etterna user
func CmdUnsetUser(bot *eb.Bot, m *discordgo.MessageCreate) {
	ok, err := bot.Users.Unregister(m.GuildID, m.Author.ID)

	if err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if ok {
		bot.Session.ChannelMessageSend(m.ChannelID,
			"Success! You are no longer registered. Use the setuser command to register "+
				"as another user.")
	} else {
		bot.Session.ChannelMessageSend(m.ChannelID, "You are not registered to an etterna user.")
	}
}

// CmdVersus compares the profiles of two users
func CmdVersus(bot *eb.Bot, m *discordgo.MessageCreate, args []string) {
	var err error
	var user1, user2 *model.EtternaUser

	if len(args) == 1 {
		bot.Session.ChannelMessageSend(m.ChannelID,
			"Usage: vs <username> [username]")
		return
	}

	if len(args) == 2 {
		user1, err = bot.Users.GetRegisteredUser(m.GuildID, m.Author.ID)

		if err != nil {
			bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		user2, err = getUserOrCreate(bot, args[1], true)

		if err != nil {
			bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
	} else {
		user1, err = getUserOrCreate(bot, args[1], true)

		if err != nil {
			bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		user2, err = getUserOrCreate(bot, args[2], true)

		if err != nil {
			bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
	}

	var description string

	description += fmt.Sprintf("   Overall:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDOverall,
		util.GetEqualitySign(user1.MSDOverall, user2.MSDOverall), user2.MSDOverall, user1.MSDOverall-user2.MSDOverall)

	description += fmt.Sprintf("    Stream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDStream,
		util.GetEqualitySign(user1.MSDStream, user2.MSDStream), user2.MSDStream, user1.MSDStream-user2.MSDStream)

	description += fmt.Sprintf("Jumpstream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDJumpstream,
		util.GetEqualitySign(user1.MSDJumpstream, user2.MSDJumpstream), user2.MSDJumpstream, user1.MSDJumpstream-user2.MSDJumpstream)

	description += fmt.Sprintf("Handstream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDHandstream,
		util.GetEqualitySign(user1.MSDHandstream, user2.MSDHandstream), user2.MSDHandstream, user1.MSDHandstream-user2.MSDHandstream)

	description += fmt.Sprintf("   Stamina:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDStamina,
		util.GetEqualitySign(user1.MSDStamina, user2.MSDStamina), user2.MSDStamina, user1.MSDStamina-user2.MSDStamina)

	description += fmt.Sprintf(" JackSpeed:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDJackSpeed,
		util.GetEqualitySign(user1.MSDJackSpeed, user2.MSDJackSpeed), user2.MSDJackSpeed, user1.MSDJackSpeed-user2.MSDJackSpeed)

	description += fmt.Sprintf(" Chordjack:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDChordjack,
		util.GetEqualitySign(user1.MSDChordjack, user2.MSDChordjack), user2.MSDChordjack, user1.MSDChordjack-user2.MSDChordjack)

	description += fmt.Sprintf(" Technical:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDTechnical,
		util.GetEqualitySign(user1.MSDTechnical, user2.MSDTechnical), user2.MSDTechnical, user1.MSDTechnical-user2.MSDTechnical)

	embed := &discordgo.MessageEmbed{
		Description: "```\n" + description + "\n```",
		Color:       embedColor,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: "https://i.imgur.com/HwIkGCk.png",
			Name:    user1.Username + " vs. " + user2.Username,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/AkfAZtJ.png",
		},
	}

	bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)
}
