// Package nlp implements main NLP processing from nlp.c
// Line-for-line translation of nlp.c lines 64-259
// With Go concurrency optimizations for performance
package nlp

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/zachbornheimer/morpheme-extraction/internal/alphabet"
	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
	"github.com/zachbornheimer/morpheme-extraction/internal/directory"
	"github.com/zachbornheimer/morpheme-extraction/internal/export"
	"github.com/zachbornheimer/morpheme-extraction/internal/file"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/graph"
	"github.com/zachbornheimer/morpheme-extraction/internal/index"
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
	maxWordLen := 0
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
				var fileMaxWordLen int
				semanticGraph, fileMaxWordLen = BuildNgramRelationships(wd, f, &ngramLength, &ng, semanticGraph, cfg)
				if fileMaxWordLen > maxWordLen {
					maxWordLen = fileMaxWordLen
				}

				if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.ON {
					cfg.VPrint("\nPROCESSING SEQUENTIALLY!")
					cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
					lexCount = FindMorphemes(semanticGraph, ng, ngramLength, maxWordLen, header, &morphemes, cfg.OutputFilename, cfg)
				}
			}
		}
	}

	if cfg.ProcessFull == constants.ON || cfg.ProcessSequentially == constants.OFF {
		header = "All Processed Files"
		cfg.VPrint("PROCESSING SIMULTANEOUSLY!")
		cfg.VPrint("Extracting Morphemes and Building Lexical Categories.")
		lexCount = FindMorphemes(semanticGraph, ng, ngramLength, maxWordLen, header, &morphemes, cfg.OutputFilename, cfg)
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

	// Export semantic graph to Neo4j if requested
	if cfg.ExportNeo4jPath != "" && semanticGraph != nil {
		cfg.VPrint(fmt.Sprintf("Exporting semantic graph to CSV files for Neo4j bulk import..."))

		nodesPath := "neo4j-nodes.csv"
		relsPath := "neo4j-relationships.csv"

		if err := export.ExportToCSV(semanticGraph, nodesPath, relsPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to export CSV: %v\n", err)
		} else {
			cfg.VPrint(fmt.Sprintf("CSV export complete: %s, %s", nodesPath, relsPath))
		}
	}

	return 0
}

// exportNeo4j wraps the export package function for convenience
func exportNeo4j(g *graph.SemanticGraph, outputPath string) error {
	return export.ExportToCypher(g, outputPath)
}

// BuildNgramRelationships builds n-gram structures (lines 113-211)
// Updated to accept Config parameter for dependency injection (no globals)
// Returns semantic graph and max word length (for candidate pool sizing)
func BuildNgramRelationships(wd, f string, ngramLength *int, ngramFull *[]*structs.Ngram, semanticGraph *graph.SemanticGraph, cfg *config.Config) (*graph.SemanticGraph, int) {
	arr := functions.ExplodeSansnullStr(f, wd)
	wordCount := len(arr)
	maxWordLen := 0

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
				// Track max word length for candidate pool sizing
				if len(arr[i]) > maxWordLen {
					maxWordLen = len(arr[i])
				}
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
				offset := (constants.NgramSize / 2) - j

				// Before context
				if index > 0 && index < len(arr) && len(arr[index]) > 0 {
					// AddNgramElement now handles everything: find-or-create WordFreq, increment frequency, track position
					ngram_t.AddNgramElement(&ngram.Before.At[j], arr[index], constants.BEFORE, offset)
				}

				// After context
				afterIndex := index + (constants.NgramSize/2) + 1
				if afterIndex <= wordCount && afterIndex < len(arr) && len(arr[afterIndex]) > 0 {
					// AddNgramElement now handles everything: find-or-create WordFreq, increment frequency, track position
					ngram_t.AddNgramElement(&ngram.After.At[j], arr[afterIndex], constants.AFTER, offset)
				}
			}
		}

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
			node := semanticGraph.AddNode(ng[i].Word.Identity)
			// Track this ngram position for fast lookup during morpheme extraction
			node.AddPosition(i)
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

	// Debug counters for context index effectiveness
	var totalCandidates int64
	var totalSimilarPairs int64
	var totalPossiblePairs int64 = int64(*ngramLength) * int64(*ngramLength-constants.NgramSize) / 2

	// Add linked list elements (similarity detection) - Parallel with chunked work distribution
	const magicMultiplier = 4 // Tunable: workers = NumCPU × magicMultiplier
	numWorkers := runtime.NumCPU() * magicMultiplier

	// Calculate chunk size: divide work evenly across workers
	chunkSize := (*ngramLength + numWorkers - 1) / numWorkers // Ceiling division

	if cfg.IsVerbose() {
		fmt.Printf("Parallelizing similarity detection: %d workers, chunk size %d\n", numWorkers, chunkSize)
		fmt.Printf("Ngrams: %d, Theoretical max pairs: %d\n", *ngramLength, totalPossiblePairs)
	}

	type workChunk struct {
		start, end int
	}

	type similarPair struct {
		k, j         int
		overlapCount float64 // Number of shared context words (neural pathway strength)
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
					// Context index DISABLED - checking all pairs to accumulate edge weights
					// Neural pathway approach: edges strengthen with repeated evidence
					// _ = contextIndex.FindCandidates(k, constants.NgramSize/2, 2)

					for j := k + constants.NgramSize; j < *ngramLength; j++ {
						atomic.AddInt64(&totalCandidates, 1)

						// Use Identity pointer comparison (O(1) vs O(N) string comparison)
						if ng[j] != nil && ng[k] != nil &&
							ng[j].Word.Identity != nil && ng[k].Word.Identity != nil &&
							ng[j].Word.Identity != ng[k].Word.Identity {

							// Full similarity check with overlap count
							isSimilar, overlapCount := ngram_t.NgramsSimilar(*ng[k], *ng[j])
							if isSimilar {
								localResults = append(localResults, similarPair{k, j, overlapCount})
								atomic.AddInt64(&totalSimilarPairs, 1)
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
	edgeAddCount := make(map[[2]int]int) // Track how many times we add each edge
	for localResults := range resultsChan {
		for _, pair := range localResults {
			ngram_t.AddSimilarNgramRef(&ng[pair.k], &ng[pair.j])

			// Add edge to semantic graph with overlap count as weight (neural pathway strength)
			if ng[pair.k].Word.Identity != nil && ng[pair.j].Word.Identity != nil {
				id1, id2 := ng[pair.k].Word.Identity.ID, ng[pair.j].Word.Identity.ID
				if id1 > id2 {
					id1, id2 = id2, id1 // Normalize order for counting
				}
				edgeAddCount[[2]int{id1, id2}]++
				// Use pair.overlapCount as edge weight (accumulates across multiple similar n-gram pairs)
				semanticGraph.AddEdge(ng[pair.k].Word.Identity.ID, ng[pair.j].Word.Identity.ID, pair.overlapCount)
			}

			totalPairs++
		}
	}

	// Debug: Print edge add statistics
	if cfg.IsVerbose() {
		var maxAdds int
		var totalAdds int
		for _, count := range edgeAddCount {
			totalAdds += count
			if count > maxAdds {
				maxAdds = count
			}
		}
		fmt.Printf("\nEdge Addition Stats:\n")
		fmt.Printf("  Unique word pairs with edges: %d\n", len(edgeAddCount))
		fmt.Printf("  Total AddEdge calls: %d\n", totalAdds)
		fmt.Printf("  Average adds per edge: %.2f\n", float64(totalAdds)/float64(len(edgeAddCount)))
		fmt.Printf("  Max adds for single edge: %d\n", maxAdds)
	}

	// Print context index effectiveness stats
	if cfg.IsVerbose() {
		fmt.Printf("\nContext Index Stats:\n")
		fmt.Printf("  Candidate pairs from index: %d (avg %.1f per ngram)\n",
			totalCandidates, float64(totalCandidates)/float64(*ngramLength))
		fmt.Printf("  Similar pairs found: %d (%.2f%% of candidates)\n",
			totalSimilarPairs, 100.0*float64(totalSimilarPairs)/float64(totalCandidates))
		fmt.Printf("  Reduction: checked %d instead of %d (%.2f%% reduction)\n",
			totalCandidates, totalPossiblePairs, 100.0*(1.0-float64(totalCandidates)/float64(totalPossiblePairs)))
	}

	if cfg.IsVerbose() {
		fmt.Printf("\nSemantic Graph Built:\n")
		fmt.Printf("  Nodes: %d (unique words)\n", semanticGraph.NodeCount())
		fmt.Printf("  Edges: %d (similar word pairs)\n", semanticGraph.EdgeCount())
	}

	cfg.VPrint("")
	*ngramFull = ng

	return semanticGraph, maxWordLen
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
// Optimized to use pooled candidates and direct morpheme map (eliminates 3M allocations)
// Now uses SemanticGraph instead of Refs arrays for neighbor traversal
func FindMorphemes(semanticGraph *graph.SemanticGraph, ng []*structs.Ngram, ngramLength int, maxWordLen int, header string, lexMap *[]structs.LexicalCategories, outputFilename string, cfg *config.Config) int {
	// Create candidate pool with triangular number capacity (1 forward + max internals per pair)
	maxCandidatesPerPair := 1 + (maxWordLen * (maxWordLen + 1) / 2)
	var candidatePool = sync.Pool{
		New: func() interface{} {
			buf := make([]structs.MorphemeCandidate, 0, maxCandidatesPerPair)
			return &buf
		},
	}

	// Deduplicate morphemes as we go (instead of collecting 3M candidates first)
	morphemeMap := make(map[string]*structs.Morpheme)

	// Debug: Check edge weight distribution
	var totalNeighbors int64
	var neighborsAboveThreshold int64
	var maxWeight float64

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
			totalNeighbors++

			// Check edge weight (neural pathway strength)
			// Only extract morphemes if we have enough evidence (multiple similar n-gram occurrences)
			edgeWeight := node.GetEdgeWeight(neighborID)
			if edgeWeight > maxWeight {
				maxWeight = edgeWeight
			}
			if edgeWeight < constants.MinEdgeWeight {
				continue // Neural pathway not strong enough
			}
			neighborsAboveThreshold++

			// Get positions where this neighbor appears (O(1) instead of O(n) scan!)
			neighborNode := semanticGraph.GetNode(neighborID)
			if neighborNode == nil {
				continue
			}

			positions := neighborNode.GetPositions()
			for _, j := range positions {
				// Skip positions before i + NgramSize (maintains original algorithm constraint)
				if j < i+constants.NgramSize {
					continue
				}

				if ng[j] == nil {
					continue
				}

				target := *ng[j]
				w2 := target.Word.StringClean()

				// Get pooled candidate buffer for this pair
				pairCandidates := candidatePool.Get().(*[]structs.MorphemeCandidate)
				*pairCandidates = (*pairCandidates)[:0] // Reset length, keep capacity

				// Collect forward morpheme
				forwardStr := morpheme_t.FindLongestMatchString(ngram.Word, target.Word)
				*pairCandidates = append(*pairCandidates, structs.MorphemeCandidate{
					Morpheme:  forwardStr,
					Word1:     w1,
					Word2:     w2,
					Pos1:      0,
					Pos2:      0,
					IsStem:    false,
					IsForward: true,
				})

				// Collect internal morphemes for this pair (optimized: string slicing, no SetWord in loops)
				morpheme_list_t.FindInternalMorphemesOptimized(ngram.Word, target.Word, pairCandidates)

				// Process candidates from this pair → create/update morphemes
				for _, cand := range *pairCandidates {
					if cand.Morpheme == "" || cand.Morpheme == constants.EmptyMorpheme {
						continue
					}

					// Normalize morpheme string (trim trailing punctuation for map key)
					normalizedKey := strings.TrimRight(cand.Morpheme, ".,!?;:'\"-()")

					// Check if morpheme already exists
					if existing, found := morphemeMap[normalizedKey]; found {
						// Merge this pair's pattern into existing morpheme
						var word1, word2 structs.Word
						word1.SetWord(cand.Word1)
						word2.SetWord(cand.Word2)
						existing.GenRegex(cand.Word1, cand.Word2, cand.Pos1, cand.Pos2)
						existing.AddWord(word1)
						existing.AddWord(word2)
						existing.Freq++
					} else {
						// First time seeing this morpheme - create it with normalized string
						m := structs.NewMorpheme(normalizedKey)
						m.WordsCount = 0
						m.Words = make([]*structs.Word, 0)

						var word1, word2 structs.Word
						word1.SetWord(cand.Word1)
						word2.SetWord(cand.Word2)
						m.GenRegex(cand.Word1, cand.Word2, cand.Pos1, cand.Pos2)
						m.AddWord(word1)
						m.AddWord(word2)

						morphemeMap[normalizedKey] = &m
					}
				}

				// Return buffer to pool
				candidatePool.Put(pairCandidates)
			}
		}
	}

	cfg.VPrint("")

	if cfg.IsVerbose() {
		morpheme_list_t.PrintDebugStats()
	}

	// Convert morpheme map to list (already deduplicated, no FuseRegex needed!)
	morphemeList := structs.MorphemeList{
		List:  make([]*structs.Morpheme, 0, len(morphemeMap)),
		Count: len(morphemeMap),
	}
	for _, m := range morphemeMap {
		morphemeList.List = append(morphemeList.List, m)
	}

	// Sort by morpheme string for deterministic order (circumfix detection is order-sensitive)
	sort.Slice(morphemeList.List, func(i, j int) bool {
		return morphemeList.List[i].Morpheme < morphemeList.List[j].Morpheme
	})

	if cfg.IsVerbose() {
		fmt.Printf("Found %d unique morphemes (deduplicated inline)\n", morphemeList.Count)
	}

	lex := make([]structs.LexicalCategories, morphemeList.Count)
	morpheme_list_t.IdentifyTrueMorphemes(&morphemeList, &lex)

	if cfg.IsVerbose() {
		fmt.Printf("\nEdge Weight Stats:\n")
		fmt.Printf("  Total neighbor edges: %d\n", totalNeighbors)
		fmt.Printf("  Edges >= threshold (%.1f): %d (%.2f%%)\n",
			constants.MinEdgeWeight, neighborsAboveThreshold,
			100.0*float64(neighborsAboveThreshold)/float64(totalNeighbors))
		fmt.Printf("  Max edge weight: %.1f\n", maxWeight)
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
		} else {
			// Internal morphemes already validated, just add them
			morpheme_list_t.AddMorpheme(&result, m)
		}
	}

	return result
}

