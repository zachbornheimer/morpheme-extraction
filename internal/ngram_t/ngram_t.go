// Package ngram_t implements n-gram data structures from ngram_t.c
// Line-for-line translation of ngram_t.c lines 1-198
package ngram_t

import (
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// contextWordPool reuses string slice buffers for NgramsSimilar
// Reduces allocations from ~100MB to near-zero for this hot path
var contextWordPool = sync.Pool{
	New: func() any {
		// Pre-allocate reasonable size (most ngrams have < 100 context words)
		buf := make([]string, 0, 100)
		return &buf
	},
}

// AddNgramElement adds element to ngram array (lines 22-28)
func AddNgramElement(atIndex **structs.WordArray, index int) int {
	(*atIndex).Count++
	newElems := make([]*structs.Word, (*atIndex).Count+1)
	copy(newElems, (*atIndex).Elems)
	(*atIndex).Elems = newElems
	return (*atIndex).Count - 1
}

// NewNgram creates a new ngram structure (lines 30-52)
func NewNgram() structs.Ngram {
	ng := structs.Ngram{}

	// Initialize before and after arrays
	// C uses OpenMP parallel for, but we'll do it serially
	for index := 0; index < (constants.NgramSize / 2); index++ {
		ng.Before.At[index] = &structs.WordArray{
			Count: 0,
			Elems: make([]*structs.Word, 1),
		}
		ng.After.At[index] = &structs.WordArray{
			Count: 0,
			Elems: make([]*structs.Word, 1),
		}
	}

	ng.RefsCount = 0
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
	// bora = before or after
	for i := 0; i < (constants.NgramSize / 2); i++ {
		count := bora.At[i].Count
		
		for j := 0; j < count; j++ {
			elems := make([]*structs.Word, bora.At[i].Count+1)
			elemsCount := 0
			elems[elemsCount] = bora.At[i].Elems[j]
			elemsCount++
			elem := elems[elemsCount-1]

			// C used OpenMP parallel for
			for k := j + 1; k < count-1; k++ {
				comp := bora.At[i].Elems[k]
				if comp.Freq != 1 {
					continue
				}
				if string(elem.Str) == string(comp.Str) {
					elem.Freq += bora.At[i].Elems[k].Freq
				} else {
					elems[elemsCount] = bora.At[i].Elems[k]
					elemsCount++
				}
			}
			bora.At[i].Elems = elems[:elemsCount]
		}
	}
}

// NgramsSimilar checks if two ngrams are similar (lines 130-178)
// Uses struct methods for abstraction while maintaining original C algorithm
func NgramsSimilar(a, b structs.Ngram) bool {
	ngramHalfSize := constants.NgramSize / 2

	// Count all elements using struct methods
	elemsCount := a.CountContextWords(ngramHalfSize) + b.CountContextWords(ngramHalfSize)

	if elemsCount == 0 {
		return false
	}

	// Get buffer from pool (optimization: reuse allocations)
	bufPtr := contextWordPool.Get().(*[]string)
	arr := *bufPtr
	arr = arr[:0] // Reset length, keep capacity

	// Ensure buffer has enough capacity
	if cap(arr) < elemsCount {
		arr = make([]string, elemsCount)
	} else {
		arr = arr[:elemsCount]
	}

	// Return buffer to pool when done
	defer func() {
		*bufPtr = arr[:0] // Reset for next use
		contextWordPool.Put(bufPtr)
	}()

	arrCount := 0

	// Collect all words using struct methods
	arrCount += a.CollectContextWords(arr[arrCount:], ngramHalfSize)
	arrCount += b.CollectContextWords(arr[arrCount:], ngramHalfSize)

	// Line 167 in C: count = uniq_words(arr, elems_count);
	// We pass arrCount for correctness (C has undefined behavior here)
	count := functions.UniqWords(arr[:arrCount], arrCount)

	// Line 170 in C: arr_count -= 1;
	arrCount--

	// Line 171 in C: percent_similar = (double) 100.00 * (((double)arr_count - (double)count)/(double)arr_count);
	if arrCount <= 0 {
		return false
	}
	percentSimilar := int(100.0 * (float64(arrCount-count) / float64(arrCount)))

	return float64(percentSimilar) >= constants.ThresholdSimilarNgrams
}

// AddSimilarNgramRef adds reference to similar ngram (lines 180-195)
// Fixed: Use append() instead of manual slice copy for O(1) amortized performance
func AddSimilarNgramRef(ngram **structs.Ngram, ngramRef **structs.Ngram) {
	(*ngram).Refs = append((*ngram).Refs, *ngramRef)
	(*ngram).RefsCount++
}

