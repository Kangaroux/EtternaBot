package bot

import (
	"fmt"
	"strings"
	"time"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/util"
	"github.com/Kangaroux/etternabot/bot/commands"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/model/service"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

const (
	defaultPrefix           = ";"             // Prefix for commands
	defaultRecentPlayMinAcc = 97.0            // Minimum acc to display a recent play
	recentPlayInterval      = 1 * time.Minute // How often to check for recent plays
)

func New(s *discordgo.Session, db *sqlx.DB, etternaAPIKey string) eb.Bot {
	bot := eb.Bot{
		DB:      db,
		API:     etterna.New(etternaAPIKey),
		Session: s,
		Servers: service.NewDiscordServerService(db),
		Users:   service.NewUserService(db),
	}

	s.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		guildCreate(&bot, g)
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		messageCreate(&bot, m)
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		trackRecentPlays(&bot)
	})

	return bot
}

func guildCreate(bot *eb.Bot, g *discordgo.GuildCreate) {
	server, err := bot.Servers.Get(g.ID)

	if err != nil {
		fmt.Println("ERROR: Failed to load guild info", g.ID, g.Name)
		return
	}

	if server == nil {
		server = &model.DiscordServer{
			CommandPrefix: defaultPrefix,
			ServerID:      g.ID,
		}

		if err := bot.Servers.Save(server); err != nil {
			fmt.Println("ERROR: Failed to insert guild into db", g.ID, g.Name, err)
			return
		} else {
			fmt.Println("Created record for server", g.Name)
		}
	}
}

func messageCreate(bot *eb.Bot, m *discordgo.MessageCreate) {
	if m.Author.ID == bot.Session.State.User.ID {
		return
	}

	server, err := bot.Servers.Get(m.GuildID)

	if err != nil {
		fmt.Println("Error looking up server", m.GuildID, err)
		return
	} else if server == nil {
		// Likely a DM, need to handle differently
		fmt.Println("Unknown server", m.GuildID)
		return
	}

	if !strings.HasPrefix(m.Message.Content, server.CommandPrefix) {
		return
	}

	cmdParts := strings.SplitN(m.Message.Content[len(server.CommandPrefix):], " ", 2)

	if cmdParts[0] == "" {
		return
	}

	switch cmdParts[0] {
	case "setuser":
		commands.SetUser(bot, m, cmdParts)
	case "unregister":
		unregisterUser(bot, m)
	// case "track":
	// 	bot.trackRecentPlays()
	case "here":
		setScoresChannel(bot, server, m)
	default:
		bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command '%s'.", cmdParts[0]))
	}
}

func setScoresChannel(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate) {
	server.ScoreChannelID.String = m.ChannelID
	server.ScoreChannelID.Valid = true

	if err := bot.Servers.Save(server); err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
}

func unregisterUser(bot *eb.Bot, m *discordgo.MessageCreate) {
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

func trackRecentPlays(bot *eb.Bot) {
	type RegisteredUser struct {
		model.EtternaUser   `db:"u"`
		model.DiscordServer `db:"s"`
	}

	type RegisteredUserServers struct {
		User    model.EtternaUser
		Servers []model.DiscordServer
	}

	var results []RegisteredUser

	query := `
		SELECT
			u.id                    "u.id",
			u.created_at            "u.created_at",
			u.updated_at            "u.updated_at",
			u.etterna_id            "u.etterna_id",
			u.avatar                "u.avatar",
			u.username              "u.username",
			u.last_recent_score_key "u.last_recent_score_key",
			u.msd_overall           "u.msd_overall",
			u.msd_stream            "u.msd_stream",
			u.msd_jumpstream        "u.msd_jumpstream",
			u.msd_handstream        "u.msd_handstream",
			u.msd_stamina           "u.msd_stamina",
			u.msd_jackspeed         "u.msd_jackspeed",
			u.msd_chordjack         "u.msd_chordjack",
			u.msd_technical         "u.msd_technical",
			s.id                    "s.id",
			s.created_at            "s.created_at",
			s.updated_at            "s.updated_at",
			s.command_prefix        "s.command_prefix",
			s.server_id             "s.server_id",
			s.score_channel_id      "s.score_channel_id"
		FROM
			etterna_users u
		INNER JOIN users_discord_servers uds ON uds.username=u.username
		INNER JOIN discord_servers s ON s.server_id=uds.server_id
		WHERE
			s.score_channel_id IS NOT NULL
	`

	if err := bot.DB.Select(&results, query); err != nil {
		fmt.Println("Failed to look up users to track recent plays", err)
		return
	}

	userMap := make(map[string]*RegisteredUserServers)

	for _, r := range results {
		if _, exists := userMap[r.Username]; !exists {
			userMap[r.Username] = &RegisteredUserServers{
				User:    r.EtternaUser,
				Servers: []model.DiscordServer{r.DiscordServer},
			}
		} else {
			userMap[r.Username].Servers = append(userMap[r.Username].Servers, r.DiscordServer)
		}
	}

	for _, v := range userMap {
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
		if gains == "" && s.Accuracy < defaultRecentPlayMinAcc {
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
