# Code Analysis API - Implementation Plan

**Status:** 📋 Planning  
**Target:** Post-validation of linguistic version  
**Estimated Effort:** 6-8 weeks  
**Priority:** Medium (after linguistic version is validated)

---

## 🎯 **Vision**

Extend the morpheme extraction algorithm to analyze code identifiers, enabling:

1. **Pattern Detection** - Auto-discover naming conventions in codebases
2. **Smart Search** - Find functions by structural patterns + semantics
3. **Convention Auditing** - Detect naming inconsistencies
4. **Refactoring Support** - Suggest renames based on detected patterns
5. **Hybrid Search** - Combine with embeddings for better code search

---

## 🏗️ **Architecture**

### **High-Level Structure**

```
github.com/zachbornheimer/morpheme-go/
├── linguistic/                    # Current implementation (Phase 1 ✅)
│   ├── pkg/
│   │   ├── config/
│   │   ├── corpus/
│   │   ├── ngram/
│   │   └── morpheme/
│   ├── cmd/morpheme-extract/
│   └── README.md
│
├── code/                          # New: Code analysis (Phase 2-5)
│   ├── pkg/
│   │   ├── extractor/             # Core pattern extraction
│   │   ├── pattern/               # PatternMatcher API
│   │   ├── index/                 # Codebase indexing
│   │   ├── convention/            # Convention analysis
│   │   └── composition/           # Identifier decomposition
│   ├── cmd/code-analyze/
│   └── README.md
│
├── shared/                        # Shared between linguistic + code
│   ├── types/
│   └── algorithms/
│
└── examples/
    ├── linguistic-examples/
    ├── code-examples/
    └── auditfix-integration/
```

### **Module Strategy**

**Option A: Monorepo** (Recommended)

```
morpheme-extraction/
├── go.work                        # Go workspace
├── linguistic/go.mod
└── code/go.mod
```

**Option B: Separate Repos**

```
github.com/zachbornheimer/morpheme-extraction  (linguistic)
github.com/zachbornheimer/morpheme-code        (code analysis)
```

**Recommendation:** Start with Option A (monorepo), split if code analysis takes off independently.

---

## 📦 **Phase Breakdown**

### **Phase 1: Foundation** ✅ COMPLETE

_Already implemented in go-rewrite branch_

- [x] Core morpheme extraction algorithm
- [x] N-gram building and similarity detection
- [x] Morpheme classification (PREFIX/SUFFIX/STEM/etc)
- [x] CLI for processing text files
- [x] Basic tests and documentation

**Deliverable:** `morpheme-extract` CLI working on natural language

---

### **Phase 2: Code Pattern Extraction** 🎯 NEXT (2 weeks)

**Goal:** Adapt linguistic algorithm for code identifiers

#### **2.1: Core Extractor** (Week 1)

```go
// code/pkg/extractor/extractor.go
package extractor

type CodeExtractor struct {
    linguistic *linguistic.Extractor
    splitter   IdentifierSplitter
}

// Extract patterns from code identifiers
func (e *CodeExtractor) Extract(identifiers []string) (*Result, error)

// Result holds detected patterns
type Result struct {
    Patterns      []Pattern
    Morphemes     []Morpheme
    Vocabulary    map[string]int
    ProcessedAt   time.Time
}

type Pattern struct {
    Template   string      // "get*ById"
    Morphemes  []string    // ["get", "*", "ById"]
    Examples   []string    // ["getUserById", "getPostById"]
    Frequency  int         // How many identifiers match
    Confidence float64     // 0-1
}
```

**Tasks:**

- [ ] Implement `IdentifierSplitter` (camelCase, snake_case, PascalCase)
- [ ] Adapt n-gram algorithm for code tokens
- [ ] Convert linguistic morphemes to code patterns
- [ ] Handle case sensitivity properly
- [ ] Add unit tests for pattern detection

**Test Cases:**

```go
Input: ["getUserById", "getPostById", "fetchUserByEmail"]
Expected:
  - Pattern: "[get|fetch]*By*" (3 matches)
  - Morphemes: "get", "fetch", "User", "Post", "By", "Id", "Email"
```

#### **2.2: Pattern Templates** (Week 2)

```go
// code/pkg/pattern/template.go
package pattern

type Template struct {
    Parts []Part
}

type Part struct {
    Type     PartType  // FIXED, VARIABLE, CHOICE
    Value    string    // "get", "*", "[get|fetch]"
    Optional bool
}

// Parse pattern string into template
func Parse(pattern string) (*Template, error)

// Examples:
Parse("get*ById")          // get + [variable] + ById
Parse("[get|fetch]*By*")   // [get OR fetch] + [var] + By + [var]
Parse("*User?")            // [var] + User + [optional]
```

**Tasks:**

- [ ] Implement pattern parser
- [ ] Support wildcards (`*`)
- [ ] Support choices (`[a|b]`)
- [ ] Support optionals (`?`)
- [ ] Add validation

**Deliverable:** CLI that shows detected patterns

```bash
$ code-analyze ./src/**/*.go --show-patterns

Detected Patterns (by frequency):
  1. get*By* (45 occurrences)
     Examples: getUserById, getPostBySlug, getCommentByID

  2. create* (23 occurrences)
     Examples: createUser, createPost, createSession

  3. update*From* (12 occurrences)
     Examples: updateUserFromRequest, updatePostFromInput
```

---

### **Phase 3: Pattern Matching & Search** 🔮 (2 weeks)

**Goal:** Enable querying code by patterns

#### **3.1: Pattern Matcher API** (Week 3)

```go
// code/pkg/pattern/matcher.go
package pattern

type Matcher interface {
    Match(identifier string) bool
    Score(identifier string) float64
    Explain(identifier string) *Explanation
}

func NewMatcher(pattern string) (Matcher, error)

type Explanation struct {
    Matched    bool
    Score      float64
    Reason     string
    Morphemes  []MatchedMorpheme
}

type MatchedMorpheme struct {
    Text     string
    Type     MorphemeType
    Position int
}
```

**Example Usage:**

```go
matcher, _ := pattern.NewMatcher("get*By*")

matcher.Match("getUserById")        // true
matcher.Match("fetchUserByEmail")   // false (doesn't start with "get")

score := matcher.Score("getUserById")  // 1.0 (perfect match)

explain := matcher.Explain("getUserById")
// explain.Reason: "Matches: get + User + By + Id"
```

**Tasks:**

- [ ] Implement fuzzy matching (80%+ = match)
- [ ] Score calculation (exact > partial > fuzzy)
- [ ] Explanation generation
- [ ] Performance optimization (compile patterns)

#### **3.2: Index API** (Week 4)

```go
// code/pkg/index/index.go
package index

type Index interface {
    // Build index
    Add(filepath string, identifiers []Identifier) error
    Build() error

    // Query
    FindByPattern(pattern string) ([]Match, error)
    FindByMorpheme(morpheme string) ([]Match, error)
    FindByPrefix(prefix string) ([]Match, error)
    FindBySuffix(suffix string) ([]Match, error)

    // Analysis
    DetectPatterns() []DetectedPattern
    DetectInconsistencies() []Inconsistency
}

type Match struct {
    Identifier   string
    FilePath     string
    Line         int
    Score        float64
    Morphemes    []Morpheme
}

type DetectedPattern struct {
    Pattern    string
    Template   *pattern.Template
    Matches    []Match
    Frequency  int
    Confidence float64
}

type Inconsistency struct {
    Identifier       string
    ExpectedPattern  string
    ActualPattern    string
    Suggestion       string
    Severity         Severity  // HIGH, MEDIUM, LOW
}
```

**Example Usage:**

```go
// Build index
idx := index.New()
idx.Add("user.go", []Identifier{
    {Name: "GetUserById", Line: 45},
    {Name: "GetPostById", Line: 78},
})
idx.Build()

// Query
matches, _ := idx.FindByPattern("get*By*")
// Returns: GetUserById, GetPostById

// Detect patterns
patterns := idx.DetectPatterns()
// Returns: "[get*By*]" with 2 matches

// Find issues
issues := idx.DetectInconsistencies()
// Returns: "fetchUser doesn't match dominant pattern 'get*'"
```

**Tasks:**

- [ ] In-memory index implementation
- [ ] Efficient lookup (trie or inverted index)
- [ ] Pattern frequency tracking
- [ ] Inconsistency detection algorithm
- [ ] Parallel indexing support

**Deliverable:** Search CLI

```bash
$ code-analyze search "get*By*" --codebase ./src/

Found 23 matches:

src/user/service.go:45
  func GetUserById(id string) (*User, error)
  Pattern: get + User + By + Id
  Score: 1.00

src/post/service.go:78
  func GetPostById(id string) (*Post, error)
  Pattern: get + Post + By + Id
  Score: 1.00
```

---

### **Phase 4: Convention Analysis** 🔍 (1 week)

**Goal:** Auto-detect and enforce naming conventions

#### **4.1: Convention Analyzer** (Week 5)

```go
// code/pkg/convention/analyzer.go
package convention

type Analyzer interface {
    Analyze(identifiers []Identifier) Convention
    Validate(identifier Identifier, convention Convention) *Violation
    Suggest(identifier Identifier, convention Convention) string
}

type Convention struct {
    Name         string
    Description  string
    Pattern      *pattern.Template
    Examples     []string
    Frequency    int
    Confidence   float64
}

type Violation struct {
    Identifier  string
    Expected    string
    Actual      string
    Suggestion  string
    Reason      string
    Severity    Severity
}
```

**Example Usage:**

```go
analyzer := convention.NewAnalyzer()

// Analyze codebase
identifiers := extractFromCodebase("./src")
conv := analyzer.Analyze(identifiers)

// Result:
// Convention{
//   Name: "Action-Entity-Criteria",
//   Pattern: "[verb]*By*",
//   Examples: ["getUserById", "getPostBySlug"],
//   Frequency: 45,
//   Confidence: 0.87,
// }

// Check compliance
violation := analyzer.Validate(
    Identifier{Name: "fetchUser"},
    conv,
)

// violation.Suggestion = "getUserById"
// violation.Reason = "Doesn't match dominant pattern '[verb]*By*'"
```

**Tasks:**

- [ ] Pattern clustering algorithm
- [ ] Convention detection (find most common patterns)
- [ ] Rename suggestion generation
- [ ] Severity scoring
- [ ] Multi-convention support (different patterns for different contexts)

**Deliverable:** Audit command

```bash
$ code-analyze audit ./src/ --auto-detect

Detected Convention: "Action-Entity-Criteria"
  Pattern: [get|fetch|create|update|delete]*By*
  Confidence: 87%
  Examples: getUserById, getPostBySlug

Violations Found: 12

HIGH: fetchUser (line 45 in user/service.go)
  ✗ Doesn't follow pattern
  ✓ Suggestion: getUserById

MEDIUM: UserGetter (line 78 in user/handlers.go)
  ✗ Uses noun instead of verb
  ✓ Suggestion: GetUser
```

---

### **Phase 5: Composition & Smart Search** 🧠 (2 weeks)

**Goal:** Understand identifier composition and enable natural language queries

#### **5.1: Composition Analysis** (Week 6)

```go
// code/pkg/composition/composer.go
package composition

type Composer interface {
    Decompose(identifier string) (*Composition, error)
    Compose(parts ...Part) (string, error)
}

type Composition struct {
    Prefix   *Morpheme
    Stems    []*Morpheme
    Suffix   *Morpheme

    // Semantic types (inferred)
    Action   string  // "get", "create", "update"
    Entity   string  // "User", "Post"
    Criteria string  // "ById", "ByEmail"
}

func NewComposer(idx *index.Index) Composer
```

**Example Usage:**

```go
composer := composition.NewComposer(idx)

comp, _ := composer.Decompose("getUserById")

// Result:
// Composition{
//   Prefix: {Text: "get", Type: ACTION},
//   Stems: [{Text: "User", Type: ENTITY}],
//   Suffix: {Text: "ById", Type: CRITERIA},
//   Action: "get",
//   Entity: "User",
//   Criteria: "ById",
// }

// Find similar
similar := idx.FindByComposition(comp)
// Returns: getPostById, getUserByEmail, fetchUserById
```

**Tasks:**

- [ ] Identifier tokenization (camelCase, snake_case)
- [ ] Morpheme type classification (ACTION, ENTITY, CRITERIA)
- [ ] Composition templates
- [ ] Similarity scoring

#### **5.2: Smart Search** (Week 7)

```go
// code/pkg/search/smart.go
package search

type SmartSearch interface {
    Search(query string) ([]Result, error)
    Explain(query string, result Result) (*Explanation, error)
}

type Result struct {
    Match      index.Match
    Score      float64
    MatchType  MatchType  // EXACT, STRUCTURAL, SEMANTIC, HYBRID
}

type MatchType int
const (
    ExactMatch      MatchType = iota
    StructuralMatch
    SemanticMatch
    HybridMatch
)
```

**Example Usage:**

```go
search := search.NewSmartSearch(idx, embeddings)

// Natural language query
results, _ := search.Search("functions that get users by their id")

// Finds (in order of relevance):
// 1. getUserById         (score: 0.95, type: HYBRID)
// 2. fetchUserById       (score: 0.88, type: HYBRID)
// 3. getUserByID         (score: 0.85, type: STRUCTURAL)
// 4. findUserById        (score: 0.78, type: SEMANTIC)

explain := search.Explain(query, results[0])
// explain.Reason:
//   "Structural match: [get] + User + By + Id
//    Semantic match: 'get' ≈ 'retrieve', 'users' ≈ 'User'
//    Combined score: 0.95"
```

**Tasks:**

- [ ] Natural language query parsing
- [ ] Morpheme extraction from queries
- [ ] Scoring algorithm (structural + semantic)
- [ ] Result ranking
- [ ] Explanation generation

**Deliverable:** Smart search command

```bash
$ code-analyze smart-search "functions that create new users"

Top 5 Results:

1. CreateUser (score: 0.94, hybrid match)
   src/user/service.go:123
   Why: Structural: [create] + User
        Semantic: "create" ≈ "new", "users" ≈ "User"

2. CreateUserAccount (score: 0.88, hybrid match)
   src/user/service.go:156
   Why: Structural: [create] + User + Account
        Semantic: "create" ≈ "new"

3. NewUser (score: 0.75, semantic match)
   src/user/model.go:45
   Why: Semantic: "new" ≈ "create", "users" ≈ "User"
```

---

### **Phase 6: Auditfix Integration** 🔌 (1 week)

**Goal:** Integrate with auditfix for hybrid code search

#### **6.1: Integration API** (Week 8)

```go
// auditfix integration
package auditfix

import (
    "github.com/zachbornheimer/morpheme-go/code/pkg/index"
    "github.com/zachbornheimer/morpheme-go/code/pkg/search"
)

// Extend ChromemIndex
type ChromemIndex struct {
    // Existing fields
    embeddings *vectorIndex

    // New: Morpheme index
    morphemes index.Index
    search    search.SmartSearch
}

// Enhanced search options
type SearchOptions struct {
    // Existing semantic options
    Query             string
    MaxResults        int

    // New: Structural options
    StructuralPattern string    // "get*ById"
    RequiredMorphemes []string  // ["get", "User"]
    ExcludedMorphemes []string  // ["All"]

    // Weighting
    StructuralWeight  float64   // 0-1, default 0.5
    SemanticWeight    float64   // 0-1, default 0.5
}

// Hybrid search
func (idx *ChromemIndex) Search(opts SearchOptions) []SearchResult {
    // Pure semantic search
    if opts.StructuralPattern == "" {
        return idx.semanticSearch(opts)
    }

    // Hybrid: combine structural + semantic
    return idx.hybridSearch(opts)
}

func (idx *ChromemIndex) hybridSearch(opts SearchOptions) []SearchResult {
    // Get candidates from both methods
    structuralResults := idx.morphemes.FindByPattern(opts.StructuralPattern)
    semanticResults := idx.semanticSearch(opts)

    // Merge and re-rank
    merged := mergeResults(structuralResults, semanticResults, opts)

    return rankByScore(merged)
}
```

**Tasks:**

- [ ] Add morpheme index to ChromemIndex
- [ ] Implement hybrid scoring
- [ ] Update SearchOptions
- [ ] Add CLI flags for structural search
- [ ] Integration tests

**Deliverable:** Enhanced auditfix search

```bash
# Before (pure semantic)
$ auditfix search "get user by id"
  → Returns all functions with similar embeddings
  → May include unrelated functions

# After (hybrid)
$ auditfix search "get user by id" --pattern "get*By*"
  → Only returns functions matching pattern
  → Ranked by semantic + structural similarity
  → More precise results
```

---

## 🧪 **Testing Strategy**

### **Unit Tests**

- [ ] Pattern parsing and matching
- [ ] Identifier splitting (camelCase, snake_case, etc.)
- [ ] Morpheme extraction from code
- [ ] Convention detection
- [ ] Composition analysis

### **Integration Tests**

- [ ] Index building from real codebases
- [ ] Search accuracy on known patterns
- [ ] Auditfix integration end-to-end

### **Benchmarks**

- [ ] Index build time (10K, 100K, 1M identifiers)
- [ ] Search latency (<10ms for simple queries)
- [ ] Memory usage (reasonable for large codebases)

### **Real-World Validation**

Test on actual projects:

- [ ] Auditfix codebase (Go)
- [ ] This morpheme-extraction project (Go)
- [ ] Popular open source (kubernetes, docker, etc.)

---

## 📊 **Success Metrics**

### **Phase 2-3 (Pattern Extraction & Search)**

- [ ] Detects 80%+ of actual naming patterns
- [ ] Search returns relevant results in top 10
- [ ] <100ms query latency

### **Phase 4 (Convention Analysis)**

- [ ] 90%+ accuracy in detecting dominant convention
- [ ] Suggestions accepted by developers 70%+ of time
- [ ] <5% false positives

### **Phase 5-6 (Smart Search & Integration)**

- [ ] Hybrid search beats pure semantic 20%+ in precision
- [ ] Natural language queries work 80%+ of time
- [ ] Integrates with auditfix without breaking changes

---

## 🚧 **Technical Challenges & Mitigations**

### **Challenge 1: Case Sensitivity**

**Problem:** Code is case-sensitive but morphemes are case-agnostic

**Solutions:**

- Normalize to lowercase for extraction
- Preserve original case in index
- Smart matching (getUserById ≈ GetUserByID)

### **Challenge 2: Abbreviations**

**Problem:** Code uses abbreviations (ID, URL, API)

**Solutions:**

- Abbreviation dictionary
- Context-aware splitting
- User-configurable mappings

### **Challenge 3: Multi-Language Support**

**Problem:** Different languages, different conventions

**Solutions:**

- Language-specific splitters
- Convention profiles per language
- Auto-detection from file extension

### **Challenge 4: Performance at Scale**

**Problem:** Large codebases (100K+ identifiers)

**Solutions:**

- Incremental indexing
- Trie-based storage
- Parallel processing
- Caching

### **Challenge 5: Semantic Type Inference**

**Problem:** Distinguishing ACTION vs ENTITY vs CRITERIA

**Solutions:**

- Heuristics (verbs vs nouns)
- Position-based (first = action, middle = entity, last = criteria)
- Optional embedding-based classification
- User annotations

---

## 📅 **Timeline Summary**

| Phase                   | Duration | Effort | Dependencies | Deliverable              |
| ----------------------- | -------- | ------ | ------------ | ------------------------ |
| 1. Foundation           | DONE     | -      | None         | `morpheme-extract` CLI   |
| 2. Pattern Extraction   | 2 weeks  | 60h    | Phase 1      | `code-analyze` CLI       |
| 3. Matching & Search    | 2 weeks  | 70h    | Phase 2      | Pattern search           |
| 4. Convention Analysis  | 1 week   | 35h    | Phase 3      | Audit command            |
| 5. Smart Search         | 2 weeks  | 65h    | Phase 3-4    | Natural language queries |
| 6. Auditfix Integration | 1 week   | 30h    | Phase 3-5    | Hybrid search            |

**Total:** 8 weeks, ~260 hours (full-time) or 16 weeks (half-time)

---

## 🎯 **Minimum Viable Product (MVP)**

If time is limited, start with just **Phases 2-3** (4 weeks):

### **MVP Scope**

1. **Extract patterns from code** (`code-analyze extract`)
2. **Search by pattern** (`code-analyze search "get*By*"`)
3. **Basic CLI** (no auditfix integration yet)

### **MVP API**

```go
// Just 3 functions
func Extract(identifiers []string) (*Result, error)
func NewMatcher(pattern string) (Matcher, error)
func (m Matcher) Match(identifier string) bool
```

**This alone would be useful** for:

- Understanding codebase conventions
- Finding similar functions
- Refactoring assistance

---

## 🔄 **Future Enhancements (Beyond Phase 6)**

### **Phase 7: Advanced Features**

- [ ] Multi-language support (Python, JavaScript, Rust)
- [ ] IDE plugins (VSCode, IntelliJ)
- [ ] Git history analysis (detect convention drift)
- [ ] Team convention profiles
- [ ] Auto-refactoring suggestions

### **Phase 8: Machine Learning**

- [ ] Learn conventions from commits
- [ ] Predict best identifier names
- [ ] Personalized suggestions per developer

### **Phase 9: Semantic Phase 2+ (Linguistic Fusion)**

- [ ] Lexical categorization (implement your 2014 paper Phase 2)
- [ ] Syntax modeling (Phase 3)
- [ ] Semantic mapping (Phase 4)
- [ ] Production/generation (Phase 5)

---

## 📚 **Documentation Plan**

### **User Documentation**

- [ ] Getting started guide
- [ ] Pattern syntax reference
- [ ] CLI command reference
- [ ] Auditfix integration guide
- [ ] FAQ

### **Developer Documentation**

- [ ] Architecture overview
- [ ] API reference (GoDoc)
- [ ] Contributing guide
- [ ] Algorithm explanation
- [ ] Performance tuning guide

### **Research Documentation**

- [ ] Comparison: code vs linguistic morphology
- [ ] Evaluation on real codebases
- [ ] Benchmark results
- [ ] Academic paper (optional)

---

## 🎬 **Getting Started (When Ready)**

### **Step 1: Validate Linguistic Version**

Before building code analysis:

1. Test morpheme-extract on multiple languages
2. Validate accuracy of morpheme detection
3. Get feedback from linguists/NLP researchers
4. Fix any bugs or issues

### **Step 2: Prototype Code Extraction**

Quick 2-day spike:

1. Extract identifiers from Go files
2. Run through linguistic extractor
3. See if results make sense
4. Decide if approach is viable

### **Step 3: Full Development**

If prototype works:

1. Follow phases 2-6 above
2. Iterate based on feedback
3. Integrate with auditfix
4. Public release

---

## 💡 **Key Design Decisions**

### **Decision 1: Separate or Unified?**

**Recommendation:** Start unified (monorepo), split if needed

**Rationale:**

- Shared algorithm benefits both
- Easier to maintain initially
- Can split later if code analysis takes off

### **Decision 2: API Design Philosophy**

**Recommendation:** Simple by default, powerful when needed

**Pattern:**

```go
// Simple (covers 80% of cases)
result, _ := code.Extract(identifiers)

// Powerful (remaining 20%)
result, _ := code.Extract(identifiers,
    code.WithSplitter(customSplitter),
    code.WithMinFrequency(5),
    code.WithPatternWeights(map[string]float64{...}),
)
```

### **Decision 3: Performance vs Features**

**Recommendation:** Start simple, optimize later

**Strategy:**

1. Make it work (naive implementation)
2. Make it right (clean API)
3. Make it fast (optimize hot paths)

Don't pre-optimize unless profiling shows bottleneck.

---

## 📝 **Notes & Considerations**

### **Linguistic Fidelity**

The code analysis uses the same core algorithm as linguistic extraction, but:

- Code identifiers are "synthetic language" (constructed, not natural)
- Patterns are more regular than natural language
- May need different thresholds (experiment!)

### **Cross-Language Challenges**

Different languages have different conventions:

- Go: `GetUserById`
- Python: `get_user_by_id`
- JavaScript: `getUserById`
- Rust: `get_user_by_id`

Need language-aware splitting and pattern normalization.

### **Integration Philosophy**

Auditfix integration should be:

- **Optional** - Works standalone
- **Non-breaking** - Existing searches still work
- **Additive** - New features don't remove old ones
- **Configurable** - Can tune structural vs semantic weights

---

## ✅ **Checklist Before Starting**

When ready to begin, verify:

- [ ] Linguistic version is stable and tested
- [ ] API design reviewed and approved
- [ ] Architecture decisions documented
- [ ] Test strategy defined
- [ ] Success metrics agreed upon
- [ ] Integration plan with auditfix clear
- [ ] Timeline realistic
- [ ] Resources available (time, focus)

---

## 🚀 **Final Thoughts**

This is an ambitious project that extends your 2013 research into a new domain (code analysis). The core insight—that patterns in output reveal underlying structure—applies equally well to code as to natural language.

**The beauty:** Same algorithm, two applications:

1. **Linguistic** → Understand human language
2. **Code** → Understand developer language

**The opportunity:** No other tool does both. This could be unique.

**The risk:** Might be over-engineering. Validate linguistic version first.

**When to start:** After linguistic version proven + auditfix showing need for better search.

---

**Status:** 📋 Plan complete, waiting for Phase 1 validation

**Next Review:** After 1 month of using linguistic version

**Decision Point:** Prototype code analysis if linguistic version successful

---

_Plan created: January 8, 2025_  
_Review by: February 8, 2025_  
_Implementation target: Q2 2025 (if approved)_
