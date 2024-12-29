package core

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

const DefaultMnemonicEntropy = 128 //256 for 24 word seed phrase

// Generate a private key from a mnemonic phrase and return it, it can create new wallets from diffrernt pkIndex
func MnemonicToPk(mnemonic string, pkIndex uint32) ([]byte, error) {
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

// FormatMnemonic splits the mnemonic into lines with a specified number of words per line.
func FormatMnemonic(mnemonic string, wordsPerLine int) string {
	words := strings.Split(mnemonic, " ")
	formatted := ""

	if len(words) != 12 && len(words) != 24 {
		return "Invalid mnemonic phrase"
	}

	for i := 0; i < len(words); i += wordsPerLine {
		end := i + wordsPerLine
		if end > len(words) {
			end = len(words)
		}
		formatted += strings.Join(words[i:end], " ")
		if end < len(words) {
			formatted += "\n"
		}
	}

	return formatted
}
