// Package word_t defines word-related functions from word_t.c
// Word methods are now defined in the structs package
package word_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// NewWord constructs a new word (lines 28-33)
func NewWord(w string) structs.Word {
	a := structs.Word{Len: 0}
	a.SetWord(w)
	return a
}

