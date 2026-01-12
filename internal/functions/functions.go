// Package functions contains utility functions from functions.c
// Line-for-line translation of functions.c
package functions

import (
	"strings"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/dictionary"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// VPrint prints only if verbose mode is ON
var VerboseMode int

func VPrint(msg string) {
	if VerboseMode == constants.ON {
		println(msg)
	}
}

// Append concatenates two strings (lines 21-29)
func Append(one, two string) string {
	return one + two
}

// InArray checks if byte c is in uniq slice (lines 31-38)
func InArray(c byte, uniq []byte, size int) int {
	for i := 0; i < size; i++ {
		if c == uniq[i] {
			return i
		}
	}
	return -1
}

// InCharArray checks if string w is in uniq slice (lines 40-47)
func InCharArray(w string, uniq []string, size int) int {
	for i := 0; i < size; i++ {
		if w == uniq[i] {
			return i
		}
	}
	return -1
}

// Uniq returns unique characters from input string (lines 52-69)
// Returns unique string and count (base 0)
func Uniq(chars string) (string, int) {
	if chars == "" {
		return "", -1
	}
	
	length := len(chars)
	u := make([]byte, length+2)
	i := 0
	
	for j := 0; j <= length; j++ {
		if j < length && chars[j] != 0 {
			if i == 0 || InArray(chars[j], u, i) == -1 {
				u[i] = chars[j]
				i++
			}
		} else {
			break
		}
	}
	
	return string(u[:i]), i - 1
}

// UniqWords returns count of unique words in array (lines 72-84)
// Returns base 0
// Optimized: Uses map for larger arrays (O(n)), linear search for small (less overhead)
func UniqWords(arr []string, size int) int {
	// For small arrays, linear search has less overhead than map
	const threshold = 50
	if size <= threshold {
		u := make([]string, size)
		i := 0
		for j := 0; j < size; j++ {
			if i == 0 || InCharArray(arr[j], u, i) == -1 {
				u[i] = arr[j]
				i++
			}
		}
		return i - 1
	}

	// For larger arrays, use map for O(n) performance
	seen := structs.NewStringSet(size)
	for j := 0; j < size; j++ {
		seen.Add(arr[j])
	}
	return seen.Len() - 1
}

// UniqIdentities returns count of unique Identity pointers in array
// Returns base 0
// Uses O(1) pointer equality checks (very fast for small arrays)
func UniqIdentities(arr []*dictionary.WordIdentity, size int) int {
	if size == 0 {
		return -1
	}

	// Simple slice-based deduplication with pointer comparison
	// Fast for typical ngram context sizes (< 50 words)
	// O(n²) but pointer comparison is extremely fast (single instruction)
	uniqueCount := 0
	for i := 0; i < size; i++ {
		if arr[i] == nil {
			continue
		}

		// Check if this pointer already seen
		seen := false
		for j := 0; j < uniqueCount; j++ {
			if arr[j] == arr[i] {  // Pointer equality: single CPU cycle
				seen = true
				break
			}
		}

		if !seen {
			// Swap to front of "unique" section
			arr[uniqueCount], arr[i] = arr[i], arr[uniqueCount]
			uniqueCount++
		}
	}

	return uniqueCount - 1
}

// ExplodeSansnullStr splits string by delimiter string (lines 88-99)
func ExplodeSansnullStr(str string, delimiter string) []string {
	// C version has a weird loop that calls explode_sansnull len(delimiter) times
	// with the same arguments. This is inefficient; we just call it once.
	return ExplodeSansnull(str, delimiter)
}

// ExplodeSansnull splits string by delimiter (lines 107-136)
// Translated from public domain code modified by Z. Bornheimer
func ExplodeSansnull(str string, delimiter string) []string {
	if str == "" || delimiter == "" {
		return []string{}
	}
	
	// Count delimiter occurrences
	count := 0
	for _, ch := range str {
		for _, delim := range delimiter {
			if ch == delim {
				count++
			}
		}
	}
	
	wc := count + 1
	arr := make([]string, 0, wc)
	
	// Split by delimiter
	parts := strings.FieldsFunc(str, func(r rune) bool {
		for _, delim := range delimiter {
			if r == delim {
				return true
			}
		}
		return false
	})
	
	// Filter out empty strings
	for _, part := range parts {
		if part != "" {
			arr = append(arr, part)
		}
	}
	
	return arr
}

// MoveChar moves character at index to next position (lines 138-153)
func MoveChar(index *int, in *string) int {
	s := []byte(*in)
	length := len(s)
	
	if s[*index+1] == 0 || *index+1 >= length {
		*index = 0
	}
	
	f := *index
	temp := make([]byte, length+1)
	
	// Copy before index
	for j := 0; j < f; j++ {
		temp[j] = s[j]
	}
	
	// Swap characters at f and f+1
	if f+1 < length {
		temp[f] = s[f+1]
		temp[f+1] = s[f]
		
		// Copy rest
		for j := f + 2; j < length; j++ {
			temp[j] = s[j]
		}
	}
	
	*in = string(temp[:length])
	(*index)++
	return 1
}

// Permute permutes the string (lines 155-164)
func Permute(str *string, i *int) string {
	if *i >= len(*str) {
		*i = 0
	}
	
	if MoveChar(i, str) == 0 {
		println("Could not move char.")
	}
	
	return *str
}

// Reverse reverses a string in place (lines 166-184)
func Reverse(str string) string {
	if len(str) < 2 {
		return str
	}

	s := []byte(str)
	length := len(s) - 1
	i := 0

	for i < (length - i) {
		temp := s[i]
		s[i] = s[length-i]
		s[length-i] = temp
		i++
	}

	return string(s)
}

// ReverseBytes reverses a byte slice in place and returns it
// Avoids string allocations - works directly with bytes
func ReverseBytes(b []byte) []byte {
	if len(b) < 2 {
		return b
	}

	length := len(b) - 1
	for i := 0; i < length/2; i++ {
		b[i], b[length-i] = b[length-i], b[i]
	}
	return b
}

