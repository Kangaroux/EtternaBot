package commands

import (
	eb "github.com/Kangaroux/etternabot"
	"github.com/Kangaroux/etternabot/model"
	"github.com/bwmarrin/discordgo"
)

// SetScoresChannel sets which discord channel the bot should post scores in
// when tracking recent plays
func SetScoresChannel(bot *eb.Bot, server *model.DiscordServer, m *discordgo.MessageCreate) {
	server.ScoreChannelID.String = m.ChannelID
	server.ScoreChannelID.Valid = true

	if err := bot.Servers.Save(server); err != nil {
		bot.Session.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
}
