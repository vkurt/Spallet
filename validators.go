package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// validateAccountInput validates a wallet name, account details, or just an account address based on the specified check type.
// Parameters:
// - savedNames: A slice of saved wallet names. It can be nil if not checking names.
// - savedAccs: A map of saved accounts, where the key is a string and the value is a Wallet. It can be nil if not checking accounts.
// - name: The wallet name to validate. It can be an empty string if not checking names.
// - checkType: A string specifying the type of validation ("name", "account", or "address").
// - checkUniqueness: A boolean indicating whether to check for uniqueness of the address.
// - accDetails: Optional parameters for account validation (accName, accWif, accAddr, chckWif).
// - Calling the function to validate the wallet name --result, err := validateAccountInput(savedNames, nil, name, "name", false)
// - Calling the function to validate account details --result, err := validateAccountInput(nil, savedAccs, "", "account", true, accName, accWif, accAddr, chckWif)
// - Calling the function to validate the address with uniqueness check --result, err := validateAccountInput(nil, savedAccs, "", "address", true, phaAddrName)

func validateAccountInput(savedNames []string, savedAccs map[string]Wallet, name string, checkType string, checkAddrUniqueness bool, accDetails ...interface{}) (string, error) {
	switch checkType {
	case "name":
		// Validate wallet name
		if len(name) < 1 {
			return "Not entered", errors.New("not entered")
		} else if len(name) > 20 {
			return "Please use max 20 letters", errors.New("max 20 characters")
		}

		for _, savedName := range savedNames {
			if savedName == name {
				return "Name already used", errors.New("already used")
			}
		}
		return "Name available", nil

	case "account":
		// Ensure accDetails has the required number of parameters
		if len(accDetails) != 4 {
			return "Missing account parameters", errors.New("missing account parameters")
		}

		// Extract account parameters
		accName, ok1 := accDetails[0].(string)
		accWif, ok2 := accDetails[1].(string)
		accAddr, ok3 := accDetails[2].(string)
		chckWif, ok4 := accDetails[3].(bool)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return "Invalid account parameters", errors.New("invalid account parameters")
		}

		// Validate account details
		for _, acc := range savedAccs {
			if accName == acc.Name {
				return "Account name already used", fmt.Errorf("account name already used")
			}
			if accWif == acc.WIF && chckWif {
				return fmt.Sprintf("Wif already used with name %s", acc.Name), fmt.Errorf("wif already used with name %s", acc.Name)
			} else if !chckWif && accAddr == acc.Address {
				return fmt.Sprintf("Address already used with name %s", acc.Name), fmt.Errorf("address already used with name %s", acc.Name)
			}
		}
		return "Account is not in the wallet data", nil

	case "address":
		// Ensure accDetails has the required number of parameters
		if len(accDetails) != 1 {
			return "Missing address parameter", errors.New("missing address parameter")
		}

		// Extract account address
		phaAddrName, ok := accDetails[0].(string)
		if !ok {
			return "Invalid address parameter", errors.New("invalid address parameter")
		}

		// Validate the address or name
		if len(phaAddrName) < 3 {
			return "Recipient address/name is too short", errors.New("recipient address/name is too short")
		} else if len(phaAddrName) <= 15 {
			noSpaces := !regexp.MustCompile(`\s`).MatchString(phaAddrName)
			matched, _ := regexp.MatchString("^[a-z][a-z0-9]{2,14}$", phaAddrName)
			if noSpaces && matched {
				return "Name usable", nil
			} else {
				return "Recipient name can't contain special characters, spaces, can't start with a number", errors.New("recipient name can't contain special characters, spaces, can't start with a number")
			}
		} else if len(phaAddrName) < 47 && len(phaAddrName) > 15 {
			noSpaces := !regexp.MustCompile(`\s`).MatchString(phaAddrName)
			matched, _ := regexp.MatchString("^[P][a-zA-Z0-9]{2,46}$", phaAddrName)
			if noSpaces && matched {
				if checkAddrUniqueness {
					for _, acc := range savedAccs {
						if phaAddrName == acc.Address {
							return fmt.Sprintf("Address already used with name %s", acc.Name), fmt.Errorf("address already used with name %s", acc.Name)
						}
					}
				}
				return "Address usable but short", fmt.Errorf("too short")
			} else {
				return "Phantasma addresses can't contain special characters, spaces, can't start with a number, and must start with 'P'", errors.New("phantasma addresses can't contain special characters, spaces, can't start with a number, and must start with 'P'")
			}
		} else if len(phaAddrName) == 47 {
			noSpaces := !regexp.MustCompile(`\s`).MatchString(phaAddrName)
			matched, _ := regexp.MatchString("^[P][a-zA-Z0-9]{2,46}$", phaAddrName)
			if noSpaces && matched {
				if checkAddrUniqueness {
					for _, acc := range savedAccs {
						if phaAddrName == acc.Address {
							return fmt.Sprintf("Address already used with name %s", acc.Name), fmt.Errorf("address already used with name %s", acc.Name)
						}
					}
				}
				return "Valid address", nil
			} else {
				return "Phantasma addresses can't contain special characters, spaces, can't start with a number, and must start with 'P'", errors.New("phantasma addresses can't contain special characters, spaces, can't start with a number, and must start with 'P'")
			}
		} else if len(phaAddrName) > 47 {
			return "Phantasma addresses are shorter than 48 characters", errors.New("phantasma addresses are shorter than 48 characters")
		}

		return "Valid", nil

	default:
		return "Invalid check type specified", errors.New("invalid check type")
	}
}

func wifValidator(wif string) (string, error) {
	noSpaces := !regexp.MustCompile(`\s`).MatchString(wif)
	matched, _ := regexp.MatchString("^[KL5][a-zA-Z0-9]{0,51}$", wif)

	if len(wif) < 1 {

		return "Please enter wif", errors.New("please enter wif")
	} else if len(wif) <= 52 {
		if !noSpaces {

			return "Wif can't contain spaces", errors.New("wif can't contain spaces")
		} else if !matched {

			return "Wif can't contain special characters and must start with 'K', 'L', or '5'", errors.New("wif can't contain special characters and must start with 'K', 'L', or '5'")
		} else if len(wif) < 51 {

			return "Wif key is too short", errors.New("wif key is too short")
		}
	} else if len(wif) > 52 {

		return "Wif key is too long", errors.New("wif key is too long")
	}

	if noSpaces && matched && len(wif) >= 51 {
		return "Wif format is correct", nil
	} else {
		return "Wif format is wrong", fmt.Errorf("wif format is wrong")
	}

}
func seedPhraseValidator(phrase string) error {
	words := strings.Split(phrase, " ")
	wordCount := len(words)

	switch {
	case wordCount < 12:

		return fmt.Errorf("seed phrase is too short, minimum 12 words")
	case wordCount == 12:
		if bip39.IsMnemonicValid(phrase) {

			return nil
		}

		return fmt.Errorf("seed phrase is invalid")
	case wordCount < 24:

		return fmt.Errorf("continue to enter your 24-word seed phrase")
	case wordCount == 24:
		if bip39.IsMnemonicValid(phrase) {

			return nil
		}

		return fmt.Errorf("seed phrase is invalid")
	default:

		return fmt.Errorf("seed phrase is invalid")
	}
}
func pwdMatch(firstPwdEntry, scndPwdEntry string) (string, error) {

	if firstPwdEntry == scndPwdEntry {
		return "Paswords mathched", nil
	} else {
		return "Password mismatch", fmt.Errorf("password mismatch")
	}
}

// checks if name is compatible with Phantasma's naming sytem for registering
func isValidName(name string) bool {
	matched, _ := regexp.MatchString("^[a-z][a-z0-9]{2,14}$", name)
	noSpaces := !regexp.MustCompile(`\s`).MatchString(name)
	return matched && noSpaces
}
