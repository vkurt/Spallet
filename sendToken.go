package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

func confirmSendToken(WIF, tokenSymbol, to string, tokenAmount *big.Int, creds Credentials, name string) error {
	// Show confirmation dialog
	sendTokenDecimal := latestAccountData.FungibleTokens[tokenSymbol].Decimals
	var confirmMessage string
	if len(name) < 3 {
		confirmMessage = fmt.Sprintf("You are about to send %s %s to %s\nAre you sure to ignite engines?", formatBalance(*tokenAmount, int(sendTokenDecimal)), tokenSymbol, to)

	} else {
		confirmMessage = fmt.Sprintf("You are about to send \n%s %s to %s' related address\n%s\nAre you sure to ignite engines?", formatBalance(*tokenAmount, int(sendTokenDecimal)), tokenSymbol, name, to)

	}
	dialog.ShowConfirm("Confirm Transaction", confirmMessage, func(confirmed bool) {
		if confirmed {
			// Proceed with transaction if user confirmed
			go sendTransaction(WIF, tokenSymbol, to, tokenAmount, creds)
		} else {
			// Handle abort action
			return
		}
	}, mainWindowGui)

	return nil
}

func sendTransaction(WIF, tokenSymbol, to string, tokenAmount *big.Int, creds Credentials) {
	keyPair, err := cryptography.FromWIF(WIF)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Transaction Failed",
			Content: fmt.Sprintf("Invalid WIF: %v", err),
		})
		return
	}

	expire := time.Now().UTC().Add(time.Second * 300).Unix()
	// fmt.Println("Sending token from ", keyPair.Address().String())
	// fmt.Println("sending token to ", to)
	// fmt.Println("Token ", tokenSymbol)
	// fmt.Println("Amount ", formatBalance(*tokenAmount, latestAccountData.FungibleTokens[tokenSymbol].Decimals))
	// fmt.Println("Gas Price ", gasPrice)
	// fmt.Println("Gas limit ", gasLimit)
	// fmt.Println("Expiration time ", time.Unix(expire, 0).Format("2006-01-02 15:04:05"))

	sb := scriptbuilder.BeginScript()
	sb.AllowGas(keyPair.Address().String(), cryptography.NullAddress().String(), gasPrice, gasLimit)
	sb.TransferTokens(tokenSymbol, keyPair.Address().String(), to, tokenAmount)
	sb.SpendGas(keyPair.Address().String())
	script := sb.EndScript()

	tx := blockchain.NewTransaction(network, chain, script, uint32(expire), payload)
	tx.Sign(keyPair)
	txHex := hex.EncodeToString(tx.Bytes())
	// fmt.Println("*****Tx: \n" + txHex)

	// Start the animation
	startAnimation("send", "Specky is delivering wait a bit...")

	// Here, you can use stopChan if needed later, for example:
	// defer close(stopChan) when you need to ensure it gets closed properly.

	// Send the transaction
	SendTransaction(txHex, creds, handleSuccess, handleFailure)

}

func showSendTokenDia(symbol string, creds Credentials, decimal int8) {
	// Usage
	askPwdDia(askPwd, creds.Password, mainWindowGui, func(correct bool) {
		fmt.Println("result", correct)
		if !correct {
			return
		}
		// Continue with your code here })
		var validAmount, validRecipient bool
		token := latestAccountData.FungibleTokens[symbol]
		recipientFirstString := ""
		bindRecipient := binding.BindString(&recipientFirstString)
		recipient := widget.NewEntryWithData(bindRecipient)
		recipient.PlaceHolder = "Please enter or select delivery address/name"
		if userSettings.SendOnly {
			recipient.Disable()
		} else {
			recipient.Enable()
		}
		var addressBookSelect *widget.Select
		var usersAddressesSelect *widget.Select

		addressBookSelect = widget.NewSelect(userAddressBook.WalletOrder, func(s string) {
			if len(s) < 1 {
				return
			}
			if usersAddressesSelect.SelectedIndex() >= 0 {
				usersAddressesSelect.ClearSelected()
			}
			bindRecipient.Set(userAddressBook.Wallets[s].Address)
			recText, _ := bindRecipient.Get()
			recipient.CursorColumn = len(recText)
			mainWindowGui.Canvas().Focus(recipient)

		})
		var usersAddresses []string
		for _, address := range creds.WalletOrder {
			if address != creds.LastSelectedWallet {
				usersAddresses = append(usersAddresses, address)
			}
		}
		usersAddressesSelect = widget.NewSelect(usersAddresses, func(s string) {
			if len(s) < 1 {
				return
			}
			if addressBookSelect.SelectedIndex() >= 0 {
				addressBookSelect.ClearSelected()
			}
			bindRecipient.Set(creds.Wallets[s].Address)
			recText, _ := bindRecipient.Get()
			recipient.CursorColumn = len(recText) // for some reason it dont wanna go end of the line sometimes goes tho strange tried refresh also but dunno
			mainWindowGui.Canvas().Focus(recipient)
		})

		usersAddressesSelect.PlaceHolder = "My Accounts"
		addressBookSelect.PlaceHolder = "Address Book"
		bindAmountString := ""
		bindAmount := binding.BindString(&bindAmountString)
		amount := widget.NewEntryWithData(bindAmount)
		amount.Bind(bindAmount)

		var amountBigInt *big.Int
		recipientSelectBox := container.NewGridWithColumns(2, addressBookSelect, usersAddressesSelect)
		recipientBox := container.NewVBox(container.NewGridWithRows(2, recipient, recipientSelectBox))

		sendButton := widget.NewButton("Engage", func() {
			// fmt.Println("Sending from WIF: ", creds.Wallets[creds.LastSelectedWallet].WIF)
			var err error
			var bindedAmount string
			var name string
			bindedAmount, err = bindAmount.Get()
			if err != nil {
				dialog.ShowError(err, mainWindowGui)
				return
			}
			recipientEntry, err := bindRecipient.Get()
			if err != nil {
				dialog.ShowError(err, mainWindowGui)
				return
			}
			if recipientEntry == "" {
				dialog.ShowError(errors.New("recipient is empty"), mainWindowGui)
				return
			} else if len(recipientEntry) <= 15 {
				name = recipientEntry
				nameToAddress, err := client.LookupName(recipientEntry)
				if err != nil {
					dialog.ShowError(fmt.Errorf("specky encountered an error while searching this name\n%s", err), mainWindowGui)
					return
				}
				recipientEntry = nameToAddress
			}

			amountBigInt, err = convertUserInputToBigInt(bindedAmount, int(decimal))
			if err != nil {
				dialog.ShowError(err, mainWindowGui)
				return
			} else {
				confirmSendToken(creds.Wallets[creds.LastSelectedWallet].WIF, symbol, recipientEntry, amountBigInt, creds, name)
			}

		})
		sendButton.Disable()

		amount.PlaceHolder = fmt.Sprintf("Please enter %s amount for delivery", symbol)

		gasLimitFloat, _ := gasLimit.Float64()
		gasSlider := widget.NewSlider(10000, 100000)
		gasSlider.Value = gasLimitFloat
		gasSliderLabel := widget.NewLabelWithStyle("Specky's energy limit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		warning := binding.NewString()
		warning.Set("You have enough Kcal to fill Specky's engines")
		warningLabel := widget.NewLabelWithData(warning)
		warningLabel.Bind(warning)

		updateSendButtonState := func() {
			var amountErr error
			amountBigInt, amountErr = convertUserInputToBigInt(amount.Text, int(decimal))
			if amountErr != nil {
				warningLabel.TextStyle.Bold = true
				warning.Set(amountErr.Error())
				return
			}

			feeAmount := new(big.Int).Mul(gasPrice, gasLimit)
			if symbol == "KCAL" {
				feeAmount = feeAmount.Add(amountBigInt, feeAmount)
			}

			err := checkFeeBalance(feeAmount)
			fmt.Printf("Update Submit Button: Amount: %s, Error: %v\n", amountBigInt.String(), err)

			if err == nil && amountBigInt != nil && amountBigInt.Cmp(big.NewInt(0)) > 0 && validAmount && validRecipient {
				sendButton.Enable()
				warningLabel.TextStyle = fyne.TextStyle{Bold: false}
				warning.Set("Specky is ready for launch!")
			} else if !validRecipient {
				sendButton.Disable()
				warningLabel.TextStyle = fyne.TextStyle{Bold: true}
				warning.Set("Delivery Address/Name is not valid!")
			} else if !validAmount {
				sendButton.Disable()
				warningLabel.TextStyle = fyne.TextStyle{Bold: true}
				warning.Set("Delivery Amount is not valid!")
			} else if err != nil {
				sendButton.Disable()
				warningLabel.TextStyle = fyne.TextStyle{Bold: true}
				warning.Set(err.Error())

			}
		}

		var amountBox *fyne.Container
		maxButton := widget.NewButton("Max", func() {
			feeAmount := new(big.Int).Mul(gasLimit, gasPrice)
			if symbol == "KCAL" {
				if feeAmount.Cmp(&token.Amount) < 0 {
					maxAmount := new(big.Int).Sub(&token.Amount, feeAmount)
					amTxt := formatBalance(*maxAmount, kcalDecimals)
					bindAmount.Set(amTxt)
					amount.CursorColumn = len(amTxt)
					mainWindowGui.Canvas().Focus(amount)
					validAmount = true
					updateSendButtonState()
				} else {
					bindAmount.Set("Dont have enough Kcal")
					validAmount = false
					updateSendButtonState()

				}

			} else {

				amTxt := formatBalance(token.Amount, token.Decimals)
				bindAmount.Set(amTxt)
				amount.CursorColumn = len(amTxt) // for some reason it dont wanna go end of the line sometimes goes tho strange tried refresh also but dunno
				validAmount = true
				updateSendButtonState()
				mainWindowGui.Canvas().Focus(amount)

			}

		})
		amountBox = container.NewVBox(container.NewBorder(nil, nil, nil, maxButton, amount))
		recipient.OnChanged = func(text string) {
			if len(text) != 47 {
				if usersAddressesSelect.SelectedIndex() > -1 {
					usersAddressesSelect.ClearSelected()
				}
				if addressBookSelect.SelectedIndex() > -1 {
					addressBookSelect.ClearSelected()
				}
			}
		}

		recipient.Validator = func(s string) error {

			if len(s) < 3 {
				validRecipient = false
				updateSendButtonState()
				return errors.New("recipient address/name is too short")
			} else if len(s) <= 15 {

				noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
				matched, _ := regexp.MatchString("^[a-z][a-z0-9]{2,14}$", s)
				if noSpaces && matched {
					validRecipient = true
					updateSendButtonState()
					return nil
				} else {
					validRecipient = false
					updateSendButtonState()
					dialog.ShowError(errors.New("recipient name cant contain \nspecial characters, \nspaces, \ncant start with number,\ncapital letters"), mainWindowGui)
					return errors.New("recipient name cant contain special chracters, spaces, cant start with number")
				}

			} else if len(s) < 47 && len(s) > 15 {
				noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
				matched, _ := regexp.MatchString("^[P][a-zA-Z0-9]{2,46}$", s)
				if noSpaces && matched {
					validRecipient = false
					updateSendButtonState()
					return nil
				} else {
					validRecipient = false
					updateSendButtonState()
					dialog.ShowError(errors.New("phantasma addresses cant contain\n special characters,\n spaces, \ncant start with number \nstarts with P"), mainWindowGui)
					return errors.New("phantasma addresses cant contain special characters, spaces, cant start with number and starts with P")
				}
			} else if len(s) == 47 {
				noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
				matched, _ := regexp.MatchString("^[P][a-zA-Z0-9]{2,46}$", s)
				if noSpaces && matched {
					validRecipient = true
					updateSendButtonState()
					return nil
				} else {
					validRecipient = false
					updateSendButtonState()
					dialog.ShowError(errors.New("phantasma addresses cant contain\n special characters,\n spaces, \ncant start with number \nstarts with P"), mainWindowGui)
					return errors.New("phantasma addresses cant contain special characters, spaces, cant start with number and starts with P")
				}

			} else if len(s) > 47 {
				validRecipient = false
				updateSendButtonState()
				dialog.ShowError(errors.New("phantasma addresses are shorter than 48 characters"), mainWindowGui)
				return errors.New("phantasma addresses are shorter than 48 characters")
			}
			validRecipient = false
			updateSendButtonState()
			return nil

		}
		gasSliderBox := container.NewVBox(container.NewGridWithRows(3, gasSliderLabel, gasSlider, warningLabel))

		amount.Validator = func(s string) error {
			validateUserAmount, _ := convertUserInputToBigInt(s, token.Decimals)
			if validateUserAmount == nil {
				validAmount = false
				warningLabel.TextStyle.Bold = true
				warning.Set("Amount is empty")
				updateSendButtonState()
				return errors.New("amount is empty")
			} else if validateUserAmount.Cmp(big.NewInt(0)) == 0 {
				validAmount = false
				warningLabel.TextStyle.Bold = true
				warning.Set("Amount is zero")
				updateSendButtonState()
				return errors.New("amount is zero")
			}
			if symbol == "KCAL" {
				feeAmount := new(big.Int).Mul(gasLimit, gasPrice)
				validateUserAmount = new(big.Int).Add(validateUserAmount, feeAmount)
			}
			if validateUserAmount.Cmp(&token.Amount) <= 0 {
				validAmount = true
				updateSendButtonState()
				return nil
			}
			fmt.Println("not enough balance", formatBalance(latestAccountData.FungibleTokens[symbol].Amount, latestAccountData.FungibleTokens[symbol].Decimals))
			validAmount = false
			updateSendButtonState()
			warningLabel.TextStyle.Bold = true
			warning.Set("Not enough balance")
			return errors.New("not enough balance")
		}

		updateSendButtonState()
		backButton := widget.NewButton("Maybe Later", func() {
			currentMainDialog.Hide()
		})

		buttonsBox := container.NewGridWithColumns(2, backButton, sendButton)

		sendTokenDiaContent := container.NewBorder(nil, buttonsBox, nil, nil, container.NewVBox(recipientBox, amountBox, gasSliderBox))
		gasSlider.OnChanged = func(value float64) {
			gasLimit.SetInt64(int64(value))
			updateSendButtonState()

		}
		sendDiaTitle := fmt.Sprintf("Specky is preparing for delivering your %s...", token.Name)
		d := dialog.NewCustomWithoutButtons(sendDiaTitle, sendTokenDiaContent, mainWindowGui)
		d.Resize(fyne.NewSize(720, 405))
		currentMainDialog = d
		currentMainDialog.Refresh()
		currentMainDialog.Show()
		mainWindowGui.Canvas().Focus(recipient)

	})

}
func checkFeeBalance(feeAmount *big.Int) error {
	kcalBalance := new(big.Int)

	if token, exists := latestAccountData.FungibleTokens["KCAL"]; exists {
		kcalBalance = &token.Amount
	} else {
		fmt.Println("Acc dont have Kcal")
		return fmt.Errorf("acc dont have Kcal to fill Specky's engines")
	}

	if kcalBalance.Cmp(big.NewInt(0)) == 0 {
		fmt.Println("Kcal balance not found")
		return fmt.Errorf("acc dont have Kcal to fill Specky's engines")
	}

	fmt.Printf("Kcal Balance: %s, Required Kcal: %s\n", formatBalance(*kcalBalance, kcalDecimals), formatBalance(*feeAmount, kcalDecimals))

	if kcalBalance.Cmp(feeAmount) < 0 {
		fmt.Printf("Insufficient Kcal: Required: %s, Available: %s\n", formatBalance(*feeAmount, kcalDecimals), formatBalance(*kcalBalance, kcalDecimals))
		return fmt.Errorf("insufficient Spark for filling Specky's engines. required: %s Kcal, available: %s Kcal", formatBalance(*feeAmount, kcalDecimals), formatBalance(*kcalBalance, kcalDecimals))
	}

	return nil
}
