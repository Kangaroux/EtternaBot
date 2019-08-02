package model

import "github.com/Kangaroux/etternabot/etterna"

type Score struct {
	DiscordID string // The discord ID of the user who played the song
	etterna.Song
	etterna.Score
}
