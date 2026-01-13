# Profiling Analysis - Post-MinHash Removal

**Date**: 2026-01-12
**Commit**: 34878c4 (MinHash removed)
**Runtime**: 41.0s (14% faster than 47.6s Phase 4 baseline)
**Output**: 3,702 morphemes, 354 unique STEMs (correct)

---

## Executive Summary

**Top 3 Bottlenecks** (with GC enabled):
1. **FindMorphemes**: 31.52s (68.7%) - morpheme extraction algorithm
2. **FindInternalMorphemes**: 13.90s (30.3%) - infix detection
3. **GC/Runtime**: 5.03s (11.0%) - memory management overhead

**Memory Usage**:
- Total: 1,221 MB peak
- buildMorphemesFromCandidatesWithRegex: 382 MB (31%)
- NewMorpheme: 266 MB (22%)
- FindMorphemes: 247 MB (20%)

---

## Part 1: With GC Enabled (Normal Operation)

### CPU Profile (45.88s total samples)

```
Function                                flat    flat%   cum     cum%
--------------------------------------------------------------------
FindMorphemes                          13.61s   29.7%  31.52s  68.7%
runtime.madvise                         5.03s   11.0%   5.03s  11.0%
runtime.stringtoslicerune               3.16s    6.9%   3.21s   7.0%
runtime.mallocgcTiny                    2.18s    4.8%   2.80s   6.1%
runtime.memclrNoHeapPointers            1.75s    3.8%   1.75s   3.8%
runtime.memmove                         1.24s    2.7%   1.24s   2.7%
runtime.slicerunetostring               1.08s    2.4%   3.15s   6.9%
runtime.encoderune                      0.86s    1.9%   0.86s   1.9%
runtime.scanobject                      0.84s    1.8%   2.74s   6.0%
FindInternalMorphemes                   0.81s    1.8%  13.90s  30.3%
runtime.pthread_cond_signal             0.81s    1.8%   0.81s   1.8%
runtime.mallocgc                        0.72s    1.6%   5.41s  11.8%
ToLowerASCII                            0.71s    1.6%   2.22s   4.8%
runtime.concatstrings                   0.71s    1.6%   3.08s   6.7%
UniqIdentities                          0.56s    1.2%   1.30s   2.8%
```

### Slowest Lines in FindMorphemes (line-level)

```
Line    Flat     Cum      Code
----------------------------------------
418     2.35s    2.38s    for j := i + constants.NgramSize; j < ngramLength; j++
419     5.08s    5.16s    if ng[j] == nil || ng[j].Word.Identity == nil
434     4.59s    4.77s    candidates = append(candidates, ...)
424     1.53s    1.57s    if ng[j].Word.Identity.ID != neighborID
445      —      13.90s    FindInternalMorphemes(...)
414      —       0.75s    neighborIDs := node.GetNeighbors()
433      —       0.17s    forwardStr := FindLongestMatchString(...)
```

**Key insights**:
- Line 419 (nil checks): 5.08s flat - unexpectedly expensive!
- Line 434 (append): 4.59s - slice growth overhead
- Line 418 (inner loop): 2.35s - iteration cost
- Line 445 (FindInternalMorphemes): 13.90s cumulative - biggest hotspot

### Memory Profile (1,221 MB peak)

```
Function                                Alloc    %
-------------------------------------------------
buildMorphemesFromCandidatesWithRegex   382 MB  31.3%
NewMorpheme (inline)                    266 MB  21.8%
FindMorphemes                           247 MB  20.3%
SemanticGraph.AddEdge                   143 MB  11.8%
RegexPattern.AddContext                  74 MB   6.0%
Word.SetWord                             52 MB   4.3%
Morpheme.AddWord                         29 MB   2.4%
ToLowerASCII                             11 MB   0.9%
```

**Key insights**:
- 382 MB in regex building (31%) - largest single allocator
- 266 MB in NewMorpheme - creating ~2.6M morpheme objects
- 247 MB in FindMorphemes - candidate slice growth

---

## Part 2: With GC Disabled (GOGC=off)

### CPU Profile (45.05s total samples)

```
Function                                flat    flat%   cum     cum%
--------------------------------------------------------------------
runtime.madvise                        33.81s   75.1%  33.81s  75.1%
FindMorphemes                           1.85s    4.1%   6.97s  15.5%
runtime.memmove                         0.80s    1.8%   0.80s   1.8%
runtime.memequal                        0.72s    1.6%   0.72s   1.6%
runtime.memclrNoHeapPointers            0.55s    1.2%   0.55s   1.2%
runtime.mallocgcTiny                    0.48s    1.1%   1.01s   2.2%
UniqIdentities                          0.37s    0.8%   0.84s   1.9%
BitSet.IsSet                            0.33s    0.7%   0.33s   0.7%
FindInternalMorphemes                   0.11s    0.2%   2.33s   5.2%
```

**Key differences from GC-enabled**:
- `runtime.madvise` dominates (75%) - kernel memory management without GC
- FindMorphemes drops to 6.97s (from 31.52s) - 78% reduction!
- FindInternalMorphemes drops to 2.33s (from 13.90s) - 83% reduction!

**Conclusion**: With GC disabled, pure algorithmic cost is ~7s for FindMorphemes.
**GC overhead**: 31.52s - 6.97s = **24.55s (77% of FindMorphemes time is GC!)**

### Slowest Lines in FindMorphemes (no GC)

```
Line    Flat     Cum      Code
----------------------------------------
419     640ms    650ms    if ng[j] == nil || ng[j].Word.Identity == nil
418     400ms    400ms    for j := i + constants.NgramSize; j < ngramLength; j++
434     580ms    750ms    candidates = append(candidates, ...)
424     220ms    220ms    if ng[j].Word.Identity.ID != neighborID
445      —      2.33s     FindInternalMorphemes(...)
414      —       90ms     neighborIDs := node.GetNeighbors()
433      —       40ms     forwardStr := FindLongestMatchString(...)
```

**Pure algorithmic costs** (without GC overhead):
- Inner loop: 400ms (was 2.35s with GC)
- Nil checks: 640ms (was 5.08s with GC)
- Append: 580ms (was 4.59s with GC)
- FindInternalMorphemes: 2.33s (was 13.90s with GC)

---

## Part 3: Optimization Opportunities

### High Impact (10-20s potential savings)

1. **Reduce GC Pressure in FindMorphemes**
   - Current: 77% of FindMorphemes is GC overhead (24.55s)
   - Strategy: Reuse candidate slice, object pooling for MorphemeCandidate
   - Expected: 10-15s savings

2. **Optimize FindInternalMorphemes** (13.90s → target 5s)
   - Current: String allocations, substring operations
   - Strategy: Use byte slices, reduce allocations
   - Expected: 8-10s savings

3. **Preallocate Candidate Slice**
   - Current: `make([]structs.MorphemeCandidate, 0, ngramLength*100)`
   - Actual usage: ~2.6M candidates added
   - Strategy: Profile actual max size, preallocate once
   - Expected: 2-3s savings (reduce append overhead)

### Medium Impact (2-5s potential savings)

4. **Optimize Nil Checks** (5.08s → target 1s)
   - Line 419 surprisingly expensive
   - Strategy: Early validation, reduce redundant checks
   - Expected: 2-3s savings

5. **Cache neighborIDs Slice**
   - Current: Allocated on each outer loop iteration
   - Strategy: Reuse slice with pooling
   - Expected: 1-2s savings

6. **String Operations** (6.9s total in rune conversions)
   - runtime.stringtoslicerune: 3.16s
   - runtime.slicerunetostring: 1.08s
   - Strategy: Work with byte slices when possible
   - Expected: 2-3s savings

### Low Impact (<2s potential savings)

7. **BitSet Already Optimized**
   - UniqIdentities: 1.30s (down from 21.7s baseline)
   - BitSet.IsSet: 0.33s
   - Already 93% reduction - little room for improvement

8. **Dictionary Normalization**
   - ToLowerASCII: 2.22s
   - Called frequently but individual calls are fast
   - Hard to optimize further without caching (trade memory for speed)

---

## Part 4: Performance Timeline

```
Baseline (Phase 0):           59s   (C implementation reference)
Phase 1-3 (Graph):            59s   (semantic graph, no performance change)
Phase 4 (BitSet):             47.6s (11.4s saved, 19% faster)
MinHash disabled:             47.0s (0.6s saved)
MinHash removed:              41.0s (6.6s additional, 14% faster total)
```

**Total improvement**: 59s → 41s = **18s saved (31% faster)**

---

## Part 5: Recommendations (Priority Order)

### Immediate (High ROI)
1. **Object pooling for MorphemeCandidate** - reduce GC in FindMorphemes
2. **Preallocate candidate slice** - profile max size, allocate once
3. **Optimize FindInternalMorphemes** - reduce string allocations

### Short-term (Medium ROI)
4. **Cache neighbor slices** - reuse allocations in graph traversal
5. **Reduce nil checks** - early validation, remove redundant checks
6. **Byte slice operations** - avoid rune conversions where possible

### Long-term (Research)
7. **Parallel morpheme extraction** - FindMorphemes is embarrassingly parallel
8. **Incremental processing** - process chunks, write results, reduce peak memory
9. **Alternative algorithms** - suffix arrays, suffix trees for faster matching

---

## Part 6: GC Analysis

**With GC (normal)**:
- Total time: 45.88s
- FindMorphemes: 31.52s (68.7%)

**Without GC**:
- Total time: 45.05s (similar, but dominated by madvise)
- FindMorphemes: 6.97s (15.5%)

**GC overhead in FindMorphemes**: 31.52s - 6.97s = **24.55s (77%)**

This means **3 out of every 4 seconds** in FindMorphemes is spent on garbage collection, not algorithm execution. This is the primary optimization target.

---

## Appendix: Profile Commands

```bash
# With GC enabled (normal)
./nlp-go -cpuprofile cpu-with-gc.prof -memprofile mem-with-gc.prof test-corpus/
go tool pprof -top cpu-with-gc.prof
go tool pprof -list=FindMorphemes cpu-with-gc.prof
go tool pprof -top mem-with-gc.prof

# With GC disabled
GOGC=off ./nlp-go -cpuprofile cpu-no-gc.prof -memprofile mem-no-gc.prof test-corpus/
go tool pprof -top cpu-no-gc.prof
go tool pprof -list=FindMorphemes cpu-no-gc.prof
```
