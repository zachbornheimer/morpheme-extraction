package morpheme_list_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
	"github.com/zachbornheimer/morpheme-extraction/internal/suffixtree"
)

// FindInternalMorphemesSuffixTree uses a suffix tree to find internal morphemes
// This is O(m+n) instead of O(m²×n²) for the nested loop approach
func FindInternalMorphemesSuffixTree(oneOrig, twoOrig structs.Word, candidates *[]structs.MorphemeCandidate) {
	debugFunctionCalls++

	// Use Word.StringClean() to get clean string without null padding
	word1 := oneOrig.StringClean()
	word2 := twoOrig.StringClean()

	// Build suffix tree and get all common substrings at once
	// This replaces the nested loops that strip chars one by one
	commonSubstrings := suffixtree.FindCommonSubstringsForPair(word1, word2, 2)

	// Track the index of the last SUCCESSFULLY ADDED candidate
	lastAddedIdx := -1
	lookback := 1
	reverseBuf := make([]byte, 0, 256)

	// Process candidates in order (sorted by pos1, pos2 to match original behavior)
	for _, substr := range commonSubstrings {
		morphemeStr := substr.Substring
		iteration1 := substr.Pos1
		iteration2 := substr.Pos2

		// Apply the same lookback deduplication logic as original
		if lastAddedIdx >= 0 {
			prevMorpheme := (*candidates)[lastAddedIdx].Morpheme

			// Check if current morpheme is different from previous
			currentPrefixOfPrev := findCommonPrefixStr(morphemeStr, prevMorpheme)

			if currentPrefixOfPrev != morphemeStr {
				// Apply lookback logic
				test := functions.ReverseInto(prevMorpheme, reverseBuf)

				if (len(test) - lookback) >= 1 {
					checkPos := len(test) - lookback
					shouldTrim := len(test) > lookback

					emptyByte := constants.EmptyMorphemeByte
					if shouldTrim && constants.HasEmptyMarker && checkPos >= 0 && checkPos < len(test) {
						if test[checkPos] == emptyByte {
							shouldTrim = false
						}
					}

					if shouldTrim {
						test = test[:len(test)-lookback]
					}
				} else {
					debugLookbackFail++
					lookback = len(morphemeStr)
					continue
				}

				test = functions.ReverseInto(test, reverseBuf)

				if test != morphemeStr {
					// Create temporary morpheme to test regex
					m := structs.NewMorpheme(morphemeStr)
					m.WordsCount = 0
					m.Words = make([]*structs.Word, 0)

					// Generate regex pattern
					m.GenRegex(word1, word2, iteration1, iteration2)

					// Only add if regex is valid
					if m.Regex() != "" {
						*candidates = append(*candidates, structs.MorphemeCandidate{
							Morpheme:  morphemeStr,
							Word1:     word1,
							Word2:     word2,
							Pos1:      iteration1,
							Pos2:      iteration2,
							IsStem:    false,
							IsForward: false,
						})
						lastAddedIdx = len(*candidates) - 1
						debugAdded++
					}

					lookback = len(morphemeStr)
				} else {
					debugTestMismatch++
					lookback = len(morphemeStr)
				}
			} else {
				debugPrefixMatch++
			}
		} else {
			// First candidate - check regex before adding
			m := structs.NewMorpheme(morphemeStr)
			m.WordsCount = 0
			m.Words = make([]*structs.Word, 0)

			m.GenRegex(word1, word2, iteration1, iteration2)

			if m.Regex() != "" {
				*candidates = append(*candidates, structs.MorphemeCandidate{
					Morpheme:  morphemeStr,
					Word1:     word1,
					Word2:     word2,
					Pos1:      iteration1,
					Pos2:      iteration2,
					IsStem:    false,
					IsForward: false,
				})
				lastAddedIdx = len(*candidates) - 1
				debugFirstAdded++
			}

			lookback = len(morphemeStr)
		}

		debugTotalChecks++
	}
}
