package etternabot

import (
	"fmt"

	"github.com/Kangaroux/etternabot/etterna"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

type Bot struct {
	db  *sqlx.DB
	ett etterna.EtternaAPI
}

func New(s *discordgo.Session, db *sqlx.DB, etternaAPIKey string) Bot {
	bot := Bot{
		db:  db,
		ett: etterna.New(etternaAPIKey),
	}

	s.AddHandler(bot.messageCreate)

	return bot
}

func (bot *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println(m.Message.Content)
}
