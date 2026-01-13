// Package alphabet implements word delimiter detection from alphabet.c
// Line-for-line translation of alphabet.c lines 1-102
package alphabet

import (
	"fmt"

	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// FindWordDelimiter retrieves word delimiter (lines 30-99)
// Updated to accept Config parameter for dependency injection (no globals)
func FindWordDelimiter(filePtr *string, cfg *config.Config) (string, error) {
	file := *filePtr

	cfg.VPrint("Creating List of Uniq Chars.")
	ret, i := functions.Uniq(file)
	if i == -1 {
		return "", constants.ErrUniq
	}

	cfg.VPrint("Constructing Word Delimiter Structures.")
	wd := make([]structs.CharDoubleton, i+1)
	
	for j := 0; j <= i; j++ {
		if cfg.IsVerbose() {
			fmt.Printf("Working on character %d/%d...\r", j, i)
		}
		
		c := string([]byte{ret[j]})
		arr := functions.ExplodeSansnull(file, c)
		freq := len(arr)
		wd[j].First = ret[j]
		wd[j].Second = freq
	}
	
	cfg.VPrint("Initial Word Delimiter Structures Constructed.")

	maxSize := 0
	index := 0
	wdReal := make([]byte, 2)

	for j := 0; j <= i; j++ {
		if wd[j].Second > maxSize {
			index = 0
			wdReal[index] = wd[j].First
			index++
			wdReal[index] = 0
			maxSize = wd[j].Second
		} else if wd[j].Second == maxSize && functions.InArray(wd[j].First, wdReal, index) == -1 {
			// Reallocate BEFORE writing if needed
			if index >= len(wdReal) {
				newWdReal := make([]byte, index+2)
				copy(newWdReal, wdReal)
				wdReal = newWdReal
			}

			wdReal[index] = wd[j].First
			index++

			// Reallocate again if needed for null terminator
			if index >= len(wdReal) {
				newWdReal := make([]byte, index+2)
				copy(newWdReal, wdReal)
				wdReal = newWdReal
			}
			wdReal[index] = 0

			uniqd, _ := functions.Uniq(string(wdReal))
			wdReal = []byte(uniqd)
		}
	}
	
	cfg.VPrint("Word Delimiter Candidates Identified.")

	wdRealStr := string(wdReal[:index])
	wdPermuted := wdRealStr
	wdFinalFreq := 0
	wdFinal := ""
	permIndex := 0
	
	if len(wdRealStr) > 1 {
		initialWd := wdRealStr
		for {
			wdPermuted = functions.Permute(&wdPermuted, &permIndex)
			if wdPermuted == initialWd {
				break
			}
			
			wdTesting := wdPermuted
			arr := functions.ExplodeSansnullStr(file, wdTesting)
			size := len(arr)
			
			if size > wdFinalFreq {
				wdFinalFreq = size
				wdFinal = wdPermuted
			} else if size == wdFinalFreq && size != 0 {
				fmt.Println("This implementation is not designed to handle multiple word delimiters.\nSkipping this file...")
				return "", constants.ErrOverruled
			}
		}
	}

	if wdFinalFreq == 0 {
		return wdRealStr, nil
	}
	return wdFinal, nil
}

