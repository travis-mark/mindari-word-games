package main

import (
	"testing"
)

// Check correctness of incoming score parsing
func TestScoreParser(t *testing.T) {
	type Case struct {
		input  string
		output Score
	}
	data := [...]Case{
		{
			input:  "Wordle 771 3/6*\r\n\r\n⬛⬛⬛⬛🟩\r\n🟨🟩⬛⬛🟩\r\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "3", GameNumber: "771", Hardmode: "*", Win: "Y"},
		},
		{
			input:  "Wordle 733 X/6\r\n\r\n🟨⬛⬛⬛⬛\r\n⬛🟩⬛🟨⬛\r\n⬛🟩⬛🟨🟩\r\n⬛🟩🟩🟩🟩\r\n⬛🟩🟩🟩🟩\r\n⬛🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "7", GameNumber: "733", Win: "N"},
		},
		{
			input:  "Wordle 1,327 4/6\n⬜🟩🟨⬜⬜\n⬜⬜🟨⬜⬜\n⬜⬜⬜⬜🟨\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "4", GameNumber: "1,327", Win: "Y"},
		},
		{
			input:  "Daily Dordle 0597 4&6/7 🟨🟨⬜⬜⬜ 🟨⬜🟨⬜⬜ ⬜⬜⬜⬜⬜ ⬜⬜⬜⬜⬜ ⬜⬜🟨🟨⬜ ⬜⬜⬜⬜⬜ 🟩🟩🟩🟩🟩 ⬜🟨⬜⬜⬜ ⬛⬛⬛⬛⬛ ⬜🟩🟨🟩🟩 ⬛⬛⬛⬛⬛ 🟩🟩🟩🟩🟩 zaratustra.itch.io/dordle",
			output: Score{Game: "Daily Dordle", GameNumber: "0597", Score: "10", Win: "Y"},
		},
		{
			input:  "Daily Dordle 1112 X&X/7\n⬜⬜⬜⬜⬜ ⬜🟨🟨⬜⬜\n⬜🟨🟩⬜⬜ ⬜⬜⬜🟨⬜\n⬜⬜⬜⬜⬜ ⬜⬜⬜⬜⬜\n🟨⬜⬜⬜⬜ ⬜🟨⬜🟨⬜\n⬜⬜⬜⬜⬜ ⬜⬜⬜⬜⬜\n🟩⬜⬜🟨⬜ ⬜⬜⬜⬜⬜\n🟨⬜⬜🟩⬜ ⬜🟨⬜⬜⬜\nzaratustra.itch.io/dordle",
			output: Score{Game: "Daily Dordle", GameNumber: "1112", Score: "14", Win: "N"},
		},
		{
			input:  "Daily Octordle #553\r\n🔟7️⃣\r\n6️⃣8️⃣\r\n3️⃣5️⃣\r\n9️⃣🕚\r\nScore: 59",
			output: Score{Game: "Daily Octordle", Score: "59", GameNumber: "553", Win: "Y"},
		},
		{
			input:  "Daily Octordle #501\r\n6️⃣🟥\r\n5️⃣8️⃣\r\n3️⃣🟥\r\n🕐🔟\r\nScore: 73",
			output: Score{Game: "Daily Octordle", Score: "73", GameNumber: "501", Win: "N"},
		},
		{ // User text in middle of string (courtesy: Josh)
			input:  "Daily Octordle #1131 7️⃣🔟 5️⃣8️⃣ 🟥🟥 oof 🟥🕐 Score: 85",
			output: Score{Game: "Daily Octordle", Score: "85", GameNumber: "1131", Win: "N"},
		},
		{
			input:  "Daily Sequence Octordle #563 4️⃣5️⃣ 7️⃣8️⃣ 9️⃣🔟 🕚🕛 Score: 66",
			output: Score{Game: "Daily Sequence Octordle", Score: "66", GameNumber: "563", Win: "Y"},
		},
		{
			input:  "Connections \nPuzzle #51\n🟨🟨🟨🟨\n🟩🟩🟩🟩\n🟪🟪🟪🟪\n🟦🟦🟦🟦",
			output: Score{Game: "Connections", Score: "4", GameNumber: "51", Win: "Y"},
		},
		{
			input:  "Connections Puzzle #59 🟦🟦🟩🟦 🟦🟦🟦🟩 🟦🟦🟨🟩 🟦🟦🟪🟩",
			output: Score{Game: "Connections", Score: "7", GameNumber: "59", Win: "N"},
		},
		{
			input:  "Connections Puzzle #608 🟪🟨🟦🟦 🟪🟪🟪🟪 🟨🟦🟦🟦 🟩🟩🟩🟩 🟨🟨🟨🟨 🟦🟦🟦🟦",
			output: Score{Game: "Connections", Score: "6", GameNumber: "608", Win: "Y"},
		},
		{
			input:  "Connections Puzzle #613 🟨🟩🟨🟨 🟨🟨🟨🟨 🟦🟪🟦🟦 🟦🟪🟪🟦 🟩🟩🟩🟩 🟦🟪🟪🟦",
			output: Score{Game: "Connections", Score: "7", GameNumber: "613", Win: "N"},
		},
		{ // User text in middle of string (courtesy: Elly)
			input:  "Connections Puzzle #623 😢😂 🟪🟪🟦🟦 🟩🟩🟩🟩 🟨🟦🟨🟨 🟨🟨🟨🟨 🟪🟪🟦🟦 🟪🟪🟦🟪",
			output: Score{Game: "Connections", Score: "7", GameNumber: "623", Win: "N"},
		},
		{
			input:  "#Tradle #527 2/6 🟩🟩🟩🟩🟨 🟩🟩🟩🟩🟩 https://oec.world/en/tradle",
			output: Score{Game: "Tradle", Score: "2", GameNumber: "527", Win: "Y"},
		},
		{
			input:  "#Tradle #1072 X/6 🟩🟩🟩🟨⬜ 🟩🟩🟨⬜⬜ 🟩🟩🟩⬜⬜ 🟩🟩🟩🟩⬜ 🟩🟩🟩⬜⬜ 🟩🟩🟩⬜⬜ https://oec.world/en/games/tradle",
			output: Score{Game: "Tradle", Score: "7", GameNumber: "1072", Win: "N"},
		},
		{
			input:  "Strands #448\n“Get out the dust buster!”\n🔵💡🔵💡\n🔵💡🔵🟡\n💡🔵",
			output: Score{Game: "Strands", Score: "4", GameNumber: "448", Win: "Y"},
		},
		{
			input:  "Strands #449\n“Body language”\n💡🔵🔵🔵\n💡🔵🟡🔵\n🔵",
			output: Score{Game: "Strands", Score: "2", GameNumber: "449", Win: "Y"},
		},
	}
	for _, item := range data {
		score, err := ParseScoreFromContent(item.input)
		if err != nil {
			t.Fatalf(`TestScoreParser("%s") returned error: %v`, item.input, err)
		}
		if score.Game != item.output.Game {
			t.Fatalf("TestScoreParser [Game]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Game, item.output.Game)
		}
		if score.Score != item.output.Score {
			t.Fatalf("TestScoreParser [Score]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Score, item.output.Score)
		}
		if score.GameNumber != item.output.GameNumber {
			t.Fatalf("TestScoreParser [GameNumber]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.GameNumber, item.output.GameNumber)
		}
		if score.Hardmode != item.output.Hardmode {
			t.Fatalf("TestScoreParser [Hardmode]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Hardmode, item.output.Hardmode)
		}
		if score.Win != item.output.Win {
			t.Fatalf("TestScoreParser [Win]\n%s\nReturned:\n%s\nExpected:\n%s", item.input, score.Win, item.output.Win)
		}
	}
}
