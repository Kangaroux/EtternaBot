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
	users service.UserService
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

func (bot *Bot) setUser(m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		bot.s.ChannelMessageSend(m.ChannelID,
			"Missing argument <username>. Usage: setuser <username>")
		return
	}

	username := strings.TrimSpace(args[1])

	// Check and see if the user is already registered with an etterna user
	registeredUser, err := bot.users.GetDiscordID(m.Author.ID)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if registeredUser != nil {
		// Already registered to this user
		if strings.ToLower(registeredUser.Username) == strings.ToLower(username) {
			bot.s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("You are already registered as '%s'.", username))
			return
		}

		bot.s.ChannelMessageSend(m.ChannelID,
			"You are currently registered as another user. Use the 'unregister' command first and try again.")
		return
	}

	// See if the user is in the database
	user, err := bot.users.GetUsername(username)

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

		user = &model.User{
			Username:      username,
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
	}

	// This user is already associated with a discord user
	if user.DiscordID.Valid {
		if user.DiscordID.String == m.Author.ID {

		}

		bot.s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Another user is already registered as '%s'.", username))
		return
	}

	user.DiscordID.String = m.Author.ID
	user.DiscordID.Valid = true

	if err := bot.users.Save(user); err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	bot.s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Success! You are now registered as '%s'.", username))
}

func (bot *Bot) unregisterUser(m *discordgo.MessageCreate) {
	user, err := bot.users.GetDiscordID(m.Author.ID)

	if err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	if user == nil {
		bot.s.ChannelMessageSend(m.ChannelID, "You are not registered to an Etterna user.")
		return
	}

	user.DiscordID.Valid = false

	if err := bot.users.Save(user); err != nil {
		bot.s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	bot.s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Success! You are no longer registered to '%s'.", user.Username))
}
