// Package morpheme_t implements morpheme extraction from morpheme_t.c
// Morpheme operations are now methods on structs.Morpheme
// This package contains utilities that operate on multiple morphemes/words
package morpheme_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// FindLongestMatchString finds longest matching prefix and returns just the string.
// This avoids allocating a full Morpheme object (saves ~500 bytes per call).
func FindLongestMatchString(one, two structs.Word) string {
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

	// Return empty string if no match
	if matchLen == 0 {
		return constants.EmptyMorpheme
	}

	// Return the common prefix
	return string(r1[:matchLen])
}

// FindLongestMatch finds longest matching prefix using runes for Unicode support
// DEPRECATED: Use FindLongestMatchString + MorphemeCandidate for better performance.
// This allocates a full Morpheme object which is expensive when called millions of times.
func FindLongestMatch(one, two structs.Word) structs.Morpheme {
	morphemeStr := FindLongestMatchString(one, two)
	return structs.NewMorpheme(morphemeStr)
}

