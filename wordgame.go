package main

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"strconv"
	"strings"
)

type Game struct {
	NumGuesses int64
	Word       string
	MaskedWord string
}

var (
	games          = map[int64]Game{}
	words          = []string{"aardvark", "labradoodle", "kittycat", "porpoise", "brontosaurus"}
	underscoreFunc = func(r rune) rune { return '_' }
)

// This ignores safety for simplicity
func UnhideByte(guess byte, word string, maskedword string) string {
	wordbytes := []byte(word)
	maskedwordbytes := []byte(maskedword)
	for pos := range wordbytes {
		if wordbytes[pos] == guess {
			maskedwordbytes[pos] = wordbytes[pos]
		}
	}
	return string(maskedwordbytes)
}

func (wg *wordgame) makeNewGame() int64 {

	// skipping some error handling for presentation reasons
	randint, _ := rand.Int(rand.Reader, big.NewInt(int64(math.MaxUint32)))

	currgame := randint.Int64()
	wordnum := currgame % int64(len(words))

	games[currgame] = Game{
		NumGuesses: 0,
		Word:       words[wordnum],
		MaskedWord: strings.Map(underscoreFunc, words[wordnum]),
	}

	wg.send <- []byte("Hello, welcome to the game!\n")
	return currgame
}

func (wg *wordgame) removeGame() {
	delete(games, wg.gameId)
}

func (wg *wordgame) processMessage(reader io.Reader) {
	// In practice you will often want to use a LimitedReader to protect against malicious/bad inputs
	message, err := ioutil.ReadAll(reader)
	if err != nil || len(message) == 0 {
		return
	}

	stringMsg := strings.TrimSpace(string(message))
	
	if len(stringMsg) == 0 {
		wg.send <- []byte("You need to give me something to work with, here...\n")
		return
	}

	game := games[wg.gameId]
	game.NumGuesses++

  if len(stringMsg) > 1 { //it's a guess at the word, not a new character
		if stringMsg == game.Word {
			wg.send <- []byte("You got it!\n")
			wg.send <- []byte("__MAGIC_CLOSE_VALUE__")
		} else {
			wg.send <- []byte("Nice try, but no cigar...that costs you a guess!\n")
		}
		return
	}

	game.MaskedWord = UnhideByte(stringMsg[0], game.Word, game.MaskedWord)
	games[wg.gameId] = game
	wg.send <- []byte("Number of guesses: " + strconv.Itoa(int(game.NumGuesses)) + "\n")
  wg.send <- []byte("Current word: " + game.MaskedWord + "\n")
	return
}
