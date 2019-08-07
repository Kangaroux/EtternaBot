package etternabot

import (
	"github.com/Kangaroux/etternabot/etterna"
	"github.com/Kangaroux/etternabot/model"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

type Bot struct {
	DB      *sqlx.DB
	API     etterna.EtternaAPI
	Session *discordgo.Session
	Servers model.DiscordServerServicer
	Users   model.EtternaUserServicer
}

type Play struct {
	Score etterna.Score
	User  model.EtternaUser
}
