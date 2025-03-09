package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Called when a message is created in a channel
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	score, err := ParseScoreFromMessage(m.Message)
	if err != nil {
		logPrintln("Parser error: %v, %v", err, m)
		return
	}
	err = AddScores([]Score{*score})
	if err != nil {
		logPrintln("AddScores error: %v, %v", err, m)
		return
	}
	logPrintln("Added score from bot: %s %s %s %s", score.Username, score.Game, score.GameNumber, score.Score)
	// TODO: Remove global ref workaround
	discordConnection.startChannelMonitor(m.ChannelID)
}

// Periodic channel scan to cover messages not received by websocket
func (dc *DiscordConnection) channelTick(channel string) error {
	before, after, err := GetScoreIDRange()
	if err != nil {
		return err
	}
	if before != "" && after != "" {
		// Incremental load
		err = dc.scanChannel(Options{Channel: channel, Before: before})
		if err != nil {
			return err
		}
		err = dc.scanChannel(Options{Channel: channel, After: after})
		if err != nil {
			return err
		}
	} else {
		// Fetch all
		err = dc.scanChannel(Options{Channel: channel})
		if err != nil {
			return err
		}
	}
	return err
}

var channelMonitors = map[string]*time.Ticker{}

// Start periodic scan
func (dc *DiscordConnection) startChannelMonitor(channel string) error {
	err := dc.channelTick(channel)
	if err != nil {
		return err
	}
	ticker := channelMonitors[channel]
	if ticker != nil {
		return nil
	}
	ticker = time.NewTicker(1 * time.Hour)
	channelMonitors[channel] = ticker
	defer ticker.Stop()
	for range ticker.C {
		err := dc.channelTick(channel)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dc *DiscordConnection) startDiscordMonitor() error {
	dc.Session.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuildMessages
	logPrintln("Starting monitor...")
	dc.Session.AddHandler(messageCreate)
	dc.onDiscordConnectionClose(func() error {
		logPrintln("Stopping monitor...")
		return nil
	})
	return nil
}
