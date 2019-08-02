package etternabot

import (
	"fmt"
	"strings"

	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/Kangaroux/etternabot/model/service"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

const (
	cmdPrefix = ";"
)

type Bot struct {
	db    *sqlx.DB
	ett   etterna.EtternaAPI
	s     *discordgo.Session
	users service.EtternaUserService
}

func New(s *discordgo.Session, db *sqlx.DB, etternaAPIKey string) Bot {
	bot := Bot{
		db:    db,
		ett:   etterna.New(etternaAPIKey),
		s:     s,
		users: service.NewUserService(db),
	}

	s.AddHandler(bot.messageCreate)

	return bot
}

func (bot *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Message.Content, cmdPrefix) {
		return
	}

	cmdParts := strings.SplitN(m.Message.Content[len(cmdPrefix):], " ", 2)

	if cmdParts[0] == "" {
		return
	}

	switch cmdParts[0] {
	case "setuser":
		bot.setUser(m, cmdParts)
	case "unregister":
		bot.unregisterUser(m)
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command '%s'.", cmdParts[0]))
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
