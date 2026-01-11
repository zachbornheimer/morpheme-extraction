# Morpheme Extraction: Optimization Roadmap

## Phase 1A: String Interning (Immediate - 40% memory reduction)

### Problem
- Same words repeated thousands of times
- Each creates new `[]byte` allocation
- `string(word.Str[:word.Len])` happens on every operation

### Solution: String Pool
```go
// internal/stringpool/stringpool.go
package stringpool

import "sync"

var (
    pool = make(map[string]string)
    mu   sync.RWMutex
)

// Intern returns a canonical copy of the string
// If string already exists, returns existing reference
func Intern(s string) string {
    mu.RLock()
    if canonical, exists := pool[s]; exists {
        mu.RUnlock()
        return canonical
    }
    mu.RUnlock()

    mu.Lock()
    defer mu.Unlock()
    // Double-check after acquiring write lock
    if canonical, exists := pool[s]; exists {
        return canonical
    }
    pool[s] = s
    return s
}

// Stats returns pool statistics
func Stats() (size int, memoryBytes int64) {
    mu.RLock()
    defer mu.RUnlock()
    size = len(pool)
    for k := range pool {
        memoryBytes += int64(len(k))
    }
    return
}
```

### Usage in structs.go
```go
// Update Word.String() to use interning
func (w *Word) String() string {
    return stringpool.Intern(string(w.Str[:w.Len]))
}

// Update NewWord to intern immediately
func NewWord(w string) Word {
    interned := stringpool.Intern(w)
    return Word{
        Len:  len(interned),
        Freq: 1,
        Str:  append([]byte(interned), 0),
    }
}
```

**Expected**: 40% memory reduction (128MB → 77MB for War of the Worlds)

---

## Phase 1B: Replace Linear Search (Immediate - 20% CPU reduction)

### Problem (functions.go:81)
```go
// O(n) linear search for EVERY comparison
if i == 0 || InCharArray(arr[j], u, i) == -1 {
    u[i] = arr[j]
    i++
}
```

### Solution: Use Map from Start
```go
// Replace entire UniqWords function
func UniqWords(arr []string, size int) int {
    if size == 0 {
        return -1
    }

    // Always use map - simpler and O(n)
    seen := structs.NewStringSet(size)
    for j := 0; j < size; j++ {
        seen.Add(arr[j])
    }
    return seen.Len() - 1
}
```

**Expected**: 20% CPU reduction (4s → 3.2s)

---

## Phase 1C: Batch Allocations (Immediate - 30MB memory reduction)

### Problem (nlp.go:236)
```go
// Every AddMorpheme appends, causing slice growth
morpheme_list_t.AddMorpheme(&internal, forward)
```

### Solution: Pre-allocate with Capacity
```go
// In FindMorphemes, estimate capacity
func FindMorphemes(...) int {
    // Pre-allocate based on expected morpheme count
    // Heuristic: ~10 morphemes per ngram reference
    estimatedMorphemes := 0
    for i := 0; i < ngramLength; i++ {
        if ng[i] != nil {
            estimatedMorphemes += ng[i].RefsCount * 10
        }
    }

    internal := structs.MorphemeList{
        Count: 0,
        List:  make([]*structs.Morpheme, 0, estimatedMorphemes),
    }

    // ... rest of function
}
```

**Expected**: 30MB reduction in morpheme list allocations

---

## Phase 2A: Ngram Hashing (Week 2 - 50% CPU reduction)

### Problem
- Comparing every ngram pair: O(N²)
- Most comparisons yield "not similar"

### Solution: Hash-Based Bucketing
```go
// internal/ngramindex/ngramindex.go
package ngramindex

import (
    "hash/fnv"
    "github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

type NgramIndex struct {
    buckets map[uint64][]*structs.Ngram
}

func NewNgramIndex(capacity int) *NgramIndex {
    return &NgramIndex{
        buckets: make(map[uint64][]*structs.Ngram, capacity/10),
    }
}

// Hash computes hash from context words
func (idx *NgramIndex) Hash(ng *structs.Ngram, ngramHalfSize int) uint64 {
    h := fnv.New64a()

    // Hash context words in sorted order for consistency
    words := make([]string, 0, 10)
    words = ng.Before.ToSlice(ngramHalfSize, words)
    words = ng.After.ToSlice(ngramHalfSize, words)

    // Sort for consistent hashing
    sort.Strings(words)

    for _, w := range words {
        h.Write([]byte(w))
    }

    return h.Sum64()
}

func (idx *NgramIndex) Add(ng *structs.Ngram, ngramHalfSize int) {
    hash := idx.Hash(ng, ngramHalfSize)
    idx.buckets[hash] = append(idx.buckets[hash], ng)
}

// GetSimilar returns ngrams in same bucket (likely similar)
func (idx *NgramIndex) GetSimilar(ng *structs.Ngram, ngramHalfSize int) []*structs.Ngram {
    hash := idx.Hash(ng, ngramHalfSize)
    return idx.buckets[hash]
}
```

### Usage in nlp.go
```go
func BuildNgramRelationships(...) {
    // ... existing ngram creation code ...

    // NEW: Build index
    index := ngramindex.NewNgramIndex(ngramLength)
    for i := 0; i < ngramLength; i++ {
        if ng[i] != nil {
            index.Add(ng[i], ngramHalfSize)
        }
    }

    // NEW: Compare only within buckets
    for i := 0; i < ngramLength; i++ {
        if ng[i] == nil {
            continue
        }

        // Only compare with ngrams in same bucket
        candidates := index.GetSimilar(ng[i], ngramHalfSize)

        for _, candidate := range candidates {
            if ng[i] == candidate {
                continue
            }
            if ngram_t.NgramsSimilar(*ng[i], *candidate) {
                ngram_t.AddSimilarNgramRef(&ng[i], &candidate)
            }
        }
    }
}
```

**Expected**: 50% CPU reduction (comparisons: N² → N log N)

---

## Phase 2B: Worker Pool (Week 3 - 4x CPU speedup)

### Solution: Parallel Morpheme Extraction
```go
// internal/workers/morpheme_worker.go
package workers

type MorphemeJob struct {
    Ngram  *structs.Ngram
    Target *structs.Ngram
}

type MorphemeResult struct {
    Forward  structs.Morpheme
    Internal []structs.Morpheme
}

func MorphemeWorker(
    jobs <-chan MorphemeJob,
    results chan<- MorphemeResult,
) {
    for job := range jobs {
        // Extract morphemes
        forward := morpheme_t.FindLongestMatch(job.Ngram.Word, job.Target.Word)
        forward.WordsCount = 0
        forward.Words = make([]*structs.Word, 0, 2)

        w1 := job.Ngram.Word.String()
        w2 := job.Target.Word.String()
        morpheme_t.GenRegex(w1, w2, 0, 0, &forward)
        morpheme_t.AddWord(&forward, job.Ngram.Word)
        morpheme_t.AddWord(&forward, job.Target.Word)

        // Find internal morphemes
        internal := structs.MorphemeList{Count: 0}
        morpheme_list_t.FindInternalMorphemes(
            job.Ngram.Word,
            job.Target.Word,
            &internal,
        )

        results <- MorphemeResult{
            Forward:  forward,
            Internal: internal.List,
        }
    }
}
```

### Usage in nlp.go
```go
func FindMorphemes(...) int {
    numWorkers := runtime.NumCPU()
    jobs := make(chan workers.MorphemeJob, 1000)
    results := make(chan workers.MorphemeResult, 1000)

    // Start workers
    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            workers.MorphemeWorker(jobs, results)
        }()
    }

    // Collector goroutine
    internal := structs.MorphemeList{Count: 0}
    var collectorWg sync.WaitGroup
    collectorWg.Add(1)
    go func() {
        defer collectorWg.Done()
        for result := range results {
            morpheme_list_t.AddMorpheme(&internal, result.Forward)
            for _, m := range result.Internal {
                morpheme_list_t.AddMorpheme(&internal, m)
            }
        }
    }()

    // Send jobs
    for i := 0; i < ngramLength; i++ {
        if ng[i] == nil {
            continue
        }
        for j := 0; j < ng[i].RefsCount; j++ {
            jobs <- workers.MorphemeJob{
                Ngram:  ng[i],
                Target: ng[i].Refs[j],
            }
        }
    }

    close(jobs)
    wg.Wait()
    close(results)
    collectorWg.Wait()

    // Rest of function...
}
```

**Expected**: 4x speedup on M1 (8 performance cores)

---

## Phase 3: Streaming + Disk Storage (Month 2)

### Problem
- Cannot fit trillion operations in memory
- Need persistent intermediate results

### Solution: SQLite Backend
```go
// internal/storage/morpheme_db.go
package storage

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type MorphemeDB struct {
    db *sql.DB
}

func NewMorphemeDB(path string) (*MorphemeDB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

    // Create schema
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS morphemes (
            id INTEGER PRIMARY KEY,
            morpheme TEXT NOT NULL,
            type INTEGER NOT NULL,
            freq INTEGER NOT NULL,
            regex TEXT,
            UNIQUE(morpheme, type)
        );

        CREATE INDEX IF NOT EXISTS idx_morpheme ON morphemes(morpheme);
        CREATE INDEX IF NOT EXISTS idx_freq ON morphemes(freq DESC);
    `)

    return &MorphemeDB{db: db}, err
}

func (db *MorphemeDB) AddMorpheme(m *structs.Morpheme) error {
    _, err := db.db.Exec(`
        INSERT INTO morphemes (morpheme, type, freq, regex)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(morpheme, type) DO UPDATE SET
            freq = freq + excluded.freq
    `, m.Morpheme, m.Type, m.Freq, m.Regex)

    return err
}

func (db *MorphemeDB) StreamMorphemes() <-chan *structs.Morpheme {
    ch := make(chan *structs.Morpheme, 100)

    go func() {
        defer close(ch)

        rows, _ := db.db.Query(`
            SELECT morpheme, type, freq, regex
            FROM morphemes
            ORDER BY freq DESC
        `)
        defer rows.Close()

        for rows.Next() {
            m := &structs.Morpheme{}
            rows.Scan(&m.Morpheme, &m.Type, &m.Freq, &m.Regex)
            ch <- m
        }
    }()

    return ch
}
```

**Expected**: Process unlimited corpus size with fixed memory (4GB)

---

## Implementation Timeline

### Week 1 (Immediate ROI)
- [ ] Day 1-2: String interning (40% memory)
- [ ] Day 3: Replace linear search (20% CPU)
- [ ] Day 4-5: Batch allocations (30MB memory)
- **Expected**: 60% memory reduction, 20% CPU reduction

### Week 2 (Algorithmic)
- [ ] Day 1-3: Ngram hashing/indexing (50% CPU)
- [ ] Day 4-5: Testing & tuning
- **Expected**: 70% total CPU reduction

### Week 3 (Concurrency)
- [ ] Day 1-3: Worker pool implementation (4x CPU)
- [ ] Day 4-5: Lock-free data structures
- **Expected**: 80% total CPU reduction

### Month 2 (Scale)
- [ ] Week 1-2: SQLite backend
- [ ] Week 3: Streaming corpus reader
- [ ] Week 4: Incremental processing
- **Expected**: Process 1M+ words

---

## Success Metrics

| Metric | Current | Week 1 | Week 3 | Month 2 |
|--------|---------|--------|--------|---------|
| **Memory (War of Worlds)** | 320MB | 128MB | 128MB | 64MB |
| **CPU Time** | 4.0s | 3.2s | 0.8s | 1.2s |
| **Max Corpus Size** | 10K words | 20K | 200K | 10M+ |
| **Throughput** | 500 words/s | 625 w/s | 2500 w/s | 100K w/s |

---

## Validation

After each phase, run:
```bash
# Correctness
./nlp-go test-corpus/
diff languageModel.zdrm languageModel-good.zdrm

# Performance
time ./nlp-go test-corpus/

# Memory
./nlp-go -memprofile=mem.prof test-corpus/
go tool pprof -top mem.prof

# Stress test
./nlp-go large-corpus/  # 10x larger
```

---

## Next Steps

1. **Start with Phase 1A** (string interning) - Biggest immediate win
2. **Measure before/after** - Use pprof to validate improvements
3. **Iterate quickly** - Small, tested changes
4. **Keep algorithm correct** - Verify output matches baseline

Ready to scale to trillions! 🚀
