# Profiling Results - Go Performance Analysis

## Executive Summary

For a 12KB corpus (2,211 words), the Go version allocates **2.66GB** of memory!

### Performance Breakdown
- **With GC**: 5.48s (C is 2.45s = 2.24x slower)
- **Without GC**: 4.31s (C is 2.45s = 1.76x slower)
- **GC Overhead**: 21% of runtime

### The Real Problem: Excessive Allocations

**GC isn't slow - we're just allocating way too much!**

---

## Memory Allocation Hotspots

| Function | Memory | % of Total | Issue |
|----------|--------|------------|-------|
| **NgramsSimilar** | 1377MB | 51.8% | Creates temp array on EVERY call |
| **UniqWords** | 544MB | 20.5% | O(n²) + creates temp array |
| **FindInternalMorphemes** | 578MB | 21.7% | String operations |
| **RemoveFirstChar** | 333MB | 12.5% | Creates new string every call |
| **Reverse** | 202MB | 7.6% | Creates new string every call |

---

## CPU Time (Without GC)

| Operation | Time | % |
|-----------|------|---|
| **Memory allocation** | 3.15s | 85% |
| FindMorphemes | 0.31s | 8.4% |
| FuseRegex | 0.19s | 5.2% |
| NgramsSimilar | 0.17s | 4.6% |
| UniqWords | 0.05s | 1.4% |

**85% of time is spent allocating memory!**

---

## Root Causes

### 1. **NgramsSimilar Creates Temp Arrays** (1377MB)

```go
func NgramsSimilar(a, b structs.Ngram) bool {
    arr := make([]string, elemsCount)  // 🔥 NEW ALLOCATION EVERY CALL!
    
    // Collect all words into arr...
    count := functions.UniqWords(arr[:arrCount], arrCount)
    // ...
}
```

**Called thousands of times, creates 1.4GB of garbage!**

### 2. **UniqWords is O(n²) with Temp Arrays** (544MB)

```go
func UniqWords(arr []string, size int) int {
    u := make([]string, size)  // 🔥 NEW ALLOCATION!
    
    for j := 0; j < size; j++ {
        if i == 0 || InCharArray(arr[j], u, i) == -1 {  // 🔥 O(n) search!
            u[i] = arr[j]
            i++
        }
    }
    return i - 1
}
```

**O(n²) time complexity + allocates temp array every call.**

### 3. **String Operations Create Copies** (535MB)

```go
// RemoveFirstChar: 333MB
func RemoveFirstChar(w Word) string {
    return string(w.Str[1:])  // 🔥 Creates new string!
}

// Reverse: 202MB
func Reverse(input string) string {
    // 🔥 Creates new string!
}
```

**Called repeatedly in loops, creates hundreds of MB of garbage.**

### 4. **Repeated string(byteSlice) Conversions**

```go
word := string(a.Before.At[i].Elems[j].Str)  // 🔥 Allocates!
```

**Every conversion allocates a new string. Called thousands of times.**

---

## Optimization Strategy

### Priority 1: Reduce Allocations (Target: 90% reduction)

#### Fix 1: Reuse Buffers in NgramsSimilar
```go
// Use sync.Pool or preallocated buffer
var stringPool = sync.Pool{
    New: func() interface{} {
        s := make([]string, 0, 1000)
        return &s
    },
}

func NgramsSimilar(a, b structs.Ngram) bool {
    arr := stringPool.Get().(*[]string)
    defer func() {
        *arr = (*arr)[:0]  // Reset
        stringPool.Put(arr)
    }()
    // Use *arr...
}
```

**Expected savings: 1377MB → ~0MB**

#### Fix 2: Use Map for UniqWords (O(n) + No Allocation)
```go
func UniqWords(arr []string, size int) int {
    seen := make(map[string]struct{}, size)
    for j := 0; j < size; j++ {
        seen[arr[j]] = struct{}{}
    }
    return len(seen)
}
```

**Expected savings: 544MB → ~50MB (just the map)**

#### Fix 3: Avoid String Copies
```go
// Work with byte slices directly, no conversions
// Or use string interning/caching
```

**Expected savings: 535MB → ~50MB**

### Priority 2: Algorithmic Improvements

#### Fix InCharArray (Already O(n), but called in O(n²) loop)
Use map for O(1) lookups in UniqWords.

---

## Expected Performance After Fixes

| Metric | Before | After (Estimated) |
|--------|--------|-------------------|
| **Memory Allocated** | 2.66GB | ~200MB (92% reduction) |
| **GC Overhead** | 21% (1.17s) | ~2% (0.1s) |
| **Runtime with GC** | 5.48s | **~3.2s** |
| **vs C (2.45s)** | 2.24x slower | **1.31x slower** |

### Optimistic Target
With perfect allocation reduction:
- Runtime: ~2.8s (1.14x slower than C)
- This would be **excellent** for a memory-safe rewrite

---

## Implementation Plan

1. ✅ Add profiling support (cpu, mem)
2. ✅ Identify bottlenecks
3. **Next: Fix NgramsSimilar allocations** (51.8% of memory!)
4. **Next: Fix UniqWords to use map** (20.5% of memory + O(n²) fix)
5. **Next: Reduce string copies** (12.5% + 7.6% of memory)
6. Benchmark and verify
7. Consider parallelization ONLY after allocations are fixed

---

## Key Lessons

1. **GC is not the enemy** - excessive allocations are
2. **Profile before optimizing** - we found the real bottlenecks
3. **Memory matters** - 2.66GB for 12KB is insane
4. **Go 1.25 GC is good** - 21% overhead is reasonable IF we don't abuse it
5. **Parallelization is premature** - fix allocations first

---

## Tools Used

```bash
# CPU profiling
GOGC=off ./morpheme-extract -c test-corpus/ -cpuprofile cpu.prof
go tool pprof -top -cum cpu.prof

# Memory profiling  
./morpheme-extract -c test-corpus/ -memprofile mem.prof
go tool pprof -top -alloc_space mem.prof

# Compare GC on/off
time GOGC=off ./morpheme-extract ...
time ./morpheme-extract ...
```

