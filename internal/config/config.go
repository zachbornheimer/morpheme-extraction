// Package config contains runtime configuration
package config

import "github.com/zachbornheimer/morpheme-extraction/internal/constants"

// Config holds runtime configuration
// Replaces global variables for better testability and encapsulation
type Config struct {
	VerboseMode         int
	ProcessSequentially int
	ProcessFull         int
	OutputFilename      string
	CorpusDir           string
}

// VPrint prints message only if verbose mode is enabled
func (c *Config) VPrint(msg string) {
	if c.VerboseMode == constants.ON {
		println(msg)
	}
}

// IsVerbose returns true if verbose mode is enabled
func (c *Config) IsVerbose() bool {
	return c.VerboseMode == constants.ON
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		VerboseMode:         constants.OFF,
		ProcessSequentially: constants.OFF,
		ProcessFull:         constants.OFF,
		OutputFilename:      "languageModel.zdrm",
		CorpusDir:           "",
	}
}
