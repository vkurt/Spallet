package core

import "math/big"

func StringToBigInt(s string) big.Int {
	n := new(big.Int)
	n.SetString(s, 10)
	return *n
}
