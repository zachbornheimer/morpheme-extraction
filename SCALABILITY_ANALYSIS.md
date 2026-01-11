# Morpheme Extraction Scalability Analysis

## Executive Summary
**Current Performance**: ~4 seconds for 2,210 words (War of the Worlds corpus)
**Memory Usage**: 320MB total allocations
**Scalability Goal**: Handle trillions of operations on MacBook Pro M1

## Profile Analysis (January 11, 2026)

### Critical Bottlenecks

#### 1. Memory Hotspots (62% of total allocations)
**FindMorphemes** (nlp.go:232-242):
- Line 236 `AddMorpheme`: **122.52MB (38.33%)** - Appending morphemes creates new slices
- Line 234 `AddWord`: **40MB (12.51%)** - Word array growth
- Line 233 `AddWord`: **36.50MB (11.42%)** - Word array growth
- Line 237 `FindInternalMorphemes`: **35.50MB (11%)** - Internal morpheme discovery
- Line 242 `FuseRegex`: **38.34MB (12%)** - Morpheme fusion/deduplication

**Problem**: O(n²) morpheme comparisons with repeated slice allocations

#### 2. CPU Hotspots
- **GC Overhead**: 25.79% (runtime.madvise) - Too much allocation churn
- **InCharArray**: 4.23% - O(n) linear search in UniqWords
- **UniqWords**: 9.94% cumulative - Called thousands of times with linear search

#### 3. String Operations
- **Reverse**: 4.50MB - Creates new string every time
- **String conversions**: `string(word.Str[:word.Len])` throughout codebase

---

## Scalability Limitations

### Current Algorithm Complexity
```
For corpus with N words:
- Ngram generation: O(N)
- Ngram comparison: O(N² × M) where M = avg context words
- Morpheme extraction: O(N² × P) where P = morpheme operations
- Overall: O(N²) with high constant factors
```

### Memory Growth Pattern
```
Small corpus (2K words):     320MB
Medium corpus (20K words):   ~3.2GB (estimated)
Large corpus (200K words):   ~32GB (estimated)
Trillion operations:         WILL NOT FIT IN MEMORY
```

**Conclusion**: Current architecture **CANNOT scale to trillions** without fundamental changes.

---

## Optimization Strategy

### Phase 1: Low-Hanging Fruit (10-20% improvement)
**Target**: Reduce allocations without algorithm changes

1. **Object Pooling** ✅ (Already done for NgramsSimilar)
   - Extend to Morpheme, Word structs
   - Pre-allocate slices with capacity

2. **String Interning**
   - Most words repeat frequently
   - Use `sync.Map` for string deduplication
   - Store strings once, use pointers

3. **Remove Linear Search** (functions.go:81)
   ```go
   // Current: O(n) InCharArray for every comparison
   // Solution: Use map[string]struct{} from the start
   ```

4. **Reduce String Conversions**
   - Cache `word.String()` result
   - Work with `[]byte` where possible

### Phase 2: Algorithmic Improvements (50-70% improvement)
**Target**: Reduce O(N²) to O(N log N) or O(N)

1. **Index Ngrams by Hash**
   ```go
   type NgramIndex map[uint64][]*Ngram

   // Instead of comparing every ngram pair (O(N²)):
   // - Hash context words
   // - Only compare ngrams with similar hashes
   // - Reduces comparisons by 100x-1000x
   ```

2. **Streaming Processing**
   - Don't load entire corpus in memory
   - Process in chunks
   - Write intermediate results to disk (SQLite/BadgerDB)

3. **Trie-Based Morpheme Storage**
   - Current: Linear list of morphemes
   - Solution: Prefix trie for O(log N) lookup
   - Reduces FuseRegex from O(N²) to O(N log N)

4. **Bloom Filters for Quick Rejection**
   - Before expensive similarity check
   - "Does this ngram possibly match?" → 99% can be rejected instantly

### Phase 3: Concurrency (3-4x improvement on M1)
**Target**: Leverage all 8-10 cores efficiently

1. **Parallel Ngram Comparison** (nlp.go:199-200)
   ```go
   // Current: Sequential
   // Solution: Worker pool with job queue

   workers := runtime.NumCPU()
   jobs := make(chan NgramPair, 1000)
   results := make(chan Morpheme, 1000)

   for w := 0; w < workers; w++ {
       go worker(jobs, results)
   }
   ```

2. **Lock-Free Data Structures**
   - Use `sync.Map` for concurrent morpheme collection
   - Atomic counters instead of mutex

3. **Pipeline Architecture**
   ```
   Read Corpus → Parse Words → Generate Ngrams → Compare → Extract Morphemes → Fuse
        ↓             ↓              ↓              ↓            ↓              ↓
    Channel1      Channel2       Channel3       Channel4     Channel5     Channel6
   ```

### Phase 4: Distributed Processing (1000x+ improvement)
**Target**: Scale beyond single machine

1. **MapReduce Architecture**
   - Map: Process chunks independently
   - Reduce: Merge morpheme results

2. **Persistent Storage**
   - BadgerDB/LevelDB for intermediate results
   - SQLite for final morpheme database

3. **Incremental Processing**
   - Process new text without reprocessing old
   - Update morpheme frequencies incrementally

---

## Recommended Implementation Order

### Immediate (This Week)
1. ✅ **Object pooling for Ngrams** (Done)
2. **String interning for Words**
3. **Replace InCharArray with map**
4. **Batch allocations in AddMorpheme**

### Short Term (This Month)
1. **Implement NgramIndex with hashing**
2. **Add worker pool for parallel comparison**
3. **Trie-based morpheme storage**
4. **Streaming corpus reader**

### Long Term (3-6 Months)
1. **SQLite backend for large corpora**
2. **Incremental processing mode**
3. **Distributed processing framework**

---

## Expected Performance Improvements

| Optimization | Memory | CPU Time | Scalability |
|--------------|--------|----------|-------------|
| **Current** | 320MB | 4s | 2K words |
| + String interning | -40% (192MB) | -10% (3.6s) | 5K words |
| + Remove linear search | -5% (182MB) | -20% (2.9s) | 5K words |
| + Ngram indexing | -20% (146MB) | -50% (1.4s) | 50K words |
| + Worker pool | ±0% (146MB) | -60% (0.6s) | 50K words |
| + Streaming + disk | -80% (32MB) | +50% (0.9s) | 1M words |
| + Distributed | -90% (16MB/node) | -95% (0.2s) | **Trillions** |

---

## Code Hotspots to Optimize

### Priority 1 (Highest ROI)
1. **nlp.go:236** - `AddMorpheme` → Use pre-allocated slice with capacity
2. **nlp.go:234** - `AddWord` → Pool Word objects
3. **functions.go:81** - `InCharArray` → Replace with map lookup
4. **nlp.go:199-200** - `BuildNgramRelationships` → Add worker pool

### Priority 2 (Medium ROI)
1. **morpheme_list_t.go** - `FuseRegex` → Use trie or hash-based dedup
2. **ngram_t.go:96** - `NgramsSimilar` → Add bloom filter for quick rejection
3. **structs.go** - Add `Word.hash()` for fast equality checks
4. **nlp.go** - Stream corpus instead of loading entirely

### Priority 3 (Long-term)
1. Add persistent storage layer (SQLite/BadgerDB)
2. Implement incremental processing
3. Add distributed processing coordinator

---

## Memory Management Strategy

### For M1 MacBook Pro (16GB RAM)
```
Target Memory Budget: 8GB (leave 8GB for system)

Allocations:
- Ngram pool: 2GB (reusable)
- Morpheme storage: 4GB (persistent)
- Working memory: 1GB (temporary)
- OS/buffers: 1GB (reserved)
```

### Adaptive Processing
```go
if corpusSize < 10K {
    // In-memory processing (current approach)
} else if corpusSize < 100K {
    // Streaming with memory buffer
} else {
    // Disk-backed with SQLite
}
```

---

## Testing Scalability

### Benchmark Targets
1. **Small**: 2K words (War of the Worlds) - **< 1s**
2. **Medium**: 20K words (10x larger) - **< 5s**
3. **Large**: 200K words (100x larger) - **< 30s**
4. **Huge**: 2M words (1000x larger) - **< 5 minutes**

### Memory Targets
1. Small: < 200MB
2. Medium: < 500MB
3. Large: < 2GB
4. Huge: < 4GB

---

## Conclusion

**Current State**: Highly optimized single-threaded algorithm, but O(N²) complexity limits scale.

**Path to Trillion-Scale**:
1. Short-term: Algorithmic improvements (hashing, indexing) → 100x scale
2. Medium-term: Concurrency + streaming → 1000x scale
3. Long-term: Distributed + persistent storage → **Trillion-scale**

**Key Insight**: The algorithm is fundamentally sound (produces correct morphemes), but needs architectural changes for massive scale. Think: "What if Netflix built this?"

**Next Steps**: Start with Priority 1 optimizations (string interning, map-based lookup, worker pools) for immediate 2-3x improvement.
