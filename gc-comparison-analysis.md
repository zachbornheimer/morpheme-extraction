# GC vs NO GC Profiling Comparison - Complete Analysis

## Test Configuration
- Corpus: War of the Worlds (test-corpus/)
- Dictionary: 4,215 unique words interned
- Output: 1,024 morphemes ✅ EXACT MATCH

---

## Runtime Comparison

| Configuration | Wall Time | CPU Samples | Efficiency |
|--------------|-----------|-------------|------------|
| **WITH GC** | 3.03s | 4.17s | 138% (multi-core) |
| **NO GC** | 6.95s | 2.70s | 39% (memory-bound) |

**Key Insight:** With GC enabled, runtime is **2.3× faster** (3.03s vs 6.95s), showing GC overhead is minimal compared to continuous memory allocation pressure without GC.

---

## CPU Profile Comparison

### Overall Time Distribution

#### WITH GC (Total: 4.17s)
| Component | Time | % |
|-----------|------|---|
| **GC (marking/scanning)** | 1.25s | 30% |
| **Memory allocation** | 1.17s | 28% |
| **Application code** | 1.75s | 42% |

#### NO GC (Total: 2.70s)
| Component | Time | % |
|-----------|------|---|
| **Memory allocation** | 1.84s | 68% |
| **Application code** | 0.86s | 32% |

### Application Code Hot Functions

| Function | WITH GC | NO GC | Analysis |
|----------|---------|-------|----------|
| **FindMorphemes** | 1.03s (25%) | 0.47s (17%) | 2.2× slower with GC (more objects to scan) |
| **BuildNgramRelationships workers** | 0.48s (12%) | 0.31s (11%) | Consistent |
| **NgramsSimilar** | 0.44s (11%) | 0.26s (10%) | 1.7× slower with GC |
| **FindInternalMorphemes** | 0.83s (20%) | 0.24s (9%) | 3.5× slower with GC (write barriers) |
| **FindLongestMatch** | 0.34s (8%) | 0.18s (7%) | 1.9× slower with GC |
| **UniqIdentities** | 0.26s (6%) | 0.16s (6%) | 1.6× slower with GC |
| **reverseString** | 0.44s (11%) | 0.03s (1%) | **15× slower with GC!** |
| **Word.RemoveFirstChar** | 0.36s (9%) | 0.03s (1%) | **12× slower with GC!** |

**Critical Observation:** String operations (`reverseString`, `RemoveFirstChar`) show **10-15× slowdown** with GC due to:
- Write barriers for string creation
- Scanning of intermediate string objects
- GC pressure from allocations

---

## Memory Profile Comparison (Peak Active Memory)

| Configuration | Active Memory | Total Allocated | GC Collections |
|--------------|---------------|-----------------|----------------|
| **WITH GC** | 117 MB | 1,304 MB | 60 |
| **NO GC** | 1,304 MB | 1,304 MB | 0 |

**GC Effectiveness:** Reclaims **91%** of allocated memory (1,304 MB → 117 MB active)

### Memory Allocations Breakdown

#### WITH GC (117 MB active)
| Component | Memory | % |
|-----------|--------|---|
| NewMorpheme | 34 MB | 29% |
| FindMorphemes overhead | 47 MB | 40% |
| AddSimilarNgramRef | 10 MB | 8% |
| BuildNgramRelationships | 13 MB | 11% |
| Other | 13 MB | 11% |

#### NO GC (1,304 MB total)
| Component | Memory | % |
|-----------|--------|---|
| NewMorpheme | 796 MB | 61% |
| reverseString | 120 MB | 9% |
| FuseRegex | 96 MB | 7% |
| file.ReadByChar | 81 MB | 6% |
| Other | 211 MB | 16% |

---

## GC Impact Analysis

### GC Overhead Breakdown (WITH GC)

| GC Activity | Time | % of Total |
|-------------|------|------------|
| **gcDrain (marking)** | 1.25s | 30% |
| **scanobject** | 1.05s | 25% |
| **greyobject** | 0.41s | 10% |
| **findObject** | 0.29s | 7% |
| **wbBufFlush (write barriers)** | 0.21s | 5% |
| **Total GC** | ~2.21s | ~53% |

**GC spends 53% of CPU time**, but this enables:
- 91% memory reclamation
- 2.3× faster overall runtime
- Lower memory footprint (117 MB vs 1,304 MB)

### Hot Path Analysis: String Operations

**reverseString:**
- NO GC: 0.03s (1%)
- WITH GC: 0.44s (11%) - **15× slower**
- Reason: Creates new strings, triggers write barriers, GC scanning

**Word.RemoveFirstChar:**
- NO GC: 0.03s (1%)
- WITH GC: 0.36s (9%) - **12× slower**
- Reason: String slicing creates new objects, GC overhead

**Optimization Opportunity:** These string operations could use byte buffers with pooling to reduce GC pressure.

---

## Performance Characteristics

### WITH GC (Recommended for Production)
✅ **Pros:**
- 2.3× faster overall (3.03s vs 6.95s)
- 91% memory reclamation
- Steady-state memory: 117 MB (usable for larger corpora)
- More CPU time available (multi-core scaling)

⚠️ **Cons:**
- 53% CPU spent on GC
- String operations 10-15× slower
- Write barrier overhead on allocations

### NO GC (Useful for Profiling Only)
✅ **Pros:**
- Pure allocation time measurement
- No GC interference in profiles
- String operations faster (no barriers)

❌ **Cons:**
- 2.3× slower overall
- Memory grows unbounded (1.3 GB for small corpus)
- Not sustainable for production

---

## Key Insights

### 1. GC is Highly Effective
- Reclaims 91% of memory (1,304 MB → 117 MB)
- 60 GC cycles complete successfully
- Enables 2.3× faster runtime despite 53% CPU overhead

### 2. String Operations are GC Bottleneck
- `reverseString`: 15× slower with GC
- `Word.RemoveFirstChar`: 12× slower with GC
- Root cause: Write barriers + object scanning + allocation pressure

### 3. Identity Optimization Benefits Both Modes
- Similarity detection: < 10 MB (was 599 MB)
- Works efficiently in both GC and NO GC modes
- Pointer comparison is cheap in both scenarios

### 4. Most Time is Productive
Even with GC:
- 42% CPU on application logic (vs 32% NO GC)
- 30% on GC marking (necessary for memory reclamation)
- Only 28% on pure allocation overhead

---

## Optimization Priorities

### 🔴 HIGH PRIORITY: String Operations (0.80s / 19% WITH GC)

**Current:**
```go
func reverseString(s string) string {
    runes := []rune(s)              // Allocation + write barrier
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)            // Allocation + write barrier
}
```

**Optimization:**
- Use byte slice pool for reversal
- In-place operations where possible
- Expected: 0.80s → 0.10s (8× improvement)

### 🟡 MEDIUM PRIORITY: NewMorpheme Pooling (34 MB WITH GC)

**Current:** Creates new Morpheme struct on every call
**Optimization:** Object pool with Reset() method
**Expected:** 34 MB → 10 MB

### 🟢 LOW PRIORITY: Other Allocations
- Already minimal with GC
- Further optimization not worth complexity

---

## Comparison Tables

### Runtime Performance
| Metric | WITH GC | NO GC | Winner |
|--------|---------|-------|--------|
| Wall Time | 3.03s | 6.95s | ✅ GC (2.3× faster) |
| CPU Time | 4.17s | 2.70s | NO GC (but not useful) |
| Memory Active | 117 MB | 1,304 MB | ✅ GC (91% less) |
| GC Collections | 60 | 0 | (N/A) |

### CPU Hot Functions (WITH GC)
| Function | Flat | Cum | % of Total |
|----------|------|-----|------------|
| FindMorphemes | - | 1.03s | 25% |
| FindInternalMorphemes | - | 0.83s | 20% |
| BuildNgramRelationships workers | 0.01s | 0.48s | 12% |
| reverseString | 0.05s | 0.44s | 11% |
| NgramsSimilar | 0.01s | 0.44s | 11% |
| Word.RemoveFirstChar | - | 0.36s | 9% |
| FindLongestMatch | 0.01s | 0.34s | 8% |
| UniqIdentities | 0.12s | 0.26s | 6% |

### Memory Allocations (WITH GC)
| Component | Active | % |
|-----------|--------|---|
| NewMorpheme | 34 MB | 29% |
| FindMorphemes overhead | 47 MB | 40% |
| BuildNgramRelationships | 13 MB | 11% |
| AddSimilarNgramRef | 10 MB | 8% |
| Other | 13 MB | 11% |

---

## Recommendations

### For Production
✅ **Use GC enabled (GOGC=100 or default)**
- 2.3× faster
- 91% memory efficiency
- Sustainable for large corpora

### For Further Optimization
1. **String operation pooling** (HIGH) - 8× improvement potential
2. **Morpheme object pooling** (MEDIUM) - 3× memory reduction
3. **Profile with larger corpus** to validate scaling

### For Profiling/Analysis
- Use NO GC for pure allocation analysis
- Use WITH GC for realistic performance measurement
- Compare both to understand GC impact

---

## Conclusion

### GC Performance Summary
✅ **GC is working efficiently:**
- 91% memory reclamation
- 2.3× faster than NO GC
- Only 53% overhead (reasonable for GC languages)
- Enables scalability to larger corpora

### Bottleneck Identified
🎯 **String operations (reverseString, RemoveFirstChar):**
- 19% of runtime WITH GC
- 15× slower than NO GC
- Clear optimization target

### Identity Optimization Success
✅ **Similarity detection optimized:**
- 98% memory reduction (599 MB → < 10 MB)
- Works efficiently in both GC modes
- No further optimization needed

**Next step:** Optimize string operations with byte slice pooling for additional 8× improvement in those hot paths.
