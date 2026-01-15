// Package ngram_t implements n-gram data structures from ngram_t.c
// Line-for-line translation of ngram_t.c lines 1-198
package ngram_t

import (
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// contextIdentityPool reuses Identity pointer slice buffers for NgramsSimilar
// Uses O(1) pointer equality instead of O(N) string comparison
// Reduces allocations from ~600MB to near-zero for this hot path
var contextIdentityPool = sync.Pool{
	New: func() any {
		// Pre-allocate reasonable size (most ngrams have < 100 context words)
		buf := make([]*dictionary.WordIdentity, 0, 100)
		return &buf
	},
}

// AddNgramElement adds or increments a word in the ngram array
// Now uses frequency-based storage: finds existing word or creates new WordFreq entry
// wordStr: the word to add
// context: constants.BEFORE or constants.AFTER
// offset: position offset from center
// Returns: index of the WordFreq entry (for compatibility)
func AddNgramElement(atIndex **structs.WordArray, wordStr string, context int, offset int) int {
	if *atIndex == nil {
		return -1
	}

	// Get word identity from dictionary
	wordIdentity := dictionary.Dictionary.Word(wordStr)
	if wordIdentity == nil {
		return -1
	}

	// Search for existing WordFreq with this identity
	for i := 0; i < len((*atIndex).Elems); i++ {
		if (*atIndex).Elems[i] != nil &&
		   (*atIndex).Elems[i].Word != nil &&
		   (*atIndex).Elems[i].Word.Identity == wordIdentity {
			// Found existing entry - increment frequency
			(*atIndex).Elems[i].Freq++
			(*atIndex).Count++ // Increment total count
			// Track positional frequency in dictionary
			wordIdentity.Frequency.IncPos(context, offset)
			return i
		}
	}

	// Not found - create new WordFreq entry
	newWord := &structs.Word{}
	newWord.SetWord(wordStr)

	newWordFreq := &structs.WordFreq{
		Word: newWord,
		Freq: 1,
	}

	(*atIndex).Elems = append((*atIndex).Elems, newWordFreq)
	(*atIndex).Count++ // Increment total count

	// Track positional frequency in dictionary
	if newWord.Identity != nil {
		newWord.Identity.Frequency.IncPos(context, offset)
	}

	return len((*atIndex).Elems) - 1
}

// NewNgram creates a new ngram structure (lines 30-52)
func NewNgram() structs.Ngram {
	ng := structs.Ngram{}

	// Initialize before and after arrays
	// C uses OpenMP parallel for, but we'll do it serially
	for index := 0; index < (constants.NgramSize / 2); index++ {
		ng.Before.At[index] = &structs.WordArray{
			Count: 0,
			Elems: make([]*structs.WordFreq, 0), // Start empty, will grow as needed
		}
		ng.After.At[index] = &structs.WordArray{
			Count: 0,
			Elems: make([]*structs.WordFreq, 0), // Start empty, will grow as needed
		}
	}

	return ng
}

// FreeNgram frees ngram memory (lines 54-64)
// In Go, this is mostly handled by GC, but we keep the function for compatibility
func FreeNgram(ng *structs.Ngram) {
	// Go's GC will handle cleanup
	// Keeping function signature for C compatibility
}

// Uniqify removes duplicates from ngram arrays (lines 68-92)
func Uniqify(bora *structs.NgramHalfArray) {
	// Note: With frequency-based storage, this function is now redundant
	// AddNgramElement already deduplicates and merges frequencies
	// Keeping this for compatibility but it's a no-op
	_ = bora // unused
}

// NgramsSimilar checks if two ngrams are similar (lines 130-178)
// Uses Identity pointers for O(1) comparison, expands frequencies to match original algorithm
// Returns: (isSimilar bool, overlapCount float64)
// overlapCount = number of shared context words (neural pathway strength)
func NgramsSimilar(a, b structs.Ngram) (bool, float64) {
	ngramHalfSize := constants.NgramSize / 2

	// Count all elements using struct methods (already accounts for frequencies via Count)
	elemsCount := a.CountContextWords(ngramHalfSize) + b.CountContextWords(ngramHalfSize)

	if elemsCount == 0 {
		return false, 0.0
	}

	// Get Identity buffer from pool (optimization: reuse allocations)
	bufPtr := contextIdentityPool.Get().(*[]*dictionary.WordIdentity)
	arr := *bufPtr
	arr = arr[:0] // Reset length, keep capacity

	// Ensure buffer has enough capacity
	if cap(arr) < elemsCount {
		arr = make([]*dictionary.WordIdentity, elemsCount)
	} else {
		arr = arr[:elemsCount]
	}

	// Return buffer to pool when done
	defer func() {
		*bufPtr = arr[:0] // Reset for next use
		contextIdentityPool.Put(bufPtr)
	}()

	arrCount := 0

	// Collect all Identity pointers (expanded by frequency) using struct methods
	arrCount += a.CollectContextIdentities(arr[arrCount:], ngramHalfSize)
	arrCount += b.CollectContextIdentities(arr[arrCount:], ngramHalfSize)

	// Count unique identities using bitset
	count := functions.UniqIdentities(arr[:arrCount], arrCount)

	// Line 170 in C: arr_count -= 1;
	arrCount--

	// Line 171 in C: percent_similar = (double) 100.00 * (((double)arr_count - (double)count)/(double)arr_count);
	if arrCount <= 0 {
		return false, 0.0
	}
	percentSimilar := int(100.0 * (float64(arrCount-count) / float64(arrCount)))

	// Calculate overlap count (neural pathway strength)
	overlapCount := float64(arrCount - count) // Number of shared context words

	isSimilar := float64(percentSimilar) >= constants.ThresholdSimilarNgrams
	return isSimilar, overlapCount
}

// AddSimilarNgramRef adds reference to similar ngram (lines 180-195)
// NOTE: This function is now a no-op since we use SemanticGraph instead of Refs
// Kept for backward compatibility during transition
func AddSimilarNgramRef(ngram **structs.Ngram, ngramRef **structs.Ngram) {
	// No-op: relationships are now stored in SemanticGraph
}

