// Package bitset provides efficient set membership operations using bit manipulation
// Used for O(1) set membership checks in UniqIdentities deduplication
package bitset

// BitSet represents a set of integers using bit manipulation
// Each bit position represents whether an ID is in the set
// Uses 64-bit words for efficient operations (64 IDs per word)
type BitSet struct {
	bits []uint64 // Bit array: each uint64 holds 64 bits
	size int      // Maximum ID supported (1-based)
}

// NewBitSet creates a bitset that can hold IDs from 1 to maxID
// Allocates enough uint64 words to cover the range
func NewBitSet(maxID int) *BitSet {
	if maxID < 0 {
		maxID = 0
	}
	// Round up to nearest 64 to ensure we have enough space
	// Add 1 because IDs are 1-based (ID 0 is unused)
	numWords := (maxID + 64) / 64
	return &BitSet{
		bits: make([]uint64, numWords),
		size: maxID,
	}
}

// Set marks an ID as present in the set
// ID must be >= 1 (WordIdentity IDs are 1-based)
func (b *BitSet) Set(id int) {
	if id < 1 || id > b.size {
		return // Ignore out of range IDs
	}
	// Convert 1-based ID to 0-based bit index
	bitIndex := id - 1
	word := bitIndex / 64
	bit := uint(bitIndex % 64)

	// Expand if needed (defensive)
	if word >= len(b.bits) {
		newBits := make([]uint64, word+1)
		copy(newBits, b.bits)
		b.bits = newBits
		b.size = (word + 1) * 64
	}

	b.bits[word] |= (1 << bit)
}

// IsSet checks if an ID is present in the set
// Returns false for out-of-range or unset IDs
func (b *BitSet) IsSet(id int) bool {
	if id < 1 || id > b.size {
		return false
	}
	// Convert 1-based ID to 0-based bit index
	bitIndex := id - 1
	word := bitIndex / 64
	bit := uint(bitIndex % 64)

	if word >= len(b.bits) {
		return false
	}

	return (b.bits[word] & (1 << bit)) != 0
}

// Clear resets all bits to 0 (empty set)
// Faster than allocating a new bitset
func (b *BitSet) Clear() {
	for i := range b.bits {
		b.bits[i] = 0
	}
}

// Count returns the number of set bits (cardinality)
// Uses population count (number of 1 bits)
func (b *BitSet) Count() int {
	count := 0
	for _, word := range b.bits {
		count += popcnt(word)
	}
	return count
}

// popcnt counts the number of 1 bits in a uint64
// Uses Brian Kernighan's algorithm: each iteration clears the lowest set bit
func popcnt(x uint64) int {
	count := 0
	for x != 0 {
		count++
		x &= x - 1 // Clear the lowest set bit
	}
	return count
}

// Size returns the maximum ID this bitset can hold
func (b *BitSet) Size() int {
	return b.size
}
