// Package structs contains all struct definitions from structs.h (lines 1-73)
package structs

// CharDoubleton represents struct char_doubleton (lines 11-14)
type CharDoubleton struct {
	First  byte // c
	Second int  // freq (must be int, not byte, to avoid overflow!)
}

// Word represents struct word_t (lines 16-20)
type Word struct {
	Len  int    // i (unsigned int)
	Freq int    // freq
	Str  []byte // word (char*)
}

// WordArray represents struct word_t_array (lines 24-27)
type WordArray struct {
	Elems []*Word // elems
	Count int     // count (unsigned int)
}

// NgramHalfArray represents struct ngram_half_array (lines 29-31)
type NgramHalfArray struct {
	At [5]*WordArray // at[NGRAM_SIZE/2+1] where NGRAM_SIZE=9, so 9/2+1=5
}

// Ngram represents struct ngram_t (lines 33-39)
type Ngram struct {
	Word      Word             // word
	Before    NgramHalfArray   // before
	After     NgramHalfArray   // after
	RefsCount int              // refs_count
	Refs      []*Ngram         // refs (struct ngram_t**)
}

// Morpheme represents struct morpheme_t (lines 45-58)
type Morpheme struct {
	Regex               string   // regex
	Morpheme            string   // morpheme
	Freq                int      // freq
	WordsCount          int      // words_count
	Words               []*Word  // words
	Type                int      // type
	FrontRegex          string   // front_regex
	FrontRegexArr       []string // front_regex_arr (char**)
	BackRegex           string   // back_regex
	BackRegexArr        []string // back_regex_arr (char**)
	FrontRegexArrIndex  int      // front_regex_arr_index
	BackRegexArrIndex   int      // back_regex_arr_index
}

// MorphemeList represents struct morpheme_list_t (lines 60-63)
type MorphemeList struct {
	List  []*Morpheme // list
	Count int         // count
}

// LexicalCategories represents struct lexical_categories_t (lines 65-69)
type LexicalCategories struct {
	Morpheme *Morpheme // morpheme
	RefCount int       // ref_count
	Refs     []*Ngram  // refs (struct ngram_t**)
}

