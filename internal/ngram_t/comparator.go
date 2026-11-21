// Provides: NGram comparison abstraction for efficient similarity detection
package ngram_t

import (
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

var similarCount = 0
var notSimilarCount = 0

// AreSimilar checks if two ngrams are similar based on context word overlap
// Uses array-based comparison to preserve word frequency information (matches C algorithm)
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
	countA := a.CollectContextWords(c.buffer, c.ngramHalfSize)
	countB := b.CollectContextWords(c.buffer[countA:], c.ngramHalfSize)
	arrCount = countA + countB
	
	// Count unique words using reusable set
	// CRITICAL: Only iterate over filled portion [:arrCount], not full buffer
	c.set.Clear()
	for _, word := range c.buffer[:arrCount] {
		c.set.Add(word)
	}
	// Match C algorithm: uniq_words returns count-1, not count
	count := c.set.Len() - 1
	
	// Match C algorithm: arr_count -= 1
	arrCount--
	
	if arrCount <= 0 {
		return false
	}
	
	// Calculate similarity: percent_similar = 100 * ((arrCount - count) / arrCount)
	percentSimilar := int(100.0 * (float64(arrCount-count) / float64(arrCount)))
	
	result := float64(percentSimilar) >= constants.ThresholdSimilarNgrams
	if result {
		similarCount++
	} else {
		notSimilarCount++
	}
	return result
}

// GetSimilarityCounts returns the number of similar and not-similar comparisons
func GetSimilarityCounts() (int, int) {
	return similarCount, notSimilarCount
}

