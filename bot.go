package etternabot

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/model/service"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

const (
	defaultPrefix             = ";"
	minAccToDisplayRecentPlay = 97.0
	recentPlayInterval        = 1 * time.Minute
)

type Bot struct {
	db      *sqlx.DB
	ett     etterna.EtternaAPI
	s       *discordgo.Session
	servers model.DiscordServerServicer
	users   model.EtternaUserServicer
}

func New(s *discordgo.Session, db *sqlx.DB, etternaAPIKey string) Bot {
	bot := Bot{
		db:      db,
		ett:     etterna.New(etternaAPIKey),
		s:       s,
		servers: service.NewDiscordServerService(db),
		users:   service.NewUserService(db),
	}

	s.AddHandler(bot.guildCreate)
	s.AddHandler(bot.messageCreate)

	// Check for recent plays periodically
	go func() {
		for {
			fmt.Println("Tracking recent plays...")
			bot.trackRecentPlays()
			<-time.After(recentPlayInterval)
		}
	}()

	return bot
}

func (bot *Bot) guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	server, err := bot.servers.Get(g.ID)

	if err != nil {
		fmt.Println("ERROR: Failed to load guild info", g.ID, g.Name)
		return
	}

	if server == nil {
		server = &model.DiscordServer{
			CommandPrefix: defaultPrefix,
			ServerID:      g.ID,
		}

		if err := bot.servers.Save(server); err != nil {
			fmt.Println("ERROR: Failed to insert guild into db", g.ID, g.Name, err)
			return
		} else {
			fmt.Println("Created record for server", g.Name)
		}
	}
}

func (bot *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	server, err := bot.servers.Get(m.GuildID)

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
		bot.setUser(m, cmdParts)
	case "unregister":
		bot.unregisterUser(m)
	// case "track":
	// 	bot.trackRecentPlays()
	case "here":
		bot.setScoresChannel(server, m)
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command '%s'.", cmdParts[0]))
	}
}

func (bot *Bot) setScoresChannel(server *model.DiscordServer, m *discordgo.MessageCreate) {
	server.ScoreChannelID.String = m.ChannelID
	server.ScoreChannelID.Valid = true

	if err := bot.servers.Save(server); err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
}

// setUser registers an etterna user with the discord user who sent the message.
// The process of registering as an etterna user is:
// 1. Check if the etterna user is already registered
// 2. Check if the discord user is already registered
// 3. Check if the etterna user exists
// 4. Register
func (bot *Bot) setUser(m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		bot.s.ChannelMessageSend(m.ChannelID,
			"Usage: setuser <username>")
		return
	}

	username := strings.TrimSpace(args[1])
	discordID, err := bot.users.GetRegisteredDiscordUserID(m.GuildID, username)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if discordID == m.Author.ID {
		bot.s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("You are already registered as '%s'.", username))
		return
	} else if discordID != "" {
		bot.s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Another user is already registered as '%s'.", username))
		return
	}

	// Get the etterna user registered with this discord user
	user, err := bot.users.GetRegisteredUser(m.GuildID, m.Author.ID)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if user != nil {
		bot.s.ChannelMessageSend(m.ChannelID,
			"You are already registered as another user. Use the 'unregister' "+
				"command first and try again.")
		return
	}

	// The discord user is not associated with any etterna users, look up
	// the etterna user with that username
	user, err = bot.users.GetUsername(username)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// User doesn't exist, try and look them up
	if user == nil {
		ettUser, err := bot.ett.GetByUsername(username)

		if err != nil {
			bot.s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		id, err := bot.ett.GetUserID(username)

		if err != nil {
			bot.s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		user = &model.EtternaUser{
			Username:      ettUser.Username,
			EtternaID:     id,
			Avatar:        ettUser.AvatarURL,
			MSDOverall:    ettUser.MSD.Overall,
			MSDStream:     ettUser.MSD.Stream,
			MSDJumpstream: ettUser.MSD.Jumpstream,
			MSDHandstream: ettUser.MSD.Handstream,
			MSDStamina:    ettUser.MSD.Stamina,
			MSDJackSpeed:  ettUser.MSD.JackSpeed,
			MSDChordjack:  ettUser.MSD.Chordjack,
			MSDTechnical:  ettUser.MSD.Technical,
		}

		if err := bot.users.Save(user); err != nil {
			bot.s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
	}

	ok, err := bot.users.Register(user.Username, m.GuildID, m.Author.ID)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if !ok {
		// This can only happen due to a data race, still worth checking though
		bot.s.ChannelMessageSend(m.ChannelID,
			"You are currently registered as another user. Use the 'unregister' command "+
				"first and try again.")
	} else {
		bot.s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Success! You are now registered as '%s'.", user.Username))
	}
}

func (bot *Bot) unregisterUser(m *discordgo.MessageCreate) {
	ok, err := bot.users.Unregister(m.GuildID, m.Author.ID)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if ok {
		bot.s.ChannelMessageSend(m.ChannelID,
			"Success! You are no longer registered. Use the setuser command to register "+
				"as another user.")
	} else {
		bot.s.ChannelMessageSend(m.ChannelID, "You are not registered to an etterna user.")
	}
}

func (bot *Bot) trackRecentPlays() {
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

	if err := bot.db.Select(&results, query); err != nil {
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
		scores, err := bot.ett.GetScores(v.User.EtternaID, 1, 0, etterna.SortDate, false)

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

		if err := bot.ett.GetScoreDetail(&s); err != nil {
			fmt.Println("Failed to get score details", s.Key, err)
			return
		}

		latestUser, err := bot.ett.GetByUsername(v.User.Username)

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

		v.User.MSDOverall = TruncateFloat(latestUser.Overall, 2)
		v.User.MSDStream = TruncateFloat(latestUser.Stream, 2)
		v.User.MSDJumpstream = TruncateFloat(latestUser.Jumpstream, 2)
		v.User.MSDHandstream = TruncateFloat(latestUser.Handstream, 2)
		v.User.MSDStamina = TruncateFloat(latestUser.Stamina, 2)
		v.User.MSDJackSpeed = TruncateFloat(latestUser.JackSpeed, 2)
		v.User.MSDChordjack = TruncateFloat(latestUser.Chordjack, 2)
		v.User.MSDTechnical = TruncateFloat(latestUser.Technical, 2)
		v.User.LastRecentScoreKey.String = s.Key
		v.User.LastRecentScoreKey.Valid = true

		bot.users.Save(&v.User)

		// If the score is invalid don't post it
		if !s.Valid {
			fmt.Println("Score isn't valid")
			continue
		}

		gains := ""

		if diffMSD.Overall >= 0.01 {
			gains = fmt.Sprintf("➤ **Overall:** %.2f (+%.2f)\n", latestUser.Overall, diffMSD.Overall)
		}

		if diffMSD.Stream >= 0.01 {
			gains = fmt.Sprintf("➤ **Stream:** %.2f (+%.2f)\n", latestUser.Stream, diffMSD.Stream)
		}

		if diffMSD.Jumpstream >= 0.01 {
			gains = fmt.Sprintf("➤ **Jumpstream:** %.2f (+%.2f)\n", latestUser.Jumpstream, diffMSD.Jumpstream)
		}

		if diffMSD.Handstream >= 0.01 {
			gains = fmt.Sprintf("➤ **Handstream:** %.2f (+%.2f)\n", latestUser.Handstream, diffMSD.Handstream)
		}

		if diffMSD.Stamina >= 0.01 {
			gains = fmt.Sprintf("➤ **Stamina:** %.2f (+%.2f)\n", latestUser.Stamina, diffMSD.Stamina)
		}

		if diffMSD.JackSpeed >= 0.01 {
			gains = fmt.Sprintf("➤ **JackSpeed:** %.2f (+%.2f)\n", latestUser.JackSpeed, diffMSD.JackSpeed)
		}

		if diffMSD.Chordjack >= 0.01 {
			gains = fmt.Sprintf("➤ **Chordjack:** %.2f (+%.2f)\n", latestUser.Chordjack, diffMSD.Chordjack)
		}

		if diffMSD.Technical >= 0.01 {
			gains = fmt.Sprintf("➤ **Technical:** %.2f (+%.2f)\n", latestUser.Technical, diffMSD.Technical)
		}

		// Only display the song if the player got above a certain acc or if they gained pp
		if gains == "" && s.Accuracy < minAccToDisplayRecentPlay {
			continue
		}

		song, err := bot.ett.GetSong(s.Song.ID)

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
			Timestamp:   s.Date.UTC().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				IconURL: "https://i.imgur.com/HwIkGCk.png",
				Text:    v.User.Username,
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://etternaonline.com/song_images/bg/" + s.Song.BackgroundURL,
			},
		}

		for _, server := range v.Servers {
			bot.s.ChannelMessageSendEmbed(server.ScoreChannelID.String, msg)
		}
	}
}
