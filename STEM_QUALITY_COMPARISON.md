# STEM Quality Comparison: C vs Go

## Executive Summary

**Go version has both MORE and HIGHER QUALITY STEMs than C.**

| Metric | C Version | Go Version | Winner |
|--------|-----------|------------|--------|
| Total STEMs | 104 | 125 | Go (+21) |
| Valid common words | 32 | 51 | Go (+19) |
| Valid % | 30.8% | 40.8% | Go (+10%) |

## Detailed Analysis

### C Version (104 STEMs)

**Valid common words: 32/104 = 30.8%**

Examples: a, about, after, all, an, and, are, as, at, be, by, do, for, he, I, in, is, it, me, my, no, of, on, one, or, our, out, so, the, to, up, was, we

**Quality issues:**
- Capitalization duplicates: For, It, The (should be lowercase)
- Proper names (NOT linguistic stems): Mars, Ogilvy, Ottershaw
- Many domain-specific or technical terms counted as STEMs

### Go Version (125 STEMs)

**Valid common words: 51/125 = 40.8%**

All 32 C valid words PLUS 19 additional valid words: am, but, end, had, his, how, if, its, men, more, pit, their, too, use, very, way, went, where, and more

**Quality issues:**
- Same capitalization duplicates as C: For, It, The
- Same proper names as C: Mars, Ogilvy, Ottershaw
- **One additional invalid**: `night,` (with comma - parsing error)

### Go's 21 Extra STEMs

**Validity: 20/21 = 95.2%**

**Valid (20):**
am, and, but, end, had, his, how, if, its, men, more, pit, possible, their, too, use, very, way, went, where

**Invalid (1):**
- `night,` - Has trailing comma, should be filtered out

## Key Findings

1. **Go finds 60% MORE valid common word stems** (51 vs 32)
2. **Go's extra STEMs are 95.2% valid** - almost all are legitimate English words
3. **Both versions share quality issues**: capitalization duplicates and proper names
4. **Go has one unique parsing bug**: `night,` with comma

## Recommendation

**Go version is superior for STEM extraction:**
- Finds significantly more valid linguistic stems
- Higher percentage of valid STEMs overall
- The one parsing bug (comma in `night,`) is trivial to fix

**Fix needed for Go:**
Filter out morphemes with punctuation characters (non-alphanumeric):
```go
if strings.ContainsAny(morpheme.Str, ",.!?;:") {
    continue  // Skip invalid morphemes with punctuation
}
```

## Linguistic Correctness

The 19 additional valid STEMs that Go finds are all high-frequency English words:
- Verbs: am, had, went, use
- Conjunctions: and, but, if
- Determiners: their
- Adverbs: how, very, way, where, too
- Nouns: men, end, pit
- Adjective: possible

These are core English vocabulary items that SHOULD be identified as stems. Go's algorithm is correctly identifying them while C is missing them.
