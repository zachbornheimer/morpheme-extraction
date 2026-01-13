// Package minhash provides locality-sensitive hashing for fast similarity estimation
// Used to filter word pairs before expensive NgramsSimilar checks
// MinHash estimates Jaccard similarity of context word sets
package minhash

import (
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// NumHashes is the number of hash functions used for MinHash signatures
// More hashes = more accurate but slower
// 128 provides good balance: ~1% error rate for similarity estimation
const NumHashes = 128

// MinHashSignature represents a MinHash signature for an ngram's context
// Each element is the minimum hash value across all context words for that hash function
type MinHashSignature [NumHashes]uint64

// fnv1aHash computes FNV-1a hash of a value with a seed
// FNV-1a is fast and provides good distribution for MinHash
// Different seeds simulate different hash functions
func fnv1aHash(data uint64, seed uint64) uint64 {
	// FNV-1a offset basis
	hash := uint64(14695981039346656037) ^ seed

	// Convert uint64 to bytes
	bytes := [8]byte{
		byte(data),
		byte(data >> 8),
		byte(data >> 16),
		byte(data >> 24),
		byte(data >> 32),
		byte(data >> 40),
		byte(data >> 48),
		byte(data >> 56),
	}

	// FNV-1a algorithm: XOR then multiply
	for _, b := range bytes {
		hash ^= uint64(b)
		hash *= 1099511628211 // FNV-1a prime
	}

	return hash
}

// ComputeMinHash computes MinHash signature from context word IDs
// Uses k-permutation MinHash with k=128 hash functions
func ComputeMinHash(contextIDs []int) MinHashSignature {
	var sig MinHashSignature

	// Initialize all hash values to max (empty set)
	for i := range sig {
		sig[i] = ^uint64(0) // Max uint64
	}

	// For each context word ID, compute k hash values
	// Keep minimum for each hash function (MinHash property)
	for _, id := range contextIDs {
		if id <= 0 {
			continue // Skip invalid IDs
		}

		// Apply k hash functions (different seeds)
		for h := 0; h < NumHashes; h++ {
			hash := fnv1aHash(uint64(id), uint64(h))
			if hash < sig[h] {
				sig[h] = hash // Keep minimum
			}
		}
	}

	return sig
}

// EstimateSimilarity estimates Jaccard similarity between two MinHash signatures
// Returns percentage (0-100) of matching hash values
// This approximates the Jaccard similarity of the original context sets
func EstimateSimilarity(sig1, sig2 MinHashSignature) float64 {
	matches := 0
	for i := 0; i < NumHashes; i++ {
		if sig1[i] == sig2[i] {
			matches++
		}
	}
	return 100.0 * float64(matches) / float64(NumHashes)
}

// ExtractContextIDs extracts all context word IDs from an ngram
// Includes both Before and After context words
func ExtractContextIDs(ng *structs.Ngram, ngramHalfSize int) []int {
	// Preallocate for typical context size
	ids := make([]int, 0, ngramHalfSize*2*4) // ~4 words per position on average

	// Collect Before context word IDs
	for i := 0; i < ngramHalfSize; i++ {
		if ng.Before.At[i] != nil {
			for j := 0; j < ng.Before.At[i].Count; j++ {
				if ng.Before.At[i].Elems[j] != nil && ng.Before.At[i].Elems[j].Identity != nil {
					ids = append(ids, ng.Before.At[i].Elems[j].Identity.ID)
				}
			}
		}
	}

	// Collect After context word IDs
	for i := 0; i < ngramHalfSize; i++ {
		if ng.After.At[i] != nil {
			for j := 0; j < ng.After.At[i].Count; j++ {
				if ng.After.At[i].Elems[j] != nil && ng.After.At[i].Elems[j].Identity != nil {
					ids = append(ids, ng.After.At[i].Elems[j].Identity.ID)
				}
			}
		}
	}

	return ids
}
