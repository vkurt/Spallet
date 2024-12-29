package core

type addressBook struct {
	WalletOrder []string
	Wallets     map[string]Wallet
}

var UserAddressBook = addressBook{
	WalletOrder: []string{},
	Wallets:     make(map[string]Wallet),
}
