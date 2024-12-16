package main

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

const defaultMnemonicEntropy = 128 //256 for 24 word seed phrase

// Generate a private key from a mnemonic phrase and return it, it can create new wallets from diffrernt pkIndex
func mnemonicToPk(mnemonic string, pkIndex uint32) ([]byte, error) {
	// Convert mnemonic to seed
	seed := bip39.NewSeed(mnemonic, "")
	// fmt.Println("Seed:", hex.EncodeToString(seed))

	// Generate master key from seed
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Master Key:", masterKey)

	// Define derivation path: m/44'/60'/0'/0
	keyPath := []uint32{44 | hdkeychain.HardenedKeyStart, 60 | hdkeychain.HardenedKeyStart, 0 | hdkeychain.HardenedKeyStart, 0}

	// Traverse the key path
	pk := masterKey
	for _, index := range keyPath {
		pk, err = pk.Child(index)
		if err != nil {
			return nil, err
		}
	}

	// Derive the key for the specified index
	keyNew, err := pk.Child(pkIndex)
	if err != nil {
		return nil, err
	}

	// Convert to ECDSA private key
	privateKey, err := keyNew.ECPrivKey()
	if err != nil {
		return nil, err
	}

	privateKeyBytes := privateKey.Serialize()
	// fmt.Println("Private Key:", hex.EncodeToString(privateKeyBytes))

	return privateKeyBytes, nil
}

// splits mnemonic to 6 word lines
func formatMnemonic(mnemonic string) string {
	words := strings.Split(mnemonic, " ")
	formatted := ""
	fmt.Println(mnemonic, len(words))

	if len(words) == 12 || len(words) == 24 {
		if len(words) == 12 {
			firstLine := strings.Join(words[:6], " ")
			secondLine := strings.Join(words[6:], " ")
			formatted = firstLine + "\n" + secondLine

		} else if len(words) == 24 {
			firstLine := strings.Join(words[:6], " ")
			secondLine := strings.Join(words[6:12], " ")
			thirdLine := strings.Join(words[12:18], " ")
			fourthLine := strings.Join(words[18:24], " ")
			formatted = firstLine + "\n" + secondLine + "\n" + thirdLine + "\n" + fourthLine
		}
	} else {
		return "Invalid mnemonic phrase"
	}

	return formatted
}
