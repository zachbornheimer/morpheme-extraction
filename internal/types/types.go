// Package types defines the core data structures for morpheme extraction
// based on Bornheimer's Linguistic Fusion theory (2013-2014).
package types

// MorphemeType represents the classification of a morpheme.
type MorphemeType int

const (
	Undefined MorphemeType = iota
	Prefix
	Suffix
	Circumfix
	Infix
	Stem
)

func (mt MorphemeType) String() string {
	switch mt {
	case Prefix:
		return "PREFIX"
	case Suffix:
		return "SUFFIX"
	case Circumfix:
		return "CIRCUMFIX"
	case Infix:
		return "INFIX"
	case Stem:
		return "STEM"
	default:
		return "UNDEFINED"
	}
}

// Word represents a word with its frequency and length.
type Word struct {
	Text      string
	Length    int
	Frequency int
}

// Ngram represents an n-gram with context windows before and after the target word.
// This implements the data structure from nlp.c:build_ngram_relationships
type Ngram struct {
	Word   Word
	Before [][]Word // NGRAM_SIZE/2 arrays of words before target
	After  [][]Word // NGRAM_SIZE/2 arrays of words after target
	Refs   []*Ngram // Similar n-grams (parallel environments)
}

// Morpheme represents an extracted morpheme with its pattern and classification.
// Based on the morpheme_t structure from the original C implementation.
type Morpheme struct {
	Text          string       // The morpheme string
	Type          MorphemeType // PREFIX, SUFFIX, STEM, etc.
	Frequency     int          // Occurrence count
	Words         []Word       // Words containing this morpheme
	Pattern       string       // Regex pattern for this morpheme
	FrontPattern  string       // Prefix variation pattern
	BackPattern   string       // Suffix variation pattern
	FrontVariants []string     // Character alternatives before morpheme
	BackVariants  []string     // Character alternatives after morpheme
}

// LanguageModel represents the complete extracted linguistic model.
// This is the top-level structure that contains all extracted information.
type LanguageModel struct {
	Delimiter  string      // Word delimiter (from alphabet.c)
	Morphemes  []Morpheme  // Extracted morphemes
	Vocabulary map[string]int // Word frequency map
	NgramSize  int         // N-gram window size (default 9)
}

