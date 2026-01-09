package morpheme_t

import (
	"testing"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

func TestFindLongestMatch(t *testing.T) {
	tests := []struct {
		name     string
		word1    string
		word2    string
		expected string
	}{
		{
			name:     "Common prefix 'hel'",
			word1:    "hello",
			word2:    "help",
			expected: "hel",
		},
		{
			name:     "No common prefix",
			word1:    "abc",
			word2:    "xyz",
			expected: "",
		},
		{
			name:     "One word is prefix of other",
			word1:    "test",
			word2:    "testing",
			expected: "test",
		},
		{
			name:     "Identical words",
			word1:    "same",
			word2:    "same",
			expected: "same",
		},
		{
			name:     "Single char match",
			word1:    "apple",
			word2:    "about",
			expected: "a",
		},
		{
			name:     "Empty strings",
			word1:    "",
			word2:    "",
			expected: "",
		},
		{
			name:     "One empty",
			word1:    "hello",
			word2:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w1 := structs.Word{Str: tt.word1, Freq: 1}
			w2 := structs.Word{Str: tt.word2, Freq: 1}
			
			result := FindLongestMatch(w1, w2)
			
			if result.Morpheme != tt.expected {
				t.Errorf("FindLongestMatch(%q, %q) = %q, want %q",
					tt.word1, tt.word2, result.Morpheme, tt.expected)
			}
		})
	}
}

