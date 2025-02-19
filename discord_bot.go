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
		ID: m.ID,
		Type: int(m.Type),
		Author: Author{
			ID: m.Author.ID,
			Username: m.Author.Username,
		},
		Content: m.Content,
	})

	if err != nil {
		log.Printf("parser error: %v, %s\n", err, m.Content)
	} else {
		log.Printf("parsed score: %v\n", score)
	}
}

func ConnectToDiscord() {
	discord, err := discordgo.New(getAuthorization())
	discord.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuildMessages
	err = discord.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", err)
	}
	discord.AddHandler(messageCreate)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	discord.Close()
}