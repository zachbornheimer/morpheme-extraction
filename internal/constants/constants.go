// Package constants contains all constant definitions from constants.h
package constants

import "errors"

// File paths
const (
	DefaultPath    = "./test-corpus/"
	ZedramOutput   = "languageModel.zdrm"
	NgramSize      = 9
)

// Switches
const (
	ON  = 1
	OFF = 0
)

// Error codes
const (
	EOverruled = 8
	ERealloc   = 9
	EMalloc    = 10
	EUniq      = 6
	EAssert    = 7
)

// Thresholds - percents
const (
	ThresholdSimilarNgrams  = 12.5
	ThresholdCircumfix      = 100.0
	ThresholdConfirmation   = 0.0  // Temporarily disabled for testing
)

// Morpheme Types
const (
	UNDEF     = 0
	PREFIX    = 1
	SUFFIX    = 2
	CIRCUMFIX = 3
	INFIX     = 4
	STEM      = 5
)

// Error variables
var (
	ErrOverruled = errors.New("overruled")
	ErrRealloc   = errors.New("realloc failed")
	ErrMalloc    = errors.New("malloc failed")
	ErrUniq      = errors.New("uniq failed")
	ErrAssert    = errors.New("assertion failed")
)

// StrType returns string representation of morpheme type
func StrType(code int) string {
	switch code {
	case UNDEF:
		return "UNDEF"
	case PREFIX:
		return "PREFIX"
	case SUFFIX:
		return "SUFFIX"
	case CIRCUMFIX:
		return "CIRCUMFIX"
	case INFIX:
		return "INFIX"
	case STEM:
		return "STEM"
	default:
		return "UNKNOWN!"
	}
}

