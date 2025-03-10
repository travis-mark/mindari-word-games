package main

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

// Cache fetched channel info
func storeChannelInfo(channel *discordgo.Channel) error {
	db, err := getDatabase()
	if err != nil {
		return err
	}
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO channels (channel_id, guild_id, name)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(channel.ID, channel.GuildID, channel.Name)
	if err != nil {
		return err
	}
	return nil
}

// Grab channel info from Discord and save to database for later use
func (dc *DiscordConnection) fetchChannelInfo(channelID string) (*discordgo.Channel, error) {
	channel, err := dc.Session.Channel(channelID)
	if err != nil {
		return nil, err
	}
	err = storeChannelInfo(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// Read channel info from storage.
//
// Defaults to a local database. Falls back to querying Discord.
func readChannelInfo(channelID string) (*discordgo.Channel, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	var channel discordgo.Channel
	row := db.QueryRow("SELECT channel_id, guild_id, name FROM channels WHERE channel_id = ?", channelID)
	err = row.Scan(&channel.ID, &channel.GuildID, &channel.Name)
	if err == sql.ErrNoRows {
		// TODO: Remove global ref workaround
		fetched_channel, err := discordConnection.fetchChannelInfo(channelID)
		return fetched_channel, err
	} else if err != nil {
		return nil, err
	} else {
		return &channel, nil
	}
}
