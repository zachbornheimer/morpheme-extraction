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
	"github.com/zachbornheimer/morpheme-extraction/internal/graph"
	"github.com/zachbornheimer/morpheme-extraction/internal/index"
	"github.com/zachbornheimer/morpheme-extraction/internal/minhash"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_list_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/ngram_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// Nlp runs the main NLP process (lines 64-111)
func Nlp(cfg *config.Config) int {
	// Set directory corpus (TODO: refactor directory package to use Config)
	directory.CorpusDir = cfg.CorpusDir

	var f string
	var header string
	index := 0
	ngramLength := 0
	lexCount := 0
	var ng []*structs.Ngram
	var morphemes []structs.LexicalCategories
	var semanticGraph *graph.SemanticGraph

	for {
		f = directory.GetFiles(&index, &header)
		if f == "" {
			break
		}

		if len(f) > 2 {
			cfg.VPrint("Looking for Word Delimiter...")
			wd, err := alphabet.FindWordDelimiter(&f, cfg)
			if cfg.IsVerbose() {
				fmt.Printf("Found Word Delimiter: %q\n", wd)
			}

			if err == constants.ErrOverruled {
				continue
			}

			if err == nil && wd != "" {
				cfg.VPrint("Building NGRAMS")
				semanticGraph = BuildNgramRelationships(wd, f, &ngramLength, &ng, semanticGraph, cfg)

				if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.ON {
					cfg.VPrint("\nPROCESSING SEQUENTIALLY!")
					cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
					lexCount = FindMorphemes(semanticGraph, ng, ngramLength, header, &morphemes, cfg.OutputFilename, cfg)
				}
			}
		}
	}

	if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.OFF {
		header = "All Processed Files"
		cfg.VPrint("PROCESSING SIMULTANEOUSLY!")
		cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
		lexCount = FindMorphemes(semanticGraph, ng, ngramLength, header, &morphemes, cfg.OutputFilename, cfg)
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

	// Print graph stats if verbose
	if cfg.IsVerbose() && semanticGraph != nil {
		fmt.Printf("\nSemantic Graph Statistics:\n")
		fmt.Printf("  Nodes:      %10d (unique words in graph)\n", semanticGraph.NodeCount())
		fmt.Printf("  Edges:      %10d (co-occurrence relationships)\n", semanticGraph.EdgeCount())
	}

	// Print memory stats if verbose
	if cfg.IsVerbose() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("\nMemory Statistics:\n")
		fmt.Printf("  Alloc:      %10d MB (currently allocated)\n", m.Alloc/1024/1024)
		fmt.Printf("  TotalAlloc: %10d MB (cumulative allocated)\n", m.TotalAlloc/1024/1024)
		fmt.Printf("  Sys:        %10d MB (obtained from system)\n", m.Sys/1024/1024)
		fmt.Printf("  NumGC:      %10d (garbage collections)\n", m.NumGC)
	}

	return 0
}

// BuildNgramRelationships builds n-gram structures (lines 113-211)
// Updated to accept Config parameter for dependency injection (no globals)
func BuildNgramRelationships(wd, f string, ngramLength *int, ngramFull *[]*structs.Ngram, semanticGraph *graph.SemanticGraph, cfg *config.Config) *graph.SemanticGraph {
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

	cfg.VPrint("Setting NGRAM elements")

	// Hash map for O(1) ngram lookup using interned Identity pointers
	// Pointer equality is O(1) vs O(N) string comparison
	ngramMap := make(map[*dictionary.WordIdentity]int, wordCount)

	// Build initial map from existing ngrams
	for idx := 0; idx < addLoc; idx++ {
		if ng[idx] != nil && ng[idx].Word.Identity != nil {
			ngramMap[ng[idx].Word.Identity] = idx
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
			// O(1) hash map lookup using interned Identity pointer
			wordIdentity := dictionary.Dictionary.Word(arr[i])
			if existingIdx, found := ngramMap[wordIdentity]; found {
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

		if cfg.IsVerbose() {
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

		// Compute MinHash signature for this ngram's context
		// Used for fast similarity filtering (Phase 5)
		contextIDs := minhash.ExtractContextIDs(&ngram, constants.NgramSize/2)
		ngram.MinHash = minhash.ComputeMinHash(contextIDs)

		if makeNew == 1 {
			ng[addLoc] = &ngram
			// Add to map so future occurrences of this word will merge into this ngram
			if ngram.Word.Identity != nil {
				ngramMap[ngram.Word.Identity] = addLoc
			}
			addLoc++
		} else {
			// Update existing ngram (line 185-186 in C)
			// In C: *ng[q] = ngram;
			// We need to update the existing ngram with our modifications
			*ng[q] = ngram
		}
	}

	*ngramLength = addLoc

	// Create semantic graph if not provided (first file), then add nodes
	if semanticGraph == nil {
		semanticGraph = graph.NewSemanticGraph()
	}
	for i := 0; i < *ngramLength; i++ {
		if ng[i] != nil && ng[i].Word.Identity != nil {
			semanticGraph.AddNode(ng[i].Word.Identity)
		}
	}

	// Build context index for O(V × candidates) similarity detection (Phase 6)
	// Reduces candidate pairs by ~90% (from O(V²) to O(V × candidates))
	contextIndex := index.NewContextIndex(ng)
	for i := 0; i < *ngramLength; i++ {
		contextIndex.Add(i, constants.NgramSize/2)
	}

	if cfg.IsVerbose() {
		contextWords, totalEntries := contextIndex.Stats()
		fmt.Printf("Context index: %d unique context words, %d total entries\n", contextWords, totalEntries)
	}

	// Add linked list elements (similarity detection) - Parallel with chunked work distribution
	const magicMultiplier = 4 // Tunable: workers = NumCPU × magicMultiplier
	numWorkers := runtime.NumCPU() * magicMultiplier

	// Calculate chunk size: divide work evenly across workers
	chunkSize := (*ngramLength + numWorkers - 1) / numWorkers // Ceiling division

	if cfg.IsVerbose() {
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
					// Phase 6: Context index - DISABLED
					// TODO: Context index needs tuning - morphologically similar words
					// may not share context words (e.g., "walking" and "talked" in different contexts)
					// With minOverlap=1, got 3,525 morphemes vs expected 3,702
					// For now, checking all pairs to ensure correctness
					_ = contextIndex.FindCandidates(k, constants.NgramSize/2, 1)

					for j := k + constants.NgramSize; j < *ngramLength; j++ {
						// Use Identity pointer comparison (O(1) vs O(N) string comparison)
						if ng[j] != nil && ng[k] != nil &&
							ng[j].Word.Identity != nil && ng[k].Word.Identity != nil &&
							ng[j].Word.Identity != ng[k].Word.Identity {

							// MinHash fast rejection filter (Phase 5) - DISABLED
							// TODO: MinHash threshold needs calibration
							// MinHash estimates Jaccard, NgramsSimilar uses different metric
							// For now, disabled to ensure correctness
							_ = minhash.EstimateSimilarity(ng[k].MinHash, ng[j].MinHash)
							// if minHashSim < threshold { continue }

							// Full similarity check
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
	totalPairs := 0
	for localResults := range resultsChan {
		for _, pair := range localResults {
			ngram_t.AddSimilarNgramRef(&ng[pair.k], &ng[pair.j])

			// Add edge to semantic graph (parallel construction)
			if ng[pair.k].Word.Identity != nil && ng[pair.j].Word.Identity != nil {
				semanticGraph.AddEdge(ng[pair.k].Word.Identity.ID, ng[pair.j].Word.Identity.ID, 1.0)
			}

			totalPairs++
		}
	}

	// Count total pairs checked
	totalChecked := 0
	for k := 0; k < *ngramLength; k++ {
		for j := k + constants.NgramSize; j < *ngramLength; j++ {
			totalChecked++
		}
	}

	fmt.Printf("DEBUG: ngram_length: %d\n", *ngramLength)
	fmt.Printf("DEBUG: Pairs checked: %d, similar: %d (%.1f%%)\n", totalChecked, totalPairs, 100.0*float64(totalPairs)/float64(totalChecked))

	if cfg.IsVerbose() {
		fmt.Printf("\nSemantic Graph Built:\n")
		fmt.Printf("  Nodes: %d (unique words)\n", semanticGraph.NodeCount())
		fmt.Printf("  Edges: %d (similar word pairs)\n", semanticGraph.EdgeCount())
	}

	cfg.VPrint("")
	*ngramFull = ng

	return semanticGraph
}

// findAllNgramsByID searches the ngram array for ALL ngrams with matching word ID
// Returns slice of ngrams (can be empty if none found)
// A word can appear at multiple positions in the corpus (multiple ngram indices)
func findAllNgramsByID(ng []*structs.Ngram, wordID int, ngramLength int) []*structs.Ngram {
	results := make([]*structs.Ngram, 0, 4) // Usually not many duplicates
	for i := 0; i < ngramLength; i++ {
		if ng[i] != nil && ng[i].Word.Identity != nil && ng[i].Word.Identity.ID == wordID {
			results = append(results, ng[i])
		}
	}
	return results
}

// FindMorphemes extracts morphemes from n-grams (lines 213-259)
// Sequential to preserve morpheme order (affects FuseRegex results)
// Updated to accept Config parameter for dependency injection (no globals)
// Optimized to use MorphemeCandidate for lazy allocation
// Now uses SemanticGraph instead of Refs arrays for neighbor traversal
func FindMorphemes(semanticGraph *graph.SemanticGraph, ng []*structs.Ngram, ngramLength int, header string, lexMap *[]structs.LexicalCategories, outputFilename string, cfg *config.Config) int {
	// Collect lightweight candidates first (no expensive allocations)
	candidates := make([]structs.MorphemeCandidate, 0, ngramLength*100)

	for i := 0; i < ngramLength; i++ {
		if ng[i] == nil {
			continue
		}

		ngram := *ng[i]
		w1 := ngram.Word.StringClean()

		// Get similar ngrams from semantic graph instead of Refs array
		node := semanticGraph.GetNode(ngram.Word.Identity.ID)
		if node == nil {
			continue
		}

		// Get neighbor IDs and iterate through them
		neighborIDs := node.GetNeighbors()
		for _, neighborID := range neighborIDs {
			// Find ngrams with this neighbor word ID at indices j >= i + NgramSize
			// This matches the original algorithm's constraint
			for j := i + constants.NgramSize; j < ngramLength; j++ {
				if ng[j] == nil || ng[j].Word.Identity == nil {
					continue
				}

				// Check if this ngram has the neighbor word ID
				if ng[j].Word.Identity.ID != neighborID {
					continue
				}

				target := *ng[j]
				w2 := target.Word.StringClean()

				// Get forward morpheme (whole word prefix/suffix match)
				// Always add (matches original behavior - no empty check)
				forwardStr := morpheme_t.FindLongestMatchString(ngram.Word, target.Word)
				candidates = append(candidates, structs.MorphemeCandidate{
					Morpheme:  forwardStr,
					Word1:     w1,
					Word2:     w2,
					Pos1:      0,
					Pos2:      0,
					IsStem:    false, // Let IdentifyTrueMorphemes classify based on regex patterns
					IsForward: true,  // Forward morphemes always added (no regex check)
				})

				// Find internal morphemes (infixes) - adds to candidates
				morpheme_list_t.FindInternalMorphemes(ngram.Word, target.Word, &candidates)
			}
		}
	}

	cfg.VPrint("")

	if cfg.IsVerbose() {
		fmt.Printf("\nCollected %d candidates\n", len(candidates))
		morpheme_list_t.PrintDebugStats()
	}

	// Build full Morpheme objects from candidates (with GenRegex)
	internal := buildMorphemesFromCandidatesWithRegex(candidates)

	if cfg.IsVerbose() {
		fmt.Printf("Built %d morphemes from candidates\n", internal.Count)
	}

	morphemeList := morpheme_list_t.FuseRegex(internal)

	if cfg.IsVerbose() {
		fmt.Printf("\nFound %d internal morphemes before fusion\n", internal.Count)
		fmt.Printf("Found %d morphemes after fusion\n", morphemeList.Count)
	}

	lex := make([]structs.LexicalCategories, morphemeList.Count)
	morpheme_list_t.IdentifyTrueMorphemes(&morphemeList, &lex)
	
	if cfg.IsVerbose() {
		fmt.Printf("Writing %d morphemes to file\n", morphemeList.Count)
	}
	
	file.WriteToFile(outputFilename, lex, morphemeList.Count, header)

	*lexMap = lex
	return morphemeList.Count
}

// buildMorphemesFromCandidatesWithRegex creates full Morpheme objects from candidates
// and generates regex patterns using GenRegex. This preserves the original algorithm
// while avoiding expensive allocations in hot loops.
func buildMorphemesFromCandidatesWithRegex(candidates []structs.MorphemeCandidate) structs.MorphemeList {
	result := structs.MorphemeList{
		List:  make([]*structs.Morpheme, 0, len(candidates)),
		Count: 0,
	}

	skippedEmpty := 0
	addedForward := 0
	addedInternal := 0

	for _, cand := range candidates {
		// Don't skip empty morphemes here - let FuseRegex handle that
		// (matching original behavior where all morphemes go to FuseRegex)

		// Create morpheme with GenRegex
		m := structs.NewMorpheme(cand.Morpheme)
		m.WordsCount = 0
		m.Words = make([]*structs.Word, 0)

		// Create Word objects for GenRegex
		var word1, word2 structs.Word
		word1.SetWord(cand.Word1)
		word2.SetWord(cand.Word2)

		// Generate regex pattern
		m.GenRegex(cand.Word1, cand.Word2, cand.Pos1, cand.Pos2)
		m.AddWord(word1)
		m.AddWord(word2)

		// Internal candidates already passed regex check in FindInternalMorphemes
		// Forward morphemes always added (no regex check needed)
		if cand.IsForward {
			morpheme_list_t.AddMorpheme(&result, m)
			addedForward++
		} else {
			// Internal morphemes already validated, just add them
			morpheme_list_t.AddMorpheme(&result, m)
			addedInternal++
		}
	}

	fmt.Printf("DEBUG: skippedEmpty=%d, addedForward=%d, addedInternal=%d\n",
		skippedEmpty, addedForward, addedInternal)

	return result
}

