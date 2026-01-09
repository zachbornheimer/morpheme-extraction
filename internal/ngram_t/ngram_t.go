// Package ngram_t implements n-gram data structures from ngram_t.c
// Line-for-line translation of ngram_t.c lines 1-198
package ngram_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

var debugCounter = 0
var debugCallCounter = 0

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

// NgramsSimilar has been replaced by NGramComparator.AreSimilar()
// The comparator approach reuses buffers to avoid allocations in hot paths.
// See comparator.go for the implementation.

// AddSimilarNgramRef adds reference to similar ngram (lines 180-195)
// Fixed: Use append() instead of manual slice copy for O(1) amortized performance
func AddSimilarNgramRef(ngram **structs.Ngram, ngramRef **structs.Ngram) {
	(*ngram).Refs = append((*ngram).Refs, *ngramRef)
	(*ngram).RefsCount++
}

