package model

import "database/sql"

type DiscordServerServicer interface {
	Get(serverID string) (*DiscordServer, error)
	Save(server *DiscordServer) error
}

type DiscordServer struct {
	BaseModel
	CommandPrefix  string         `db:"command_prefix"` // Prefix for using bot commands
	ServerID       string         `db:"server_id"`
	ScoreChannelID sql.NullString `db:"score_channel_id"` // The channel to post recent plays in
}
