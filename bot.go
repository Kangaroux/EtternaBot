package etternabot

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	db *sql.DB
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

}
