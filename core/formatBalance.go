package core

import (
	"math"
	"math/big"
	"strings"
)

func FormatBalance(balance *big.Int, decimals int) string {
	factor := new(big.Rat).SetFloat64(math.Pow10(decimals))
	if balance == nil {
		balance = big.NewInt(0)
	}
	humanReadable := new(big.Rat).SetInt(balance).Quo(new(big.Rat).SetInt(balance), factor)
	floatStr := humanReadable.FloatString(decimals)

	if decimals > 0 {
		floatStr = strings.TrimRight(floatStr, "0")
	}

	floatStr = strings.TrimSuffix(floatStr, ".")
	// Add thousands separator
	parts := strings.Split(floatStr, ".")
	intPart := parts[0]
	var result strings.Builder

	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(digit)
	}
	if len(parts) > 1 {
		result.WriteString(".")
		result.WriteString(parts[1])
	}

	return result.String()
}
