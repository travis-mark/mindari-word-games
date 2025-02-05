package main

import (
	"testing"
)

// Checkd correctness of incoming score parsing
func TestScoreParser(t *testing.T) {
	type Case struct {
		input  string
		output Score
	}
	// TODO: Get other test cases from Bytebinder
	// TODO: Hardmode, Game No, Losses
	data := [...]Case{
		{
			input:  "Wordle 771 3/6*\r\n\r\n⬛⬛⬛⬛🟩\r\n🟨🟩⬛⬛🟩\r\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "3"},
		},
		{
			input:  "Wordle 1,327 4/6\n⬜🟩🟨⬜⬜\n⬜⬜🟨⬜⬜\n⬜⬜⬜⬜🟨\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "4"},
		},
		{
			input:  "Daily Octordle #553\r\n🔟7️⃣\r\n6️⃣8️⃣\r\n3️⃣5️⃣\r\n9️⃣🕚\r\nScore: 59",
			output: Score{Game: "Daily Octordle", Score: "59"},
		},
	}
	for _, item := range data {
		message := Message{Content: item.input}
		score, err := ParseScoreFromMessage(message)
		if err != nil {
			t.Fatalf(`TestScoreParser("%s") returned error: %v`, item.input, err)
		}
		if score.Game != item.output.Game {
			t.Fatalf("TestScoreParser [Game]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Game, item.output.Game)
		}
		if score.Score != item.output.Score {
			t.Fatalf("TestScoreParser [Score]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Score, item.output.Score)
		}
	}
}
