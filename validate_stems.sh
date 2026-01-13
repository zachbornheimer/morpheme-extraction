#!/bin/bash

echo "=== Analyzing STEM Quality ==="
echo ""

# Common English words list (basic check)
echo "C version (104 STEMs):"
echo "  Valid English words: $(grep -E '^(a|about|after|all|an|and|are|as|at|be|by|do|for|he|I|in|is|it|me|my|no|of|on|one|or|our|out|so|the|to|up|was|we)[ ]*$' c_stems.txt | wc -l) common words"
echo ""

# Check for issues
echo "  Capitalization duplicates:"
grep -E '^(For|It|The)[ ]*$' c_stems.txt | sed 's/ *$//' | xargs -I {} echo "    - {}"
echo ""
echo "  Proper names:"
grep -E '^(Ogilvy|Ottershaw|Mars)[ ]*$' c_stems.txt | sed 's/ *$//' | xargs -I {} echo "    - {}"
echo ""

echo "Go version (125 STEMs):"
echo "  Valid English words: $(grep -E '^(a|about|after|all|am|an|and|are|as|at|be|but|by|do|end|for|had|he|his|how|I|if|in|is|it|its|me|men|more|my|no|of|on|one|or|our|out|pit|so|the|their|to|too|up|use|very|was|way|we|went|where)[ ]*$' go_stems.txt | wc -l) common words"
echo ""
echo "  Same issues as C plus:"
echo "    - night, (with comma - INVALID)"
echo ""

echo "=== 21 Extra STEMs in Go ==="
comm -23 go_stems.txt c_stems.txt > go_extra_stems.txt
echo ""
echo "Valid English words:"
valid_count=$(grep -E '^(am|and|but|end|had|his|how|if|its|men|more|pit|possible|their|too|use|very|way|went|where)[ ]*$' go_extra_stems.txt | wc -l)
echo "  $valid_count out of 21 = $(echo "scale=1; 100*$valid_count/21" | bc)%"
echo ""
echo "Invalid:"
grep -v -E '^(am|and|but|end|had|his|how|if|its|men|more|pit|possible|their|too|use|very|way|went|where)[ ]*$' go_extra_stems.txt | sed 's/ *$//' | head -10
