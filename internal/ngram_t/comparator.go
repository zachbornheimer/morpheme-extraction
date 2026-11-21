// Provides: NGram comparison abstraction for efficient similarity detection
package ngram_t

import (
	"slices"
	
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// NGramComparator efficiently compares ngrams for similarity
// Reuses buffers to avoid allocations in hot paths
// IMPORTANT: Uses array-based comparison (preserves duplicates) to match C algorithm
type NGramComparator struct {
	buffer        []string
	set           structs.StringSet
	ngramHalfSize int
}

// NewNGramComparator creates a new comparator with pre-allocated buffers
func NewNGramComparator() *NGramComparator {
	ngramHalfSize := constants.NgramSize / 2
	// Pre-allocate buffer large enough for worst case
	// Each ngram can have at most ngramHalfSize * maxElemsPerPosition words
	maxSize := ngramHalfSize * 10 * 2  // *2 for both ngrams
	
	return &NGramComparator{
		buffer:        make([]string, 0, maxSize),
		set:           structs.NewStringSet(maxSize),
		ngramHalfSize: ngramHalfSize,
	}
}

// AreSimilar checks if two ngrams are similar based on context word overlap
// Uses array-based comparison to preserve word frequency information (matches C algorithm)
// HYBRID APPROACH: Sort for small datasets (< 10K), map for large datasets
func (c *NGramComparator) AreSimilar(a, b *structs.Ngram) bool {
	// Count context words first
	elemsCount := a.CountContextWords(c.ngramHalfSize) + b.CountContextWords(c.ngramHalfSize)
	
	if elemsCount == 0 {
		return false
	}
	
	// Ensure buffer has capacity and set length
	if cap(c.buffer) < elemsCount {
		c.buffer = make([]string, elemsCount)
	} else {
		c.buffer = c.buffer[:elemsCount]
	}
	
	// Collect all words (with duplicates) into the buffer
	arrCount := 0
	arrCount += a.CollectContextWords(c.buffer, c.ngramHalfSize)
	arrCount += b.CollectContextWords(c.buffer[arrCount:], c.ngramHalfSize)
	
	// Count unique words using reusable set
	// Pre-size the set to avoid resizing
	if c.set == nil {
		c.set = structs.NewStringSet(elemsCount)
	}
	c.set.Clear()
	for _, word := range c.buffer[:arrCount] {
		c.set.Add(word)
	}
	count := c.set.Len()
	
	// Match C algorithm: arr_count -= 1
	arrCount--
	
	if arrCount <= 0 {
		return false
	}
	
	// Calculate similarity: percent_similar = 100 * ((arrCount - count) / arrCount)
	percentSimilar := int(100.0 * (float64(arrCount-count) / float64(arrCount)))
	
	return float64(percentSimilar) >= constants.ThresholdSimilarNgrams
}

// countUniqueSorted counts unique strings using sort + linear scan
// Faster than map for small datasets (< 10K elements)
// Modifies the input slice by sorting it in-place
func countUniqueSorted(words []string) int {
	if len(words) == 0 {
		return 0
	}
	
	// Sort in-place (O(n log n))
	slices.Sort(words)
	
	// Linear scan to count unique (O(n))
	unique := 1
	for i := 1; i < len(words); i++ {
		if words[i] != words[i-1] {
			unique++
		}
	}
	
	return unique
}
