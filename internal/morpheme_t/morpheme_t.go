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

// FindLongestMatch finds longest matching prefix (lines 25-60)
func FindLongestMatch(one, two structs.Word) structs.Morpheme {
	var longest, shortest structs.Word
	empty := structs.Morpheme{Morpheme: constants.EmptyMorpheme, Freq: 1}
	realLength := 0

	if one.Len >= two.Len {
		longest = one
		shortest = two
	} else {
		longest = two
		shortest = one
	}

	str := make([]byte, shortest.Len+2)
	shortestStr := string(shortest.Str)
	longestStr := string(longest.Str)

	// Check for empty morpheme marker during matching
	emptyByte := constants.EmptyMorphemeByte
	hasEmptyMarker := constants.HasEmptyMarker

	for realLength < len(shortestStr) && realLength < len(longestStr) {
		// Stop if we hit the empty marker (if it exists)
		if hasEmptyMarker && shortestStr[realLength] == emptyByte {
			break
		}
		// Stop if characters don't match
		if shortestStr[realLength] != longestStr[realLength] {
			break
		}
		str[realLength] = shortestStr[realLength]
		realLength++
	}
	// Add empty marker if it exists
	if hasEmptyMarker {
		str[realLength] = emptyByte
	}

	if realLength == 0 {
		return empty
	}

	longestMatch := structs.Morpheme{
		Freq:     1,
		Morpheme: string(str[:realLength]),
	}

	return longestMatch
}

// GenRegex generates regex for morpheme (lines 62-234)
func GenRegex(one, two string, placeOne, placeTwo int, morpheme *structs.Morpheme) {
	// Step 1: Confirm Morpheme Placement
	if morpheme.Morpheme == constants.EmptyMorpheme || len(morpheme.Morpheme) == 0 ||
		!(len(morpheme.Morpheme) <= len(one) && len(morpheme.Morpheme) <= len(two)) {
		return
	}

	i := 0
	word1 := make([]byte, len(morpheme.Morpheme)+1)
	word2 := make([]byte, len(morpheme.Morpheme)+1)

	for i = placeOne; i < (placeOne + len(morpheme.Morpheme)); i++ {
		word1[i-placeOne] = one[i]
	}
	// Add empty marker if it exists
	emptyByte := constants.EmptyMorphemeByte
	if emptyByte != 0 {
		word1[i-placeOne] = emptyByte
	}

	for i = placeTwo; i < (placeTwo + len(morpheme.Morpheme)); i++ {
		word2[i-placeTwo] = two[i]
	}
	// Add empty marker if it exists
	if emptyByte != 0 {
		word2[i-placeTwo] = emptyByte
	}

	if !(string(word1[:len(morpheme.Morpheme)]) == string(word2[:len(morpheme.Morpheme)]) &&
		string(word1[:len(morpheme.Morpheme)]) == morpheme.Morpheme) {
		return
	}

	// Step 1.5: If Either Word = Morpheme, it's a STEM
	morpheme.Type = constants.UNDEF

	if one == morpheme.Morpheme || two == morpheme.Morpheme {
		morpheme.Regex = "^" + morpheme.Morpheme + "$"
		morpheme.FrontRegex = "^"
		morpheme.BackRegex = "$"
		morpheme.Type = constants.STEM
		return
	}

	// Step 2: Extract Chars on Each Side
	firstHalf1 := one[:placeOne]
	secondHalf1 := ""
	if placeOne+len(morpheme.Morpheme) <= len(one) {
		secondHalf1 = one[placeOne+len(morpheme.Morpheme):]
	}

	firstHalf2 := two[:placeTwo]
	secondHalf2 := ""
	if placeTwo+len(morpheme.Morpheme) <= len(two) {
		secondHalf2 = two[placeTwo+len(morpheme.Morpheme):]
	}

	// Step 3: Generate Regex for Each Side
	front := ")"
	back := ")"

	firstHalf1 = functions.Reverse(firstHalf1)
	firstHalf2 = functions.Reverse(firstHalf2)

	maxFrontLen := len(firstHalf1)
	if len(firstHalf2) > maxFrontLen {
		maxFrontLen = len(firstHalf2)
	}

	morpheme.FrontRegexArr = make([]string, maxFrontLen*2)
	morpheme.FrontRegexArrIndex = -1

	i = 0
	for {
		if i >= len(firstHalf1) && i >= len(firstHalf2) {
			break
		}

		morpheme.FrontRegexArr[i] = ""
		front += "]"

		if i < len(firstHalf2) {
			work := string(firstHalf2[i])
			front += work
			morpheme.FrontRegexArr[i] += work
		}

		if i < len(firstHalf1) {
			work := string(firstHalf1[i])
			front += work
			morpheme.FrontRegexArr[i] += work
		}

		morpheme.FrontRegexArrIndex = i
		front += "["
		i++
	}

	firstHalf1 = functions.Reverse(firstHalf1)
	firstHalf2 = functions.Reverse(firstHalf2)
	front += "("
	front = functions.Reverse(front)

	// Process back half
	secondHalf1 = functions.Reverse(secondHalf1)
	secondHalf2 = functions.Reverse(secondHalf2)

	maxBackLen := len(secondHalf1)
	if len(secondHalf2) > maxBackLen {
		maxBackLen = len(secondHalf2)
	}

	morpheme.BackRegexArr = make([]string, maxBackLen*4)
	morpheme.BackRegexArrIndex = -1

	i = 0
	for {
		if i >= len(secondHalf1) && i >= len(secondHalf2) {
			break
		}

		back += "]"
		morpheme.BackRegexArr[i] = ""

		if i < len(secondHalf1) {
			work := string(secondHalf1[i])
			back += work
			morpheme.BackRegexArr[i] += work
		}

		if i < len(secondHalf2) {
			work := string(secondHalf2[i])
			back += work
			morpheme.BackRegexArr[i] += work
		}

		morpheme.BackRegexArrIndex = i
		back += "["
		i++
	}

	secondHalf1 = functions.Reverse(secondHalf1)
	secondHalf2 = functions.Reverse(secondHalf2)
	back += "("
	back = functions.Reverse(back)

	// Step 4: Store Regex
	morpheme.Regex = front + morpheme.Morpheme + back
	morpheme.FrontRegex = front
	morpheme.BackRegex = back
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

