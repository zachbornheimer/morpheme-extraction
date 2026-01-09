// Package structs contains all struct definitions from structs.h (lines 1-73)
package structs

import (
	"unicode/utf8"
)

// CharDoubleton represents struct char_doubleton (lines 11-14)
type CharDoubleton struct {
	First  byte // c
	Second int  // freq (must be int, not byte, to avoid overflow!)
}

// Word represents struct word_t (lines 16-20)
// Optimized: Uses native Go string instead of []byte
type Word struct {
	Str  string // word - native Go string (knows its own length)
	Freq int    // freq
}

// String returns the word as a string (zero-cost with native string)
func (w *Word) String() string {
	return w.Str
}

// Len returns the length of the word (for C compatibility)
func (w *Word) Len() int {
	return len(w.Str)
}

// Set sets the word string and resets frequency to 1
// Pointer receiver: modifies the Word in place
func (w *Word) Set(s string) {
	w.Str = s
	w.Freq = 1
}

// Reverse reverses the word string in place using built-in rune iteration
// Pointer receiver: modifies the Word in place
func (w *Word) Reverse() {
	runes := []rune(w.Str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	w.Str = string(runes)
}

// RemoveFirstChar removes the first UTF-8 character from the word
// Pointer receiver: modifies the Word in place
func (w *Word) RemoveFirstChar() {
	if len(w.Str) == 0 {
		return
	}
	// Find the byte index of the second rune
	_, i := utf8.DecodeRuneInString(w.Str)
	w.Str = w.Str[i:]
}

// WordArray represents struct word_t_array (lines 24-27)
type WordArray struct {
	Elems []*Word // elems
	Count int     // count (unsigned int)
}

// ToSlice converts WordArray to []string, collecting non-nil words
// Uses raw Str (including nulls) to match C behavior
func (wa *WordArray) ToSlice() []string {
	if wa == nil {
		return nil
	}
	result := make([]string, 0, wa.Count)
	for i := 0; i < wa.Count; i++ {
		if wa.Elems[i] != nil {
			word := string(wa.Elems[i].Str)
			if word != "" {
				result = append(result, word)
			}
		}
	}
	return result
}

// ToSet converts WordArray to StringSet (deduplicates)
func (wa *WordArray) ToSet() StringSet {
	if wa == nil {
		return NewStringSet(0)
	}
	result := NewStringSet(wa.Count)
	for i := 0; i < wa.Count; i++ {
		if wa.Elems[i] != nil {
			word := string(wa.Elems[i].Str)
			if word != "" {
				result.Add(word)
			}
		}
	}
	return result
}

// NgramHalfArray represents struct ngram_half_array (lines 29-31)
type NgramHalfArray struct {
	At [5]*WordArray // at[NGRAM_SIZE/2+1] where NGRAM_SIZE=9, so 9/2+1=5
}

// CountWords returns total count of words in this half-array
func (h *NgramHalfArray) CountWords(ngramHalfSize int) int {
	count := 0
	for i := 0; i < ngramHalfSize; i++ {
		if h.At[i] != nil {
			count += h.At[i].Count
		}
	}
	return count
}

// CollectWordsInto appends words from this half-array into the provided slice
// Returns the number of words added
func (h *NgramHalfArray) CollectWordsInto(arr []string, startIdx int, ngramHalfSize int) int {
	idx := startIdx
	for i := 0; i < ngramHalfSize; i++ {
		if h.At[i] != nil {
			for j := 0; j < h.At[i].Count; j++ {
				if h.At[i].Elems[j] != nil {
					word := string(h.At[i].Elems[j].Str)
					if word != "" {
						arr[idx] = word
						idx++
					}
				}
			}
		}
	}
	return idx - startIdx
}

// Ngram represents struct ngram_t (lines 33-39)
type Ngram struct {
	Word      Word             // word
	Before    NgramHalfArray   // before
	After     NgramHalfArray   // after
	RefsCount int              // refs_count
	Refs      []*Ngram         // refs (struct ngram_t**)
}

// CountContextWords returns total count of context words (Before + After)
func (ng *Ngram) CountContextWords(ngramHalfSize int) int {
	return ng.Before.CountWords(ngramHalfSize) + ng.After.CountWords(ngramHalfSize)
}

// CollectContextWords collects all context words into a pre-allocated slice
// Returns the number of words collected
func (ng *Ngram) CollectContextWords(arr []string, ngramHalfSize int) int {
	count := 0
	count += ng.Before.CollectWordsInto(arr, count, ngramHalfSize)
	count += ng.After.CollectWordsInto(arr, count, ngramHalfSize)
	return count
}

// Morpheme represents struct morpheme_t (lines 45-58)
type Morpheme struct {
	Regex               string   // regex
	Morpheme            string   // morpheme
	Freq                int      // freq
	WordsCount          int      // words_count
	Words               []*Word  // words
	Type                int      // type
	FrontRegex          string   // front_regex
	FrontRegexArr       []string // front_regex_arr (char**)
	BackRegex           string   // back_regex
	BackRegexArr        []string // back_regex_arr (char**)
	FrontRegexArrIndex  int      // front_regex_arr_index
	BackRegexArrIndex   int      // back_regex_arr_index
}

// WordsToSlice returns morpheme's words as a string slice
func (m *Morpheme) WordsToSlice() []string {
	result := make([]string, 0, m.WordsCount)
	for i := 0; i < m.WordsCount; i++ {
		if m.Words[i] != nil {
			word := string(m.Words[i].Str)
			if word != "" {
				result = append(result, word)
			}
		}
	}
	return result
}

// WordsToSet returns morpheme's words as a set (deduplicates)
func (m *Morpheme) WordsToSet() StringSet {
	result := NewStringSet(m.WordsCount)
	for i := 0; i < m.WordsCount; i++ {
		if m.Words[i] != nil {
			word := string(m.Words[i].Str)
			if word != "" {
				result.Add(word)
			}
		}
	}
	return result
}

// MorphemeList represents struct morpheme_list_t (lines 60-63)
type MorphemeList struct {
	List  []*Morpheme // list
	Count int         // count
}

// LexicalCategories represents struct lexical_categories_t (lines 65-69)
type LexicalCategories struct {
	Morpheme *Morpheme // morpheme
	RefCount int       // ref_count
	Refs     []*Ngram  // refs (struct ngram_t**)
}

// StringSet is a set of strings (memory-efficient, idiomatic Go)
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet with optional initial capacity
func NewStringSet(capacity int) StringSet {
	return make(StringSet, capacity)
}

// Add adds a string to the set
func (s StringSet) Add(str string) {
	s[str] = struct{}{}
}

// Contains checks if a string is in the set
func (s StringSet) Contains(str string) bool {
	_, exists := s[str]
	return exists
}

// Len returns the number of elements in the set
func (s StringSet) Len() int {
	return len(s)
}

// AddAll adds all strings from a slice
func (s StringSet) AddAll(strs []string) {
	for _, str := range strs {
		s.Add(str)
	}
}

