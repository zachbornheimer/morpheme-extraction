package bitset

import "sync"

// bitsetPool provides object pooling for BitSet instances
// Reduces GC pressure by reusing allocated bitsets
// sync.Pool is thread-safe and managed by the runtime
var bitsetPool = sync.Pool{
	New: func() interface{} {
		// Preallocate for typical vocabulary size
		// 10,000 words = ~157 uint64 words = ~1.2KB per bitset
		return NewBitSet(10000)
	},
}

// Get retrieves a cleared BitSet from the pool
// BitSet is guaranteed to be empty (all bits cleared)
func Get() *BitSet {
	bs := bitsetPool.Get().(*BitSet)
	bs.Clear()
	return bs
}

// Put returns a BitSet to the pool for reuse
// Caller must not use the bitset after putting it back
func Put(bs *BitSet) {
	if bs != nil {
		bitsetPool.Put(bs)
	}
}
