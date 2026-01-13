// Package structs contains all struct definitions from structs.h (lines 1-73)
package structs

import (
	"strings"
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
)

// CharDoubleton represents struct char_doubleton (lines 11-14)
type CharDoubleton struct {
	First  byte // c
	Second int  // freq (must be int, not byte, to avoid overflow!)
}

// Word represents struct word_t (lines 16-20)
// Future: Will be replaced by direct WordIdentity references for full interning
type Word struct {
	Identity *dictionary.WordIdentity // If set, use interned identity (future: always set)
	Str      string                   // word (string, not []byte - idiomatic Go) [Deprecated: use Identity]
	Len      int                      // i (unsigned int) - redundant with len(Str) but kept for compatibility
	Freq     int                      // freq [Deprecated: use Identity.Frequency]
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
// Uses dictionary for interning (future-proof)
func (w *Word) SetWord(str string) {
	// Get interned identity from dictionary
	w.Identity = dictionary.Dictionary.Word(str)

	// Keep legacy fields for compatibility (will be removed later)
	w.Str = str + constants.EmptyMorpheme
	w.Len = len(str)
	w.Freq = 1 // initialized (deprecated, use Identity.Frequency)
}

// GetIdentity returns the word identity, creating it if needed
func (w *Word) GetIdentity() *dictionary.WordIdentity {
	if w.Identity == nil && w.Str != "" {
		// Migrate legacy Word to use identity
		cleanStr := w.StringClean()
		w.Identity = dictionary.Dictionary.Word(cleanStr)
	}
	return w.Identity
}

// IncFreq increments frequency at center position
// Delegates to identity if available, otherwise uses legacy Freq
func (w *Word) IncFreq() {
	if w.Identity != nil {
		w.Identity.Frequency.IncCenter()
	} else {
		w.Freq++
	}
}

// GetFreq returns frequency at center position
func (w *Word) GetFreq() int {
	if w.Identity != nil {
		return w.Identity.Frequency.GetCenter()
	}
	return w.Freq
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
// Uses idiomatic Go approach with UTF-8/Unicode support
func (w Word) RemoveFirstChar() string {
	// Use StringClean to get string without marker
	word := w.StringClean()
	if len(word) == 0 {
		return ""
	}

	// Convert to runes for proper Unicode handling
	runes := []rune(word)
	if len(runes) <= 1 {
		return ""
	}

	// Simply slice from second rune onwards (idiomatic Go)
	return string(runes[1:])
}

// runeSlicePool reuses rune slice allocations for string reversal
// Reduces GC pressure by pooling temporary buffers
var runeSlicePool = sync.Pool{
	New: func() any {
		// Pre-allocate reasonable size for most words
		s := make([]rune, 0, 128)
		return &s
	},
}

// reverseString reverses a string with Unicode/UTF-8 support
// Uses sync.Pool to reduce allocations and GC pressure
func reverseString(s string) string {
	if s == "" {
		return ""
	}

	// Get pooled rune slice
	runesPtr := runeSlicePool.Get().(*[]rune)
	runes := (*runesPtr)[:0] // Reset length, keep capacity

	// Convert string to runes
	for _, r := range s {
		runes = append(runes, r)
	}

	// Reverse in place
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Convert back to string
	result := string(runes)

	// Return slice to pool for reuse
	*runesPtr = runes[:0]
	runeSlicePool.Put(runesPtr)

	return result
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

// CollectIdentitiesInto appends Identity pointers from this half-array into the provided slice
// Returns the number of identities added (uses O(1) pointer operations)
func (h *NgramHalfArray) CollectIdentitiesInto(arr []*dictionary.WordIdentity, startIdx int, ngramHalfSize int) int {
	idx := startIdx
	for i := 0; i < ngramHalfSize; i++ {
		if h.At[i] != nil {
			for j := 0; j < h.At[i].Count; j++ {
				if h.At[i].Elems[j] != nil && h.At[i].Elems[j].Identity != nil {
					arr[idx] = h.At[i].Elems[j].Identity
					idx++
				}
			}
		}
	}
	return idx - startIdx
}

// Ngram represents struct ngram_t (lines 33-39)
// NOTE: Refs and RefsCount removed - now using SemanticGraph for relationships
type Ngram struct {
	Word    Word           // word
	Before  NgramHalfArray // before
	After   NgramHalfArray // after
	MinHash [128]uint64    // MinHash signature for fast similarity filtering
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

// CollectContextIdentities collects all context word identities into a pre-allocated slice
// Returns the number of identities collected (O(1) pointer operations)
func (ng *Ngram) CollectContextIdentities(arr []*dictionary.WordIdentity, ngramHalfSize int) int {
	count := 0
	count += ng.Before.CollectIdentitiesInto(arr, count, ngramHalfSize)
	count += ng.After.CollectIdentitiesInto(arr, count, ngramHalfSize)
	return count
}

// RegexPattern encapsulates a regex pattern built from character alternatives at each position.
//
// Example:
//   After adding "walk": charArray = ["w", "a", "l", "k"]
//   After adding "run":  charArray = ["wr", "au", "ln", "k"]
//   Pattern: ([wr])([au])([ln])([k])
type RegexPattern struct {
	charArray []string // Character alternatives at each position (e.g., ["wr", "au", "ln", "k"])
	lastIndex int      // Highest index with data in charArray (-1 if empty)
	fixed     string   // Fixed pattern for STEM markers ("^" or "$")
}

// NewFixedPattern creates a pattern with a fixed string (for STEM markers)
func NewFixedPattern(pattern string) *RegexPattern {
	return &RegexPattern{fixed: pattern, lastIndex: -1}
}

// AddContext adds a context string to the character array.
// Characters are merged position-by-position (reversed for morpheme variation capture).
//
// Example:
//   Pattern.AddContext("walk") → charArray = ["w", "a", "l", "k"]
//   Pattern.AddContext("run")  → charArray = ["wr", "au", "ln", "kn"]
func (rp *RegexPattern) AddContext(context string) {
	if rp == nil {
		return
	}

	// Empty context is valid - it just doesn't add any characters
	// (morpheme appears at word boundary)
	if context == "" {
		return
	}

	// Reverse context to build from end (where morphemes vary most)
	reversed := reverseString(context)
	runes := []rune(reversed)

	// Ensure charArray is large enough
	if len(runes) > len(rp.charArray) {
		newArray := make([]string, len(runes))
		copy(newArray, rp.charArray)
		rp.charArray = newArray
	}

	// Merge characters at each position (avoid duplicates)
	for i, char := range runes {
		charStr := string(char)
		if !containsChar(rp.charArray[i], charStr) {
			rp.charArray[i] += charStr
		}
		// Update lastIndex to track highest used position
		if i > rp.lastIndex {
			rp.lastIndex = i
		}
	}
}

// String returns the regex pattern string, built lazily from character array.
func (rp *RegexPattern) String() string {
	if rp == nil {
		return ""
	}

	// Return fixed pattern for STEM markers
	if rp.fixed != "" {
		return rp.fixed
	}

	// Build pattern from charArray: )][chars][chars][(  → reversed → ([chars])([chars])
	// Even empty charArray produces "()" pattern (not "")
	// IMPORTANT: Include ALL slots up to lastIndex, even if empty, to match old MergeRules behavior

	// Use strings.Builder to avoid repeated string allocations
	var builder strings.Builder
	builder.Grow((rp.lastIndex + 1) * 4) // Pre-allocate: "][" + avg chars per slot
	builder.WriteByte(')')
	for i := 0; i <= rp.lastIndex; i++ {
		builder.WriteByte(']')
		builder.WriteString(rp.charArray[i]) // Include even if empty
		builder.WriteByte('[')
	}
	builder.WriteByte('(')

	return reverseString(builder.String())
}

// MergeWith merges another RegexPattern's character alternatives into this one.
// Handles alignment when patterns have different lengths (right-aligned).
// Used by FuseRegex to combine patterns from multiple word observations.
//
// Example merging "walk" and "run" contexts:
//   this:  ["w", "a", "l", "k"] (lastIndex=3)
//   other: ["r", "u", "n"]      (lastIndex=2)
//   Result: ["wr", "au", "ln", "k"] (lastIndex=3, right-aligned)
func (rp *RegexPattern) MergeWith(other *RegexPattern) {
	if rp == nil || other == nil {
		return
	}

	// Can't merge fixed patterns (^ or $)
	if rp.fixed != "" || other.fixed != "" {
		return
	}

	// Calculate lengths (lastIndex is 0-based, so add 1)
	thisLen := rp.lastIndex + 1
	otherLen := other.lastIndex + 1

	// Determine which is longer
	longest := thisLen
	if otherLen > longest {
		longest = otherLen
	}

	// Ensure our charArray is large enough
	if longest > len(rp.charArray) {
		newArray := make([]string, longest)
		// Copy existing data
		for i := 0; i < len(rp.charArray); i++ {
			newArray[i] = rp.charArray[i]
		}
		rp.charArray = newArray
	}

	// Merge characters position-by-position (right-aligned)
	// Positions are aligned from the right because morphemes vary at the end
	for j := 0; j < longest; j++ {
		// Calculate indices into each array (right-aligned)
		thisIdx := j - (longest - thisLen)
		otherIdx := j - (longest - otherLen)

		// Initialize position if needed
		if j >= len(rp.charArray) {
			break // Safety check
		}

		// Build merged string for this position
		merged := ""

		// Add characters from this array
		if thisIdx >= 0 && thisIdx < len(rp.charArray) {
			merged += rp.charArray[thisIdx]
		}

		// Add characters from other array
		if otherIdx >= 0 && otherIdx <= other.lastIndex && otherIdx < len(other.charArray) {
			merged += other.charArray[otherIdx]
		}

		// Deduplicate merged string (matches original FuseRegex behavior exactly)
		if merged != "" {
			deduplicated := deduplicateChars(merged)
			if thisIdx >= 0 && thisIdx < len(rp.charArray) {
				rp.charArray[thisIdx] = deduplicated
			} else if j < len(rp.charArray) {
				rp.charArray[j] = deduplicated
			}
		}
	}

	// Update lastIndex to the maximum
	if longest-1 > rp.lastIndex {
		rp.lastIndex = longest - 1
	}
}

// containsChar checks if a character already exists in the string
func containsChar(s, char string) bool {
	for _, c := range s {
		if string(c) == char {
			return true
		}
	}
	return false
}

// deduplicateChars removes duplicate characters from a string, preserving order.
// Matches the behavior of functions.Uniq() to ensure exact compatibility.
func deduplicateChars(s string) string {
	if s == "" {
		return ""
	}

	seen := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			break
		}
		// Check if we've seen this character before
		found := false
		for _, b := range seen {
			if b == s[i] {
				found = true
				break
			}
		}
		if !found {
			seen = append(seen, s[i])
		}
	}
	return string(seen)
}

// Regex encapsulates the complete morpheme regex pattern.
// Composes front pattern + morpheme + back pattern on demand.
type Regex struct {
	Front    *RegexPattern // Pattern for context before morpheme (e.g., "([rw])")
	Back     *RegexPattern // Pattern for context after morpheme (e.g., "([sk])")
	morpheme string        // The morpheme itself (e.g., "ing")
}

// String returns the complete regex pattern
func (r *Regex) String() string {
	if r == nil {
		return ""
	}
	return r.Front.String() + r.morpheme + r.Back.String()
}

// SetStem configures as a STEM pattern (exact match only)
func (r *Regex) SetStem(morpheme string) {
	r.morpheme = morpheme
	r.Front = NewFixedPattern("^")
	r.Back = NewFixedPattern("$")
}

// Morpheme represents struct morpheme_t (lines 45-58)
type Morpheme struct {
	Pattern    *Regex // Complete regex pattern (encapsulates front + morpheme + back)
	Morpheme   string // The morpheme string (e.g., "ing")
	Freq       int    // Frequency count
	WordsCount int    // Number of words containing this morpheme
	Words      []*Word // Words containing this morpheme
	Type       int    // Morpheme type (STEM, PREFIX, SUFFIX, INFIX)
}

// Legacy accessors for compatibility with file writing code
// TODO: Remove once file writing uses Pattern directly

// Regex returns the complete regex pattern as a string
func (m *Morpheme) Regex() string {
	if m.Pattern == nil {
		return ""
	}
	return m.Pattern.String()
}

// Public accessors for regex patterns (used by classification logic)

func (m *Morpheme) GetFrontRegex() string {
	if m.Pattern == nil || m.Pattern.Front == nil {
		return ""
	}
	return m.Pattern.Front.String()
}

func (m *Morpheme) GetBackRegex() string {
	if m.Pattern == nil || m.Pattern.Back == nil {
		return ""
	}
	return m.Pattern.Back.String()
}

func (m *Morpheme) SetFrontRegex(pattern string) {
	if m.Pattern != nil {
		m.Pattern.Front = NewFixedPattern(pattern)
	}
}

func (m *Morpheme) SetBackRegex(pattern string) {
	if m.Pattern != nil {
		m.Pattern.Back = NewFixedPattern(pattern)
	}
}

// NewMorpheme creates a new morpheme with the given morpheme string
func NewMorpheme(morpheme string) Morpheme {
	return Morpheme{
		Morpheme:   morpheme,
		Freq:       1,
		WordsCount: 0,
		Words:      make([]*Word, 0),
		Pattern: &Regex{
			morpheme: morpheme,
			Front:    &RegexPattern{lastIndex: -1},
			Back:     &RegexPattern{lastIndex: -1},
		},
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
	beforeMorpheme1, afterMorpheme1 := splitAroundMorpheme(one, placeOne, len(m.Morpheme))
	beforeMorpheme2, afterMorpheme2 := splitAroundMorpheme(two, placeTwo, len(m.Morpheme))

	// Add contexts to build character class patterns
	m.Pattern.Front.AddContext(beforeMorpheme1)
	m.Pattern.Front.AddContext(beforeMorpheme2)
	m.Pattern.Back.AddContext(afterMorpheme1)
	m.Pattern.Back.AddContext(afterMorpheme2)
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
		m.Pattern.SetStem(m.Morpheme)
	}
}

// MergeRules is a no-op with the new design.
// Patterns are built lazily via String(), so no rebuild needed.
// Kept for compatibility with existing code.
func (m *Morpheme) MergeRules() {
	// Pattern.String() builds regex on-demand from charArray
	// No explicit rebuild needed
}

// MergePatternsWith merges regex patterns from another morpheme into this one.
// Used by FuseRegex to combine patterns from multiple observations of the same morpheme.
//
// Example:
//   Morpheme "ing" seen in "running" → Front pattern for "runn"
//   Morpheme "ing" seen in "walking" → Front pattern for "walk"
//   After merge: Front pattern matches both "runn" and "walk"
func (m *Morpheme) MergePatternsWith(other *Morpheme) {
	if m == nil || other == nil || m.Pattern == nil || other.Pattern == nil {
		return
	}

	// Merge front patterns (context before morpheme)
	if m.Pattern.Front != nil && other.Pattern.Front != nil {
		m.Pattern.Front.MergeWith(other.Pattern.Front)
	}

	// Merge back patterns (context after morpheme)
	if m.Pattern.Back != nil && other.Pattern.Back != nil {
		m.Pattern.Back.MergeWith(other.Pattern.Back)
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

// MorphemeList represents struct morpheme_list_t (lines 60-63)
type MorphemeList struct {
	List  []*Morpheme // list
	Count int         // count
}

// MorphemeCandidate is a lightweight representation of a potential morpheme
// before full Morpheme object allocation. Saves ~450 bytes per candidate.
type MorphemeCandidate struct {
	Morpheme  string // The morpheme string (e.g., "walk")
	Word1     string // First word containing morpheme
	Word2     string // Second word containing morpheme
	Pos1      int    // Position of morpheme in word1
	Pos2      int    // Position of morpheme in word2
	IsStem    bool   // True if morpheme equals entire word
	IsForward bool   // True if forward morpheme (no regex check), false if internal (needs regex check)
}

// CandidateList holds lightweight morpheme candidates before validation
type CandidateList struct {
	Candidates []MorphemeCandidate
	Count      int
}

// BuildMorphemesFromCandidates creates full Morpheme objects from lightweight candidates.
// This deduplicates by morpheme string BEFORE allocating expensive Regex structures,
// saving ~720MB by only creating 1,024 morphemes instead of 1.6 million.
func (cl *CandidateList) BuildMorphemesFromCandidates() MorphemeList {
	// Group candidates by morpheme string
	unique := make(map[string][]MorphemeCandidate)
	for i := 0; i < cl.Count; i++ {
		cand := cl.Candidates[i]
		if cand.Morpheme == "" || cand.Morpheme == constants.EmptyMorpheme {
			continue // Skip empty morphemes
		}
		unique[cand.Morpheme] = append(unique[cand.Morpheme], cand)
	}

	// Build full Morpheme objects for unique morphemes only
	result := MorphemeList{
		List:  make([]*Morpheme, 0, len(unique)),
		Count: 0,
	}

	for morphemeStr, candidates := range unique {
		// Create ONE morpheme object for all candidates with this string
		m := NewMorpheme(morphemeStr)
		m.Freq = len(candidates) // Frequency = number of times seen

		// Build regex patterns by adding ALL word pair contexts
		for _, cand := range candidates {
			// Check if it's a STEM (morpheme equals entire word)
			if cand.IsStem {
				m.SetType(constants.STEM)
				break // STEM doesn't need pattern building
			}

			// Extract contexts around morpheme in both words
			before1, after1 := splitAroundMorpheme(cand.Word1, cand.Pos1, len(morphemeStr))
			before2, after2 := splitAroundMorpheme(cand.Word2, cand.Pos2, len(morphemeStr))

			// Add contexts to pattern (incremental building)
			if m.Pattern != nil {
				if m.Pattern.Front != nil {
					m.Pattern.Front.AddContext(before1)
					m.Pattern.Front.AddContext(before2)
				}
				if m.Pattern.Back != nil {
					m.Pattern.Back.AddContext(after1)
					m.Pattern.Back.AddContext(after2)
				}
			}
		}

		// Set type to UNDEF for non-stems (will be classified later)
		if m.Type != constants.STEM {
			m.Type = constants.UNDEF
		}

		result.List = append(result.List, &m)
		result.Count++
	}

	return result
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

