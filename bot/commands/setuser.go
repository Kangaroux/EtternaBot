package commands

import (
	"fmt"
	"strings"

	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
	"github.com/bwmarrin/discordgo"
)

// SetUser links a discord user with an etterna user. Only one discord user
// can be linked to a given etterna user at a time in a server. Likewise, discord
// users can only be linked to one etterna user at a time in a server.
func SetUser(bot *eb.Bot, m *discordgo.MessageCreate, args []string) {
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
			"You are already registered as another user. Use the 'unregister' "+
				"command first and try again.")
		return
	}

	// The discord user is not associated with any etterna users, look up
	// the etterna user with that username
	user, err = getUserOrCreate(bot, username)

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
			"You are currently registered as another user. Use the 'unregister' command "+
				"first and try again.")
	} else {
		bot.Session.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Success! You are now registered as '%s'.", user.Username))
	}
}

// getUserOrCreate returns the etterna user with the given username, inserting the user into the
// database automatically if they don't already exist
func getUserOrCreate(bot *eb.Bot, username string) (*model.EtternaUser, error) {
	etternaUser, err := bot.API.GetByUsername(username)

	if err != nil {
		return nil, err
	}

	id, err := bot.API.GetUserID(username)

	if err != nil {
		return nil, err
	}

	user := &model.EtternaUser{
		Username:      etternaUser.Username,
		EtternaID:     id,
		Avatar:        etternaUser.AvatarURL,
		MSDOverall:    etternaUser.MSD.Overall,
		MSDStream:     etternaUser.MSD.Stream,
		MSDJumpstream: etternaUser.MSD.Jumpstream,
		MSDHandstream: etternaUser.MSD.Handstream,
		MSDStamina:    etternaUser.MSD.Stamina,
		MSDJackSpeed:  etternaUser.MSD.JackSpeed,
		MSDChordjack:  etternaUser.MSD.Chordjack,
		MSDTechnical:  etternaUser.MSD.Technical,
	}

	if err := bot.Users.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
