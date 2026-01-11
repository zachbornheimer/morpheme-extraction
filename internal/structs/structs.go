// Package structs contains all struct definitions from structs.h (lines 1-73)
package structs

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
)

// CharDoubleton represents struct char_doubleton (lines 11-14)
type CharDoubleton struct {
	First  byte // c
	Second int  // freq (must be int, not byte, to avoid overflow!)
}

// Word represents struct word_t (lines 16-20)
type Word struct {
	Str  string // word (string, not []byte - idiomatic Go)
	Len  int    // i (unsigned int) - redundant with len(Str) but kept for compatibility
	Freq int    // freq
}

// String returns the word as a string (includes full Str with null terminator for compatibility)
func (w *Word) String() string {
	return string(w.Str)
}

// StringClean returns the word as a clean string without null terminator (uses Len field)
func (w *Word) StringClean() string {
	if w.Len > len(w.Str) {
		return string(w.Str)
	}
	return string(w.Str[:w.Len])
}

// SetWord sets the word string, length, and frequency
func (w *Word) SetWord(str string) {
	w.Str = str + constants.EmptyMorpheme
	w.Len = len(str)
	w.Freq = 1 // initialized
}

// Reverse returns a new Word with reversed string
func (w Word) Reverse() Word {
	temp := Word{}
	temp.Len = w.Len
	temp.Freq = w.Freq
	temp.Str = reverseString(w.Str)
	return temp
}

// RemoveFirstChar removes first character and returns clean string without marker
// Returns the result WITHOUT the marker byte (caller should use SetWord to add it back)
func (w Word) RemoveFirstChar() string {
	// Use StringClean to get string without marker
	word := w.StringClean()
	if len(word) == 0 {
		return ""
	}
	finalized := reverseString(word)
	if len(finalized) > 0 {
		finalized = finalized[:len(finalized)-1]
	}
	return reverseString(finalized)
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
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

// NewMorpheme creates a new morpheme with the given morpheme string
func NewMorpheme(morpheme string) Morpheme {
	return Morpheme{
		Morpheme:   morpheme,
		Freq:       1,
		WordsCount: 0,
		Words:      make([]*Word, 0),
	}
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

// AddWord adds word to morpheme's word list
func (m *Morpheme) AddWord(w Word) {
	m.WordsCount++
	newWords := make([]*Word, m.WordsCount)
	copy(newWords, m.Words)
	newWords[m.WordsCount-1] = &w
	m.Words = newWords
}

// GenRegex generates regex pattern for this morpheme based on two example words
func (m *Morpheme) GenRegex(one, two string, placeOne, placeTwo int) {
	// Validate morpheme placement
	if !m.isValidPlacement(one, two, placeOne, placeTwo) {
		return
	}

	m.Type = constants.UNDEF

	// Handle STEM case (morpheme equals entire word)
	if one == m.Morpheme || two == m.Morpheme {
		m.setStemRegex()
		return
	}

	// Extract context (prefix/suffix) from both words
	prefix1, suffix1 := extractContext(one, placeOne, len(m.Morpheme))
	prefix2, suffix2 := extractContext(two, placeTwo, len(m.Morpheme))

	// Build character class regex patterns
	factory := newRegexFactory()
	m.FrontRegex, m.FrontRegexArr, m.FrontRegexArrIndex = factory.buildCharacterClass(prefix1, prefix2, true)
	m.BackRegex, m.BackRegexArr, m.BackRegexArrIndex = factory.buildCharacterClass(suffix1, suffix2, false)
	m.Regex = m.FrontRegex + m.Morpheme + m.BackRegex
}

// MergeRules merges regex rules from multiple morpheme observations
func (m *Morpheme) MergeRules() {
	front := ")"
	for i := 0; i <= m.FrontRegexArrIndex; i++ {
		front += "]"
		front += m.FrontRegexArr[i]
		front += "["
	}
	front += "("
	front = reverseString(front)

	back := ")"
	for i := 0; i <= m.BackRegexArrIndex; i++ {
		back += "]"
		back += m.BackRegexArr[i]
		back += "["
	}
	back += "("
	back = reverseString(back)

	m.Regex = ""
	if m.FrontRegex != "^" {
		m.FrontRegex = front
		m.Regex += front
	} else {
		m.Regex += m.FrontRegex
	}

	m.Regex += m.Morpheme

	if m.BackRegex != "$" {
		m.BackRegex = back
		m.Regex += back
	} else {
		m.Regex += m.BackRegex
	}
}

// isValidPlacement validates morpheme can be placed at specified positions
func (m *Morpheme) isValidPlacement(one, two string, placeOne, placeTwo int) bool {
	// Check morpheme is not empty
	if m.Morpheme == constants.EmptyMorpheme || len(m.Morpheme) == 0 {
		return false
	}

	// Check morpheme fits in both words
	if len(m.Morpheme) > len(one) || len(m.Morpheme) > len(two) {
		return false
	}

	// Verify morpheme matches at specified positions
	morphemeLen := len(m.Morpheme)

	extractedOne := extractMorphemeAt(one, placeOne, morphemeLen)
	extractedTwo := extractMorphemeAt(two, placeTwo, morphemeLen)

	return extractedOne == m.Morpheme && extractedTwo == m.Morpheme
}

// setStemRegex configures morpheme as a STEM (whole word)
func (m *Morpheme) setStemRegex() {
	m.Regex = "^" + m.Morpheme + "$"
	m.FrontRegex = "^"
	m.BackRegex = "$"
	m.Type = constants.STEM
}

// extractMorphemeAt extracts substring at position with marker byte handling
func extractMorphemeAt(word string, position, length int) string {
	if position+length > len(word) {
		return ""
	}

	// Extract the morpheme substring
	extracted := make([]byte, length+1)
	for i := 0; i < length; i++ {
		extracted[i] = word[position+i]
	}

	// Add marker byte if configured
	if constants.EmptyMorphemeByte != 0 {
		extracted[length] = constants.EmptyMorphemeByte
	}

	return string(extracted[:length])
}

// extractContext splits word into prefix and suffix around morpheme
func extractContext(word string, position, morphemeLen int) (prefix, suffix string) {
	prefix = word[:position]

	if position+morphemeLen <= len(word) {
		suffix = word[position+morphemeLen:]
	}

	return prefix, suffix
}

// regexFactory builds character class regex patterns
type regexFactory struct{}

func newRegexFactory() *regexFactory {
	return &regexFactory{}
}

// buildCharacterClass creates regex pattern from two context strings
// isPrefix determines if building front (true) or back (false) pattern
func (rf *regexFactory) buildCharacterClass(ctx1, ctx2 string, isPrefix bool) (regex string, charArray []string, lastIndex int) {
	// Reverse for processing (builds right-to-left)
	ctx1 = reverseString(ctx1)
	ctx2 = reverseString(ctx2)

	maxLen := len(ctx1)
	if len(ctx2) > maxLen {
		maxLen = len(ctx2)
	}

	// Allocate array (different sizes for front vs back in original C)
	arraySize := maxLen * 2
	if !isPrefix {
		arraySize = maxLen * 4
	}
	charArray = make([]string, arraySize)
	lastIndex = -1

	// Build character class pattern
	pattern := ")"
	for i := 0; i < maxLen; i++ {
		if i >= len(ctx1) && i >= len(ctx2) {
			break
		}

		charArray[i] = ""
		pattern += "]"

		// Add characters from both contexts at this position
		if i < len(ctx2) {
			char := string(ctx2[i])
			pattern += char
			charArray[i] += char
		}

		if i < len(ctx1) {
			char := string(ctx1[i])
			pattern += char
			charArray[i] += char
		}

		lastIndex = i
		pattern += "["
	}

	pattern += "("
	regex = reverseString(pattern)

	return regex, charArray, lastIndex
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

