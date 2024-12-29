package core

import "math/big"

type Wallet struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	WIF      string `json:"wif"`
	Mnemonic string `json:"mnemonic"`
}

type Credentials struct {
	Password           string            `json:"password"`
	Wallets            map[string]Wallet `json:"wallets"`
	WalletOrder        []string          `json:"wallet_order"`
	LastSelectedWallet string            `json:"last_selected_wallet"`
}

type Stake struct {
	Amount    *big.Int
	Time      uint
	Unclaimed *big.Int
}
