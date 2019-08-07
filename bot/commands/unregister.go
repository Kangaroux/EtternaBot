package commands

import (
	eb "github.com/Kangaroux/etternabot"
	"github.com/bwmarrin/discordgo"
)

// UnregisterUser unregisters the given discord user from an etterna user
func UnregisterUser(bot *eb.Bot, m *discordgo.MessageCreate) {
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