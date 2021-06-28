package libnumber

import (
	"fmt"
	"math"
)

// Suffix - Convert number to suffix K, M, B
func Suffix(input float64) string {
	// Under 10,000 (9999)
	if input < 10000 {
		return fmt.Sprintf("%v", input)
	}

	// Under 1,000,000 (999K)
	if input < 1000000 {
		input = math.Floor(input / 1000)
		return fmt.Sprintf("%vK", input)
	}

	// Under 1,000,000,000 (999M)
	if input < 1000000000 {
		input = math.Floor(input / 1000000)
		return fmt.Sprintf("%vM", input)
	}

	input = math.Floor(input / 1000000000)
	return fmt.Sprintf("%vB", input)
}

// SuffixByte - Convert number to suffix K, M, B
func SuffixByte(input float64) string {
	// Under 10,000 (9999)
	if input < 10000 {
		return fmt.Sprintf("%v B", input)
	}

	// Under 1,000,000 (999K)
	if input < 1000000 {
		input = math.Floor(input / 1000)
		return fmt.Sprintf("%v KB", input)
	}

	// Under 1,000,000,000 (999M)
	if input < 1000000000 {
		input = math.Floor(input / 1000000)
		return fmt.Sprintf("%v MB", input)
	}

	input = math.Floor(input / 1000000000)
	return fmt.Sprintf("%v GB", input)
}
