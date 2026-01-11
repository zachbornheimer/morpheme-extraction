// Package morpheme_list_t implements morpheme list operations from morpheme_list_t.c
// Line-for-line translation of morpheme_list_t.c lines 1-333
package morpheme_list_t

import (
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

// FindInternalMorphemes finds morphemes within words (lines 24-93)
func FindInternalMorphemes(oneOrig, twoOrig structs.Word, internal *structs.MorphemeList) {
	// Use Word.StringClean() to get clean string without null padding
	word1 := oneOrig.StringClean()
	word2 := twoOrig.StringClean()
	one := oneOrig
	two := twoOrig

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

			lm := morpheme_t.FindLongestMatch(one, two)
			lm.WordsCount = 0

			if len(lm.Morpheme) >= 2 {
				if internal.Count > 0 {
					var current, previous structs.Word
					current.SetWord(lm.Morpheme)
					previous.SetWord(internal.List[internal.Count-1].Morpheme)

					if morpheme_t.FindLongestMatch(current, previous).Morpheme != lm.Morpheme {
						test := string(previous.Str)
						test = functions.Reverse(test)

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
							continue
						}

						test = functions.Reverse(test)

						if test != string(current.Str) {
							lm.Words = make([]*structs.Word, 0)
							lm.AddWord(oneOrig)
							lm.AddWord(twoOrig)
							lm.GenRegex(word1, word2, iteration1, iteration2)

							if lm.Regex() != "" {
								AddMorpheme(internal, lm)
							}

							lookback = len(lm.Morpheme)
						}
					}
				} else {
					lm.Words = make([]*structs.Word, 0)
					lm.AddWord(oneOrig)
					lm.AddWord(twoOrig)
					lm.GenRegex(word1, word2, iteration1, iteration2)

					if lm.Regex() != "" {
						AddMorpheme(internal, lm)
					}

					lookback = len(lm.Morpheme)
				}
			} else {
				lookback = 1
			}
		}

		two.SetWord(word2)
		one.SetWord(one.RemoveFirstChar())
		iteration1++
		iteration2 = 0
	}
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
				continue
			}

			if orig.Type == constants.STEM {
				orig.Freq++
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

		(*lex)[i].Morpheme = list.List[i]
	}
}

