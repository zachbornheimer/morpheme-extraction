// Package word_t defines word data structures and functions from word_t.c
// Line-for-line translation of word_t.c lines 1-56
package word_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// SetWord sets the word in word_t struct (lines 19-24)
func SetWord(word *structs.Word, w string) {
	word.Len = len(w)
	word.Freq = 1 // initialized
	// Convert string to []byte with null terminator
	word.Str = append([]byte(w), 0)
}

// NewWord constructs a new word (lines 28-33)
func NewWord(w string) structs.Word {
	a := structs.Word{Len: 0}
	SetWord(&a, w)
	return a
}

// ReverseWord reverses a word (lines 36-44)
func ReverseWord(w structs.Word) structs.Word {
	temp := structs.Word{}
	temp.Len = w.Len
	temp.Freq = w.Freq
	temp.Str = []byte(functions.Reverse(string(w.Str)))
	return temp
}

// RemoveFirstChar removes first character from word (lines 47-53)
func RemoveFirstChar(w structs.Word) string {
	word := string(w.Str)
	if len(word) == 0 {
		return ""
	}
	finalized := functions.Reverse(word)
	if len(finalized) > 0 {
		finalized = finalized[:len(finalized)-1]
	}
	return functions.Reverse(finalized)
}

