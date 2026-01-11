# Path to 0.4s: Aggressive Optimization Plan

## Current State
- **C version**: 2.9s
- **Go version**: 3.7s (no GC), 4.1s (with GC)
- **Target**: 0.4s (10x faster than current Go!)

## The Bottlenecks

### Critical O(N²) Issues

**1. BuildNgramRelationships Lines 115-124**
```go
for q = 0; q < addLoc; q++ {
    if string(ng[q].Word.Str) == arr[i] {  // LINEAR SEARCH!
```
- **Complexity**: O(N²) for N words
- **For 2,210 words**: ~4.88M comparisons
- **Cost**: ~2s

**2. Ngram Similarity Comparisons (lines 199-200)**
```go
for i := 0; i < ngramLength; i++ {
    for j := i + 1; j < ngramLength; j++ {
        if ngram_t.NgramsSimilar(*ng[i], *ng[j])
```
- **Complexity**: O(N²) comparisons
- **For 2,210 ngrams**: 2.4M comparisons
- **Cost**: ~1s

**3. UniqWords Temporary Arrays**
- **614MB** allocated across 50K calls
- **Cost**: 300ms just in malloc

**4. FindInternalMorphemes String Operations**
- Double-reverse operations
- String ↔ []byte conversions
- **Cost**: 200ms

## Optimization Strategy

### Phase 1: Eliminate O(N²) → O(N log N) (Target: 1.5s, 60% reduction)

#### 1.1 Hash Map for Ngram Lookup (Replace lines 115-124)
```go
// BEFORE: O(N²) linear search
for q = 0; q < addLoc; q++ {
    if string(ng[q].Word.Str) == arr[i] {
        ngram = *ng[q]
        ...
    }
}

// AFTER: O(1) hash lookup
type NgramMap map[string]*structs.Ngram

ngramMap := make(NgramMap, wordCount)

// Build phase
if existing, found := ngramMap[arr[i]]; found {
    ngram = *existing
    ngramCount--
    makeNew = 0
} else {
    makeNew = 1
}
```

**Expected**: 2s → 0.2s (90% reduction in this section)

#### 1.2 LSH (Locality-Sensitive Hashing) for Ngram Similarity

Instead of comparing ALL pairs, hash ngrams by their context and only compare within buckets.

```go
type NgramBucket struct {
    hash    uint64
    ngrams  []*structs.Ngram
}

func HashNgram(ng *structs.Ngram) uint64 {
    h := fnv.New64a()
    // Hash sorted context words
    words := collectAndSortContextWords(ng)
    for _, w := range words {
        h.Write([]byte(w))
    }
    return h.Sum64()
}

// Only compare ngrams with same hash
buckets := make(map[uint64][]*structs.Ngram)
for i := 0; i < ngramLength; i++ {
    hash := HashNgram(ng[i])
    buckets[hash] = append(buckets[hash], ng[i])
}

// Compare only within buckets
for _, bucket := range buckets {
    for i := 0; i < len(bucket); i++ {
        for j := i + 1; j < len(bucket); j++ {
            if ngram_t.NgramsSimilar(*bucket[i], *bucket[j]) {
                ...
            }
        }
    }
}
```

**Expected**: 1s → 0.1s (reduces comparisons by 100x for typical data)

### Phase 2: Parallelize Everything (Target: 0.5s, 67% reduction)

#### 2.1 Parallel Morpheme Extraction

```go
func FindMorphemes(...) int {
    numWorkers := runtime.NumCPU()
    jobs := make(chan MorphemeJob, 10000)
    results := make(chan MorphemeResult, 10000)

    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go morphemeWorker(jobs, results, &wg)
    }

    // Distribute work
    go func() {
        for i := 0; i < ngramLength; i++ {
            if ng[i] == nil {
                continue
            }
            for j := 0; j < ng[i].RefsCount; j++ {
                jobs <- MorphemeJob{
                    ngram: ng[i],
                    target: ng[i].Refs[j],
                }
            }
        }
        close(jobs)
    }()

    // Collect results
    go func() {
        for result := range results {
            morpheme_list_t.AddMorpheme(&internal, result.morpheme)
        }
    }()

    wg.Wait()
    close(results)
}
```

**Expected on M1 (8 cores)**: 1.5s → 0.3s (5x speedup)

#### 2.2 Parallel Ngram Comparison

```go
// Distribute buckets across workers
for hash, bucket := range buckets {
    jobs <- ComparisonJob{hash: hash, ngrams: bucket}
}
```

**Expected**: 0.1s → 0.02s

### Phase 3: Memory Pool Everything (Target: 0.4s, 20% reduction)

#### 3.1 Word Pool
```go
var wordPool = sync.Pool{
    New: func() interface{} {
        return &structs.Word{Str: make([]byte, 100)}
    },
}
```

#### 3.2 Morpheme Pool
```go
var morphemePool = sync.Pool{
    New: func() interface{} {
        return &structs.Morpheme{
            Words: make([]*structs.Word, 0, 10),
        }
    },
}
```

#### 3.3 StringSet Pool (for UniqWords)
```go
var stringSetPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]struct{}, 50)
    },
}

func UniqWords(arr []string, size int) int {
    seen := stringSetPool.Get().(map[string]struct{})
    defer func() {
        for k := range seen {
            delete(seen, k)  // Clear map
        }
        stringSetPool.Put(seen)
    }()

    for j := 0; j < size; j++ {
        seen[arr[j]] = struct{}{}
    }
    return len(seen) - 1
}
```

**Expected**: 0.5s → 0.4s (eliminates 100MB of allocations)

### Phase 4: Aggressive Micro-optimizations (Target: <0.4s)

#### 4.1 Replace string() conversions with unsafe pointers
```go
import "unsafe"

func bytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
```

#### 4.2 Pre-compute string hashes
```go
type HashedWord struct {
    Word structs.Word
    Hash uint64
}
```

#### 4.3 Use byte comparisons instead of string comparisons
```go
// Instead of: if string(ng[q].Word.Str) == arr[i]
// Use: if bytes.Equal(ng[q].Word.Str[:ng[q].Word.Len], arr[i])
```

---

## Implementation Order

### Week 1: Hash Maps (Days 1-2)
- [ ] Replace linear search in BuildNgramRelationships with map
- [ ] Test correctness
- **Expected**: 3.7s → 1.7s (2s saved)

### Week 1: LSH for Ngram Similarity (Days 3-4)
- [ ] Implement LSH bucketing
- [ ] Only compare within buckets
- [ ] Test correctness
- **Expected**: 1.7s → 0.7s (1s saved)

### Week 1: Memory Pools (Day 5)
- [ ] Add sync.Pool for common allocations
- [ ] Test correctness
- **Expected**: 0.7s → 0.6s (100ms saved)

### Week 2: Parallelization (Days 1-3)
- [ ] Parallel morpheme extraction
- [ ] Parallel ngram comparison
- [ ] Lock-free data structures
- **Expected**: 0.6s → 0.15s (4x speedup on M1)

### Week 2: Micro-optimizations (Days 4-5)
- [ ] Unsafe string conversions
- [ ] Pre-computed hashes
- [ ] Byte-level comparisons
- **Expected**: 0.15s → <0.1s

---

## Timeline to 0.4s

| Week | Optimization | Time | Reduction |
|------|-------------|------|-----------|
| **Baseline** | Current Go (no GC) | 3.7s | - |
| **Week 1, Day 2** | + Hash maps | 1.7s | -2.0s |
| **Week 1, Day 4** | + LSH bucketing | 0.7s | -1.0s |
| **Week 1, Day 5** | + Memory pools | 0.6s | -0.1s |
| **Week 2, Day 3** | + Parallelization (8 cores) | **0.15s** | -0.45s |
| **Week 2, Day 5** | + Micro-opts | **<0.1s** | -0.05s |

**Target achieved in 2 weeks!**

---

## Code Changes Required

### File: internal/nlp/nlp.go
- Lines 115-124: Replace with hash map
- Lines 199-200: Add LSH bucketing
- Lines 216-238: Add worker pool

### File: internal/functions/functions.go
- Line 78: Add sync.Pool for UniqWords

### File: internal/ngram_t/ngram_t.go
- Add LSH hash function
- Add parallel comparison logic

### File: internal/structs/structs.go
- Add Hash field to Word
- Add sync.Pool support

---

## Risk Mitigation

After EVERY change:
```bash
go build -o nlp-go ./cmd/morpheme-extract
./nlp-go test-corpus/
diff -a languageModel.zdrm languageModel-good.zdrm
```

If diff fails, revert and debug.

---

## Alternative: Radical Rethink

If we can't hit 0.4s with these optimizations, consider:

1. **Streaming algorithm**: Process corpus in chunks, never load all in memory
2. **Trie-based morpheme storage**: O(log N) lookup instead of O(N)
3. **Approximate algorithms**: Trade 1% accuracy for 10x speed
4. **GPU acceleration**: Parallel string comparisons on Metal (M1)
5. **SIMD**: Use ARM NEON instructions for byte comparisons

---

## Next Step

Start with the hash map replacement (Day 1-2 task). This alone should get us from 3.7s → 1.7s (54% reduction) and proves the approach works.

Ready to implement?
