# C vs Go Morpheme Extraction Comparison

## Executive Summary

After reverting C to the original algorithm (with only the minimal STEM fix), the two versions produce **significantly different results**:

| Metric | C Version | Go Version | Difference |
|--------|-----------|------------|------------|
| **Total morphemes** | 506 | 1,091 | +585 (+116%) |
| **Unique morphemes** | 504 | 1,089 | +585 (+116%) |
| **Shared morphemes** | 449 | 449 | - |
| **Unique to version** | 55 | 640 | - |

**Go finds 2.16× more morphemes than C.**

---

## Morpheme Type Classification

### Overall Distribution

| Type | C Count | C % | Go Count | Go % | Change |
|------|---------|-----|----------|------|--------|
| STEM | 104 | 20.6% | 125 | 11.5% | +21 (+20%) |
| PREFIX | 214 | 42.3% | 129 | 11.8% | -85 (-40%) |
| SUFFIX | 7 | 1.4% | 113 | 10.4% | +106 (+1514%) |
| INFIX | 178 | 35.2% | 721 | 66.1% | +543 (+305%) |
| **Total** | **503** | **100%** | **1,088** | **100%** | **+585** |

### Key Classification Differences

1. **C classifies heavily as PREFIX** (42.3%)
   - Most morphemes starting at word boundaries → PREFIX

2. **Go classifies heavily as INFIX** (66.1%)
   - Most morphemes found internally → INFIX

3. **Go finds 16× more SUFFIXes** (113 vs 7)
   - C appears to under-classify suffixes

---

## Morpheme Discovery Patterns

### Morphemes Go Finds But C Doesn't (640 morphemes)

**Categories:**

1. **Punctuation/numbers** (10-15%):
   - `,000`, `'s`, `00`, `000`, `0,000`, `00,000`
   - `ace,`, `ace.`, `--an`

2. **Very short fragments** 2-3 chars (30-40%):
   - `ai`, `ak`, `ami`, `ani`, `ape`, `ard`, `ail`, `aily`
   - Most are linguistically questionable

3. **Common affixes** (20-30%):
   - `able`, `ance`, `ances`, `and`, `ange`, `ant`, `ants`
   - `tion`, `ness`, `ment`, `ing`
   - These are valid linguistic morphemes

4. **Small-offset internal morphemes** (20-30%):
   - `lanet` (planet-1), `elescope` (telescope-1)
   - `ntelligence` (intelligence-1)

**Examples:**
```
Go-only: 's, 00, 000, ai, ak, able, ace., and, ange, anger, ant, ants,
         ape, aper, elescope, lanet, ntelligence
```

### Morphemes C Finds But Go Doesn't (55 morphemes)

**Pattern**: Internal fragments with **larger offsets** from word start

**Examples:**
```
C-only: anet (planet-2), lescope (telescope-2), lligence (intelligence-3),
        credibl (incredible-3), dnight (midnight-2), istence (existence-2),
        member (remember-2), ntur (century-2)
```

All are fragments at offset 2+ from the beginning of longer words.

---

## Root Cause Analysis

### Different Offset Discovery Behavior

For the same word pairs, C and Go find morphemes at **different offsets**:

| Word | C Finds | Go Finds | Pattern |
|------|---------|----------|---------|
| planet | `anet` (offset 2) | `lanet` (offset 1) | Go: smaller offset |
| telescope | `lescope` (offset 2) | `elescope` (offset 1) | Go: smaller offset |
| intelligence | `lligence` (offset 3) | `ntelligence` (offset 1) | Go: smaller offset |

### Hypothesis

The **lookback logic** in `FindInternalMorphemes` behaves differently:

1. **C behavior**: Lookback allows discovery at larger offsets
2. **Go behavior**: Lookback filters more aggressively, favoring smaller offsets
3. **Effect**: Once a morpheme is found at offset 1, subsequent checks skip offset 2+
4. **Result**: Cascading difference in which morphemes are discovered

**Code location**: `morpheme_list_t.c:39-66` vs `internal/morpheme_list_t/morpheme_list_t.go:153-196`

The lookback comparison logic:
```c
// C version
if (strcmp(find_longest_match(current, previous).morpheme, lm.morpheme) != 0) {
    // Complex lookback test with reverse and truncation
    // ...
}
```

This logic is identical in both versions, but subtle differences in string handling or iteration order may cause divergence.

---

## Quality Assessment

### Which is "Better"?

**Go's advantages:**
- ✅ Finds 60% more valid STEMs (51 vs 32 common words)
- ✅ Identifies more valid affixes (`able`, `tion`, `ness`, `ment`)
- ✅ More thorough discovery (1,089 vs 504 unique morphemes)
- ✅ Better SUFFIX detection (113 vs 7)

**C's advantages:**
- ✅ Finds longer internal fragments (more linguistically interesting?)
- ✅ Fewer spurious punctuation morphemes
- ✅ Less "noise" from 2-character fragments

**Go's disadvantages:**
- ❌ Includes punctuation (`,000`, `'s`, `00`) - should be filtered
- ❌ Many 2-3 character fragments of questionable linguistic value
- ❌ Classifies too many things as INFIX (66% vs 35%)

**C's disadvantages:**
- ❌ Misses valid affixes that Go finds
- ❌ Misses valid STEMs that Go finds
- ❌ Very poor SUFFIX detection (only 7 found)
- ❌ Over-classifies as PREFIX (42% vs 12%)

---

## Recommendations

### For Production Use

**Use Go version** as the base, with these filters:

1. **Filter punctuation morphemes**:
   ```go
   if strings.ContainsAny(morpheme.Str, ",.!?;:'-") {
       continue  // Skip morphemes with punctuation
   }
   ```

2. **Filter very short morphemes** (optional):
   ```go
   if len(morpheme.Str) < 3 && morpheme.Type != STEM {
       continue  // Keep short STEMs but filter short affixes
   }
   ```

3. **Review PREFIX/INFIX classification**:
   - C's 42% PREFIX seems too high
   - Go's 66% INFIX seems too high
   - Consider adjusting thresholds in `identify_true_morphemes`

### For Research

Both versions provide complementary insights:
- **C**: Conservative, finds longer fragments
- **Go**: Aggressive, finds more candidates

Running both and taking the intersection would provide high-confidence morphemes.

---

## Next Steps

1. **Investigate lookback logic divergence** - Why do identical algorithms produce different offsets?
2. **Add punctuation filtering** to Go version
3. **Review classification thresholds** - Why such different PREFIX/INFIX ratios?
4. **Consider hybrid approach** - Use Go's thoroughness + C's conservative filtering

---

## Test Output Files

- `c-output.zdrm`: C version output (506 lines)
- `go-output.zy`: Go version output (1,091 lines)
- `c-morphemes.txt`: 504 unique C morphemes
- `go-morphemes.txt`: 1,089 unique Go morphemes
