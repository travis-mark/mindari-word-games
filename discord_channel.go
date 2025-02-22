package main

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

// Cache fetched channel info
func storeChannelInfo(channel *discordgo.Channel) error {
	db, err := GetDatabase()
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
func FetchChannelInfo(channelID string) (*discordgo.Channel, error) {
	authorization, err := getAuthorization()
	if err != nil {
		return nil, err
	}
	discord, err := discordgo.New(authorization)
	if err != nil {
		return nil, err
	}
	channel, err := discord.Channel(channelID)
	if err != nil {
		return nil, err
	}
	err = storeChannelInfo(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// Read channel info from storage without hitting discord
func ReadChannelInfo(channelID string) (*discordgo.Channel, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, err
	}
	var channel discordgo.Channel
	row := db.QueryRow("SELECT channel_id, guild_id, name FROM channels WHERE channel_id = ?", channelID)
	err = row.Scan(&channel.ID, &channel.GuildID, &channel.Name)
	if err == sql.ErrNoRows {
		fetched_channel, err := FetchChannelInfo(channelID)
		return fetched_channel, err
	} else if err != nil {
		return nil, err
	} else {
		return &channel, nil
	}
}
