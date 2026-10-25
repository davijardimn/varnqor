package settings

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseTokenAmount accepts plain integer token counts or human-friendly
// suffixes such as 500k, 10m, and 1.5g. Suffixes use decimal multipliers.
func ParseTokenAmount(value string) (int64, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return 0, fmt.Errorf("token amount is required")
	}
	multiplier := float64(1)
	switch value[len(value)-1] {
	case 'k':
		multiplier = 1_000
		value = value[:len(value)-1]
	case 'm':
		multiplier = 1_000_000
		value = value[:len(value)-1]
	case 'g':
		multiplier = 1_000_000_000
		value = value[:len(value)-1]
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid token amount")
	}
	result := number * multiplier
	if math.IsInf(result, 0) || result > float64(math.MaxInt64) {
		return 0, fmt.Errorf("token amount is too large")
	}
	return int64(math.Round(result)), nil
}
