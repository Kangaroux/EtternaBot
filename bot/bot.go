package bot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/model/service"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

const (
	defaultPrefix           = ";"  // Prefix for commands
	defaultRecentPlayMinAcc = 99.5 // Minimum acc to display a recent play
	embedColor              = 8519899
	recentPlayInterval      = 2 * time.Minute // How often to check for recent plays
)

var (
	reCommand  = regexp.MustCompile(`^[a-z\d\.@]+$`)
	reScoreURL = regexp.MustCompile(`etternaonline\.com\/score\/view\/(S[a-f0-9]+)`)
)

// New returns a new discord bot instance that is ready to be started
func New(s *discordgo.Session, db *sqlx.DB, etternaAPIKey string) eb.Bot {
	bot := eb.Bot{
		DB:      db,
		API:     etterna.New(etternaAPIKey),
		Session: s,
		Servers: service.NewDiscordServerService(db),
		Songs:   service.NewSongService(db),
		Users:   service.NewUserService(db),
	}

	s.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		guildCreate(&bot, g)
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		messageCreate(&bot, m)
	})

	s.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		ready(&bot, r)
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
		}

		fmt.Println("Created record for server", g.Name)
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

	// Parse the message even if it doesn't look like a command
	if !strings.HasPrefix(m.Message.Content, server.CommandPrefix) {
		parseMessageNoCmd(bot, m)
		return
	}

	cmdParts := strings.Split(m.Message.Content[len(server.CommandPrefix):], " ")
	cmdParts[0] = strings.ToLower(cmdParts[0])

	if cmdParts[0] == "" || !reCommand.MatchString(cmdParts[0]) {
		return
	}

	switch cmdParts[0] {
	case "compare":
		CmdCompare(bot, server, m, cmdParts)
	case "help":
		CmdHelp(bot, server, m)
	case "profile":
		CmdProfile(bot, m, cmdParts)
	case "recent":
		CmdRecentPlay(bot, server, m, cmdParts)
	case "setuser":
		CmdSetUser(bot, m, cmdParts)
	case "unset":
		CmdUnsetUser(bot, m)
	case "vs":
		CmdVersus(bot, m, cmdParts)
	case "here":
		CmdSetScoresChannel(bot, server, m)
	default:
		if strings.HasPrefix(cmdParts[0], "compare@") {
			CmdCompareRate(bot, server, m, cmdParts)
		} else {
			bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command '%s'.", cmdParts[0]))
		}
	}
}

func ready(bot *eb.Bot, r *discordgo.Ready) {
	// Periodically set the bot status
	go func() {
		for {
			bot.Session.UpdateStatus(0, ";help")
			<-time.After(1 * time.Hour)
		}
	}()

	// Periodically check for recent plays
	go func() {
		for {
			TrackAllRecentPlays(bot, defaultRecentPlayMinAcc)
			<-time.After(recentPlayInterval)
		}
	}()
}

func parseMessageNoCmd(bot *eb.Bot, m *discordgo.MessageCreate) {
	handleScoreURLs(bot, m)
}

func handleScoreURLs(bot *eb.Bot, m *discordgo.MessageCreate) {
	match := reScoreURL.FindStringSubmatch(m.Message.Content)

	if match == nil {
		return
	} else if len(match[1]) < 41 {
		bot.Session.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Score URL does not look correct (did you copy it right?)"),
		)
		return
	}

	bot.Session.ChannelTyping(m.ChannelID)

	key := match[1]
	detail, err := bot.API.GetScoreDetail(key)

	if err != nil {
		bot.Session.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Failed to get score (%s)", err.Error()),
		)
		return
	}

	scores, err := bot.API.GetScores(detail.User.ID, detail.Song.Name, 100, 0, etterna.SortNerf, false)

	if err != nil {
		bot.Session.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Failed to get score (%s)", err.Error()),
		)
		return
	}

	var score *etterna.Score

	// Find the matching score and set the nerf rating
	for _, s := range scores {
		if s.Key == key[:41] {
			detail.Nerfed = s.Nerfed
			score = &s
			break
		}
	}

	if score == nil {
		bot.Session.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Failed to get score (could not be found)"),
		)
		return
	}

	score.MaxCombo = detail.MaxCombo
	score.MinesHit = detail.MinesHit
	score.Mods = detail.Mods
	score.Date = detail.Date
	user, err := getUserOrCreate(bot, detail.User.Username, false)

	if err != nil {
		bot.Session.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Could not find user %s (%s)", detail.User.Username, err.Error()),
		)
		return
	}

	embed, err := getPlaySummaryAsDiscordEmbed(bot, score, user)

	if err != nil {
		fmt.Println(err)
		return
	}

	embed.Author.Name = "Played by " + user.Username
	_, err = bot.Session.ChannelMessageSendEmbed(m.ChannelID, embed)

	if err != nil {
		fmt.Println(err)
		return
	}

	server, _ := bot.Servers.Get(m.GuildID)
	server.LastSongID.Int64 = int64(score.Song.ID)
	server.LastSongID.Valid = true
	bot.Servers.Save(server)
}
