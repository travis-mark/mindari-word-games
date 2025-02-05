package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Score struct {
	ID       string
	Username string
	Game     string
	Score    string
	Content  string
}

// Parse message to extract score
func ParseScoreFromMessage(msg Message) (*Score, error) {
	if msg.Type != 0 {
		return nil, fmt.Errorf("Message is not a score")
	}
	lines := strings.Split(msg.Content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("Message content is blank")
	}
	game := ""
	score_value := lines[0]
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?P<game>Wordle) (?P<game_no>[\d,]+) (?P<score>\w)\/6(?P<hardmode>[*]?)`),
		regexp.MustCompile(`(?s)(?P<game>[A-Za-z ]*Octordle) #(?P<game_no>\d+).*Score[:] (?P<score>\d+)`),
	}
	for _, re := range patterns {
		match := re.FindStringSubmatch(msg.Content)
		if match != nil {
			names := re.SubexpNames()
			for i, name := range names {
				switch name {
				case "game":
					game = match[i]
				case "score":
					score_value = match[i]
				}
			}
			break // patterns
		}
	}
	score := Score{
		ID:       msg.ID,
		Username: msg.Author.Username,
		Game:     game,
		Score:    score_value,
		Content:  msg.Content,
	}
	return &score, nil
}

func ParseScores(messages []Message) ([]Score, error) {
	scores := make([]Score, 0, len(messages))

	for _, msg := range messages {
		score, err := ParseScoreFromMessage(msg)
		if err != nil {
			// Drop errors, TODO: Log?
			fmt.Printf("ParseScoreFromMessage error = %v", err)
			continue
		}
		scores = append(scores, *score)
	}
	return scores, nil
}
