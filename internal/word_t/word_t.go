// Package word_t provides legacy function wrappers for Word methods
// Methods are now defined on structs.Word in the structs package
package word_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// NewWord constructs a new word (constructor function)
func NewWord(w string) structs.Word {
	return structs.Word{
		Str:  w,
		Freq: 1,
	}
}

// SetWord is a legacy wrapper for Word.Set()
func SetWord(word *structs.Word, w string) {
	word.Set(w)
}

// ReverseWord is a legacy wrapper that returns a new reversed Word
func ReverseWord(w structs.Word) structs.Word {
	result := w
	result.Reverse()
	return result
}

