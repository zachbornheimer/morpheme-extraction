// Package morpheme_list_t implements morpheme list operations from morpheme_list_t.c
// Line-for-line translation of morpheme_list_t.c lines 1-333
package morpheme_list_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
	"github.com/zachbornheimer/morpheme-extraction/internal/word_t"
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

	for len(one.Str) > 0 {
		for len(two.Str) > 0 {
			two.Str = word_t.RemoveFirstChar(two)
			iteration2++

			if len(two.Str) < 2 {
				break
			}

			lm := morpheme_t.FindLongestMatch(one, two)
			lm.WordsCount = 0

			if len(lm.Morpheme) >= 2 {
				if internal.Count > 0 {
					var current, previous structs.Word
					word_t.SetWord(&current, lm.Morpheme)
					word_t.SetWord(&previous, internal.List[internal.Count-1].Morpheme)

					if morpheme_t.FindLongestMatch(current, previous).Morpheme != lm.Morpheme {
						test := string(previous.Str)
						test = functions.Reverse(test)

						if (len(test) - lookback) >= 1 {
							if len(test) > lookback && test[len(test)-lookback] != 0 {
								test = test[:len(test)-lookback]
							}
						} else {
							continue
						}

						test = functions.Reverse(test)

						if test != string(current.Str) {
							lm.Words = make([]*structs.Word, 0)
							morpheme_t.AddWord(&lm, oneOrig)
							morpheme_t.AddWord(&lm, twoOrig)
							morpheme_t.GenRegex(word1, word2, iteration1, iteration2, &lm)

							if lm.Regex != "" {
								AddMorpheme(internal, lm)
							}

							lookback = len(lm.Morpheme)
						}
					}
				} else {
					lm.Words = make([]*structs.Word, 0)
					morpheme_t.AddWord(&lm, oneOrig)
					morpheme_t.AddWord(&lm, twoOrig)
					morpheme_t.GenRegex(word1, word2, iteration1, iteration2, &lm)

					if lm.Regex != "" {
						AddMorpheme(internal, lm)
					}

					lookback = len(lm.Morpheme)
				}
			} else {
				lookback = 1
			}
		}

		two.Str = word2
		one.Str = word_t.RemoveFirstChar(one)
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
		if original.List[i].Morpheme == "" {
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

			// Extract Parenthetical Sections
			working := *new.List[currentIndex]
			dupRegex := dup.Regex
			origRegex := orig.Regex
			origFrontIndex := orig.FrontRegexArrIndex
			dupFrontIndex := dup.FrontRegexArrIndex
			origBackIndex := orig.BackRegexArrIndex
			dupBackIndex := dup.BackRegexArrIndex

			if origRegex == dupRegex {
				new.List[currentIndex].Freq++
				continue
			}

			// Merge regex patterns
			regexFront := make([]string, (dupFrontIndex+1)+(origFrontIndex+1)+1)
			regexBack := make([]string, (dupBackIndex+1)+(origBackIndex+1)+1)

			// Merge front blocks
			j := 0
			for {
				size := 2
				longest := dupFrontIndex
				if origFrontIndex > longest {
					longest = origFrontIndex
				}

				if j-(longest-origFrontIndex) >= 0 && j-(longest-origFrontIndex) <= origFrontIndex {
					size += len(orig.FrontRegexArr[j-(longest-origFrontIndex)])
				}

				if j-(longest-dupFrontIndex) >= 0 && j-(longest-dupFrontIndex) <= dupFrontIndex {
					size += len(dup.FrontRegexArr[j-(longest-dupFrontIndex)])
				}

				regexFront[j] = ""

				if size == 2 {
					break
				}

				if j >= longest-origFrontIndex && origFrontIndex != -1 {
					regexFront[j] += orig.FrontRegexArr[j-(longest-origFrontIndex)]
				}

				if j >= longest-dupFrontIndex && dupFrontIndex != -1 {
					regexFront[j] += dup.FrontRegexArr[j-(longest-dupFrontIndex)]
				}

				arr, _ := functions.Uniq(regexFront[j])
				regexFront[j] = arr
				j++
			}

			working.FrontRegexArrIndex = j - 1
			working.FrontRegexArr = regexFront

			// Merge back blocks
			j = 0
			for {
				size := 2
				longest := dupBackIndex
				if origBackIndex > longest {
					longest = origBackIndex
				}

				if j-(longest-origBackIndex) >= 0 && j-(longest-origBackIndex) <= origBackIndex {
					size += len(orig.BackRegexArr[j-(longest-origBackIndex)])
				}

				if j-(longest-dupBackIndex) >= 0 && j-(longest-dupBackIndex) <= dupBackIndex {
					size += len(dup.BackRegexArr[j-(longest-dupBackIndex)])
				}

				regexBack[j] = ""

				if size == 2 {
					break
				}

				if j >= longest-origBackIndex && origBackIndex != -1 {
					regexBack[j] += orig.BackRegexArr[j-(longest-origBackIndex)]
				}

				if j >= longest-dupBackIndex && dupBackIndex != -1 {
					regexBack[j] += dup.BackRegexArr[j-(longest-dupBackIndex)]
				}

				arr, _ := functions.Uniq(regexBack[j])
				regexBack[j] = arr
				j++
			}

			working.BackRegexArrIndex = j - 1
			working.BackRegexArr = regexBack

			// Reconstruct and restore
			morpheme_t.MergeRules(&working)
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
			// Prefix Identification
			if morpheme.FrontRegex == "^" || morpheme.FrontRegex == "()" {
				if morpheme.BackRegex == "$" {
					list.List[i].Type = constants.STEM
				} else {
					list.List[i].Type = constants.PREFIX
				}
			}

			// Suffix Identification
			if morpheme.BackRegex == "$" || morpheme.BackRegex == "()" {
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
							list.List[i].FrontRegex = list.List[i].Regex
							list.List[i].BackRegex = list.List[j].Regex
							list.List[j].FrontRegex = list.List[i].Regex
							list.List[j].BackRegex = list.List[j].Regex

							grapheme := list.List[i].Morpheme + "(.*)" + list.List[j].Morpheme
							list.List[i].Morpheme = grapheme
							list.List[j].Morpheme = grapheme
						} else {
							list.List[i].FrontRegex = list.List[j].Regex
							list.List[i].BackRegex = list.List[i].Regex
							list.List[j].FrontRegex = list.List[j].Regex
							list.List[j].BackRegex = list.List[i].Regex

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

