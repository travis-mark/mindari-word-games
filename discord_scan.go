package main

import "github.com/bwmarrin/discordgo"

type Options struct {
	Channel string // Discord Channel ID
	Before  string // Backward search pointer
	After   string // Forward search pointer
}

// Fetch messages from Discord, parse for puzzles and save to DB
func ScanChannel(options Options) error {
	authorization, err := getAuthorization()
	if err != nil {
		return err
	}
	discord, err := discordgo.New(authorization)
	if err != nil {
		return err
	}
	messages, err := discord.ChannelMessages(options.Channel, 50, options.Before, options.After, "")
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
		ScanChannel(prev_page)
	}
	if options.After != "" || options.Before == "" && options.After == "" {
		last_id := messages[0].ID
		next_page := options
		next_page.After = last_id
		ScanChannel(next_page)
	}
	return nil
}
