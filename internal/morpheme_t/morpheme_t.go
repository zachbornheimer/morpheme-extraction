// Package morpheme_t implements morpheme extraction from morpheme_t.c
// Line-for-line translation of morpheme_t.c lines 1-331
package morpheme_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// AddWord adds word to morpheme's word list (lines 13-18)
func AddWord(morpheme *structs.Morpheme, w structs.Word) {
	morpheme.WordsCount++
	newWords := make([]*structs.Word, morpheme.WordsCount)
	copy(newWords, morpheme.Words)
	newWords[morpheme.WordsCount-1] = &w
	morpheme.Words = newWords
}

// FreeMorpheme frees morpheme memory (lines 20-23)
// Kept for C compatibility, Go GC handles this
func FreeMorpheme(morpheme *structs.Morpheme) {
	// Go GC handles cleanup
}

// FindLongestMatch finds longest matching prefix using runes for Unicode support
func FindLongestMatch(one, two structs.Word) structs.Morpheme {
	// Get clean strings without marker bytes
	s1 := one.StringClean()
	s2 := two.StringClean()

	// Convert to runes for proper Unicode handling
	r1 := []rune(s1)
	r2 := []rune(s2)

	// Find the minimum length
	minLen := len(r1)
	if len(r2) < minLen {
		minLen = len(r2)
	}

	// Find longest common prefix
	matchLen := 0
	for i := 0; i < minLen; i++ {
		if r1[i] != r2[i] {
			break
		}
		matchLen++
	}

	// Return empty morpheme if no match
	if matchLen == 0 {
		return structs.NewMorpheme(constants.EmptyMorpheme)
	}

	// Return the common prefix as a morpheme
	return structs.NewMorpheme(string(r1[:matchLen]))
}

// GenRegex generates regex for morpheme (lines 62-234)
func GenRegex(one, two string, placeOne, placeTwo int, morpheme *structs.Morpheme) {
	// Validate morpheme placement
	if !isValidMorphemePlacement(one, two, placeOne, placeTwo, morpheme) {
		return
	}

	morpheme.Type = constants.UNDEF

	// Handle STEM case (morpheme equals entire word)
	if one == morpheme.Morpheme || two == morpheme.Morpheme {
		setStemRegex(morpheme)
		return
	}

	// Extract context (prefix/suffix) from both words
	prefix1, suffix1 := extractContext(one, placeOne, len(morpheme.Morpheme))
	prefix2, suffix2 := extractContext(two, placeTwo, len(morpheme.Morpheme))

	// Build character class regex patterns
	factory := newRegexFactory()
	frontRegex, frontArr, frontIdx := factory.buildCharacterClass(prefix1, prefix2, true)
	backRegex, backArr, backIdx := factory.buildCharacterClass(suffix1, suffix2, false)

	// Store results
	morpheme.FrontRegex = frontRegex
	morpheme.FrontRegexArr = frontArr
	morpheme.FrontRegexArrIndex = frontIdx

	morpheme.BackRegex = backRegex
	morpheme.BackRegexArr = backArr
	morpheme.BackRegexArrIndex = backIdx

	morpheme.Regex = frontRegex + morpheme.Morpheme + backRegex
}

// isValidMorphemePlacement validates morpheme can be placed at specified positions
func isValidMorphemePlacement(one, two string, placeOne, placeTwo int, morpheme *structs.Morpheme) bool {
	// Check morpheme is not empty
	if morpheme.Morpheme == constants.EmptyMorpheme || len(morpheme.Morpheme) == 0 {
		return false
	}

	// Check morpheme fits in both words
	if len(morpheme.Morpheme) > len(one) || len(morpheme.Morpheme) > len(two) {
		return false
	}

	// Verify morpheme matches at specified positions
	morphemeLen := len(morpheme.Morpheme)

	extractedOne := extractMorphemeAt(one, placeOne, morphemeLen)
	extractedTwo := extractMorphemeAt(two, placeTwo, morphemeLen)

	return extractedOne == morpheme.Morpheme && extractedTwo == morpheme.Morpheme
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

// setStemRegex configures morpheme as a STEM (whole word)
func setStemRegex(morpheme *structs.Morpheme) {
	morpheme.Regex = "^" + morpheme.Morpheme + "$"
	morpheme.FrontRegex = "^"
	morpheme.BackRegex = "$"
	morpheme.Type = constants.STEM
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
	// Reverse for prefix processing (builds right-to-left)
	if isPrefix {
		ctx1 = functions.Reverse(ctx1)
		ctx2 = functions.Reverse(ctx2)
	} else {
		ctx1 = functions.Reverse(ctx1)
		ctx2 = functions.Reverse(ctx2)
	}

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
	regex = functions.Reverse(pattern)

	return regex, charArray, lastIndex
}

// MergeRules merges regex rules (lines 236-312)
func MergeRules(morpheme *structs.Morpheme) {
	i := 0
	frontSize := 0
	backSize := 0

	for i = 0; i <= morpheme.FrontRegexArrIndex; i++ {
		if len(morpheme.FrontRegexArr[i]) != 0 {
			frontSize += len(morpheme.FrontRegexArr[i])
		}
	}
	frontSize++

	for i = 0; i <= morpheme.BackRegexArrIndex; i++ {
		if len(morpheme.BackRegexArr[i]) != 0 {
			backSize += len(morpheme.BackRegexArr[i])
		}
	}
	backSize++

	front := ")"
	back := ")"

	i = 0
	for {
		if i > morpheme.FrontRegexArrIndex {
			break
		}
		front += "]"
		front += morpheme.FrontRegexArr[i]
		front += "["
		i++
	}
	front += "("
	front = functions.Reverse(front)

	i = 0
	for {
		if i > morpheme.BackRegexArrIndex {
			break
		}
		back += "]"
		back += morpheme.BackRegexArr[i]
		back += "["
		i++
	}
	back += "("
	back = functions.Reverse(back)

	morpheme.Regex = ""
	if morpheme.FrontRegex != "^" {
		morpheme.FrontRegex = front
		morpheme.Regex += front
	} else {
		morpheme.Regex += morpheme.FrontRegex
	}

	morpheme.Regex += morpheme.Morpheme

	if morpheme.BackRegex != "$" {
		morpheme.BackRegex = back
		morpheme.Regex += back
	} else {
		morpheme.Regex += morpheme.BackRegex
	}
}

