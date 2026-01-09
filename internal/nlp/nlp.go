// Package nlp implements main NLP processing from nlp.c
// Line-for-line translation of nlp.c lines 64-259
// With Go concurrency optimizations for performance
package nlp

import (
	"fmt"

	"github.com/zachbornheimer/morpheme-extraction/internal/alphabet"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/directory"
	"github.com/zachbornheimer/morpheme-extraction/internal/file"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_list_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/ngram_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
	"github.com/zachbornheimer/morpheme-extraction/internal/word_t"
)

// Nlp runs the main NLP process (lines 64-111)
func Nlp(verboseMode, processSequentially, processFull int, outputFilename, corpusDir string) int {
	// Set globals
	functions.VerboseMode = verboseMode
	directory.CorpusDir = corpusDir
	alphabet.VerboseMode = &verboseMode

	var f string
	var header string
	index := 0
	ngramLength := 0
	lexCount := 0
	var ng []*structs.Ngram
	var morphemes []structs.LexicalCategories

	for {
		f = directory.GetFiles(&index, &header)
		if f == "" {
			break
		}

		if len(f) > 2 {
			functions.VPrint("Looking for Word Delimiter...")
			wd, err := alphabet.FindWordDelimiter(&f)
			if functions.VerboseMode == constants.ON {
				fmt.Printf("Found Word Delimiter: %q\n", wd)
			}

			if err == constants.ErrOverruled {
				continue
			}

			if err == nil && wd != "" {
				functions.VPrint("Building NGRAMS")
				BuildNgramRelationships(wd, f, &ngramLength, &ng)

				if processFull == constants.ON || processSequentially == constants.ON {
					functions.VPrint("\nPROCESSING SEQUENTIALLY!")
					functions.VPrint("Extracting Morphemes and Building Lexical Categories.")
					lexCount = FindMorphemes(ng, ngramLength, header, &morphemes, outputFilename)
				}
			}
		}
	}

	if processFull == constants.ON || processSequentially == constants.OFF {
		header = "All Processed Files"
		functions.VPrint("PROCESSING SIMULTANEOUSLY!")
		functions.VPrint("Extracting Morphemes and Building Lexical Categories.")
		lexCount = FindMorphemes(ng, ngramLength, header, &morphemes, outputFilename)
	}

	// Cleanup
	for i := lexCount - 1; i >= 0; i-- {
		// Go GC handles cleanup
	}

	for i := ngramLength - 1; i >= 0; i-- {
		ngram_t.FreeNgram(ng[i])
	}

	return 0
}

// BuildNgramRelationships builds n-gram structures (lines 113-211)
func BuildNgramRelationships(wd, f string, ngramLength *int, ngramFull *[]*structs.Ngram) {
	arr := functions.ExplodeSansnullStr(f, wd)
	wordCount := len(arr)

	i := 0
	ngramCount := wordCount
	makeNew := 1
	addLoc := *ngramLength

	ng := *ngramFull

	if addLoc == 0 {
		ng = make([]*structs.Ngram, wordCount+1)
	}

	functions.VPrint("Setting NGRAM elements")

	// Declare q outside loop so it persists (line 125 in C: int i = 0, q = 0, j = 0;)
	var q int
	
	for i = 0; i <= wordCount; i++ {
		if i >= len(arr) || arr[i] == "" || len(arr[i]) == 0 {
			continue
		}

		ngram := ngram_t.NewNgram()

		if i != 0 {
			for q = 0; q < addLoc; q++ {
				if string(ng[q].Word.Str) == arr[i] {
					ngram = *ng[q]
					ngramCount--
					makeNew = 0
					break
				} else {
					makeNew = 1
				}
			}
		}

		if makeNew == 1 {
			ngram = ngram_t.NewNgram()
			if len(arr[i]) > 0 {
				word_t.SetWord(&ngram.Word, arr[i])
			}
		}

		if functions.VerboseMode == constants.ON {
			fmt.Printf("  Processing Word: %d:%d     \r", i, wordCount)
		}

		if len(arr[i]) > 0 {
			ngram.Word.Freq++

			for j := 0; j < (constants.NgramSize / 2); j++ {
				index := i - (constants.NgramSize / 2) + j
				var elemID int

				if index > 0 && index < len(arr) && len(arr[index]) > 0 {
					elemID = ngram_t.AddNgramElement(&ngram.Before.At[j], j)
					if ngram.Before.At[j].Elems == nil {
						ngram.Before.At[j].Elems = make([]*structs.Word, 1)
					}
					if elemID >= len(ngram.Before.At[j].Elems) {
						newElems := make([]*structs.Word, elemID+1)
						copy(newElems, ngram.Before.At[j].Elems)
						ngram.Before.At[j].Elems = newElems
					}
					ngram.Before.At[j].Elems[elemID] = &structs.Word{}
					word_t.SetWord(ngram.Before.At[j].Elems[elemID], arr[index])
				}

				afterIndex := index + (constants.NgramSize/2) + 1
				if afterIndex <= wordCount && afterIndex < len(arr) && len(arr[afterIndex]) > 0 {
					elemID = ngram_t.AddNgramElement(&ngram.After.At[j], j)
					if ngram.After.At[j].Elems == nil {
						ngram.After.At[j].Elems = make([]*structs.Word, 1)
					}
					if elemID >= len(ngram.After.At[j].Elems) {
						newElems := make([]*structs.Word, elemID+1)
						copy(newElems, ngram.After.At[j].Elems)
						ngram.After.At[j].Elems = newElems
					}
					ngram.After.At[j].Elems[elemID] = &structs.Word{}
					word_t.SetWord(ngram.After.At[j].Elems[elemID], arr[afterIndex])
				}
			}
		}

		if makeNew == 1 {
			ng[addLoc] = &ngram
			addLoc++
		} else {
			// Update existing ngram (line 185-186 in C)
			// In C: *ng[q] = ngram;
			// We need to update the existing ngram with our modifications
			*ng[q] = ngram
		}
	}

	*ngramLength = addLoc

	// Add linked list elements (similarity detection) - Sequential version
	for k := 0; k < *ngramLength; k++ {
		for j := k; j < *ngramLength; j++ {
			if j-k < constants.NgramSize {
				continue
			}

			if ng[j] != nil && ng[k] != nil &&
				string(ng[j].Word.Str) != "" && string(ng[k].Word.Str) != "" &&
				string(ng[j].Word.Str) != string(ng[k].Word.Str) {
				if ngram_t.NgramsSimilar(*ng[k], *ng[j]) {
					ngram_t.AddSimilarNgramRef(&ng[k], &ng[j])
				}
			}
		}
	}

	functions.VPrint("")
	*ngramFull = ng
}

// FindMorphemes extracts morphemes from n-grams (lines 213-259)
// Sequential to preserve morpheme order (affects FuseRegex results)
func FindMorphemes(ng []*structs.Ngram, ngramLength int, header string, lexMap *[]structs.LexicalCategories, outputFilename string) int {
	internal := structs.MorphemeList{Count: 0}
	internal.List = make([]*structs.Morpheme, 0, ngramLength*100) // Pre-allocate

	for i := 0; i < ngramLength; i++ {
		if ng[i] == nil {
			continue
		}

		ngram := *ng[i]

		for j := 0; j < ngram.RefsCount; j++ {
			target := *ngram.Refs[j]
			// Direct string access (no conversion needed with native strings)
			w1 := ngram.Word.Str
			w2 := target.Word.Str

			forward := morpheme_t.FindLongestMatch(ngram.Word, target.Word)
			forward.WordsCount = 0
			forward.Words = make([]*structs.Word, 0)
			morpheme_t.GenRegex(w1, w2, 0, 0, &forward)
			morpheme_t.AddWord(&forward, ngram.Word)
			morpheme_t.AddWord(&forward, target.Word)

			morpheme_list_t.AddMorpheme(&internal, forward)
			morpheme_list_t.FindInternalMorphemes(ngram.Word, target.Word, &internal)
		}
	}

	functions.VPrint("")
	morphemeList := morpheme_list_t.FuseRegex(internal)

	if functions.VerboseMode == constants.ON {
		fmt.Printf("\nFound %d internal morphemes before fusion\n", internal.Count)
		fmt.Printf("Found %d morphemes after fusion\n", morphemeList.Count)
	}

	lex := make([]structs.LexicalCategories, morphemeList.Count)
	morpheme_list_t.IdentifyTrueMorphemes(&morphemeList, &lex)
	
	if functions.VerboseMode == constants.ON {
		fmt.Printf("Writing %d morphemes to file\n", morphemeList.Count)
	}
	
	file.WriteToFile(outputFilename, lex, morphemeList.Count, header)

	*lexMap = lex
	return morphemeList.Count
}

