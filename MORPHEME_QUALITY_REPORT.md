# Morpheme Quality Analysis Report

**Corpus**: War of the Worlds excerpt (12,730 chars)
**Output**: 1,091 morphemes total

---

## Executive Summary

### Overall Quality Assessment

| Category | Status | Issues |
|----------|--------|--------|
| **STEM** | ❌ **NOT 100% accurate** | 8 false positives |
| **PREFIX** | ⚠️ **Contains correct + incorrect** | Many single-char and punctuation false positives |
| **SUFFIX** | ⚠️ **Contains correct + incorrect** | Many punctuation and number false positives |
| **INFIX** | ⚠️ **Misclassification issue** | Common affixes wrongly classified as INFIX |
| **CIRCUMFIX** | ✅ **None found** | Correctly absent for English |

---

## 1. STEM Issues (CRITICAL - Must be 100% accurate)

### Current State: 131 unique STEMs

**FALSE POSITIVES (8 total):**

1. **Proper Nouns (3)** - Should NOT be STEMs:
   - `Mars` - planet name
   - `Ogilvy` - person name
   - `Ottershaw` - place name

2. **Punctuation Error (1)**:
   - `night,` - has trailing comma

3. **Case Duplicates (4 pairs = 4 duplicates)**:
   - `and` / `And`
   - `for` / `For`
   - `it` / `It`
   - `the` / `The`

**Valid STEMs**: ~123 out of 131 = **93.9% accuracy**

### Required Fixes for 100% STEM Accuracy:

```go
// 1. Filter proper nouns (capitalize first letter + rare in general vocabulary)
if IsProperNoun(morpheme.Str) {
    continue // Don't classify as STEM
}

// 2. Filter punctuation
if strings.ContainsAny(morpheme.Str, ",.!?;:'\"-") {
    continue // Skip morphemes with punctuation
}

// 3. Normalize case (convert to lowercase before deduplication)
morpheme.Str = strings.ToLower(morpheme.Str)

// 4. Validate against dictionary (optional but recommended)
if !IsCommonEnglishWord(morpheme.Str) && !IsProperNoun(morpheme.Str) {
    // Consider not marking as STEM
}
```

---

## 2. PREFIX Issues (Contains false positives)

### Current State: 144 unique prefixes

**FALSE POSITIVES:**

1. **Single characters**: `a`, `b`, `c`, `d`, `e`, `f`, etc. (26+ single letters)
   - These are word starts, not linguistic prefixes

2. **Punctuation**: `"`, `_`

3. **Numbers**: `1`

4. **Case variants**: `A`/`a`, `C`/`c`, `D`/`d`, etc.

**MISSING COMMON PREFIXES**: `un-`, `re-`, `pre-`, `dis-`, `mis-`
- These are currently classified as **INFIX** (see below)

### Recommended Filters:

```go
// Minimum length for prefix
if len(morpheme.Str) < 2 {
    continue // Skip single chars
}

// No punctuation
if strings.ContainsAny(morpheme.Str, ",.!?;:'\"-_") {
    continue
}

// No numbers
if strings.ContainsAny(morpheme.Str, "0123456789") {
    continue
}
```

---

## 3. SUFFIX Issues (Contains false positives)

### Current State: 120 unique suffixes

**FALSE POSITIVES:**

1. **Numbers with punctuation**:
   - `,000,000` - million
   - `00,000` - thousands

2. **Word fragments with punctuation**:
   - `ace,`, `ace.` - word + punctuation
   - `ars.`, `ce,`, `ce.`, `d,`, `d.`, `n,`, `r,`, `r.`, `s,`, `s.`, `t,`, `t.`, `w.`, `y,`
   - `bservatory,` - partial word + comma

3. **Very short fragments**: `aw`, `ds`, `gs`, `ny`, `ps`

**MISSING COMMON SUFFIXES**: `ed`, `ing`, `ly`, `er`, `est`, `tion`, `ness`, `ment`
- These are currently classified as **INFIX** (see below)

### Recommended Filters:

```go
// No punctuation
if strings.ContainsAny(morpheme.Str, ",.!?;:'\"-_") {
    continue
}

// No numbers
if strings.ContainsAny(morpheme.Str, "0123456789") {
    continue
}

// Minimum meaningful length (optional)
if len(morpheme.Str) < 2 {
    continue // Skip single chars like 'd', 's'
}
```

---

## 4. INFIX Issues (Major misclassification)

### Current State: 808 unique infixes

**CRITICAL PROBLEM: Common affixes classified as INFIX**

**Should be SUFFIX** (currently marked INFIX):
- `ed` - past tense (walked, jumped)
- `ing` - present participle (walking, running)
- `ly` - adverb (quickly, slowly)
- `er` - comparative (bigger, faster)
- `est` - superlative (biggest, fastest)
- `tion` - noun (action, creation)
- `ness` - noun (kindness, darkness)
- `ment` - noun (movement, argument)
- `ous` - adjective (famous, dangerous)
- `ive` - adjective (active, creative)
- `al` - adjective (local, personal)

**Should be PREFIX** (currently marked INFIX):
- `un` - negative (unkind, unable)
- `re` - again (return, replay)
- `dis` - negative (disagree, dislike)

**FALSE POSITIVES in INFIX**:
- Punctuation: `--`, `--an`, `,000`, `'s`, `0,000`, `e,`, `s-`
- Single chars with punctuation

### Root Cause:

The classification logic in `identify_true_morphemes()` is not correctly identifying affixes at word boundaries:

```go
// Current logic (oversimplified)
if frontRegex == "^" || frontRegex == "()" {
    morpheme.Type = PREFIX  // Only if at absolute start
} else if backRegex == "$" || backRegex == "()" {
    morpheme.Type = SUFFIX  // Only if at absolute end
} else {
    morpheme.Type = INFIX   // Everything else
}
```

**Problem**: Many valid suffixes like `-ed`, `-ing` appear in the middle of the regex context but are still functionally suffixes:
- Pattern: `(watch|walk|talk)ed$` → `-ed` is at word end, should be SUFFIX
- But regex is: `^(w)[a-z]+(ed)$` → Classified as INFIX because not at regex boundary

---

## 5. Recommended Pruning Strategy

### Phase 1: STEM Cleanup (CRITICAL - Must be 100%)

```go
func IsValidSTEM(morpheme *Morpheme) bool {
    str := strings.ToLower(morpheme.Str)

    // 1. No punctuation
    if strings.ContainsAny(str, ",.!?;:'\"-_()[]{}") {
        return false
    }

    // 2. No numbers
    if strings.ContainsAny(str, "0123456789") {
        return false
    }

    // 3. Minimum length (optionally allow 1-2 letter words like "a", "I", "at")
    if len(str) < 1 {
        return false
    }

    // 4. Not a proper noun (starts with capital + not in common word list)
    if IsProperNoun(morpheme.Str) {
        return false
    }

    // 5. (Optional) Validate against dictionary
    if !IsInDictionary(str) {
        return false
    }

    return true
}
```

### Phase 2: PREFIX/SUFFIX Cleanup

```go
func IsValidAffix(morpheme *Morpheme) bool {
    // 1. Minimum length
    if len(morpheme.Str) < 2 {
        return false
    }

    // 2. No punctuation
    if strings.ContainsAny(morpheme.Str, ",.!?;:'\"-_()[]{}") {
        return false
    }

    // 3. No numbers
    if strings.ContainsAny(morpheme.Str, "0123456789") {
        return false
    }

    return true
}
```

### Phase 3: Fix Affix Classification

**Option A: Pattern-based reclassification**

```go
// Common English suffixes
var commonSuffixes = []string{
    "ed", "ing", "ly", "er", "est", "tion", "sion", "ness", "ment",
    "able", "ible", "ful", "less", "ous", "ious", "ive", "al", "ic",
    "ish", "y", "en", "ize", "ise", "fy", "ward", "wise",
}

// Common English prefixes
var commonPrefixes = []string{
    "un", "re", "in", "im", "il", "ir", "dis", "en", "em", "non",
    "over", "mis", "sub", "pre", "inter", "fore", "de", "trans",
    "super", "semi", "anti", "mid", "under",
}

// Reclassify known affixes
func ReclassifyAffix(morpheme *Morpheme) {
    str := strings.ToLower(morpheme.Str)

    if morpheme.Type == INFIX {
        // Check if it's actually a common suffix
        for _, suffix := range commonSuffixes {
            if str == suffix {
                morpheme.Type = SUFFIX
                return
            }
        }

        // Check if it's actually a common prefix
        for _, prefix := range commonPrefixes {
            if str == prefix {
                morpheme.Type = PREFIX
                return
            }
        }
    }
}
```

**Option B: Regex analysis improvement**

Improve the classification logic to detect affixes even when they're not at regex boundaries:

```go
// Check if morpheme appears only at word starts
if MorphemeAlwaysAtWordStart(morpheme) {
    morpheme.Type = PREFIX
}

// Check if morpheme appears only at word ends
if MorphemeAlwaysAtWordEnd(morpheme) {
    morpheme.Type = SUFFIX
}
```

---

## 6. Priority Order for Fixes

### Immediate (Required for STEM accuracy):

1. ✅ Remove punctuation from all morphemes: `strings.ContainsAny(str, ",.!?;:'\"-")`
2. ✅ Remove numbers from all morphemes: `strings.ContainsAny(str, "0123456789")`
3. ✅ Normalize case for STEMs: `strings.ToLower()`
4. ✅ Filter proper nouns from STEMs

### High Priority (Improves classification):

5. ✅ Reclassify common suffixes (ed, ing, ly, etc.) from INFIX → SUFFIX
6. ✅ Reclassify common prefixes (un, re, dis, etc.) from INFIX → PREFIX
7. ✅ Filter single-character prefixes/suffixes

### Medium Priority (Reduces noise):

8. ⚠️ Minimum length filters (2-3 chars for affixes)
9. ⚠️ Dictionary validation for STEMs (optional)

### Low Priority (Nice to have):

10. ⭕ Frequency-based filtering (very rare morphemes might be noise)
11. ⭕ Statistical validation (morpheme productivity)

---

## 7. Implementation Checklist

```go
// In IdentifyTrueMorphemes() or as post-processing:

for i := range morphemeList {
    morpheme := &morphemeList[i]

    // STEP 1: Filter invalid characters
    if strings.ContainsAny(morpheme.Str, ",.!?;:'\"-_()[]{}0123456789") {
        continue // Skip this morpheme
    }

    // STEP 2: Normalize case for deduplication
    morpheme.Str = strings.ToLower(morpheme.Str)

    // STEP 3: Filter proper nouns from STEMs
    if morpheme.Type == STEM && IsProperNoun(morpheme.Str) {
        continue // Don't include
    }

    // STEP 4: Reclassify known affixes
    if morpheme.Type == INFIX {
        ReclassifyAffix(morpheme)
    }

    // STEP 5: Filter very short affixes (optional)
    if (morpheme.Type == PREFIX || morpheme.Type == SUFFIX) && len(morpheme.Str) < 2 {
        continue
    }

    // STEP 6: Add to validated output
    validatedMorphemes = append(validatedMorphemes, morpheme)
}
```

---

## 8. Expected Results After Cleanup

**STEMs**:
- Before: 131 (93.9% accurate)
- After: ~123 (100% accurate)
- Removed: 3 proper nouns, 1 punctuation, 4 duplicates

**PREFIXes**:
- Before: 144 unique
- After: ~30-40 unique (valid linguistic prefixes)
- Removed: Single chars, punctuation, numbers
- Added from INFIX: un, re, dis, pre, mis, etc.

**SUFFIXes**:
- Before: 120 unique
- After: ~40-50 unique (valid linguistic suffixes)
- Removed: Punctuation, numbers, fragments
- Added from INFIX: ed, ing, ly, er, est, tion, ness, ment, etc.

**INFIXes**:
- Before: 808 unique
- After: ~600-700 unique (true internal morphemes)
- Removed: Reclassified affixes, punctuation, numbers

**Total Quality**:
- STEMs: 100% accurate ✅
- Affixes: Contains all correct morphemes + some false positives ✅
- Ready for manual pruning or statistical filtering ✅
