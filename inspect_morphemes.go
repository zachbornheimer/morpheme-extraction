package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: inspect_morphemes <file> <morpheme>")
		os.Exit(1)
	}
	
	file, _ := os.Open(os.Args[1])
	defer file.Close()
	
	target := os.Args[2]
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, " " + target + " ") {
			fmt.Println(line)
		}
	}
}
