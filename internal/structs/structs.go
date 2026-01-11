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

// GenRegex generates a regex pattern for matching words containing this morpheme.
//
// Algorithm:
// 1. Takes two example words where the morpheme appears at different positions
// 2. Extracts what comes before/after the morpheme in each word
// 3. Builds character class patterns [xy] matching either context
// 4. Stores character arrays for later merging with additional examples
//
// Example:
//   Morpheme: "ing"
//   Word 1: "running"  (morpheme at position 4)
//   Word 2: "walking"  (morpheme at position 4)
//
//   Before morpheme: "runn" vs "walk"
//   After morpheme:  "" vs ""
//
//   Result: Pattern matching variations of characters before "ing"
//
// Why store both Regex and RegexArr?
// - Regex: The current regex pattern as a string (e.g., "([rnuwalkh])ing")
// - RegexArr: Character alternatives at each position for merging
// - When we see more examples, we merge character sets without regenerating from scratch
func (m *Morpheme) GenRegex(one, two string, placeOne, placeTwo int) {
	// Validate morpheme can be placed at specified positions
	if !m.isValidPlacement(one, two, placeOne, placeTwo) {
		return
	}

	m.Type = constants.UNDEF

	// STEM: morpheme equals entire word (e.g., "the" in word "the")
	if one == m.Morpheme || two == m.Morpheme {
		m.SetType(constants.STEM)
		return
	}

	// Extract surrounding context from both example words
	// beforeMorpheme1: characters before morpheme in word 1
	// afterMorpheme1:  characters after morpheme in word 1
	beforeMorpheme1, afterMorpheme1 := splitAroundMorpheme(one, placeOne, len(m.Morpheme))
	beforeMorpheme2, afterMorpheme2 := splitAroundMorpheme(two, placeTwo, len(m.Morpheme))

	// Build character class patterns for before/after morpheme
	factory := newRegexFactory()
	m.FrontRegex, m.FrontRegexArr, m.FrontRegexArrIndex = factory.buildCharacterClass(beforeMorpheme1, beforeMorpheme2)
	m.BackRegex, m.BackRegexArr, m.BackRegexArrIndex = factory.buildCharacterClass(afterMorpheme1, afterMorpheme2)

	// Complete pattern: (before)morpheme(after)
	m.Regex = m.FrontRegex + m.Morpheme + m.BackRegex
}

// SetType sets the morpheme type and configures appropriate regex pattern.
//
// Type constants:
// - STEM: Morpheme is a complete word (e.g., "the")
//   Pattern: ^morpheme$ (exact match only)
// - PREFIX/SUFFIX/INFIX: Determined later based on position analysis
//   Pattern: Character class matching context variations
func (m *Morpheme) SetType(morphemeType int) {
	m.Type = morphemeType

	if morphemeType == constants.STEM {
		// STEM regex: exact match only (^word$)
		m.Regex = "^" + m.Morpheme + "$"
		m.FrontRegex = "^"
		m.BackRegex = "$"
	}
}

// MergeRules rebuilds regex patterns after merging character arrays from multiple observations.
//
// When we see a morpheme in multiple words, we accumulate character alternatives:
//   "running" → FrontRegexArr[0] = "r", [1] = "u", [2] = "n", [3] = "n"
//   "walking" → FrontRegexArr[0] = "rw", [1] = "ua", [2] = "nl", [3] = "nk"
//
// This function reconstructs the regex from the merged arrays:
//   Result: ([rw])([ua])([nl])([nk])ing
//
// Why not regenerate from scratch?
// - We may have merged 10+ words and don't want to re-scan all of them
// - The arrays already contain the merged character sets
func (m *Morpheme) MergeRules() {
	// Rebuild front pattern from character array
	front := ")"
	for i := 0; i <= m.FrontRegexArrIndex; i++ {
		front += "]"
		front += m.FrontRegexArr[i]
		front += "["
	}
	front += "("
	front = reverseString(front)

	// Rebuild back pattern from character array
	back := ")"
	for i := 0; i <= m.BackRegexArrIndex; i++ {
		back += "]"
		back += m.BackRegexArr[i]
		back += "["
	}
	back += "("
	back = reverseString(back)

	// Assemble complete regex
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

// isValidPlacement validates that the morpheme actually appears at the specified positions in both words.
//
// Example:
//   morpheme: "ing"
//   word1: "running" at position 4 → "runn" + "ing" ✓
//   word2: "walking" at position 4 → "walk" + "ing" ✓
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

// extractMorphemeAt extracts substring at position with marker byte handling.
// Includes marker byte compatibility for C interop (adds \x00 terminator if configured).
func extractMorphemeAt(word string, position, length int) string {
	if position+length > len(word) {
		return ""
	}

	// Extract the morpheme substring
	extracted := make([]byte, length+1)
	for i := 0; i < length; i++ {
		extracted[i] = word[position+i]
	}

	// Add marker byte if configured (C compatibility)
	if constants.EmptyMorphemeByte != 0 {
		extracted[length] = constants.EmptyMorphemeByte
	}

	return string(extracted[:length])
}

// splitAroundMorpheme splits a word into the parts before and after the morpheme.
//
// Example:
//   word: "running", position: 4, morphemeLen: 3
//   Returns: beforeMorpheme="runn", afterMorpheme=""
//
//   word: "reinvent", position: 2, morphemeLen: 2
//   Returns: beforeMorpheme="re", afterMorpheme="invent"
func splitAroundMorpheme(word string, position, morphemeLen int) (beforeMorpheme, afterMorpheme string) {
	beforeMorpheme = word[:position]

	if position+morphemeLen <= len(word) {
		afterMorpheme = word[position+morphemeLen:]
	}

	return beforeMorpheme, afterMorpheme
}

// regexFactory builds character class regex patterns from context strings.
//
// Character class pattern:
// Given two contexts "abc" and "xyz", builds: ([ax])([by])([cz])
// This matches either character at each position.
type regexFactory struct{}

func newRegexFactory() *regexFactory {
	return &regexFactory{}
}

// buildCharacterClass creates a regex pattern matching either context at each character position.
//
// Algorithm:
// 1. Takes two context strings (e.g., "runn" and "walk" before "ing")
// 2. Reverses them for right-to-left processing
// 3. At each position i, creates character class [xy] matching either char
// 4. Stores character alternatives in array for later merging
//
// Example:
//   context1: "runn"
//   context2: "walk"
//
//   Reversed: "nnur" vs "klaw"
//   Position 0: [nk]  (matches 'n' or 'k')
//   Position 1: [nl]  (matches 'n' or 'l')
//   Position 2: [ua]  (matches 'u' or 'a')
//   Position 3: [rw]  (matches 'r' or 'w')
//
//   Pattern: ([rw])([ua])([nl])([nk])
//   Array: ["rw", "ua", "nl", "nk"] for merging with more examples later
//
// Why reverse?
// - Morphemes often vary at the end ("running", "walking" → "nn" vs "lk" before "ing")
// - Building from the end captures the most relevant variations
// - Historical: C code built patterns this way for prefix/suffix detection
//
// Returns:
// - regex: The character class pattern string
// - charArray: Character alternatives at each position (for merging)
// - lastIndex: Highest index in charArray (-1 if empty)
func (rf *regexFactory) buildCharacterClass(context1, context2 string) (regex string, charArray []string, lastIndex int) {
	// Reverse strings to build pattern from end to start
	context1 = reverseString(context1)
	context2 = reverseString(context2)

	// Find maximum length
	maxLen := len(context1)
	if len(context2) > maxLen {
		maxLen = len(context2)
	}

	// Allocate character array (overallocate for safety, original C used 2x or 4x)
	charArray = make([]string, maxLen*4)
	lastIndex = -1

	// Build pattern backwards: )][char2char1][
	// Will be reversed at the end to: ([char1char2])
	pattern := ")"
	for i := 0; i < maxLen; i++ {
		if i >= len(context1) && i >= len(context2) {
			break
		}

		charArray[i] = ""
		pattern += "]"

		// Accumulate characters from both contexts at position i
		if i < len(context2) {
			char := string(context2[i])
			pattern += char
			charArray[i] += char
		}

		if i < len(context1) {
			char := string(context1[i])
			pattern += char
			charArray[i] += char
		}

		lastIndex = i
		pattern += "["
	}

	pattern += "("
	// Reverse to get final pattern: ([char1char2])([...])
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

