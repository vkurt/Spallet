package core

import (
	"fmt"
	"math/big"
)

const SoulDecimals = 8
const KcalDecimals = 10

func CheckFeeBalance(feeAmount *big.Int) error {
	kcalBalance := new(big.Int)

	if token, exists := LatestAccountData.FungibleTokens["KCAL"]; exists {
		kcalBalance = token.Amount
	} else {
		fmt.Println("Acc dont have Kcal")
		return fmt.Errorf("acc dont have Kcal to fill Specky's engines")
	}

	if kcalBalance.Cmp(big.NewInt(0)) == 0 {
		fmt.Println("Kcal balance not found")
		return fmt.Errorf("acc dont have Kcal to fill Specky's engines")
	}

	fmt.Printf("Kcal Balance: %s, Required Kcal: %s\n", FormatBalance(kcalBalance, KcalDecimals), FormatBalance(feeAmount, KcalDecimals))

	if kcalBalance.Cmp(feeAmount) < 0 {
		fmt.Printf("Insufficient Kcal: Required: %s, Available: %s\n", FormatBalance(feeAmount, KcalDecimals), FormatBalance(kcalBalance, KcalDecimals))
		return fmt.Errorf("insufficient Spark for filling Specky's engines. required: %s Kcal, available: %s Kcal", FormatBalance(feeAmount, KcalDecimals), FormatBalance(kcalBalance, KcalDecimals))
	}

	return nil
}
