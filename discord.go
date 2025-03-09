package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// TODO: Consolidate commands here

// Wrapper for connection to Discord
type DiscordConnection struct {
	ApplicationID   string
	Authorization   string
	Session         *discordgo.Session
	onCloseHandlers []func() error
}

var discordConnection *DiscordConnection

func initDiscordConnection() (*DiscordConnection, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	if discordConnection != nil {
		return discordConnection, nil
	}
	applicationID := os.Getenv("DISCORD_APPLICATION_ID")
	if applicationID == "" {
		return nil, fmt.Errorf("environment variable DISCORD_APPLICATION_ID not set")
	}
	authorization := os.Getenv("DISCORD_AUTHORIZATION")
	if authorization == "" {
		return nil, fmt.Errorf("environment variable DISCORD_AUTHORIZATION not set")
	}
	session, err := discordgo.New(authorization)
	if err != nil {
		return nil, err
	}
	err = session.Open()
	if err != nil {
		return nil, err
	}
	discordConnection = &DiscordConnection{
		ApplicationID: applicationID,
		Authorization: authorization,
		Session:       session,
	}
	return discordConnection, nil
}

func (dc *DiscordConnection) close() {
	for _, handler := range dc.onCloseHandlers {
		err := handler()
		if err != nil {
			log.Panicf("error during discord connection close: %v", err)
		}
	}
	err := dc.Session.Close()
	if err != nil {
		log.Panicf("error during discord connection close: %v", err)
	}
}

func (dc *DiscordConnection) onDiscordConnectionClose(handler func() error) {
	dc.onCloseHandlers = append(dc.onCloseHandlers, handler)
}

func handleEchoCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Hey there! Congratulations, you just executed your first slash command",
		},
	})
}

func (dc *DiscordConnection) enableCommand(name string, description string, handler func(s *discordgo.Session, i *discordgo.InteractionCreate)) (ccmd *discordgo.ApplicationCommand, err error) {
	cmd := discordgo.ApplicationCommand{
		Name:        name,
		Description: description,
	}
	ccmd, err = dc.Session.ApplicationCommandCreate(dc.ApplicationID, "", &cmd)
	if err != nil {
		return nil, err
	}
	dc.Session.AddHandler(handler)
	dc.onDiscordConnectionClose(func() error {
		return dc.Session.ApplicationCommandDelete(dc.ApplicationID, "", ccmd.ID)
	})
	return ccmd, err
}

func (dc *DiscordConnection) enableEchoCommand() (ccmd *discordgo.ApplicationCommand, err error) {
	return dc.enableCommand("echo", "Test Command", handleEchoCommand)
}

type Options struct {
	Channel string // Discord Channel ID
	Before  string // Backward search pointer
	After   string // Forward search pointer
}

// Fetch messages from Discord, parse for puzzles and save to DB
func (dc *DiscordConnection) scanChannel(options Options) error {
	messages, err := dc.Session.ChannelMessages(options.Channel, 50, options.Before, options.After, "")
	if err != nil {
		return err
	}
	count := len(messages)
	if count == 0 {
		logPrintln("No new records found.")
		return nil
	}
	scores, err := ParseScores(messages)
	if err != nil {
		return err
	}
	err = AddScores(scores)
	if err != nil {
		return err
	}
	logPrintln("%d records updated (%s - %s)", count, messages[0].ID, messages[len(messages)-1].ID)
	// Check other pages
	// Unsure if assumption about message order is safe
	if options.Before != "" || options.Before == "" && options.After == "" {
		first_id := messages[count-1].ID
		prev_page := options
		prev_page.Before = first_id
		dc.scanChannel(prev_page)
	}
	if options.After != "" || options.Before == "" && options.After == "" {
		last_id := messages[0].ID
		next_page := options
		next_page.After = last_id
		dc.scanChannel(next_page)
	}
	return nil
}
