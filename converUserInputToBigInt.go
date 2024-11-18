package main

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
)

// convertUserInputToBigInt converts a user input string to a big.Int, handling decimals, invalid characters, and nil strings.
func convertUserInputToBigInt(amount string, decimals int) (*big.Int, error) {

	if amount == "" {
		return big.NewInt(0), nil
	}

	// Allow only digits, spaces, periods, and commas
	if matched, _ := regexp.MatchString(`^[\d\s.,]+$`, amount); !matched {
		return big.NewInt(0), fmt.Errorf("invalid amount format: contains invalid characters")
	}

	// Remove all spaces
	amount = strings.Replace(amount, " ", "", -1)

	// Replace commas with periods to standardize the decimal separator
	amount = strings.Replace(amount, ",", ".", -1)

	// Check if the cleaned amount has more than one decimal point
	if strings.Count(amount, ".") > 1 {
		return nil, fmt.Errorf("invalid amount format: multiple decimal points")
	}

	// Ensure there's at least one valid numeric character left
	if amount == "" || amount == "." {
		return nil, fmt.Errorf("invalid amount format: no valid numeric characters")
	}

	// Parse the cleaned amount string into a big.Float
	amountFloat, _, err := big.ParseFloat(amount, 10, 0, big.ToNearestAway)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// Scale the float amount to the specified number of decimals
	factor := new(big.Float).SetFloat64(math.Pow10(decimals))
	amountBigFloat := new(big.Float).Mul(amountFloat, factor)

	// Convert the big.Float to big.Int
	amountBigInt := new(big.Int)
	amountBigFloat.Int(amountBigInt)

	return amountBigInt, nil

}
