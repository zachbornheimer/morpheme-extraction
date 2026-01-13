// Package morpheme_list_t implements morpheme list operations from morpheme_list_t.c
// Line-for-line translation of morpheme_list_t.c lines 1-333
package morpheme_list_t

import (
	"fmt"
	"strings"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// AddMorpheme adds morpheme to list (lines 15-22)
func AddMorpheme(morphemes *structs.MorphemeList, morpheme structs.Morpheme) {
	// Use append instead of manual copy to avoid O(n²) performance
	morphemes.List = append(morphemes.List, &morpheme)
	morphemes.Count++
}

// Global debug counters (temporary for debugging)
var (
	debugTotalChecks      int
	debugTooShort         int
	debugPrefixMatch      int
	debugLookbackFail     int
	debugTestMismatch     int
	debugAdded            int
	debugFirstAdded       int
	debugFunctionCalls    int
)

// FindInternalMorphemes finds morphemes within words (lines 24-93)
// Optimized to use MorphemeCandidate for lazy allocation
func FindInternalMorphemes(oneOrig, twoOrig structs.Word, candidates *[]structs.MorphemeCandidate) {
	debugFunctionCalls++

	// Use Word.StringClean() to get clean string without null padding
	word1 := oneOrig.StringClean()
	word2 := twoOrig.StringClean()
	one := oneOrig
	two := twoOrig

	// Track the index of the last SUCCESSFULLY ADDED candidate from this word pair
	// This matches the original's behavior of comparing against internal.List[internal.Count-1]
	// IMPORTANT: Only candidates that pass regex check should update this!
	// Start at -1 to indicate no candidates added yet (matches internal.Count == 0)
	lastAddedIdx := -1

	iteration1 := 0
	iteration2 := 0
	lookback := 1

	for one.Len > 0 {
		for two.Len > 0 {
			two.SetWord(two.RemoveFirstChar())
			iteration2++

			if two.Len < (1 + len(constants.EmptyMorpheme)) {
				break
			}

			// Use FindLongestMatchString instead of FindLongestMatch (no allocation)
			morphemeStr := morpheme_t.FindLongestMatchString(one, two)
			debugTotalChecks++

			if len(morphemeStr) >= 2 {
				// Check if we have successfully added a previous candidate from THIS word pair
				// This matches the original's `if internal.Count > 0` check
				if lastAddedIdx >= 0 {
					// Get previous morpheme string from last SUCCESSFULLY ADDED candidate
					// This matches internal.List[internal.Count-1] in the original
					prevMorpheme := (*candidates)[lastAddedIdx].Morpheme

					// Check if current morpheme is different from previous
					// Use string comparison instead of creating Word objects
					currentPrefixOfPrev := findCommonPrefixStr(morphemeStr, prevMorpheme)

					if currentPrefixOfPrev != morphemeStr {
						// Apply lookback logic (string operations only)
						test := functions.Reverse(prevMorpheme)

						if (len(test) - lookback) >= 1 {
							// Check if we'd trim into the empty morpheme marker
							checkPos := len(test) - lookback
							shouldTrim := len(test) > lookback

							// If EmptyMorphemeByte is not zero, check if we'd hit it
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
							// Original updates lookback even on fail, but outside the outer if
							lookback = len(morphemeStr)
							continue
						}

						test = functions.Reverse(test)

						if test != morphemeStr {
							// CRITICAL: Match original behavior - check regex BEFORE adding to list
							// Create temporary morpheme to test regex
							m := structs.NewMorpheme(morphemeStr)
							m.WordsCount = 0
							m.Words = make([]*structs.Word, 0)

							// Create Word objects for GenRegex
							var w1, w2 structs.Word
							w1.SetWord(word1)
							w2.SetWord(word2)

							// Generate regex pattern
							m.GenRegex(word1, word2, iteration1, iteration2)

							// Only add if regex is valid (matches original line 57-58)
							if m.Regex() != "" {
								*candidates = append(*candidates, structs.MorphemeCandidate{
									Morpheme:  morphemeStr,
									Word1:     word1,
									Word2:     word2,
									Pos1:      iteration1,
									Pos2:      iteration2,
									IsStem:    false,
									IsForward: false, // Internal morphemes need regex check
								})
								// Update last added index to point to this new candidate
								lastAddedIdx = len(*candidates) - 1
								debugAdded++
							}

							// Original updates lookback OUTSIDE the regex check (line 61)
							lookback = len(morphemeStr)
						} else {
							debugTestMismatch++
							// Original still updates lookback here too
							lookback = len(morphemeStr)
						}
					} else {
						debugPrefixMatch++
					}
				} else {
					// First candidate - check regex before adding (matches original line 72-75)
					m := structs.NewMorpheme(morphemeStr)
					m.WordsCount = 0
					m.Words = make([]*structs.Word, 0)

					// Create Word objects for GenRegex
					var w1, w2 structs.Word
					w1.SetWord(word1)
					w2.SetWord(word2)

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
							IsForward: false, // Internal morphemes need regex check
						})
						// Update last added index to point to this first candidate
						lastAddedIdx = len(*candidates) - 1
						debugFirstAdded++
					}

					// Original updates lookback OUTSIDE the regex check (line 75)
					lookback = len(morphemeStr)
				}
			} else {
				debugTooShort++
				lookback = 1
			}
		}

		two.SetWord(word2)
		one.SetWord(one.RemoveFirstChar())
		iteration1++
		iteration2 = 0
	}
}

// PrintDebugStats prints debug statistics for FindInternalMorphemes
func PrintDebugStats() {
	fmt.Printf("FindInternalMorphemes Debug Stats:\n")
	fmt.Printf("  Function calls:   %d\n", debugFunctionCalls)
	fmt.Printf("  Total checks:     %d (avg %.1f per call)\n", debugTotalChecks, float64(debugTotalChecks)/float64(debugFunctionCalls))
	fmt.Printf("  Too short (<2):   %d (%.1f%%)\n", debugTooShort, 100.0*float64(debugTooShort)/float64(debugTotalChecks))
	fmt.Printf("  Prefix match:     %d (currentPrefixOfPrev == morphemeStr)\n", debugPrefixMatch)
	fmt.Printf("  Lookback fail:    %d\n", debugLookbackFail)
	fmt.Printf("  Test mismatch:    %d (test == morphemeStr)\n", debugTestMismatch)
	fmt.Printf("  Added (with prev):%d\n", debugAdded)
	fmt.Printf("  Added (first):    %d\n", debugFirstAdded)
	fmt.Printf("  Total added:      %d (avg %.1f per call)\n", debugAdded+debugFirstAdded, float64(debugAdded+debugFirstAdded)/float64(debugFunctionCalls))
	fmt.Printf("  Filtered:         %d (%.1f%% of >= 2 chars)\n",
		debugPrefixMatch+debugLookbackFail+debugTestMismatch,
		100.0*float64(debugPrefixMatch+debugLookbackFail+debugTestMismatch)/float64(debugTotalChecks-debugTooShort))
}

// findCommonPrefixStr returns longest common prefix between two strings
// Used for lookback filtering in FindInternalMorphemes
func findCommonPrefixStr(s1, s2 string) string {
	r1 := []rune(s1)
	r2 := []rune(s2)

	minLen := len(r1)
	if len(r2) < minLen {
		minLen = len(r2)
	}

	matchLen := 0
	for i := 0; i < minLen; i++ {
		if r1[i] != r2[i] {
			break
		}
		matchLen++
	}

	if matchLen == 0 {
		return ""
	}
	return string(r1[:matchLen])
}

// FuseRegex fuses regex patterns for duplicate morphemes (lines 95-244)
func FuseRegex(original structs.MorphemeList) structs.MorphemeList {
	morphemeList := make([]string, original.Count)
	new := structs.MorphemeList{Count: 0}
	// Use capacity-only allocation so append works correctly
	new.List = make([]*structs.Morpheme, 0, original.Count)
	morphemeListCount := 0
	currentIndex := -1

	for i := 0; i < original.Count; i++ {
		// Skip empty morphemes (no common substring found)
		if original.List[i].Morpheme == "" || original.List[i].Morpheme == constants.EmptyMorpheme {
			continue
		}

		if morphemeListCount == 0 {
			morphemeList[morphemeListCount] = original.List[i].Morpheme
			AddMorpheme(&new, *original.List[i])
			morphemeListCount++
			continue
		}

		morpheme := original.List[i].Morpheme
		currentIndex = functions.InCharArray(morpheme, morphemeList, morphemeListCount)

		if currentIndex != -1 {
			dup := *original.List[i]
			orig := *new.List[currentIndex]

			if dup.Type == constants.STEM {
				orig = dup
				orig.Freq++
				// BUGFIX: Must update new.List to persist changes
				new.List[currentIndex] = &orig
				continue
			}

			if orig.Type == constants.STEM {
				orig.Freq++
				// BUGFIX: Must update new.List to persist changes
				new.List[currentIndex] = &orig
				continue
			}

			// Merge regex patterns from duplicate morpheme into existing one
			working := *new.List[currentIndex]

			// Check if patterns are already identical (no merge needed)
			if orig.Regex() == dup.Regex() {
				new.List[currentIndex].Freq++
				continue
			}

			// Merge patterns using encapsulated method (replaces 90+ lines of manual merging)
			working.MergePatternsWith(&dup)
			working.Freq++
			new.List[currentIndex] = &working

		} else {
			morphemeList[morphemeListCount] = original.List[i].Morpheme
			AddMorpheme(&new, *original.List[i])
			morphemeListCount++
		}

		currentIndex = -1
	}

	return new
}

// IdentifyTrueMorphemes classifies morpheme types (lines 246-332)
func IdentifyTrueMorphemes(list *structs.MorphemeList, lex *[]structs.LexicalCategories) {
	freqThresh := int(float64(list.Count) * (constants.ThresholdConfirmation / 100))

	kept := 0
	filtered := 0

	for i := 0; i < list.Count; i++ {
		morpheme := list.List[i]

		if morpheme.Freq <= freqThresh && morpheme.Type != constants.STEM {
			// Mark as UNDEF so WriteToFile skips it (matches C behavior)
			list.List[i].Type = constants.UNDEF
			filtered++
			continue
		}
		kept++

		if morpheme.Type != constants.STEM {
			front := morpheme.GetFrontRegex()
			back := morpheme.GetBackRegex()

			// Prefix Identification
			if front == "^" || front == "()" {
				if back == "$" {
					list.List[i].Type = constants.STEM
				} else {
					list.List[i].Type = constants.PREFIX
				}
			}

			// Suffix Identification
			if back == "$" || back == "()" {
				if morpheme.Type == constants.PREFIX {
					list.List[i].Type = constants.STEM
				} else {
					list.List[i].Type = constants.SUFFIX
				}
			}

			// Circumfix Identification
			for j := 0; j < i; j++ {
				if list.List[i].Type == constants.SUFFIX || list.List[i].Type == constants.PREFIX {
					break
				}

				if !((list.List[i].Type == constants.SUFFIX && list.List[j].Type == constants.PREFIX) ||
					(list.List[i].Type == constants.PREFIX && list.List[j].Type == constants.SUFFIX)) {
					continue
				}

				if list.List[j].Freq == list.List[i].Freq {
					combined := -1
					listCombined := make([]string, list.List[i].WordsCount+list.List[j].WordsCount)

					for k := 0; k < list.List[j].WordsCount; k++ {
						combined++
						listCombined[combined] = string(list.List[j].Words[k].Str)
					}

					for k := 0; k < list.List[i].WordsCount; k++ {
						combined++
						listCombined[combined] = string(list.List[i].Words[k].Str)
					}

					count := functions.UniqWords(listCombined, combined)
					percentSimilar := int(100.0 * (float64(combined-count) / float64(combined)))

					if float64(percentSimilar) >= constants.ThresholdCircumfix {
						list.List[i].Type = constants.CIRCUMFIX
						list.List[j].Type = constants.CIRCUMFIX

						if list.List[i].Type == constants.PREFIX {
							list.List[i].SetFrontRegex(list.List[i].Regex())
							list.List[i].SetBackRegex(list.List[j].Regex())
							list.List[j].SetFrontRegex(list.List[i].Regex())
							list.List[j].SetBackRegex(list.List[j].Regex())

							grapheme := list.List[i].Morpheme + "(.*)" + list.List[j].Morpheme
							list.List[i].Morpheme = grapheme
							list.List[j].Morpheme = grapheme
						} else {
							list.List[i].SetFrontRegex(list.List[j].Regex())
							list.List[i].SetBackRegex(list.List[i].Regex())
							list.List[j].SetFrontRegex(list.List[j].Regex())
							list.List[j].SetBackRegex(list.List[i].Regex())

							grapheme := list.List[j].Morpheme + "(.*)" + list.List[i].Morpheme
							list.List[i].Morpheme = grapheme
							list.List[j].Morpheme = grapheme
						}
					}
				}
			}
		}

		// Infix Identification
		if list.List[i].Type == constants.UNDEF {
			list.List[i].Type = constants.INFIX
		}

		// Reclassify based on morpheme position in actual words
		// Check if this morpheme appears predominantly at word boundaries
		if list.List[i].Type == constants.INFIX && list.List[i].WordsCount > 0 {
			morphemeStr := list.List[i].Morpheme
			atStart := 0
			atEnd := 0

			// Check all words containing this morpheme
			for j := 0; j < list.List[i].WordsCount; j++ {
				// Use StringClean() to get proper word without null padding
				wordStr := list.List[i].Words[j].StringClean()
				// Trim whitespace and punctuation for boundary detection
				wordStr = strings.TrimSpace(wordStr)
				wordStr = strings.Trim(wordStr, ".,!?;:'\"")

				// Check if morpheme is at the start of the word
				if strings.HasPrefix(wordStr, morphemeStr) {
					atStart++
				}

				// Check if morpheme is at the end of the word
				if strings.HasSuffix(wordStr, morphemeStr) {
					atEnd++
				}
			}

			total := float64(list.List[i].WordsCount)
			startPct := float64(atStart) / total
			endPct := float64(atEnd) / total

			// If >50% of occurrences are at a boundary, reclassify
			// Lower threshold aligns with paper's focus on parallel environments
			const threshold = 0.5

			if startPct > threshold && endPct <= 0.2 {
				// Predominantly at word start → PREFIX
				list.List[i].Type = constants.PREFIX
			} else if endPct > threshold && startPct <= 0.2 {
				// Predominantly at word end → SUFFIX
				list.List[i].Type = constants.SUFFIX
			} else if startPct > threshold && endPct > threshold {
				// Appears at both start and end → likely a STEM
				list.List[i].Type = constants.STEM
			}
		}

		(*lex)[i].Morpheme = list.List[i]
	}
}

