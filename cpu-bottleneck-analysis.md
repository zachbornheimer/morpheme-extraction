# CPU Profiling Analysis - Moby Dick (50KB) + War of the Worlds

**Total Runtime:** 59.64s  
**Total CPU Time:** 68.56s

## Top Bottlenecks

### 1. UniqIdentities - 21.70s (31.65%)
**Function:** `internal/functions.UniqIdentities`  
**Issue:** O(n²) deduplication algorithm

**Hot code (lines 109-110):**
```go
for j := 0; j < uniqueCount; j++ {
    if arr[j] == arr[i] {  // Takes 21.35s combined
        seen = true
        break
    }
}
```

**Why it's slow:**
- Called frequently during ngram similarity checking
- O(n²) algorithm: for each element, scans all previous unique elements
- Inner loop executed millions of times
- With 3,095 unique words, typical ngram contexts of ~20-50 words

**Optimization:** Use a hash set (map) for O(1) lookups instead of O(n) linear scan

### 2. Runtime Memory Management - 8.98s (13.10%)
**Function:** `runtime.madvise`  
**Issue:** Memory allocation overhead from GC and heap management

**Why it's slow:**
- Creating many small allocations (strings, slices, Word objects)
- GC pressure from temporary objects

**Optimization:** 
- Pool reusable objects (sync.Pool)
- Reduce allocations in hot paths

### 3. String/Rune Conversions - 5.60s (8.17%)
**Functions:**
- `runtime.stringtoslicerune`: 4.09s
- `runtime.slicerunetostring`: 1.51s

**Why it's slow:**
- Converting between strings and []rune for character manipulation
- Each conversion allocates new memory

**Where:** Likely in `RemoveFirstChar`, `Reverse`, and morpheme comparison

### 4. FindInternalMorphemes - 18.73s cumulative (27.32%)
**Function:** `internal/morpheme_list_t.FindInternalMorphemes`  
**Issue:** Nested loops with O(n²) character comparisons

**Breakdown:**
- Calls `Word.RemoveFirstChar` repeatedly: 4.81s cumulative
- Calls `Word.SetWord` repeatedly: 9.31s cumulative
- String allocations and manipulations

### 5. NgramsSimilar - 24.71s cumulative (36.04%)
**Function:** `internal/ngram_t.NgramsSimilar`  
**Issue:** Called for every pair of ngrams (O(V²) where V = vocabulary size)

**Stats from run:**
- Moby Dick: 3,360,528 pairs checked, 1,753,438 similar (52.2%)
- War of the Worlds: 4,763,241 pairs checked, 2,576,781 similar (54.1%)
- **Total: 8,123,769 pair comparisons**

Most time spent in:
- `UniqIdentities` for deduplicating context words
- String comparisons

## Summary of Bottlenecks

| Function | Time | % | Type | Fix Difficulty |
|----------|------|---|------|----------------|
| UniqIdentities | 21.70s | 31.65% | Algorithm O(n²) | Easy (use map) |
| Memory Management | 8.98s | 13.10% | Allocations | Medium (pooling) |
| String↔Rune | 5.60s | 8.17% | Type conversions | Medium (optimize) |
| FindInternalMorphemes | 18.73s | 27.32% | Nested loops | Hard (algorithmic) |
| NgramsSimilar | 24.71s | 36.04% | O(V²) pairs | Hard (algorithmic) |

## Recommended Optimizations (Priority Order)

### High Priority (Easy wins):

1. **Replace UniqIdentities O(n²) with map-based O(n)**
   - Expected savings: ~20s
   - Implementation: Use `map[*WordIdentity]struct{}` for seen tracking
   
2. **Optimize normalization (ToLowerASCII)**
   - Currently: 3.02s
   - Cache normalized strings in WordIdentity

### Medium Priority:

3. **Reduce string allocations in RemoveFirstChar/SetWord**
   - Use string slicing instead of allocations where possible
   - Pool temporary buffers

4. **Add object pooling (sync.Pool)**
   - Pool Word objects
   - Pool temporary slices

### Lower Priority (Bigger changes):

5. **Optimize FindInternalMorphemes algorithm**
   - Current: nested loops with repeated string operations
   - Consider suffix array or other string matching structures

6. **Optimize NgramsSimilar**
   - Current: O(V²) all-pairs comparison
   - Consider locality-sensitive hashing or other approximate methods

## Expected Results After Optimization

Conservative estimates:
- Fix UniqIdentities: 59s → ~38s (-35%)
- Add pooling + reduce allocations: 38s → ~30s (-20%)
- **Total speedup: ~50% faster**

This would make full Moby Dick (1.1MB) feasible within reasonable time.
