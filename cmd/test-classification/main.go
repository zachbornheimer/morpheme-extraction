package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/nlp"
)

// Known English affixes for validation
var knownSuffixes = map[string]bool{
	"ed": true, "ing": true, "ly": true, "er": true, "est": true,
	"tion": true, "sion": true, "ness": true, "ment": true,
	"able": true, "ible": true, "ful": true, "less": true,
	"ous": true, "ious": true, "ive": true, "al": true, "ic": true,
	"ish": true, "y": true, "en": true, "ize": true, "ise": true,
	"fy": true, "ward": true, "wise": true, "s": true, "es": true,
	"ity": true, "ty": true, "age": true, "ance": true, "ence": true,
	"ant": true, "ent": true, "dom": true, "hood": true, "ship": true,
}

var knownPrefixes = map[string]bool{
	"un": true, "re": true, "in": true, "im": true, "il": true, "ir": true,
	"dis": true, "en": true, "em": true, "non": true, "over": true,
	"mis": true, "sub": true, "pre": true, "inter": true, "fore": true,
	"de": true, "trans": true, "super": true, "semi": true, "anti": true,
	"mid": true, "under": true, "out": true, "up": true, "be": true,
}

type MorphemeInfo struct {
	Str        string
	Type       int
	FrontRegex string
	BackRegex  string
	Freq       int
}

func main() {
	// Run NLP to generate morphemes
	cfg := &config.Config{
		VerboseMode:        constants.OFF,
		ProcessSequentially: constants.OFF,
		ProcessFull:        constants.OFF,
		OutputFilename:     "test-output.zdrm",
		CorpusDir:          "./test-corpus/",
	}

	fmt.Println("Running morpheme extraction...")
	nlp.Nlp(cfg)

	// Parse the output file
	fmt.Println("\n=== Parsing output file ===")
	morphemes := parseOutputFile("test-output.zdrm")

	// Group by type
	stems := []MorphemeInfo{}
	prefixes := []MorphemeInfo{}
	suffixes := []MorphemeInfo{}
	infixes := []MorphemeInfo{}

	for _, m := range morphemes {
		switch m.Type {
		case constants.STEM:
			stems = append(stems, m)
		case constants.PREFIX:
			prefixes = append(prefixes, m)
		case constants.SUFFIX:
			suffixes = append(suffixes, m)
		case constants.INFIX:
			infixes = append(infixes, m)
		}
	}

	fmt.Printf("\n=== Distribution ===\n")
	fmt.Printf("STEMs: %d\n", len(stems))
	fmt.Printf("PREFIXes: %d\n", len(prefixes))
	fmt.Printf("SUFFIXes: %d\n", len(suffixes))
	fmt.Printf("INFIXes: %d\n", len(infixes))

	// Test 1: Check for misclassified known suffixes
	fmt.Println("\n=== TEST 1: Known Suffixes Misclassified as INFIX ===")
	misclassifiedSuffixes := []MorphemeInfo{}
	for _, m := range infixes {
		if knownSuffixes[strings.ToLower(m.Str)] {
			misclassifiedSuffixes = append(misclassifiedSuffixes, m)
		}
	}

	if len(misclassifiedSuffixes) > 0 {
		fmt.Printf("❌ FAILED: Found %d known suffixes classified as INFIX\n", len(misclassifiedSuffixes))
		// Show first 10 examples
		for i, m := range misclassifiedSuffixes {
			if i >= 10 {
				fmt.Printf("... and %d more\n", len(misclassifiedSuffixes)-10)
				break
			}
			fmt.Printf("  '%s' (freq=%d) - FrontRegex='%s', BackRegex='%s'\n",
				m.Str, m.Freq, truncate(m.FrontRegex, 50), truncate(m.BackRegex, 50))
		}
	} else {
		fmt.Println("✅ PASSED: No known suffixes misclassified as INFIX")
	}

	// Test 2: Check for misclassified known prefixes
	fmt.Println("\n=== TEST 2: Known Prefixes Misclassified as INFIX ===")
	misclassifiedPrefixes := []MorphemeInfo{}
	for _, m := range infixes {
		if knownPrefixes[strings.ToLower(m.Str)] {
			misclassifiedPrefixes = append(misclassifiedPrefixes, m)
		}
	}

	if len(misclassifiedPrefixes) > 0 {
		fmt.Printf("❌ FAILED: Found %d known prefixes classified as INFIX\n", len(misclassifiedPrefixes))
		for i, m := range misclassifiedPrefixes {
			if i >= 10 {
				fmt.Printf("... and %d more\n", len(misclassifiedPrefixes)-10)
				break
			}
			fmt.Printf("  '%s' (freq=%d) - FrontRegex='%s', BackRegex='%s'\n",
				m.Str, m.Freq, truncate(m.FrontRegex, 50), truncate(m.BackRegex, 50))
		}
	} else {
		fmt.Println("✅ PASSED: No known prefixes misclassified as INFIX")
	}

	// Test 3: Check for known suffixes in SUFFIX category
	fmt.Println("\n=== TEST 3: Known Suffixes Correctly Classified as SUFFIX ===")
	correctSuffixes := []string{}
	for _, m := range suffixes {
		if knownSuffixes[strings.ToLower(m.Str)] {
			correctSuffixes = append(correctSuffixes, m.Str)
		}
	}
	sort.Strings(correctSuffixes)
	fmt.Printf("Found %d known suffixes correctly classified: %v\n", len(correctSuffixes), correctSuffixes)

	// Test 4: Check for known prefixes in PREFIX category
	fmt.Println("\n=== TEST 4: Known Prefixes Correctly Classified as PREFIX ===")
	correctPrefixes := []string{}
	for _, m := range prefixes {
		if knownPrefixes[strings.ToLower(m.Str)] {
			correctPrefixes = append(correctPrefixes, m.Str)
		}
	}
	sort.Strings(correctPrefixes)
	fmt.Printf("Found %d known prefixes correctly classified: %v\n", len(correctPrefixes), correctPrefixes)

	// Test 5: STEM quality (punctuation, case issues)
	fmt.Println("\n=== TEST 5: STEM Quality Issues ===")
	punctuationStems := []string{}
	caseIssues := make(map[string][]string)

	for _, m := range stems {
		if strings.ContainsAny(m.Str, ",.!?;:'\"-_") {
			punctuationStems = append(punctuationStems, m.Str)
		}

		lower := strings.ToLower(m.Str)
		caseIssues[lower] = append(caseIssues[lower], m.Str)
	}

	if len(punctuationStems) > 0 {
		fmt.Printf("❌ STEMs with punctuation: %v\n", punctuationStems)
	} else {
		fmt.Println("✅ No STEMs with punctuation")
	}

	caseDuplicates := []string{}
	for _, variants := range caseIssues {
		if len(variants) > 1 {
			caseDuplicates = append(caseDuplicates, strings.Join(variants, "/"))
		}
	}
	if len(caseDuplicates) > 0 {
		sort.Strings(caseDuplicates)
		fmt.Printf("❌ Case duplicates in STEMs: %v\n", caseDuplicates)
	} else {
		fmt.Println("✅ No case duplicates in STEMs")
	}

	// Summary
	fmt.Println("\n=== SUMMARY ===")
	totalIssues := len(misclassifiedSuffixes) + len(misclassifiedPrefixes) + len(punctuationStems) + len(caseDuplicates)
	if totalIssues == 0 {
		fmt.Println("✅ ALL TESTS PASSED!")
	} else {
		fmt.Printf("❌ Found %d total issues\n", totalIssues)
		fmt.Printf("  - %d known suffixes misclassified as INFIX\n", len(misclassifiedSuffixes))
		fmt.Printf("  - %d known prefixes misclassified as INFIX\n", len(misclassifiedPrefixes))
		fmt.Printf("  - %d STEMs with punctuation\n", len(punctuationStems))
		fmt.Printf("  - %d case duplicate pairs in STEMs\n", len(caseDuplicates))
	}
}

func parseOutputFile(filename string) []MorphemeInfo {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	morphemes := []MorphemeInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "All Processed") || strings.HasPrefix(line, "===") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		var mType int
		switch parts[0] {
		case "STEM:":
			mType = constants.STEM
		case "PREFIX:":
			mType = constants.PREFIX
		case "SUFFIX:":
			mType = constants.SUFFIX
		case "INFIX:":
			mType = constants.INFIX
		case "CIRCUMFIX:":
			mType = constants.CIRCUMFIX
		default:
			continue
		}

		morphemes = append(morphemes, MorphemeInfo{
			Str:  parts[1],
			Type: mType,
			// Note: We can't get regex patterns from the output file
			// The file format doesn't include them
		})
	}

	return morphemes
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
	default:
		return "UNDEF"
	}
}
