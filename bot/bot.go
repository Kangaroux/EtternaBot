package bot

import (
	"fmt"
	"strings"
	"time"

	eb "github.com/Kangaroux/etternabot"
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

	// s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
	// 	go func() {
	// 		commands.TrackRecentPlays(&bot, defaultRecentPlayMinAcc)
	// 		<-time.After(recentPlayInterval)
	// 	}()
	// })

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
		commands.UnregisterUser(bot, m)
	// case "track":
	// 	bot.trackRecentPlays()
	case "here":
		commands.SetScoresChannel(bot, server, m)
	default:
		bot.Session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command '%s'.", cmdParts[0]))
	}
}
