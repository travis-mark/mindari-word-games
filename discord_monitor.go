package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Called when a message is created in a channel
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	score, err := ParseScoreFromMessage(Message{
		ID:   m.ID,
		Type: int(m.Type),
		Author: Author{
			ID:       m.Author.ID,
			Username: m.Author.Username,
		},
		Content: m.Content,
	})
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
	startChannelMonitor(m.ChannelID)
}

// Periodic channel scan to cover messages not received by websocket
func channelTick(channel string) error {
	before, after, err := GetScoreIDRange()
	if err != nil {
		return err
	}
	if before != "" && after != "" {
		// Incremental load
		err = ScanChannel(Options{Channel: channel, Before: before})
		if err != nil {
			return err
		}
		err = ScanChannel(Options{Channel: channel, After: after})
		if err != nil {
			return err
		}
	} else {
		// Fetch all
		err = ScanChannel(Options{Channel: channel})
		if err != nil {
			return err
		}
	}
	return err
}

var channelMonitors = map[string]*time.Ticker{}

// Start periodic scan
func startChannelMonitor(channel string) error {
	err := channelTick(channel)
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
		err := channelTick(channel)
		if err != nil {
			return err
		}
	}
	return nil
}

func MonitorDiscord() error {
	authorization, err := getAuthorization()
	if err != nil {
		return err
	}
	discord, err := discordgo.New(authorization)
	if err != nil {
		return err
	}
	discord.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuildMessages
	logPrintln("Starting monitor...")
	err = discord.Open()
	if err != nil {
		return err
	}
	discord.AddHandler(messageCreate)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	logPrintln("Stopping monitor...")
	discord.Close()
	return nil
}
