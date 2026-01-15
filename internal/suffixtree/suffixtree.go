// Package suffixtree provides a generalized suffix tree for finding common substrings
package suffixtree

import (
	"strings"
)

// SuffixTree represents a generalized suffix tree for finding common substrings between words
type SuffixTree struct {
	root *Node
}

// Node represents a node in the suffix tree
type Node struct {
	children  map[byte]*Node
	positions map[int][]Position // wordID -> list of positions where this substring appears
	depth     int                 // String depth (length of path from root)
}

// Position tracks where a substring appears in a word
type Position struct {
	StartPos int // Position in original word (chars stripped from start)
}

// CommonSubstring represents a substring found in both words
type CommonSubstring struct {
	Substring string
	Pos1      int // Position in word1 (chars stripped from start)
	Pos2      int // Position in word2 (chars stripped from start)
	Length    int // Length of the common substring
}

// NewSuffixTree creates a new empty suffix tree
func NewSuffixTree() *SuffixTree {
	return &SuffixTree{
		root: &Node{
			children:  make(map[byte]*Node),
			positions: make(map[int][]Position),
			depth:     0,
		},
	}
}

// AddWord adds all suffixes of a word to the tree
// wordID identifies which word (0 or 1)
func (st *SuffixTree) AddWord(word string, wordID int) {
	// Add all suffixes: for "run" add "run", "un", "n"
	for startPos := 0; startPos < len(word); startPos++ {
		suffix := word[startPos:]
		st.addSuffix(suffix, wordID, startPos)
	}
}

// addSuffix inserts a suffix into the tree
func (st *SuffixTree) addSuffix(suffix string, wordID int, startPos int) {
	node := st.root

	for i := 0; i < len(suffix); i++ {
		char := suffix[i]

		// Create child if it doesn't exist
		if node.children[char] == nil {
			node.children[char] = &Node{
				children:  make(map[byte]*Node),
				positions: make(map[int][]Position),
				depth:     node.depth + 1,
			}
		}

		node = node.children[char]

		// Track this position for this word
		node.positions[wordID] = append(node.positions[wordID], Position{
			StartPos: startPos,
		})
	}
}

// GetCommonSubstrings returns all substrings that appear in both words
// minLen is the minimum substring length to consider
func (st *SuffixTree) GetCommonSubstrings(minLen int) []CommonSubstring {
	var results []CommonSubstring
	var path []byte
	st.traverse(st.root, path, &results, minLen)
	return results
}

// traverse recursively walks the tree to find common substrings
func (st *SuffixTree) traverse(node *Node, path []byte, results *[]CommonSubstring, minLen int) {
	// If this node has positions from both word 0 and word 1, it's a common substring
	_, hasWord0 := node.positions[0]
	_, hasWord1 := node.positions[1]

	if hasWord0 && hasWord1 && len(path) >= minLen {
		// Found a common substring!
		substring := string(path)

		// Generate all combinations of positions from word1 and word2
		// This matches the original algorithm's behavior of checking all position pairs
		for _, pos1 := range node.positions[0] {
			for _, pos2 := range node.positions[1] {
				*results = append(*results, CommonSubstring{
					Substring: substring,
					Pos1:      pos1.StartPos,
					Pos2:      pos2.StartPos,
					Length:    len(substring),
				})
			}
		}
	}

	// Recurse into children
	for char, child := range node.children {
		newPath := append(path, char)
		st.traverse(child, newPath, results, minLen)
	}
}

// FindLongestCommonPrefix finds the longest common prefix between two strings
// This is used for the forward morpheme matching
func FindLongestCommonPrefix(word1, word2 string) string {
	minLen := len(word1)
	if len(word2) < minLen {
		minLen = len(word2)
	}

	for i := 0; i < minLen; i++ {
		if word1[i] != word2[i] {
			return word1[:i]
		}
	}

	return word1[:minLen]
}

// FindCommonSubstringsForPair is a convenience function that builds a suffix tree
// for two words and returns all common substrings
func FindCommonSubstringsForPair(word1, word2 string, minLen int) []CommonSubstring {
	tree := NewSuffixTree()
	tree.AddWord(word1, 0)
	tree.AddWord(word2, 1)

	substrings := tree.GetCommonSubstrings(minLen)

	// Sort by position1, then position2 for deterministic order
	// (matches original algorithm's sequential discovery order)
	sortCommonSubstrings(substrings)

	return substrings
}

// sortCommonSubstrings sorts by pos1, then pos2 to match original ordering
func sortCommonSubstrings(subs []CommonSubstring) {
	// Simple bubble sort for now - will optimize if needed
	for i := 0; i < len(subs); i++ {
		for j := i + 1; j < len(subs); j++ {
			if subs[i].Pos1 > subs[j].Pos1 ||
				(subs[i].Pos1 == subs[j].Pos1 && subs[i].Pos2 > subs[j].Pos2) {
				subs[i], subs[j] = subs[j], subs[i]
			}
		}
	}
}

// DeduplicateSubstrings removes duplicate substrings based on the lookback logic
// This mimics the original algorithm's deduplication behavior
func DeduplicateSubstrings(subs []CommonSubstring) []CommonSubstring {
	if len(subs) == 0 {
		return subs
	}

	result := make([]CommonSubstring, 0, len(subs))

	for i := 0; i < len(subs); i++ {
		// Check if this substring is a prefix of the previous one
		if len(result) > 0 {
			prev := result[len(result)-1].Substring
			curr := subs[i].Substring

			// If current is a prefix of previous, skip it
			if strings.HasPrefix(prev, curr) {
				continue
			}
		}

		result = append(result, subs[i])
	}

	return result
}
