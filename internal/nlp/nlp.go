// Package nlp implements main NLP processing from nlp.c
// Line-for-line translation of nlp.c lines 64-259
// With Go concurrency optimizations for performance
package nlp

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/alphabet"
	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
	"github.com/zachbornheimer/morpheme-extraction/internal/directory"
	"github.com/zachbornheimer/morpheme-extraction/internal/file"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_list_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/ngram_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// Nlp runs the main NLP process (lines 64-111)
func Nlp(cfg *config.Config) int {
	// Set globals (TODO: remove these once all packages use Config)
	functions.VerboseMode = cfg.VerboseMode
	directory.CorpusDir = cfg.CorpusDir
	alphabet.VerboseMode = &cfg.VerboseMode

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
			cfg.VPrint("Looking for Word Delimiter...")
			wd, err := alphabet.FindWordDelimiter(&f)
			if cfg.IsVerbose() {
				fmt.Printf("Found Word Delimiter: %q\n", wd)
			}

			if err == constants.ErrOverruled {
				continue
			}

			if err == nil && wd != "" {
				cfg.VPrint("Building NGRAMS")
				BuildNgramRelationships(wd, f, &ngramLength, &ng)

				if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.ON {
					cfg.VPrint("\nPROCESSING SEQUENTIALLY!")
					cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
					lexCount = FindMorphemes(ng, ngramLength, header, &morphemes, cfg.OutputFilename)
				}
			}
		}
	}

	if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.OFF {
		header = "All Processed Files"
		cfg.VPrint("PROCESSING SIMULTANEOUSLY!")
		cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
		lexCount = FindMorphemes(ng, ngramLength, header, &morphemes, cfg.OutputFilename)
	}

	// Cleanup
	for i := lexCount - 1; i >= 0; i-- {
		// Go GC handles cleanup
	}

	for i := ngramLength - 1; i >= 0; i-- {
		ngram_t.FreeNgram(ng[i])
	}

	// Print dictionary stats if verbose
	cfg.VPrint(fmt.Sprintf("Dictionary: %d unique words interned", dictionary.Dictionary.Size()))

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

	// Hash map for O(1) ngram lookup (replaces O(N²) linear search)
	ngramMap := make(map[string]int, wordCount)

	// Build initial map from existing ngrams
	for idx := 0; idx < addLoc; idx++ {
		if ng[idx] != nil {
			// Word.String() now returns full string with null (matches original behavior)
			ngramMap[ng[idx].Word.String()] = idx
		}
	}

	// Declare q outside loop so it persists (line 125 in C: int i = 0, q = 0, j = 0;)
	var q int

	for i = 0; i <= wordCount; i++ {
		if i >= len(arr) || arr[i] == "" || len(arr[i]) == 0 {
			continue
		}

		ngram := ngram_t.NewNgram()

		if i != 0 {
			// O(1) hash map lookup with clean strings (no null terminators)
			if existingIdx, found := ngramMap[arr[i]]; found {
				q = existingIdx
				ngram = *ng[q]
				ngramCount--
				makeNew = 0
			} else {
				makeNew = 1
			}
		}

		if makeNew == 1 {
			ngram = ngram_t.NewNgram()
			if len(arr[i]) > 0 {
				ngram.Word.SetWord(arr[i])
			}
		}

		if functions.VerboseMode == constants.ON {
			fmt.Printf("  Processing Word: %d:%d     \r", i, wordCount)
		}

		if len(arr[i]) > 0 {
			// Track center word frequency in dictionary
			if ngram.Word.Identity != nil {
				ngram.Word.Identity.Frequency.IncCenter()
			}

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
					ngram.Before.At[j].Elems[elemID].SetWord(arr[index])
					// Track positional frequency: BEFORE context at offset (NgramSize/2 - j)
					if ngram.Before.At[j].Elems[elemID].Identity != nil {
						offset := (constants.NgramSize / 2) - j
						ngram.Before.At[j].Elems[elemID].Identity.Frequency.IncPos(constants.BEFORE, offset)
					}
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
					ngram.After.At[j].Elems[elemID].SetWord(arr[afterIndex])
					// Track positional frequency: AFTER context at offset (NgramSize/2 - j)
					if ngram.After.At[j].Elems[elemID].Identity != nil {
						offset := (constants.NgramSize / 2) - j
						ngram.After.At[j].Elems[elemID].Identity.Frequency.IncPos(constants.AFTER, offset)
					}
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

	// Add linked list elements (similarity detection) - Parallel with chunked work distribution
	const magicMultiplier = 4 // Tunable: workers = NumCPU × magicMultiplier
	numWorkers := runtime.NumCPU() * magicMultiplier

	// Calculate chunk size: divide work evenly across workers
	chunkSize := (*ngramLength + numWorkers - 1) / numWorkers // Ceiling division

	if functions.VerboseMode == constants.ON {
		fmt.Printf("Parallelizing similarity detection: %d workers, chunk size %d\n", numWorkers, chunkSize)
	}

	type workChunk struct {
		start, end int
	}

	type similarPair struct {
		k, j int
	}

	workChan := make(chan workChunk, numWorkers)
	resultsChan := make(chan []similarPair, numWorkers)
	var wg sync.WaitGroup

	// Start worker pool - each processes a chunk of k values
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localResults := make([]similarPair, 0, 1000) // Collect locally (no mutex!)

			for chunk := range workChan {
				// Process this worker's chunk of k values
				for k := chunk.start; k < chunk.end && k < *ngramLength; k++ {
					for j := k + constants.NgramSize; j < *ngramLength; j++ {
						if ng[j] != nil && ng[k] != nil &&
							ng[j].Word.Str != "" && ng[k].Word.Str != "" &&
							ng[j].Word.Str != ng[k].Word.Str {
							if ngram_t.NgramsSimilar(*ng[k], *ng[j]) {
								localResults = append(localResults, similarPair{k, j})
							}
						}
					}
				}
			}

			resultsChan <- localResults
		}()
	}

	// Distribute work chunks to workers
	for start := 0; start < *ngramLength; start += chunkSize {
		end := start + chunkSize
		if end > *ngramLength {
			end = *ngramLength
		}
		workChan <- workChunk{start, end}
	}
	close(workChan)

	// Wait for all workers to finish and close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from all workers (sequential merge, but only happens once)
	for localResults := range resultsChan {
		for _, pair := range localResults {
			ngram_t.AddSimilarNgramRef(&ng[pair.k], &ng[pair.j])
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

	// Vocabulary-based deduplication: track which word pairs we've already compared
	// Maps word1 -> word2 -> bool to avoid redundant comparisons
	// This changes O(N²) where N=occurrences to O(V²) where V=vocabulary
	seen := make(map[string]map[string]bool)

	for i := 0; i < ngramLength; i++ {
		if ng[i] == nil {
			continue
		}

		ngram := *ng[i]
		w1 := ngram.Word.StringClean()

		// Initialize inner map for this word if needed
		if seen[w1] == nil {
			seen[w1] = make(map[string]bool)
		}

		for j := 0; j < ngram.RefsCount; j++ {
			target := *ngram.Refs[j]
			w2 := target.Word.StringClean()

			// Skip if we've already compared this unique word pair
			if seen[w1][w2] {
				continue
			}
			seen[w1][w2] = true

			forward := morpheme_t.FindLongestMatch(ngram.Word, target.Word)
			forward.WordsCount = 0
			forward.Words = make([]*structs.Word, 0)
			forward.GenRegex(w1, w2, 0, 0)
			forward.AddWord(ngram.Word)
			forward.AddWord(target.Word)

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

