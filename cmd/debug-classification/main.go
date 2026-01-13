package main

import (
	"fmt"
	"strings"

	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/directory"
	"github.com/zachbornheimer/morpheme-extraction/internal/alphabet"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/morpheme_list_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/ngram_t"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

func main() {
	cfg := &config.Config{
		VerboseMode:        constants.OFF,
		ProcessSequentially: constants.OFF,
		ProcessFull:        constants.OFF,
		OutputFilename:     "debug-output.zdrm",
		CorpusDir:          "./test-corpus/",
	}

	directory.CorpusDir = cfg.CorpusDir

	var f string
	var header string
	index := 0
	ng := []*structs.Ngram{}
	ngramLength := 0

	// Process corpus
	for {
		f, header = directory.Getfiles(&index)
		if f == "" {
			break
		}

		if len(f) > 2 {
			wd, err := alphabet.FindWordDelimiter(&f, cfg)
			if err == nil && wd != "" {
				ngram_t.BuildNgramRelationships(wd, f, &ngramLength, &ng, cfg)
			}
		}
	}

	// Find morphemes
	internal := structs.MorphemeList{Count: 0}
	internal.List = make([]*structs.Morpheme, 0, ngramLength*100)

	for i := 0; i < ngramLength; i++ {
		ngram := ng[i]
		for j := 0; j < ngram.RefsCount; j++ {
			target := ngram.Refs[j]

			// Forward morpheme
			forward := functions.FindLongestMatchString(ngram.Word.StringClean(), target.Word.StringClean())
			if forward != "" && len(forward) >= 2 {
				morpheme := &structs.Morpheme{
					Morpheme:   forward,
					WordsCount: 0,
					Words:      make([]structs.Word, 0),
				}
				morpheme.GenRegex(ngram.Word.StringClean(), target.Word.StringClean(), 0, 0)
				morpheme.AddWord(ngram.Word)
				morpheme.AddWord(target.Word)
				internal.List = append(internal.List, morpheme)
				internal.Count++
			}

			// Internal morphemes
			morpheme_list_t.FindInternalMorphemes(ngram.Word, target.Word, &internal)
		}
	}

	fmt.Printf("Found %d morphemes before fusion\n", internal.Count)

	// Fuse regex
	morphemeList := morpheme_list_t.FuseRegex(internal)
	fmt.Printf("Found %d morphemes after fusion\n", morphemeList.Count)

	// Identify types
	lex := make([]structs.LexicalCategories, morphemeList.Count)
	morpheme_list_t.IdentifyTrueMorphemes(&morphemeList, &lex)

	// Analyze misclassifications
	fmt.Println("\n=== Analyzing Misclassified Morphemes ===\n")

	knownSuffixes := []string{"ed", "ing", "ly", "er", "est", "al"}
	knownPrefixes := []string{"un", "re", "dis", "pre"}

	for _, suffix := range knownSuffixes {
		analyzeOne(&morphemeList, suffix, "SUFFIX")
	}

	for _, prefix := range knownPrefixes {
		analyzeOne(&morphemeList, prefix, "PREFIX")
	}
}

func analyzeOne(list *structs.MorphemeList, morphemeStr string, expectedType string) {
	for i := 0; i < list.Count; i++ {
		m := list.List[i]
		if strings.ToLower(m.Morpheme) == morphemeStr {
			actualType := typeToString(m.Type)
			status := "✅"
			if actualType != expectedType {
				status = "❌"
			}

			fmt.Printf("%s '%s' → %s (expected %s)\n", status, morphemeStr, actualType, expectedType)
			fmt.Printf("   FrontRegex: '%s'\n", m.GetFrontRegex())
			fmt.Printf("   BackRegex:  '%s'\n", m.GetBackRegex())
			fmt.Printf("   Full Regex: '%s'\n", m.Regex())

			// Check the conditions
			front := m.GetFrontRegex()
			back := m.GetBackRegex()

			fmt.Printf("   Checks:\n")
			fmt.Printf("     front == '^' or '()': %v\n", front == "^" || front == "()")
			fmt.Printf("     back == '$' or '()':  %v\n", back == "$" || back == "()")
			fmt.Printf("     Type before classification: %s\n", typeToString(m.Type))
			fmt.Println()

			// Only show first occurrence
			break
		}
	}
}

func typeToString(t int) string {
	switch t {
	case constants.STEM:
		return "STEM"
	case constants.PREFIX:
		return "PREFIX"
	case constants.SUFFIX:
		return "SUFFIX"
	case constants.INFIX:
		return "INFIX"
	case constants.CIRCUMFIX:
		return "CIRCUMFIX"
	case constants.UNDEF:
		return "UNDEF"
	default:
		return "UNKNOWN"
	}
}
