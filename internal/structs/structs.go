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

// AddElement adds a new element slot to the array and returns its index
// Matches AddNgramElement from ngram_t.c (lines 22-28)
func (wa *WordArray) AddElement() int {
	wa.Count++
	newElems := make([]*Word, wa.Count+1)
	copy(newElems, wa.Elems)
	wa.Elems = newElems
	return wa.Count - 1
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

// CollectWordsIntoSet collects all words directly into a StringSet
// This avoids intermediate slice allocations for performance
func (h *NgramHalfArray) CollectWordsIntoSet(set *StringSet, ngramHalfSize int) {
	for i := 0; i < ngramHalfSize; i++ {
		if h.At[i] != nil {
			for j := 0; j < h.At[i].Count; j++ {
				if h.At[i].Elems[j] != nil {
					word := h.At[i].Elems[j].Str
					if word != "" {
						set.Add(word)
					}
				}
			}
		}
	}
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

// CollectContextWordsIntoSet collects all context words directly into a StringSet
// This avoids intermediate slice allocations for performance
func (ng *Ngram) CollectContextWordsIntoSet(set *StringSet, ngramHalfSize int) {
	ng.Before.CollectWordsIntoSet(set, ngramHalfSize)
	ng.After.CollectWordsIntoSet(set, ngramHalfSize)
}

// AddSimilarRef adds a reference to a similar ngram
// Matches AddSimilarNgramRef from ngram_t.c (lines 180-195)
// Uses append() instead of manual slice copy for O(1) amortized performance
func (ng *Ngram) AddSimilarRef(ngramRef *Ngram) {
	ng.Refs = append(ng.Refs, ngramRef)
	ng.RefsCount++
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

// AddWord adds a word to the morpheme's word list
// Simple data manipulation - no external dependencies
func (m *Morpheme) AddWord(w Word) {
	m.WordsCount++
	newWords := make([]*Word, m.WordsCount)
	copy(newWords, m.Words)
	newWords[m.WordsCount-1] = &w
	m.Words = newWords
}

// GenRegex and MergeRules are defined in morpheme_t package
// (they require functions and constants imports)

// MorphemeList represents struct morpheme_list_t (lines 60-63)
type MorphemeList struct {
	List  []*Morpheme // list
	Count int         // count
}

// Add adds a morpheme to the list
// Translated from morpheme_list_t.c add_morpheme (lines 15-22)
// Uses append instead of manual copy to avoid O(n²) performance
func (ml *MorphemeList) Add(morpheme Morpheme) {
	ml.List = append(ml.List, &morpheme)
	ml.Count++
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

// Clear removes all elements from the set
func (s StringSet) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// SimilarityPercent calculates the percentage of overlap between two sets
// Returns the percentage of non-unique elements (duplicates between sets)
// This matches the original C algorithm's calculation
func (s StringSet) SimilarityPercent(other *StringSet) float64 {
	// totalCount = total words from both ngrams
	totalCount := s.Len() + other.Len()
	
	if totalCount == 0 {
		return 0
	}
	
	// uniqueCount = number of distinct words across both sets
	uniqueCount := s.Len()  // Start with all elements from set A
	for k := range *other {
		if !s.Contains(k) {
			uniqueCount++  // Add unique elements from set B
		}
	}
	
	// Subtract 1 from total (line 170 in C: arr_count -= 1;)
	totalCount--
	
	if totalCount == 0 {
		return 0
	}
	
	// Calculate similarity: (total - unique) / total * 100
	// Line 171 in C: percent_similar = 100.00 * ((arr_count - count) / arr_count)
	return 100.0 * (float64(totalCount-uniqueCount) / float64(totalCount))
}

