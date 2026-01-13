# Morpheme Output Differences: C vs Go

## Summary Statistics

| Metric | C Version | Go Version | Difference |
|--------|-----------|------------|------------|
| Total morphemes | 503 | 1,088 | +585 (+116%) |
| Shared morphemes | 448 | 448 | - |
| Unique to version | 55 | 640 | - |

**Go finds 2.16× more morphemes than C**

## Pattern Analysis

### Morphemes Go Finds But C Doesn't (640 morphemes)

**Categories:**
1. **Punctuation/numbers**: `,000`, `'s`, `00`, `000`, `ace,`, `ace.`
2. **Short fragments (2-3 chars)**: `ai`, `ak`, `ami`, `ani`, `ape`, `ard`
3. **Common affixes**: `able`, `ance`, `ances`, `and`, `ange`, `ant`, `ants`
4. **Words with small offsets**: `lanet` (planet-1), `elescope` (telescope-1)

**Examples:**
```
Go-only: 's, ai, ak, able, ace., and, ange, anger, ant, ants, ape, aper
```

### Morphemes C Finds But Go Doesn't (55 morphemes)

**Pattern**: Internal fragments with **larger offsets** from word start

**Examples:**
```
C-only: anet (planet-2), lescope (telescope-2), credibl (incredible-3),
        dnight (midnight-2), istence (existence-2), lligence (intelligence-3)
```

## Key Finding: Different Offset Behavior

For the same words, C and Go find morphemes at **different offsets**:

| Word | C Finds | Go Finds | Pattern |
|------|---------|----------|---------|
| planet | `anet` (offset 2) | `lanet` (offset 1) | Go: smaller offset |
| telescope | `lescope` (offset 2) | `elescope` (offset 1) | Go: smaller offset |

**Hypothesis**: The regex validation added to FindInternalMorphemes in Go is filtering
morphemes differently than C, which affects the lookback logic and subsequent
morpheme discovery in the same word pair.

## Root Cause

When I optimized FindInternalMorphemes to check regex validity BEFORE adding candidates
(to avoid allocations), this changed the filtering behavior:

1. **C behavior**: Adds all morphemes first, validates regex later
2. **Go behavior** (after optimization): Validates regex immediately, only adds valid ones
3. **Effect**: Different morphemes pass validation, affecting lookback comparisons
4. **Result**: Different morphemes discovered at different offsets

## Impact Assessment

**Which is more correct?**
- Go finds more morphemes (1,088 vs 503) - more thorough?
- C finds longer internal fragments - more linguistically meaningful?
- Both have systematic patterns, neither is obviously "wrong"

The question is: should we prioritize **finding more morphemes** (Go) or **finding
more linguistically significant fragments** (C)?

## Next Steps

1. Review whether early regex validation is correct
2. Consider removing regex check from FindInternalMorphemes (match C exactly)
3. Or investigate why regex validation produces different results for different offsets
4. Verify which output is more linguistically correct
