// Package index provides context-based indexing for ngram similarity optimization
// Reduces O(V²) comparisons to O(V × candidates) by indexing shared context words
package index

import (
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// ContextIndex indexes ngrams by their context word IDs
// Stores ngram *indices* (not pointers) for O(1) lookups
type ContextIndex struct {
	index  map[int][]int         // WordIdentity.ID → ngram indices containing that word in context
	ngrams []*structs.Ngram       // Reference to full ngram array
	mu     sync.RWMutex
}

// NewContextIndex creates a new empty context index
// ngrams is a reference to the full ngram array for context extraction
func NewContextIndex(ngrams []*structs.Ngram) *ContextIndex {
	return &ContextIndex{
		index:  make(map[int][]int),
		ngrams: ngrams,
	}
}

// Add indexes an ngram by all its context word IDs
// ngramIdx is the index of this ngram in the ngrams array
func (ci *ContextIndex) Add(ngramIdx int, ngramHalfSize int) {
	if ngramIdx < 0 || ngramIdx >= len(ci.ngrams) {
		return
	}

	ng := ci.ngrams[ngramIdx]
	if ng == nil || ng.Word.Identity == nil {
		return
	}

	// Collect unique context word IDs (use map to deduplicate)
	contextIDs := make(map[int]bool)

	// Before context - iterate over unique WordFreq entries
	for i := 0; i < ngramHalfSize; i++ {
		if ng.Before.At[i] != nil {
			for j := 0; j < len(ng.Before.At[i].Elems); j++ {
				elem := ng.Before.At[i].Elems[j]
				if elem != nil && elem.Word != nil && elem.Word.Identity != nil {
					contextIDs[elem.Word.Identity.ID] = true
				}
			}
		}
	}

	// After context - iterate over unique WordFreq entries
	for i := 0; i < ngramHalfSize; i++ {
		if ng.After.At[i] != nil {
			for j := 0; j < len(ng.After.At[i].Elems); j++ {
				elem := ng.After.At[i].Elems[j]
				if elem != nil && elem.Word != nil && elem.Word.Identity != nil {
					contextIDs[elem.Word.Identity.ID] = true
				}
			}
		}
	}

	// Add ngram index to index for each unique context word
	ci.mu.Lock()
	defer ci.mu.Unlock()

	for id := range contextIDs {
		ci.index[id] = append(ci.index[id], ngramIdx)
	}
}

// FindCandidates returns ngram indices that share at least minOverlap context words with ngramIdx
// This reduces the candidate set from O(V) to O(candidates) where candidates << V
// Returns indices in ascending order, all >= ngramIdx
func (ci *ContextIndex) FindCandidates(ngramIdx int, ngramHalfSize int, minOverlap int) []int {
	if ngramIdx < 0 || ngramIdx >= len(ci.ngrams) {
		return nil
	}

	ng := ci.ngrams[ngramIdx]
	if ng == nil || ng.Word.Identity == nil {
		return nil
	}

	// Count how many context words each candidate shares with ng
	candidates := make(map[int]int) // ngram index → overlap count

	// Collect context word IDs from ng
	contextIDs := []int{}

	// Before context - expand by frequency
	for i := 0; i < ngramHalfSize; i++ {
		if ng.Before.At[i] != nil {
			for j := 0; j < len(ng.Before.At[i].Elems); j++ {
				elem := ng.Before.At[i].Elems[j]
				if elem != nil && elem.Word != nil && elem.Word.Identity != nil {
					// Add ID once per frequency count
					for f := 0; f < elem.Freq; f++ {
						contextIDs = append(contextIDs, elem.Word.Identity.ID)
					}
				}
			}
		}
	}

	// After context - expand by frequency
	for i := 0; i < ngramHalfSize; i++ {
		if ng.After.At[i] != nil {
			for j := 0; j < len(ng.After.At[i].Elems); j++ {
				elem := ng.After.At[i].Elems[j]
				if elem != nil && elem.Word != nil && elem.Word.Identity != nil {
					// Add ID once per frequency count
					for f := 0; f < elem.Freq; f++ {
						contextIDs = append(contextIDs, elem.Word.Identity.ID)
					}
				}
			}
		}
	}

	// For each context word, find all ngram indices that also have that word in their context
	ci.mu.RLock()
	for _, id := range contextIDs {
		for _, candidateIdx := range ci.index[id] {
			// Skip self, skip if same word, skip if before current ngram
			if candidateIdx == ngramIdx || candidateIdx < ngramIdx {
				continue
			}
			if ci.ngrams[candidateIdx] != nil && ci.ngrams[candidateIdx].Word.Identity != nil {
				if ci.ngrams[candidateIdx].Word.Identity.ID != ng.Word.Identity.ID {
					candidates[candidateIdx]++
				}
			}
		}
	}
	ci.mu.RUnlock()

	// Return candidate indices with sufficient overlap
	result := make([]int, 0, len(candidates))
	for candIdx, count := range candidates {
		if count >= minOverlap {
			result = append(result, candIdx)
		}
	}

	return result
}

// Size returns the number of unique context words in the index
func (ci *ContextIndex) Size() int {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	return len(ci.index)
}

// Stats returns indexing statistics for debugging
func (ci *ContextIndex) Stats() (contextWords int, totalEntries int) {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	contextWords = len(ci.index)
	for _, ngrams := range ci.index {
		totalEntries += len(ngrams)
	}
	return contextWords, totalEntries
}
