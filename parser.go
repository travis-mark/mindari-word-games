package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Score struct {
	ID         string
	ChannelID  string
	Username   string
	Game       string
	GameNumber string
	Score      string
	Win        string
	Hardmode   string
}

// Parse a score from a text message (string)
func ParseScoreFromContent(content string) (*Score, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("message content is blank")
	}
	game := ""
	game_no := ""
	hardmode := ""
	win := ""
	score_value := lines[0]
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)(?P<game>Wordle) (?P<game_no>[\d,]+) (?P<score>\w)\/6(?P<hardmode>[*]?)`),
		regexp.MustCompile(`(?s)(?P<game>[A-Za-z ]*Dordle) (?P<game_no>\d+) (?P<left>\w)[&](?P<right>\w)\/7`),
		regexp.MustCompile(`(?s)(?P<game>[A-Za-z ]*Octordle) #(?P<game_no>\d+).*Score[:] (?P<score>\d+)`),
		regexp.MustCompile(`(?s)(?P<game>Connections).*Puzzle #(?P<game_no>\d+)`),
		regexp.MustCompile(`(?s)(?P<game>Tradle) #(?P<game_no>\d+).*(?P<score>[123456X])\/6`),
		regexp.MustCompile(`(?s)(?P<game>Strands) #(?P<game_no>\d+).*`),
		regexp.MustCompile(`(?s).*(?P<game>Animal) #(?P<game_no>\d+).*`),
	}
	var captures map[string]string
	for _, re := range patterns {
		match := re.FindStringSubmatch(content)
		if match != nil {
			captures = make(map[string]string)
			names := re.SubexpNames()
			for i, name := range names {
				switch name {
				case "game":
					game = match[i]
				case "game_no":
					game_no = match[i]
				case "hardmode":
					hardmode = match[i]
				case "score":
					score_value = match[i]
				}
				captures[name] = match[i]
			}
			break // patterns
		}
	}
	if captures == nil {
		return nil, fmt.Errorf("message did not parse: %s", content)
	}
	switch {
	case game == "Wordle" || game == "Tradle":
		if score_value == "X" {
			score_value = "7"
			win = "N"
		} else {
			win = "Y"
		}
	case strings.Contains(game, "Dordle"):
		left := captures["left"]
		right := captures["right"]
		value := 0
		if left == "X" {
			value += 7
			win = "N"
		} else {
			left_value, _ := strconv.Atoi(left)
			value += left_value
		}
		if right == "X" {
			value += 7
			win = "N"
		} else {
			right_value, _ := strconv.Atoi(right)
			value += right_value
		}
		if win == "" {
			win = "Y"
		}
		score_value = strconv.Itoa(value)
	case strings.Contains(game, "Octordle"):
		if strings.Contains(content, "🟥") {
			win = "N"
		} else {
			win = "Y"
		}
	case game == "Connections":
		re := regexp.MustCompile("(?s)[🟨🟩🟪🟦]+")
		lines := re.FindAllString(content, 64)
		match := 0
		total := 0
		for _, line := range lines {
			if line == "🟨🟨🟨🟨" || line == "🟩🟩🟩🟩" || line == "🟪🟪🟪🟪" || line == "🟦🟦🟦🟦" {
				match++
			}
			total++
		}
		if match == 4 {
			win = "Y"
			score_value = strconv.Itoa(total)
		} else {
			win = "N"
			score_value = "7"
		}
	case game == "Strands":
		score_value = strconv.Itoa(strings.Count(content, "💡"))
		win = "Y"
	case game == "Animal":
		score_value = strconv.Itoa(strings.Count(content, "🟧") + strings.Count(content, "🟩") + strings.Count(content, "🟥"))
		if score_value == "20" {
			win = "N"
		} else {
			win = "Y"
		}
	}

	score := Score{
		Game:       game,
		GameNumber: game_no,
		Hardmode:   hardmode,
		Score:      score_value,
		Win:        win,
	}
	return &score, nil
}

// Parse Discord message to extract score
func ParseScoreFromMessage(msg *discordgo.Message) (*Score, error) {
	if msg.Type != 0 {
		return nil, fmt.Errorf("message is not a score")
	}
	score, err := ParseScoreFromContent(msg.Content)
	if err != nil {
		return nil, err
	}
	score.ID = msg.ID
	score.ChannelID = msg.ChannelID
	score.Username = msg.Author.Username
	return score, nil
}

func ParseScores(messages []*discordgo.Message) ([]Score, error) {
	scores := make([]Score, 0, len(messages))
	for _, msg := range messages {
		score, err := ParseScoreFromMessage(msg)
		if err != nil {
			logPrintln("%v", err)
			continue
		}
		scores = append(scores, *score)
	}
	return scores, nil
}
