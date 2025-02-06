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
	// TODO: Test cases for other games
	//   [
	//     "Connections \nPuzzle #51\n🟨🟨🟨🟨\n🟩🟩🟩🟩\n🟪🟪🟪🟪\n🟦🟦🟦🟦",
	//     %{"game" => "Connections", "game_no" => "51", "score" => "0", "win" => true}
	//   ],
	//   [
	//     "Connections Puzzle #59 🟦🟦🟩🟦 🟦🟦🟦🟩 🟦🟦🟨🟩 🟦🟦🟪🟩",
	//     %{"game" => "Connections", "game_no" => "59", "score" => "4", "win" => false}
	//   ],
	//   [
	//     "#Tradle #527 2/6 🟩🟩🟩🟩🟨 🟩🟩🟩🟩🟩 https://oec.world/en/tradle",
	//     %{"game" => "Tradle", "game_no" => "527", "score" => "2", "win" => true}
	//   ],
	//   [
	//     "Daily Dordle 0597 4&6/7 🟨🟨⬜⬜⬜ 🟨⬜🟨⬜⬜ ⬜⬜⬜⬜⬜ ⬜⬜⬜⬜⬜ ⬜⬜🟨🟨⬜ ⬜⬜⬜⬜⬜ 🟩🟩🟩🟩🟩 ⬜🟨⬜⬜⬜ ⬛⬛⬛⬛⬛ ⬜🟩🟨🟩🟩 ⬛⬛⬛⬛⬛ 🟩🟩🟩🟩🟩 zaratustra.itch.io/dordle",
	//     %{"game" => "Daily Dordle", "game_no" => "0597", "score" => "10", "win" => true}
	//   ]
	data := [...]Case{
		{
			input:  "Wordle 771 3/6*\r\n\r\n⬛⬛⬛⬛🟩\r\n🟨🟩⬛⬛🟩\r\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "3", GameNumber: "771", Hardmode: "*", Win: "Y"},
		},
		{
			input:  "Wordle 733 X/6\r\n\r\n🟨⬛⬛⬛⬛\r\n⬛🟩⬛🟨⬛\r\n⬛🟩⬛🟨🟩\r\n⬛🟩🟩🟩🟩\r\n⬛🟩🟩🟩🟩\r\n⬛🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "7", GameNumber: "733", Hardmode: "", Win: "N"},
		},
		{
			input:  "Wordle 1,327 4/6\n⬜🟩🟨⬜⬜\n⬜⬜🟨⬜⬜\n⬜⬜⬜⬜🟨\n🟩🟩🟩🟩🟩",
			output: Score{Game: "Wordle", Score: "4", GameNumber: "1,327", Hardmode: "", Win: "Y"},
		},
		{
			input:  "Daily Octordle #553\r\n🔟7️⃣\r\n6️⃣8️⃣\r\n3️⃣5️⃣\r\n9️⃣🕚\r\nScore: 59",
			output: Score{Game: "Daily Octordle", Score: "59", GameNumber: "553", Hardmode: "", Win: "Y"},
		},
		{
			input:  "Daily Octordle #501\r\n6️⃣🟥\r\n5️⃣8️⃣\r\n3️⃣🟥\r\n🕐🔟\r\nScore: 73",
			output: Score{Game: "Daily Octordle", Score: "73", GameNumber: "501", Hardmode: "", Win: "N"},
		},
		{
			input:  "Daily Sequence Octordle #563 4️⃣5️⃣ 7️⃣8️⃣ 9️⃣🔟 🕚🕛 Score: 66",
			output: Score{Game: "Daily Sequence Octordle", Score: "66", GameNumber: "563", Hardmode: "", Win: "Y"},
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
