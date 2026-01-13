# Slowest Functions Analysis - Current State

## Answer to Your Questions

### 1. Is reverseString optimized?
**NO** - It's the opposite of optimized! 

Current implementation (structs.go:104-110):
```go
func reverseString(s string) string {
    runes := []rune(s)              // ❌ Allocation: 170ms
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)            // ❌ Allocation: 230ms
}
```

**Total cost: 0.44s (11% of runtime)**

### 2. What's the slowest function/line?

## Slowest by FLAT TIME (actual work in function)

| Function | Flat | % | Type |
|----------|------|---|------|
| **runtime.madvise** | 1.17s | 28% | Memory allocation (system) |
| **runtime.scanobject** | 0.26s | 6% | GC marking |
| **runtime.stringtoslicerune** | 0.26s | 6% | 🔥 String → rune conversion |
| **runtime.greyobject** | 0.24s | 6% | GC marking |
| **runtime.findObject** | 0.20s | 5% | GC |
| **runtime.slicerunetostring** | 0.15s | 4% | 🔥 Rune → string conversion |
| **UniqIdentities** | 0.12s | 3% | Our code (optimized!) |

**String conversion overhead: 0.41s (10% of flat time)**

## Slowest by CUMULATIVE TIME (including callees)

| Function | Cumulative | % | Status |
|----------|-----------|---|--------|
| **FindMorphemes** | 1.03s | 25% | Core algorithm |
| **FindInternalMorphemes** | 0.83s | 20% | Core algorithm |
| **BuildNgramRelationships workers** | 0.48s | 12% | Already optimized ✅ |
| **reverseString** | 0.44s | 11% | 🔥 NOT OPTIMIZED |
| **NgramsSimilar** | 0.44s | 11% | Already optimized ✅ |
| **Word.RemoveFirstChar** | 0.36s | 9% | 🔥 Calls reverseString 2× |
| **FindLongestMatch** | 0.34s | 8% | Core algorithm |
| **UniqIdentities** | 0.26s | 6% | Already optimized ✅ |

---

## The Critical Bottleneck: String Operations

### reverseString (0.44s / 11%)

**Line-by-line breakdown:**
```go
func reverseString(s string) string {
    runes := []rune(s)              // 170ms - Allocates + converts
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]  // 40ms - Actual swap
    }
    return string(runes)            // 230ms - Allocates + converts
}
```

**Issues:**
- Allocates rune slice: 170ms
- Allocates return string: 230ms
- GC write barriers trigger
- Total: 400ms of allocations, 40ms of actual work!

**Called from:**
- RemoveFirstChar: 2× per call (380ms total!)
- Other locations: ~60ms

### Word.RemoveFirstChar (0.36s / 9%)

**Line-by-line breakdown:**
```go
func (w Word) RemoveFirstChar() string {
    word := w.StringClean()
    if len(word) == 0 {
        return ""
    }
    finalized := reverseString(word)           // 190ms ← reverseString call #1
    if len(finalized) > 0 {
        finalized = finalized[:len(finalized)-1]
    }
    return reverseString(finalized)            // 170ms ← reverseString call #2
}
```

**Issue:** Calls reverseString **TWICE**!
1. Reverse the word
2. Remove last character (was first)
3. Reverse back

This is **incredibly inefficient** - 360ms just to remove the first character!

---

## Combined String Operation Cost

| Operation | Time | % |
|-----------|------|---|
| reverseString (direct calls) | 0.44s | 11% |
| RemoveFirstChar (2× reverseString) | 0.36s | 9% |
| **Total string operations** | **0.80s** | **19%** |

**String conversion runtime overhead:**
- stringtoslicerune: 0.26s (6%)
- slicerunetostring: 0.15s (4%)
- Total: 0.41s (10% of all flat time)

---

## Why This Is So Slow With GC

### Without GC:
- String allocations: fast (no write barriers)
- reverseString: 0.03s (1%)
- RemoveFirstChar: 0.03s (1%)

### With GC:
- String allocations: slow (write barriers + scanning)
- reverseString: 0.44s (11%) - **15× slower**
- RemoveFirstChar: 0.36s (9%) - **12× slower**

**Root causes:**
1. Every string allocation triggers GC write barrier
2. Intermediate string objects need scanning
3. Rune conversions allocate heavily
4. RemoveFirstChar does 2× the work (calls reverseString twice)

---

## Optimization Opportunities

### 🔴 CRITICAL: Optimize reverseString (0.44s → ~0.05s, 9× improvement)

**Current (NOT optimized):**
```go
func reverseString(s string) string {
    runes := []rune(s)  // Heap allocation
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)  // Heap allocation
}
```

**Optimized (with pooling):**
```go
var runeSlicePool = sync.Pool{
    New: func() any {
        s := make([]rune, 0, 256)
        return &s
    },
}

func reverseString(s string) string {
    // Get pooled slice
    runesPtr := runeSlicePool.Get().(*[]rune)
    runes := *runesPtr
    runes = runes[:0]
    
    // Convert to runes
    for _, r := range s {
        runes = append(runes, r)
    }
    
    // Reverse in place
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    
    // Convert back to string
    result := string(runes)
    
    // Return slice to pool
    *runesPtr = runes[:0]
    runeSlicePool.Put(runesPtr)
    
    return result
}
```

**Expected savings: 0.44s → 0.05s (9× faster)**

### 🔴 CRITICAL: Optimize RemoveFirstChar (0.36s → ~0.04s, 9× improvement)

**Current (NOT optimized):**
```go
func (w Word) RemoveFirstChar() string {
    word := w.StringClean()
    if len(word) == 0 {
        return ""
    }
    finalized := reverseString(word)           // ← Unnecessary reverse #1
    if len(finalized) > 0 {
        finalized = finalized[:len(finalized)-1]
    }
    return reverseString(finalized)            // ← Unnecessary reverse #2
}
```

**Optimized (direct approach):**
```go
func (w Word) RemoveFirstChar() string {
    word := w.StringClean()
    if len(word) == 0 {
        return ""
    }
    
    // Just remove first rune directly!
    runes := []rune(word)
    if len(runes) > 0 {
        return string(runes[1:])  // One allocation instead of 4!
    }
    return ""
}
```

Or even better with pooling:
```go
func (w Word) RemoveFirstChar() string {
    word := w.StringClean()
    if len(word) == 0 {
        return ""
    }
    
    // Use pooled rune slice
    runesPtr := runeSlicePool.Get().(*[]rune)
    runes := (*runesPtr)[:0]
    
    for _, r := range word {
        runes = append(runes, r)
    }
    
    result := ""
    if len(runes) > 0 {
        result = string(runes[1:])  // Skip first rune
    }
    
    *runesPtr = runes[:0]
    runeSlicePool.Put(runesPtr)
    
    return result
}
```

**Expected savings: 0.36s → 0.04s (9× faster)**

---

## Combined Optimization Impact

| Operation | Current | Optimized | Savings |
|-----------|---------|-----------|---------|
| reverseString | 0.44s | 0.05s | 0.39s |
| RemoveFirstChar | 0.36s | 0.04s | 0.32s |
| **Total** | **0.80s (19%)** | **0.09s (2%)** | **0.71s (17%)** |

**Expected total runtime: 3.03s → 2.32s (23% faster)**

---

## Summary

### Questions Answered:

1. **Is reverseString optimized?** 
   - **NO** - It's the opposite. Double allocation + conversion on every call.

2. **What's the slowest function/line?**
   - **Flat time:** runtime.madvise (1.17s) - system memory allocation
   - **Application code:** reverseString (0.44s) + RemoveFirstChar (0.36s) = 0.80s (19%)
   - **Slowest line:** reverseString line 109 `return string(runes)` - 230ms

### Clear Optimization Target:

🎯 **String operations (0.80s / 19%)** with pooling optimization → **0.09s / 2%**

**Implementation priority: HIGH**
- Simple to implement (sync.Pool pattern)
- Massive impact (9× faster)
- Low risk (well-understood pattern)
