# FindMorphemes Detailed Breakdown

**Function**: `FindMorphemes` in `internal/nlp/nlp.go:395`
**Total Time**: 28.06s (65.2% of total runtime with Green Tea GC)
**Pure Algorithm**: ~7s (from GOGC=off test)
**GC Overhead**: ~21s (75% of function time!)

---

## Part 1: Line-by-Line Time Breakdown

### With Green Tea GC (Production)

```
Line    Flat     Cum      Code                                              Notes
--------------------------------------------------------------------------------
399     —        —        for i := 0; i < ngramLength; i++                   Outer loop (2,592 iterations)
414     750ms    750ms    neighborIDs := node.GetNeighbors()                 Allocates slice per iteration
415     50ms     50ms     for _, neighborID := range neighborIDs             Neighbor loop (~2,596,286 pairs)
418     2.35s    2.38s    for j := i + constants.NgramSize; j < ngramLength  Inner loop (hot!)
419     5.08s    5.16s    if ng[j] == nil || ng[j].Word.Identity == nil      Nil checks (expensive!)
424     1.53s    1.57s    if ng[j].Word.Identity.ID != neighborID            ID comparison
428     10ms     10ms     target := *ng[j]                                   Struct copy
429     —        —        w2 := target.Word.StringClean()                    String allocation
433     170ms    170ms    forwardStr := FindLongestMatchString(...)          String matching
434     4.59s    4.77s    candidates = append(candidates, ...)               Slice growth (GC trigger!)
445     —        13.90s   FindInternalMorphemes(...)                         Heaviest call (30.3% total)
--------------------------------------------------------------------------------
Total:  14.49s   28.06s   (flat vs cumulative)
```

### Without GC (GOGC=off - Pure Algorithm Cost)

```
Line    Flat     Cum      Code                                              Change from GC
--------------------------------------------------------------------------------
414     90ms     90ms     neighborIDs := node.GetNeighbors()                 -660ms (88% faster)
415     10ms     10ms     for _, neighborID := range neighborIDs             -40ms (80% faster)
418     400ms    400ms    for j := i + constants.NgramSize; j < ngramLength  -1.95s (83% faster)
419     640ms    650ms    if ng[j] == nil || ng[j].Word.Identity == nil      -4.44s (87% faster!)
424     220ms    220ms    if ng[j].Word.Identity.ID != neighborID            -1.31s (86% faster)
434     580ms    750ms    candidates = append(candidates, ...)               -3.81s (83% faster!)
445     —        2.33s    FindInternalMorphemes(...)                         -11.57s (83% faster!)
--------------------------------------------------------------------------------
Total:  1.94s    4.45s    (pure algorithm cost)
```

**Key Insight**: Every line is 5-8x slower with GC due to allocation pressure!

---

## Part 2: Algorithm Structure

### Triple Nested Loop

```go
// OUTER LOOP: For each ngram (i)
for i := 0; i < ngramLength; i++ {                    // 2,592 iterations (Moby Dick)
    ngram := *ng[i]
    neighborIDs := node.GetNeighbors()                // Allocates slice

    // MIDDLE LOOP: For each neighbor word
    for _, neighborID := range neighborIDs {          // ~1,000 neighbors per ngram average

        // INNER LOOP: Find all ngrams with that neighbor word
        for j := i + constants.NgramSize; j < ngramLength; j++ {  // ~2,000 iterations
            if ng[j].Word.Identity.ID != neighborID {
                continue                              // Skip non-matches
            }

            // Found a match - extract morphemes
            forwardStr := FindLongestMatchString(...)  // Allocate string
            candidates = append(candidates, ...)       // Allocate struct, grow slice
            FindInternalMorphemes(...)                 // Heavy nested processing
        }
    }
}
```

### Iteration Counts

**Moby Dick file** (2,592 ngrams):
- Outer loop: 2,592 iterations
- Neighbor loop: ~1,000 avg neighbors × 2,592 = **2.6M iterations**
- Inner loop: ~2,000 avg × 2.6M = **5.2B checks** (most filtered by ID check)
- Actual morpheme pairs: ~2.6M (from 1.75M similar pairs)

---

## Part 3: Memory Allocation Hotspots

### Major Allocators (from memory profile)

```
Operation                           Allocation      Frequency               Total Impact
--------------------------------------------------------------------------------------
NewMorpheme (in GenRegex)           266 MB (22%)    ~2.6M calls            High - GC thrashing
candidates slice growth             247 MB (20%)    ~2.6M appends          High - reallocations
neighborIDs slice                   ~50 MB          2,592 allocations      Medium
String allocations                  ~100 MB         Millions               High - GC trigger
Word struct copies                  ~80 MB          ~5M copies             Medium
--------------------------------------------------------------------------------------
Total:                              ~743 MB         Billions of ops        CRITICAL
```

### GC Trigger Points

1. **Line 434: `candidates = append(...)`**
   - Allocates `MorphemeCandidate` struct
   - Triggers slice reallocation when capacity exceeded
   - Called ~2.6M times
   - **4.59s with GC → 580ms without (8x slower!)**

2. **Line 445: `FindInternalMorphemes(...)`**
   - Calls `NewMorpheme()` for each internal morpheme found
   - Creates temporary Word structs for regex generation
   - String manipulations (Reverse, SetWord)
   - **13.90s with GC → 2.33s without (6x slower!)**

3. **Line 414: `neighborIDs := node.GetNeighbors()`**
   - Allocates new slice on each outer loop iteration
   - 2,592 allocations
   - **750ms with GC → 90ms without (8x slower!)**

---

## Part 4: FindInternalMorphemes Deep Dive

**Time**: 13.90s (30.3% of total, 49.5% of FindMorphemes)
**Pure algorithm**: 2.33s (without GC)
**GC overhead**: 11.57s (83% of function time!)

### What It Does

Finds morphemes **inside** words (infixes):
```
"walking" vs "talked"
  -> Extract "alk" at position 1,1
```

### Algorithm

```go
func FindInternalMorphemes(word1, word2 Word, candidates *[]MorphemeCandidate) {
    // Double nested loop - trim characters from start
    for word1.Len > 0 {                              // Outer: trim word1
        for word2.Len > 0 {                          // Inner: trim word2
            word2 = word2.RemoveFirstChar()          // Allocates string!

            morphemeStr := FindLongestMatchString(word1, word2)

            if len(morphemeStr) >= 2 {
                // Check against previous morpheme
                if lastAddedIdx >= 0 {
                    prevMorpheme := candidates[lastAddedIdx].Morpheme

                    // Reverse previous morpheme
                    test := Reverse(prevMorpheme)    // Allocates string!

                    // Trim lookback
                    if shouldTrim {
                        test = test[:len(test)-lookback]  // Allocates string!
                    }

                    // Reverse again
                    test = Reverse(test)             // Allocates string!

                    if test != morphemeStr {
                        // Create temporary morpheme for regex check
                        m := NewMorpheme(morphemeStr)     // MAJOR ALLOCATION!
                        m.GenRegex(word1, word2, ...)     // Complex regex generation

                        if m.Regex() != "" {
                            candidates = append(...)      // Add if valid
                        }
                    }
                }
            }
        }
        word1 = word1.RemoveFirstChar()              // Allocates string!
    }
}
```

### Allocation Points in FindInternalMorphemes

```
Operation                       Per Call    Calls           Total       GC Impact
----------------------------------------------------------------------------------
RemoveFirstChar()               ~100 B      Millions        ~100 MB     High
Reverse() (2x per check)        ~100 B      Millions        ~200 MB     High
NewMorpheme()                   ~1 KB       ~437,976        ~430 MB     CRITICAL
GenRegex()                      ~500 B      ~437,976        ~215 MB     High
String allocations              Various     Millions        ~400 MB     High
----------------------------------------------------------------------------------
Total:                                                      ~1.3 GB     CRITICAL
```

**Key Problem**: Every iteration allocates multiple strings + Morpheme objects!

---

## Part 5: Optimization Opportunities

### High Impact (10-15s potential savings)

#### 1. Object Pool for MorphemeCandidate (Priority 1)

**Problem**:
```go
candidates = append(candidates, structs.MorphemeCandidate{...})  // Allocates 2.6M times
```

**Solution**:
```go
var candidatePool = sync.Pool{
    New: func() interface{} {
        return &structs.MorphemeCandidate{}
    },
}

// Usage
cand := candidatePool.Get().(*structs.MorphemeCandidate)
cand.Morpheme = morphemeStr
cand.Word1 = w1
// ... set fields
candidates = append(candidates, cand)

// After use
defer candidatePool.Put(cand)
```

**Expected**: 4.59s → ~1s (75% reduction in line 434)

#### 2. Reuse neighborIDs Slice (Priority 2)

**Problem**:
```go
neighborIDs := node.GetNeighbors()  // Allocates 2,592 times
```

**Solution**:
```go
// Pre-allocate reusable buffer
neighborBuf := make([]int, 0, 1000)

for i := 0; i < ngramLength; i++ {
    // Reuse buffer instead of allocating
    neighborBuf = node.GetNeighborsInto(neighborBuf[:0])
    // ...
}
```

**Expected**: 750ms → ~100ms (87% reduction)

#### 3. Byte Slices Instead of Strings in FindInternalMorphemes (Priority 3)

**Problem**:
```go
test := Reverse(prevMorpheme)        // Allocates
test = test[:len(test)-lookback]    // Allocates
test = Reverse(test)                 // Allocates
```

**Solution**:
```go
// Use []byte buffer, modify in-place
testBuf := []byte(prevMorpheme)
ReverseBytes(testBuf)               // In-place, no allocation
if shouldTrim {
    testBuf = testBuf[:len(testBuf)-lookback]  // Slice, no allocation
}
ReverseBytes(testBuf)               // In-place, no allocation
```

**Expected**: 11.57s GC overhead → ~3s (74% reduction in FindInternalMorphemes)

### Medium Impact (2-5s potential savings)

#### 4. Preallocate Candidates Slice (Priority 4)

**Problem**:
```go
candidates := make([]structs.MorphemeCandidate, 0, ngramLength*100)  // Under-allocated
// Actual: ~2.6M candidates, capacity ~260K → many reallocations
```

**Solution**:
```go
// Profile actual max size first
candidates := make([]structs.MorphemeCandidate, 0, 3000000)  // 3M capacity
```

**Expected**: 1-2s savings from avoiding reallocations

#### 5. Avoid Struct Copies (Priority 5)

**Problem**:
```go
ngram := *ng[i]           // Copies entire Ngram struct
target := *ng[j]          // Copies entire Ngram struct
```

**Solution**:
```go
ngram := ng[i]            // Use pointer directly
target := ng[j]           // Use pointer directly
```

**Expected**: 500ms-1s savings

### Low Impact (<1s potential savings)

#### 6. Cache StringClean() Results

**Problem**:
```go
w1 := ngram.Word.StringClean()    // Called once per outer loop
w2 := target.Word.StringClean()   // Called millions of times
```

**Solution**: Cache in outer loop, pass to inner loops

**Expected**: 200-500ms savings

---

## Part 6: Priority Roadmap

### Phase 1: Quick Wins (2-3 hours work, 5-8s savings)

1. ✅ Enable Green Tea GC (done - 1.9s savings)
2. Reuse neighborIDs slice (simple - 650ms savings)
3. Preallocate candidates slice (simple - 1-2s savings)
4. Avoid struct copies (simple - 500ms-1s savings)

**Subtotal**: 3.0-4.5s savings → **35s runtime**

### Phase 2: Memory Optimization (1-2 days work, 10-12s savings)

5. Object pooling for MorphemeCandidate (moderate complexity - 3.5s savings)
6. Byte slices in FindInternalMorphemes (moderate complexity - 8-9s savings)

**Subtotal**: 11.5-12.5s additional savings → **22-23s runtime**

### Phase 3: Advanced (3-5 days work, 5-7s additional savings)

7. Cache String.StringClean() results
8. Parallel morpheme extraction (careful - order matters)
9. Better data structures (custom allocators, arena allocation)

**Potential final**: **15-20s runtime** (2.5-3x faster than current)

---

## Part 7: Code Location Summary

### Files to Modify

1. **internal/nlp/nlp.go** (FindMorphemes)
   - Lines 414, 434: Object pooling, slice reuse
   - Lines 428-429: Avoid struct copies

2. **internal/morpheme_list_t/morpheme_list_t.go** (FindInternalMorphemes)
   - Lines 57, 82, 98, 107: Byte slices instead of strings
   - Lines 112, 152: Pool NewMorpheme allocations

3. **internal/graph/graph.go** (SemanticNode.GetNeighbors)
   - Add `GetNeighborsInto(buf []int) []int` method for reuse

4. **internal/functions/functions.go**
   - Add `ReverseBytes([]byte) []byte` for in-place reversal

---

## Summary

**Current**: 28.06s FindMorphemes (21s GC overhead)
**With Quick Wins**: ~24s (3-4s savings)
**With Memory Opts**: ~12s (16s savings)
**With Advanced Opts**: ~8-10s (18-20s savings)

**The bottleneck is clear**: Memory allocations causing GC thrashing. Fix allocations, get 2-3x speedup.
