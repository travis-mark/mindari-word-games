package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Called when a message is created in a channel
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Temp guard for all of the random discords I'm in
	if m.ChannelID != getDefaultChannel() {
		return
	}

	// TODO: Change parser to use discordgo's struct
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
		log.Printf("Parser error: %v, %v\n", err, m)
		return
	}

	err = AddScores([]Score{*score})
	if err != nil {
		log.Printf("AddScores error: %v, %v\n", err, m)
		return
	}

	log.Printf("Added score from bot: %s %s %s %s\n", score.Username, score.Game, score.GameNumber, score.Score)
}

func ConnectToDiscord() {
	authorization, err := getAuthorization()
	if err != nil {
		log.Fatal("Error getting authorization: ", err)
	}
	discord, err := discordgo.New(authorization)
	discord.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuildMessages
	log.Printf("Starting bot...\n")
	err = discord.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", err)
	}
	discord.AddHandler(messageCreate)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Printf("Stopping bot...\n")
	discord.Close()
}
