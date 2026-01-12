// Package dictionary provides a singleton word identity store with positional frequency tracking
package dictionary

import (
	"sync"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
)

// Frequency encapsulates positional frequency tracking
// Array positions map directly to ngram positions:
//   [0..NgramSize/2-1] = before positions (far to near)
//   [NgramSize/2]      = center position
//   [NgramSize/2+1..N] = after positions (near to far)
type Frequency struct {
	positions [constants.NgramSize]int
	mu        sync.Mutex
}

// GetCenter returns frequency at center position
func (f *Frequency) GetCenter() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.positions[constants.NgramSize/2]
}

// IncCenter increments frequency at center position
func (f *Frequency) IncCenter() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.positions[constants.NgramSize/2]++
}

// GetPos returns frequency at specified context position
// context: constants.BEFORE, constants.CENTER, constants.AFTER
// offset: distance from center (1 to NgramSize/2)
func (f *Frequency) GetPos(context int, offset int) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	idx := f.calcIndex(context, offset)
	if idx == -1 {
		return 0 // Invalid position
	}
	return f.positions[idx]
}

// IncPos increments frequency at specified context position
func (f *Frequency) IncPos(context int, offset int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	idx := f.calcIndex(context, offset)
	if idx != -1 {
		f.positions[idx]++
	}
}

// DecPos decrements frequency at specified context position
func (f *Frequency) DecPos(context int, offset int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	idx := f.calcIndex(context, offset)
	if idx != -1 && f.positions[idx] > 0 {
		f.positions[idx]--
	}
}

// calcIndex converts semantic position to array index
// Direct mapping: BEFORE subtracts from center, AFTER adds to center
func (f *Frequency) calcIndex(context int, offset int) int {
	center := constants.NgramSize / 2

	switch context {
	case constants.CENTER:
		return center
	case constants.BEFORE:
		idx := center - offset
		if idx >= 0 && idx < constants.NgramSize {
			return idx
		}
	case constants.AFTER:
		idx := center + offset
		if idx >= 0 && idx < constants.NgramSize {
			return idx
		}
	}
	return -1 // Invalid
}

// WordIdentity represents a unique word with positional frequency tracking
// One instance per unique string (interned)
type WordIdentity struct {
	Str       string
	Frequency Frequency
}

// String returns the word string (for compatibility)
func (w *WordIdentity) String() string {
	return w.Str
}

// dict is the singleton dictionary
type dict struct {
	words map[string]*WordIdentity
	mu    sync.RWMutex
}

// Dictionary is the global singleton word identity store
var Dictionary = &dict{
	words: make(map[string]*WordIdentity),
}

// Word gets or creates a word identity (factory pattern)
// Returns pointer for O(1) equality checking
// Thread-safe for concurrent access
func (d *dict) Word(str string) *WordIdentity {
	// Fast path: read lock
	d.mu.RLock()
	if word, exists := d.words[str]; exists {
		d.mu.RUnlock()
		return word
	}
	d.mu.RUnlock()

	// Slow path: write lock
	d.mu.Lock()
	defer d.mu.Unlock()

	// Double-check after acquiring write lock
	if word, exists := d.words[str]; exists {
		return word
	}

	// Create new identity
	word := &WordIdentity{
		Str:       str,
		Frequency: Frequency{},
	}
	d.words[str] = word
	return word
}

// Size returns number of unique words in dictionary
func (d *dict) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.words)
}

// Clear resets the dictionary (for testing)
func (d *dict) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.words = make(map[string]*WordIdentity)
}
