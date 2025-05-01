package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

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

// Close connection to discord
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

// Add a handler to fire before connection closes
func (dc *DiscordConnection) onDiscordConnectionClose(handler func() error) {
	dc.onCloseHandlers = append(dc.onCloseHandlers, handler)
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

// Sample command for testing
func (dc *DiscordConnection) enableEchoCommand() (ccmd *discordgo.ApplicationCommand, err error) {
	return dc.enableCommand("echo", "Test Command", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hey there! Congratulations, you just executed your first slash command",
			},
		})
	})
}

func (dc *DiscordConnection) enableStatsCommand() (ccmd *discordgo.ApplicationCommand, err error) {
	games, _ := getGameList("", "", "", "")
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	if len(games) > 0 {
		for _, game := range games {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  game,
				Value: game,
			})
		}
	}
	cmd := discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "stats",
		Description: "Show Game Stats",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "game",
				Description: "Name of the game",
				Required:    true,
				Choices:     choices,
			},
		},
	}
	ccmd, err = dc.Session.ApplicationCommandCreate(dc.ApplicationID, "", &cmd)
	if err != nil {
		return nil, err
	}
	dc.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		game := "Wordle"
		for _, option := range data.Options {
			switch option.Name {
			case "game":
				game = option.StringValue()
			}
		}
		stats, err := getStats(game, i.GuildID, "", "")
		var content string
		if err != nil {
			content = fmt.Sprintf("Error getting stats: %v", err)
		} else {
			content = SPrintStatsMarkdownDiscord(stats)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
	})
	dc.onDiscordConnectionClose(func() error {
		return dc.Session.ApplicationCommandDelete(dc.ApplicationID, "", ccmd.ID)
	})
	return ccmd, err
}

func (dc *DiscordConnection) enableSeasonCommand() (ccmd *discordgo.ApplicationCommand, err error) {
	games, _ := getGameList("", "", "", "")
	cmd := discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "season",
		Description: "Show Season Report (all games)",
	}
	ccmd, err = dc.Session.ApplicationCommandCreate(dc.ApplicationID, "", &cmd)
	if err != nil {
		return nil, err
	}
	dc.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		content := ""
		for _, game := range games {
			stats, err := getStats(game, i.GuildID, "", "")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			} else {
				content = content + "# " + game + "\n"
				content = content + SPrintStatsMarkdownDiscord(stats) + "\n"
			}
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
	})
	dc.onDiscordConnectionClose(func() error {
		return dc.Session.ApplicationCommandDelete(dc.ApplicationID, "", ccmd.ID)
	})
	return ccmd, err
}

func (dc *DiscordConnection) enableSlashCommands() (err error) {
	_, err = dc.enableStatsCommand()
	if err != nil {
		return err
	}
	logPrintln("/stats added")
	_, err = dc.enableSeasonCommand()
	if err != nil {
		return err
	}
	logPrintln("/season added")
	return nil
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
	err = addScores(scores)
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

// Periodic channel scan to cover messages not received by websocket
func (dc *DiscordConnection) channelTick(channel string) error {
	before, after, err := getScoreIDRange()
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
	// Called when a message is created in a channel
	dc.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		score, err := ParseScoreFromMessage(m.Message)
		if err != nil {
			logPrintln("Parser error: %v, %v", err, m)
			return
		}
		err = addScores([]Score{*score})
		if err != nil {
			logPrintln("addScores error: %v, %v", err, m)
			return
		}
		logPrintln("Added score from bot: %s %s %s %s", score.Username, score.Game, score.GameNumber, score.Score)
		dc.startChannelMonitor(m.ChannelID)
	})
	dc.onDiscordConnectionClose(func() error {
		logPrintln("Stopping monitor...")
		return nil
	})
	return nil
}

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
	if err == sql.ErrNoRows && discordConnection != nil {
		fetched_channel, err := discordConnection.fetchChannelInfo(channelID)
		return fetched_channel, err
	} else if err != nil {
		return nil, err
	} else {
		return &channel, nil
	}
}
