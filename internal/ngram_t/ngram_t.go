// Package ngram_t implements n-gram data structures from ngram_t.c
// Line-for-line translation of ngram_t.c lines 1-198
package ngram_t

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
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

// NgramsSimilar checks if two ngrams are similar (lines 130-178)
func NgramsSimilar(a, b structs.Ngram) bool {
	percentSimilar := 0
	elemsCount := 0
	arrCount := 0

	// Count all elements
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if a.Before.At[i] != nil {
			elemsCount += a.Before.At[i].Count
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if a.After.At[i] != nil {
			elemsCount += a.After.At[i].Count
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if b.Before.At[i] != nil {
			elemsCount += b.Before.At[i].Count
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if b.After.At[i] != nil {
			elemsCount += b.After.At[i].Count
		}
	}

	if elemsCount == 0 {
		return false
	}

	arr := make([]string, elemsCount)

	// Collect all words
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if a.Before.At[i] != nil {
			for j := 0; j < a.Before.At[i].Count; j++ {
				if a.Before.At[i].Elems[j] != nil {
					word := string(a.Before.At[i].Elems[j].Str)
					if word != "" {
						arr[arrCount] = word
						arrCount++
					}
				}
			}
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if a.After.At[i] != nil {
			for j := 0; j < a.After.At[i].Count; j++ {
				if a.After.At[i].Elems[j] != nil {
					word := string(a.After.At[i].Elems[j].Str)
					if word != "" {
						arr[arrCount] = word
						arrCount++
					}
				}
			}
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if b.Before.At[i] != nil {
			for j := 0; j < b.Before.At[i].Count; j++ {
				if b.Before.At[i].Elems[j] != nil {
					word := string(b.Before.At[i].Elems[j].Str)
					if word != "" {
						arr[arrCount] = word
						arrCount++
					}
				}
			}
		}
	}
	for i := 0; i < (constants.NgramSize / 2); i++ {
		if b.After.At[i] != nil {
			for j := 0; j < b.After.At[i].Count; j++ {
				if b.After.At[i].Elems[j] != nil {
					word := string(b.After.At[i].Elems[j].Str)
					if word != "" {
						arr[arrCount] = word
						arrCount++
					}
				}
			}
		}
	}

	// Line 167 in C: count = uniq_words(arr, elems_count);
	// C passes elems_count, but this causes performance issues if elems_count >> arr_count
	// The C code likely has undefined behavior here (accessing uninitialized memory)
	// We'll pass arr_count for correctness and performance
	count := functions.UniqWords(arr[:arrCount], arrCount)
	
	// Line 170 in C: arr_count -= 1;
	arrCount--
	
	// Line 171 in C: percent_similar = (double) 100.00 * (((double)arr_count - (double)count)/(double)arr_count);
	if arrCount <= 0 {
		return false
	}
	percentSimilar = int(100.0 * (float64(arrCount-count) / float64(arrCount)))

	if float64(percentSimilar) >= constants.ThresholdSimilarNgrams {
		return true
	}
	return false
}

// AddSimilarNgramRef adds reference to similar ngram (lines 180-195)
// Fixed: Use append() instead of manual slice copy for O(1) amortized performance
func AddSimilarNgramRef(ngram **structs.Ngram, ngramRef **structs.Ngram) {
	(*ngram).Refs = append((*ngram).Refs, *ngramRef)
	(*ngram).RefsCount++
}

