package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

func confirmSendToken(WIF, tokenSymbol, to string, tokenAmount *big.Int, creds core.Credentials, name string, tokenFeeLimit *big.Int) error {
	// Show confirmation dialog
	sendTokenDecimal := core.LatestAccountData.FungibleTokens[tokenSymbol].Decimals
	var confirmMessage string
	if len(name) < 3 {
		confirmMessage = fmt.Sprintf("You are about to send %s %s to %s\nAre you sure to ignite engines?", core.FormatBalance(tokenAmount, int(sendTokenDecimal)), tokenSymbol, to)

	} else {
		confirmMessage = fmt.Sprintf("You are about to send \n%s %s to %s' related address\n%s\nAre you sure to ignite engines?", core.FormatBalance(tokenAmount, int(sendTokenDecimal)), tokenSymbol, name, to)

	}
	dialog.ShowConfirm("Confirm Transaction", confirmMessage, func(confirmed bool) {
		if confirmed {
			// Proceed with transaction if user confirmed
			go buildTokenTransaction(WIF, tokenSymbol, to, tokenAmount, creds, tokenFeeLimit)
		} else {
			// Handle abort action
			return
		}
	}, mainWindow)

	return nil
}

func buildTokenTransaction(WIF, tokenSymbol, to string, tokenAmount *big.Int, creds core.Credentials, tokenFeeLimit *big.Int) {
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
	// fmt.Println("Amount ", core.FormatBalance(*tokenAmount, core.LatestAccountData.FungibleTokens[tokenSymbol].Decimals))
	// fmt.Println("Gas Price ", gasPrice)
	// fmt.Println("Gas limit ", gasLimit)
	// fmt.Println("Expiration time ", time.Unix(expire, 0).Format("2006-01-02 15:04:05"))

	sb := scriptbuilder.BeginScript()
	sb.AllowGas(keyPair.Address().String(), cryptography.NullAddress().String(), core.UserSettings.GasPrice, tokenFeeLimit)
	sb.TransferTokens(tokenSymbol, keyPair.Address().String(), to, tokenAmount)
	sb.SpendGas(keyPair.Address().String())
	script := sb.EndScript()

	tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Transfer"))
	tx.Sign(keyPair)
	txHex := hex.EncodeToString(tx.Bytes())

	sendTransaction(txHex, creds)

}

func showSendTokenDia(symbol string, creds core.Credentials, decimal int8) {
	// Usage

	tokenFeeLimit := new(big.Int).Set(core.UserSettings.DefaultGasLimit)

	// Continue with your code here })
	var validAmount, validRecipient bool
	token := core.LatestAccountData.FungibleTokens[symbol]
	recipientFirstString := ""
	bindRecipient := binding.BindString(&recipientFirstString)
	recipient := widget.NewEntryWithData(bindRecipient)
	recipient.PlaceHolder = "Please enter or select delivery address/name"
	if core.UserSettings.SendOnly {
		recipient.Disable()
	} else {
		recipient.Enable()
	}
	var addressBookSelect *widget.Select
	var usersAddressesSelect *widget.Select

	addressBookSelect = widget.NewSelect(core.UserAddressBook.WalletOrder, func(s string) {
		if len(s) < 1 {
			return
		}
		if usersAddressesSelect.SelectedIndex() >= 0 {
			usersAddressesSelect.ClearSelected()
		}
		bindRecipient.Set(core.UserAddressBook.Wallets[s].Address)
		recText, _ := bindRecipient.Get()
		recipient.CursorColumn = len(recText)
		mainWindow.Canvas().Focus(recipient)

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
		mainWindow.Canvas().Focus(recipient)
	})

	usersAddressesSelect.PlaceHolder = "My Accounts"
	addressBookSelect.PlaceHolder = "Address Book"
	bindAmountString := ""
	bindAmount := binding.BindString(&bindAmountString)
	amount := widget.NewEntryWithData(bindAmount)
	amount.Bind(bindAmount)
	var amountBigInt *big.Int

	pasteBtn := widget.NewButtonWithIcon("", theme.ContentPasteIcon(), func() {
		recipient.SetText(spallet.Driver().AllWindows()[0].Clipboard().Content())
	})
	recipientSelectBox := container.NewGridWithColumns(2, addressBookSelect, usersAddressesSelect)
	recipientLyt := container.NewBorder(nil, nil, nil, pasteBtn, recipient)
	recipientBox := container.NewVBox(container.NewGridWithRows(2, recipientLyt, recipientSelectBox))

	sendButton := widget.NewButton("Engage", func() {
		// fmt.Println("Sending from WIF: ", creds.Wallets[creds.LastSelectedWallet].WIF)
		var err error
		var bindedAmount string
		var name string
		bindedAmount, err = bindAmount.Get()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		recipientEntry, err := bindRecipient.Get()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		if recipientEntry == "" {
			dialog.ShowError(errors.New("recipient is empty"), mainWindow)
			return
		} else if len(recipientEntry) <= 15 {
			name = recipientEntry
			nameToAddress, err := core.Client.LookupName(recipientEntry)
			if err != nil {
				dialog.ShowError(fmt.Errorf("specky encountered an error while searching this name\n%s", err), mainWindow)
				return
			}
			recipientEntry = nameToAddress
		}

		amountBigInt, err = core.ConvertUserInputToBigInt(bindedAmount, int(decimal))
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		} else {
			confirmSendToken(creds.Wallets[creds.LastSelectedWallet].WIF, symbol, recipientEntry, amountBigInt, creds, name, tokenFeeLimit)
		}

	})
	sendButton.Disable()

	amount.PlaceHolder = fmt.Sprintf("%s amount for delivery", symbol)

	gasLimitFloat, _ := tokenFeeLimit.Float64()
	gasSlider := widget.NewSlider(core.UserSettings.GasLimitSliderMin, core.UserSettings.GasLimitSliderMax)
	gasSlider.Value = gasLimitFloat
	gasSliderLabel := widget.NewLabelWithStyle("Specky's energy limit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	warning := binding.NewString()
	warning.Set("You have enough Kcal to fill Specky's engines")
	warningLabel := widget.NewLabelWithData(warning)
	warningLabel.Bind(warning)

	updateSendButtonState := func() {
		var amountErr error
		amountBigInt, amountErr = core.ConvertUserInputToBigInt(amount.Text, int(decimal))
		if amountErr != nil {
			warningLabel.TextStyle.Bold = true
			warning.Set(amountErr.Error())
			return
		}

		feeAmount := new(big.Int).Mul(core.UserSettings.GasPrice, tokenFeeLimit)
		if symbol == "KCAL" {
			feeAmount = feeAmount.Add(amountBigInt, feeAmount)
		}

		err := core.CheckFeeBalance(feeAmount)
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
		feeAmount := new(big.Int).Mul(tokenFeeLimit, core.UserSettings.GasPrice)
		if symbol == "KCAL" {
			if feeAmount.Cmp(token.Amount) < 0 {
				maxAmount := new(big.Int).Sub(token.Amount, feeAmount)
				amTxt := core.FormatBalance(maxAmount, core.KcalDecimals)
				bindAmount.Set(amTxt)
				amount.CursorColumn = len(amTxt)
				mainWindow.Canvas().Focus(amount)
				validAmount = true
				updateSendButtonState()
			} else {
				bindAmount.Set("Dont have enough Kcal")
				validAmount = false
				updateSendButtonState()

			}

		} else {

			amTxt := core.FormatBalance(token.Amount, token.Decimals)
			bindAmount.Set(amTxt)
			amount.CursorColumn = len(amTxt) // for some reason it dont wanna go end of the line sometimes goes tho strange tried refresh also but dunno
			validAmount = true
			updateSendButtonState()
			mainWindow.Canvas().Focus(amount)

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
				dialog.ShowError(errors.New("recipient name cant contain \nspecial characters, \nspaces, \ncant start with number,\ncapital letters"), mainWindow)
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
				dialog.ShowError(errors.New("phantasma addresses cant contain\n special characters,\n spaces, \ncant start with number \nstarts with P"), mainWindow)
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
				dialog.ShowError(errors.New("phantasma addresses cant contain\n special characters,\n spaces, \ncant start with number \nstarts with P"), mainWindow)
				return errors.New("phantasma addresses cant contain special characters, spaces, cant start with number and starts with P")
			}

		} else if len(s) > 47 {
			validRecipient = false
			updateSendButtonState()
			dialog.ShowError(errors.New("phantasma addresses are shorter than 48 characters"), mainWindow)
			return errors.New("phantasma addresses are shorter than 48 characters")
		}
		validRecipient = false
		updateSendButtonState()
		return nil

	}
	gasSliderBox := container.NewVBox(container.NewGridWithRows(3, gasSliderLabel, gasSlider, warningLabel))

	amount.Validator = func(s string) error {
		validateUserAmount, _ := core.ConvertUserInputToBigInt(s, token.Decimals)
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
			feeAmount := new(big.Int).Mul(tokenFeeLimit, core.UserSettings.GasPrice)
			validateUserAmount = new(big.Int).Add(validateUserAmount, feeAmount)
		}
		if validateUserAmount.Cmp(token.Amount) <= 0 {
			validAmount = true
			updateSendButtonState()
			return nil
		}
		fmt.Println("not enough balance", core.FormatBalance(core.LatestAccountData.FungibleTokens[symbol].Amount, core.LatestAccountData.FungibleTokens[symbol].Decimals))
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
		tokenFeeLimit.SetInt64(int64(value))
		updateSendButtonState()

	}
	sendDiaTitle := fmt.Sprintf("Send %s...", token.Name)

	d := dialog.NewCustomWithoutButtons(sendDiaTitle, sendTokenDiaContent, mainWindow)
	width := mainWindowLyt.Selected().Content.Size().Width
	d.Resize(fyne.NewSize(width, 0))
	currentMainDialog = d
	currentMainDialog.Refresh()
	currentMainDialog.Show()
	mainWindow.Canvas().Focus(recipient)

}
