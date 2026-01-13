package dictionary

// ToLowerASCII converts ASCII string to lowercase using fast bit operations
// For ASCII: uppercase 'A'-'Z' (0x41-0x5A) → lowercase 'a'-'z' (0x61-0x7A)
// Difference is bit 5 (0x20), so we set it: ch | 0x20
// Non-ASCII (UTF-8) falls back to slower byte-by-byte conversion
func ToLowerASCII(s string) string {
	// Fast path: Check if string is pure ASCII
	hasNonASCII := false
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			hasNonASCII = true
			break
		}
	}

	// Pure ASCII: fast bit manipulation
	if !hasNonASCII {
		b := []byte(s)
		for i := 0; i < len(b); i++ {
			// If uppercase (0x41-0x5A), set bit 5 to make lowercase
			if b[i] >= 'A' && b[i] <= 'Z' {
				b[i] |= 0x20 // Set bit 5: 'A' (0x41) → 'a' (0x61)
			}
		}
		return string(b)
	}

	// UTF-8: slower byte-by-byte with rune support
	// For simplicity, only convert ASCII portion (common in most text)
	// Full Unicode lowercase would require unicode.ToLower (slower)
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] |= 0x20
		}
		// Non-ASCII bytes pass through unchanged
	}
	return string(b)
}

// isPunctuation checks if a byte is common punctuation
func isPunctuation(ch byte) bool {
	return ch == '.' || ch == ',' || ch == '!' || ch == '?' ||
		ch == ';' || ch == ':' || ch == '\'' || ch == '"'
}

// Normalize prepares a word for dictionary interning
// - Converts to lowercase (case-insensitive matching)
// - Strips leading punctuation (handles "The, 1st,, etc.)
// - Strips trailing punctuation (.,!?;:'")
func Normalize(s string) string {
	// Convert to lowercase
	s = ToLowerASCII(s)

	// Strip leading punctuation
	for len(s) > 0 {
		if isPunctuation(s[0]) {
			s = s[1:]
		} else {
			break
		}
	}

	// Strip trailing punctuation
	for len(s) > 0 {
		if isPunctuation(s[len(s)-1]) {
			s = s[:len(s)-1]
		} else {
			break
		}
	}

	return s
}
