# Morpheme Extraction Optimization: Findings & Roadmap to Quintillion Scale

## Current Performance Baseline

**Small Corpus (War of the Worlds, 972 unique words):**
- Runtime: 2.46s
- CPU: 213% (multi-core utilization working)
- Memory: ~773MB peak
- Output: 1024 morphemes (683 INFIX, 138 PREFIX, 114 SUFFIX, 86 STEM)

**Bottleneck Analysis (from profiling):**
- 33% garbage collection
- 22% FindMorphemes (vocabulary deduplication)
- 18% FindInternalMorphemes
- 14% NgramsSimilar (O(V²) similarity detection) ← **Scaling killer**

## Optimization Phases Completed

### Phase 1: Parallelization Attempts
**Attempts:**
1. ❌ Naive goroutine per k value → 4.1s (overhead dominated)
2. ❌ Worker pool with mutex → 8.7s (lock contention)
3. ❌ Lock-free batching, poor distribution → 11.5s
4. ✅ Chunked workers (CPU × 4) → 2.46s (correct approach)

**Key Learning:** For small corpus (V=972), parallelization overhead ≈ sequential work. But for large V, O(V²) work will completely dominate overhead, making parallelization essential.

**Implementation:**
- Workers: NumCPU × magicMultiplier (32 workers on M2 Mac)
- Each worker processes chunk of k values
- Lock-free local result collection
- Sequential merge only at end

### Phase 2: Word Interning Attempt
**Goal:** Reduce memory by sharing Word structs across contexts

**Result:** ❌ 2048 morphemes instead of 1024 (broken)

**Root Cause:** Word.Freq is per-context frequency, not global. Sharing pointers made all contexts share the same Freq value, breaking deduplication logic in ngram_t.go:80:
```go
elem.Freq += bora.At[i].Elems[k].Freq
```

**Fix Required:** Separate word identity from context state:
```go
type WordIdentity struct {
    Str string  // Shared, immutable
}
type WordContext struct {
    Identity *WordIdentity  // Pointer to shared identity
    Freq int                // Per-context (NOT shared)
}
```

### Phase 3: Chunked Parallelization
**Result:** ✅ 2.46s, correct output, 213% CPU utilization

**Fixed Bug:** Removed redundant `string()` conversions on Word.Str (already string type), which was allocating in hot loop.

## Scaling Analysis: The O(V²) Problem

### Current Reality
For vocabulary size V:
- V=1K: 1M comparisons (~1 second) ✓
- V=10K: 100M comparisons (~100 seconds) ⚠️
- V=100K: 10B comparisons (~3 hours) ✗
- V=1M: 1T comparisons (~11 days) ✗✗
- V=1B: 10^18 comparisons (~30,000 years) ✗✗✗

### Target: 10^18 Words → ~10^9 Vocabulary
With current O(V²) algorithm:
- Comparisons: (10^9)² = 10^18
- At 1M comparisons/sec: 31,000 YEARS
- Memory: 500GB with naive approach

**Verdict:** Current algorithm CANNOT scale to quintillion. Need fundamental architectural changes.

## Recommended Path to Quintillion Scale

### Short-Term (Enables millions of words)

#### 1. MinHash Pre-Filtering ⭐ HIGH PRIORITY
**Goal:** Reduce O(V²) to O(V×B²) where B << V

**Approach:**
- Compute MinHash signature for each ngram's context
- Quick rejection: if signatures differ significantly, skip full comparison
- Full NgramsSimilar only for candidates

**Expected Impact:**
- V=1M: 1T → ~1B comparisons (1000× reduction)
- Runtime: days → hours

**Implementation:**
```go
type Ngram struct {
    Word Word
    Before, After NgramElements
    MinHashSig [16]uint32  // NEW: 16-signature MinHash
    // ... rest
}

func ComputeMinHash(ng *Ngram) {
    // Hash all context words, keep k smallest hashes
}

func MinHashSimilar(a, b Ngram, threshold float64) bool {
    // Count matching signatures: O(16) instead of O(N)
    matches := 0
    for i := 0; i < 16; i++ {
        if a.MinHashSig[i] == b.MinHashSig[i] {
            matches++
        }
    }
    return float64(matches)/16.0 >= threshold
}

// In similarity loop:
if MinHashSimilar(*ng[k], *ng[j], 0.10) {  // Pre-filter
    if ngram_t.NgramsSimilar(*ng[k], *ng[j]) {  // Full check
        // ... add similarity
    }
}
```

#### 2. Memory-Efficient Word Interning v2
**Goal:** Reduce memory from 773MB to <100MB

**Fix Phase 2 Failure:**
```go
type WordIdentity struct {
    Str string  // Shared across all contexts
}

type WordContext struct {
    Identity *WordIdentity  // Pointer (8 bytes)
    Len int                 // Keep for compatibility
    Freq int                // Per-context (unique per usage)
}
```

**Expected Impact:**
- 972 unique words × 40 bytes = 0.04MB for identities
- Context structs: 16 bytes vs 40+ bytes per occurrence
- ~70% memory reduction

### Medium-Term (Enables billions of words)

#### 3. Streaming Corpus Processing
**Goal:** Process corpus in chunks, merge results incrementally

**Approach:**
- Split corpus into N chunks
- Process each chunk independently (parallel or sequential)
- Merge morpheme results with conflict resolution
- Never load full corpus in memory

**Expected Impact:**
- Memory: constant (per-chunk) instead of O(corpus size)
- Enables processing datasets that don't fit in RAM

#### 4. LSH (Locality-Sensitive Hashing) for Similarity
**Goal:** O(V) similarity detection instead of O(V²)

**Approach:**
- Hash ngrams to buckets based on context similarity
- Only compare ngrams in same/nearby buckets
- Trade-off: May miss some morphemes (tunable)

**Expected Impact:**
- V=1B: 10^18 → ~10^9 comparisons (billion× reduction)
- Enables billion-word vocabularies

### Long-Term (Quintillion scale)

#### 5. Distributed Processing
**Goal:** Horizontal scaling across machines

**Approach:**
- Partition vocabulary across nodes
- Each node computes local similarities
- Coordinator merges results
- Map-Reduce pattern

**Expected Impact:**
- Linear scaling with machines
- V=1B across 1000 machines = V=1M per machine

#### 6. Approximate Morpheme Detection
**Goal:** Trade accuracy for massive scale

**Approach:**
- Sampling: only compare random subset of word pairs
- Probabilistic sufficiency: high-frequency morphemes found quickly
- Rare morphemes: may be missed (acceptable for NLP applications)

**Expected Impact:**
- Process quintillion words in reasonable time
- May miss <1% of rare morphemes

## Recommended Implementation Order

### Phase 4 (Next): MinHash Pre-Filtering ⭐
**Priority:** HIGHEST - Addresses fundamental O(V²) bottleneck
**Risk:** LOW - Pure optimization, doesn't change algorithm
**Files:** internal/nlp/nlp.go, internal/structs/structs.go, internal/ngram_t/ngram_t.go
**Effort:** ~200 lines of code

### Phase 5: Word Interning v2
**Priority:** HIGH - Significant memory savings
**Risk:** MEDIUM - Must fix Freq semantics correctly
**Files:** internal/structs/structs.go, internal/word_t/word_t.go, internal/vocabulary/vocabulary.go
**Effort:** ~150 lines of code

### Phase 6: Streaming Processing
**Priority:** MEDIUM - Enables larger-than-RAM datasets
**Risk:** MEDIUM - Merge logic must preserve correctness
**Files:** New package internal/streaming/, refactor internal/nlp/nlp.go
**Effort:** ~500 lines of code

### Phase 7: LSH Similarity
**Priority:** MEDIUM - Further O(V²) reduction
**Risk:** HIGH - May affect morpheme quality
**Files:** internal/nlp/nlp.go, internal/ngram_t/ngram_t.go
**Effort:** ~300 lines of code

## Key Architectural Insights

### 1. Symbolic Morphology Requires Exact Matching
Your preference for symbolic reasoning over embeddings is well-founded for morphology:
- Morphemes are discrete, compositional units
- "walk + ing = walking" is rule-based, not continuous
- Distributional similarity (manual or via embeddings) is for CANDIDATE SELECTION
- Validation (GenRegex) is still exact string matching

### 2. Threshold-Based Similarity Is Working
The 12.5% Jaccard threshold successfully implements distributional hypothesis:
- Words in similar contexts likely share morphemes
- Not mathematical abstraction - explicit context overlap
- Interpretable and tunable

### 3. Sequential State Must Be Preserved
The lookback filter is inherently sequential:
- Prevents cascading suffix redundancy ("ering" → "ring" → "ing")
- Cannot be trivially parallelized or collected as candidates
- Must remain in sequential FindMorphemes loop

## Success Metrics

### Correctness (MANDATORY)
- ✅ 1024 morphemes (683 INFIX, 138 PREFIX, 114 SUFFIX, 86 STEM)
- ✅ Exact binary match with baseline
- ✅ All optimizations preserve this

### Performance (Target)
- Current: 2.46s for V=972
- Short-term goal: <1s for V=972, <10min for V=100K
- Long-term goal: <1 day for V=1M, <1 week for V=1B

### Scalability (Ultimate Goal)
- ✅ Phase 3: Parallel-ready for multi-core (213% CPU)
- Phase 4: MinHash enables V=1M in hours
- Phase 5: Memory-efficient for V=10M+
- Phase 6+: Streaming + LSH + distributed enables V=1B+

## Next Immediate Action

Implement MinHash pre-filtering in similarity detection loop to address the O(V²) bottleneck before it becomes intractable at larger vocabularies.
