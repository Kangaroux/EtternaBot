package model

// UserServer is a join table that maps discord users to etterna users
// on a per-server basis. This table has two constraints in the database:
// (ServerID, Username) and (ServerID, DiscordID).
// These constraints ensure that, for each discord server the bot is in,
// an etterna user can only be registered to one discord user at a time,
// and each discord user can only be registered to one etterna user at
// a time. In other words, enforcing a One-to-One relationship within the
// context of a single discord server.
type UserServer struct {
	ID       int    `db:"id"`
	ServerID string `db:"server_id"`       // References `discord_servers.server_id`
	Username string `db:"user_id"`         // References `users.username`
	UserID   string `db:"discord_user_id"` // The discord ID of a discord user
}
