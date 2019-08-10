package bot

import (
	"fmt"
	"strings"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
	"github.com/bwmarrin/discordgo"
)

// CmdHelp prints some help text for the user
func CmdHelp(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate) {
	p := server.CommandPrefix

	bot.Session.ChannelMessageSend(m.ChannelID,
		"I'm a bot for tracking Etterna plays. https://etternaonline.com\n\n"+
			"For commands, use this prefix: `"+p+"`\n\n"+
			"Command list:\n\n"+
			"**help**\n"+
			"\tShows this help text. Cool.\n\n"+
			"**profile**\n"+
			"\tGets a summary of your current ranks and ratings.\n\n"+
			"**recent [username]**\n"+
			"\tGets a summary of your latest play, or the play of the player you provide.\n\n"+
			"**setuser <username>**\n"+
			"\tLinks an etterna user to you. This will cause your recent plays to be tracked automatically.\n\n"+
			"**unset**\n"+
			"\tUnlinks you from any etterna users. Your recent plays will no longer be tracked.\n\n"+
			"**vs <username> [username]**\n"+
			"\tCompares two user's profiles. Uses your profile if you only give one username.")
}

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
	}

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
func CmdRecentPlay(bot *eb.Bot, m *discordgo.MessageCreate, args []string) {
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
	}

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

// SetUser links a discord user with an etterna user. Only one discord user
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
		getEqualitySign(user1.MSDOverall, user2.MSDOverall), user2.MSDOverall, user1.MSDOverall-user2.MSDOverall)

	description += fmt.Sprintf("    Stream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDStream,
		getEqualitySign(user1.MSDStream, user2.MSDStream), user2.MSDStream, user1.MSDStream-user2.MSDStream)

	description += fmt.Sprintf("Jumpstream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDJumpstream,
		getEqualitySign(user1.MSDJumpstream, user2.MSDJumpstream), user2.MSDJumpstream, user1.MSDJumpstream-user2.MSDJumpstream)

	description += fmt.Sprintf("Handstream:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDHandstream,
		getEqualitySign(user1.MSDHandstream, user2.MSDHandstream), user2.MSDHandstream, user1.MSDHandstream-user2.MSDHandstream)

	description += fmt.Sprintf("   Stamina:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDStamina,
		getEqualitySign(user1.MSDStamina, user2.MSDStamina), user2.MSDStamina, user1.MSDStamina-user2.MSDStamina)

	description += fmt.Sprintf(" JackSpeed:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDJackSpeed,
		getEqualitySign(user1.MSDJackSpeed, user2.MSDJackSpeed), user2.MSDJackSpeed, user1.MSDJackSpeed-user2.MSDJackSpeed)

	description += fmt.Sprintf(" Chordjack:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDChordjack,
		getEqualitySign(user1.MSDChordjack, user2.MSDChordjack), user2.MSDChordjack, user1.MSDChordjack-user2.MSDChordjack)

	description += fmt.Sprintf(" Technical:  %5.2f  %c  %5.2f  (%+.2f)\n", user1.MSDTechnical,
		getEqualitySign(user1.MSDTechnical, user2.MSDTechnical), user2.MSDTechnical, user1.MSDTechnical-user2.MSDTechnical)

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

func getEqualitySign(a, b float64) rune {
	if a > b {
		return '>'
	} else if a < b {
		return '<'
	} else {
		return '='
	}
}
