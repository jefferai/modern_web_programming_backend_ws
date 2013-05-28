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

// STARTWGMNG OMIT
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
// ENDWGMNG OMIT

func (wg *wordgame) processMessage(reader io.Reader) {
	return
}
