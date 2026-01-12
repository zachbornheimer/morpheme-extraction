package dictionary

// ToUpperASCII converts ASCII string to uppercase using fast bit operations
// For ASCII: lowercase 'a'-'z' (0x61-0x7A) → uppercase 'A'-'Z' (0x41-0x5A)
// Difference is bit 5 (0x20), so we clear it: ch & 0xDF
// Non-ASCII (UTF-8) falls back to slower byte-by-byte conversion
func ToUpperASCII(s string) string {
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
			// If lowercase (0x61-0x7A), clear bit 5 to make uppercase
			if b[i] >= 'a' && b[i] <= 'z' {
				b[i] &= 0xDF // Clear bit 5: 'a' (0x61) → 'A' (0x41)
			}
		}
		return string(b)
	}

	// UTF-8: slower byte-by-byte with rune support
	// For simplicity, only convert ASCII portion (common in most text)
	// Full Unicode uppercase would require unicode.ToUpper (slower)
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] &= 0xDF
		}
		// Non-ASCII bytes pass through unchanged
	}
	return string(b)
}

// Normalize prepares a word for dictionary interning
// - Converts to uppercase (case-insensitive matching)
// - Future: Could add other normalizations (strip diacritics, etc.)
func Normalize(s string) string {
	return ToUpperASCII(s)
}
