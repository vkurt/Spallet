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

func sendNFTConfirm(WIF string, nftsToSend map[string][]string, to string, name string, creds core.Credentials, nftFeeLimit *big.Int) error {
	// Construct the confirmation message
	var nftDetails string
	var confirmMessage string
	for id, symbols := range nftsToSend {
		nftDetails += fmt.Sprintf("%s (%s)\n", id, symbols[0])
	}

	if len(name) > 2 {

		confirmMessage = fmt.Sprintf("You are about to send the following NFTs to %s's\n address:%s\n%s\nAre you sure to ignite engines?", name, to, nftDetails)
	} else {
		confirmMessage = fmt.Sprintf("You are about to send the following NFTs to %s:\n%s\nAre you sure to ignite engines?", to, nftDetails)
	}

	// Show confirmation dialog
	dialog.ShowConfirm("Confirm Transaction", confirmMessage, func(confirmed bool) {
		if confirmed {
			// Proceed with transaction if user confirmed
			go buildNftTransaction(WIF, nftsToSend, to, creds, nftFeeLimit)
		} else {
			// Handle abort action

		}
	}, mainWindow)

	return nil
}

func buildNftTransaction(WIF string, nftsToSend map[string][]string, to string, creds core.Credentials, nftFeeLimit *big.Int) {
	keyPair, err := cryptography.FromWIF(WIF)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Transaction Failed",
			Content: fmt.Sprintf("Invalid WIF: %v", err),
		})
		return
	}
	from := keyPair.Address().String()

	expire := time.Now().UTC().Add(time.Second * 300).Unix()
	fmt.Println("expiration time: ", expire)
	sb := scriptbuilder.BeginScript()
	sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, nftFeeLimit)
	for id, symbol := range nftsToSend {
		fmt.Printf("Sending nft : %v %v\n", symbol[0], id)
		sb.CallInterop("Runtime.TransferToken", from, to, symbol[0], id)
	}
	sb.SpendGas(keyPair.Address().String())
	script := sb.EndScript()
	tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Transfer"))
	tx.Sign(keyPair)
	txHex := hex.EncodeToString(tx.Bytes())

	sendTransaction(txHex, creds)
}

func showSendNFTDia(symbol string, creds core.Credentials) {
	// Usage

	nftFeeLimit := new(big.Int).Set(core.UserSettings.DefaultGasLimit)

	// Continue with your code here })
	token := core.LatestAccountData.NonFungible[symbol]
	var validRecipient bool
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
	pasteBtn := widget.NewButtonWithIcon("", theme.ContentPasteIcon(), func() {
		recipient.SetText(spallet.Driver().AllWindows()[0].Clipboard().Content())
	})

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

	usersAddressesSelect.PlaceHolder = "My Accounts"
	addressBookSelect.PlaceHolder = "Address Book"

	recipientSelectBox := container.NewGridWithColumns(2, addressBookSelect, usersAddressesSelect)
	recipientLyt := container.NewBorder(nil, nil, nil, pasteBtn, recipient)
	recipientBox := container.NewVBox(container.NewGridWithRows(2, recipientLyt, recipientSelectBox))

	nftSelections := container.NewGridWithColumns(2)
	var selectedNFTs = make(map[string][]string)

	gasLimitFloat, _ := core.UserSettings.DefaultGasLimit.Float64()
	gasSlider := widget.NewSlider(core.UserSettings.GasLimitSliderMin, core.UserSettings.GasLimitSliderMax)
	gasSlider.Value = gasLimitFloat
	gasSliderLabel := widget.NewLabelWithStyle("Specky's energy limit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	warning := binding.NewString()
	warning.Set("You have enough Kcal to fill Specky's engines")
	warningLabel := widget.NewLabelWithData(warning)
	warningLabel.Bind(warning)

	amount := 0
	sendButton := widget.NewButton("Engage", func() {
		// fmt.Println("Sending from WIF: ", creds.Wallets[creds.LastSelectedWallet].WIF) // Ensure this is the correct WIF
		var name string

		if recipient.Text == "" {
			dialog.ShowError(errors.New("recipient is empty"), mainWindow)
			return
		} else if len(recipient.Text) <= 15 {
			name = recipient.Text
			nameToAddress, err := core.Client.LookupName(recipient.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("specky encountered an error while searching this name\n%s", err), mainWindow)
				return
			}
			recipient.Text = nameToAddress
		}

		sendNFTConfirm(creds.Wallets[creds.LastSelectedWallet].WIF, selectedNFTs, recipient.Text, name, creds, nftFeeLimit)
	})
	sendButton.Disable()
	updateSendButtonState := func() {
		feeAmount := new(big.Int).Mul(nftFeeLimit, core.UserSettings.GasPrice)
		err := core.CheckFeeBalance(feeAmount) // Call the function with parentheses
		fmt.Printf("Update Submit Button: Amount: %v, Error: %v\n", amount, err)
		if err != nil {
			warningLabel.TextStyle.Bold = true
			warning.Set(err.Error())
			sendButton.Disable()
		} else if !validRecipient {
			warningLabel.TextStyle.Bold = true
			warning.Set("Please enter a valid address/name")
			sendButton.Disable()
		} else if amount == 0 {
			warningLabel.TextStyle.Bold = true
			warning.Set("Please select Nft from below list")
			sendButton.Disable()
		} else {
			warningLabel.TextStyle.Bold = false
			warning.Set("Specky is ready for launch!")
			sendButton.Enable()
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
				return errors.New("phantasma addresses cant contain special characters, spaces, cant start with number and starts with 'P'")
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
	gasSlider.OnChanged = func(value float64) {
		nftFeeLimit.SetInt64(int64(value))
		updateSendButtonState()
	}

	updateSendButtonState()
	backButton := widget.NewButton("Maybe Later", func() {
		currentMainDialog.Hide()
	})
	buttonsBox := container.NewGridWithColumns(2, backButton, sendButton)
	gasSliderBox := container.NewVBox(container.NewBorder(nil, nil, gasSliderLabel, nil, gasSlider), warningLabel)
	// Create checkboxes for each NFT ID
	for _, nft := range core.LatestAccountData.NonFungible {
		if nft.Symbol == symbol {

			img := getResourceByName(symbol)
			for _, id := range nft.Ids {
				title := id
				//  if len(title) > 20 {
				// 	title = title[:17] + "..."
				// }
				subTitle := id
				// if len(subTitle) > 20 {
				// 	subTitle = subTitle[:17] + "..."
				// }
				check := core.NewSelectableCard(title, subTitle, img, "Details", func() {}, func(selected bool) {
					if selected {
						fmt.Printf("Selected NFT: %s\n", id)
						selectedNFTs[id] = append(selectedNFTs[id], symbol)
						fmt.Printf("Nft List \n %v\n", selectedNFTs)
						amount++
						updateSendButtonState()
					} else {
						fmt.Printf("Deselected NFT: %s\n", id)
						delete(selectedNFTs, id)
						fmt.Printf("Nft List \n %v\n", selectedNFTs)
						amount--
						updateSendButtonState()
					}

				})
				nftSelections.Add(check)
			}
		}
	}
	senNftDiaTitle := fmt.Sprintf("Send %s...", token.Name)
	scrollContainer := container.NewVScroll(nftSelections)
	nftDiaContent := container.NewBorder(container.NewVBox(recipientBox, gasSliderBox), buttonsBox, nil, nil, scrollContainer)
	d := dialog.NewCustomWithoutButtons(senNftDiaTitle, nftDiaContent, mainWindow)
	d.Resize(mainWindowLyt.Selected().Content.Size())
	currentMainDialog = d
	currentMainDialog.Refresh()
	currentMainDialog.Show()
	mainWindow.Canvas().Focus(recipient)

}
