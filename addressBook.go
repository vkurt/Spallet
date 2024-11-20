package main

import (
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type addressBook struct {
	WalletOrder []string
	Wallets     map[string]Wallet
}

var userAddressBook = addressBook{
	WalletOrder: []string{},
	Wallets:     make(map[string]Wallet)}

func adddressBookDia(pwd string) {

	// Usage
	askPwdDia(userSettings.AskPwd, pwd, mainWindowGui, func(correct bool) {
		fmt.Println("result", correct)
		if !correct {
			return
		}
		// Continue with your code here
		walletButtons := container.NewVBox()
		var adddressBookDia dialog.Dialog
		// var maxWidth float32 = 0.0
		var buildWalletButtons func()
		// for _, walletName := range userAddressBook.WalletOrder {
		// 	wallet := userAddressBook.Wallets[walletName]
		// 	btn := widget.NewButton(wallet.Name+"\n"+wallet.Address, func() {})
		// 	btnSize := btn.MinSize()
		// 	if btnSize.Width > maxWidth {
		// 		maxWidth = btnSize.Width
		// 	}
		// }

		walletScroll := container.NewVScroll(walletButtons)

		moveUp := func(index int) {
			if index > 0 {
				userAddressBook.WalletOrder[index], userAddressBook.WalletOrder[index-1] = userAddressBook.WalletOrder[index-1], userAddressBook.WalletOrder[index]
				if err := saveAddressBook(userAddressBook, pwd); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
				}
				buildWalletButtons()
				walletScroll.Content.Refresh()
			}
		}

		moveDown := func(index int) {
			if index < len(userAddressBook.WalletOrder)-1 {
				userAddressBook.WalletOrder[index], userAddressBook.WalletOrder[index+1] = userAddressBook.WalletOrder[index+1], userAddressBook.WalletOrder[index]
				if err := saveAddressBook(userAddressBook, pwd); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
				}
				buildWalletButtons()
				walletScroll.Content.Refresh()
			}
		}

		moveTop := func(index int) {
			wallet := userAddressBook.WalletOrder[index]
			userAddressBook.WalletOrder = append(userAddressBook.WalletOrder[:index], userAddressBook.WalletOrder[index+1:]...)
			userAddressBook.WalletOrder = append([]string{wallet}, userAddressBook.WalletOrder...)
			if err := saveAddressBook(userAddressBook, pwd); err != nil {
				log.Println("Failed to save address book:", err)
				dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
			}
			buildWalletButtons()
			walletScroll.Content.Refresh()
		}

		moveBottom := func(index int) {
			wallet := userAddressBook.WalletOrder[index]
			userAddressBook.WalletOrder = append(userAddressBook.WalletOrder[:index], userAddressBook.WalletOrder[index+1:]...)
			userAddressBook.WalletOrder = append(userAddressBook.WalletOrder, wallet)
			if err := saveAddressBook(userAddressBook, pwd); err != nil {
				log.Println("Failed to save address book:", err)
				dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
			}
			buildWalletButtons()
			walletScroll.Content.Refresh()
		}

		buildWalletButtons = func() {
			walletButtons.Objects = nil
			for index, walletName := range userAddressBook.WalletOrder {
				wallet := userAddressBook.Wallets[walletName]
				walletBtn := widget.NewButton(wallet.Name+"\n"+wallet.Address, func() {

				})
				walletBtn.Importance = widget.HighImportance

				moveUpBttn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
					moveUp(index)
				})
				moveDownBttn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
					moveDown(index)
				})
				moveTopBttn := widget.NewButtonWithIcon("", theme.UploadIcon(), func() {
					moveTop(index)
				})
				moveBotBttn := widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
					moveBottom(index)
				})
				if index == len(userAddressBook.WalletOrder)-1 {
					moveBotBttn.Disable()
					moveDownBttn.Disable()

				} else if index == 0 {
					moveTopBttn.Disable()
					moveUpBttn.Disable()
				} else {
					moveTopBttn.Enable()
					moveUpBttn.Enable()
					moveBotBttn.Enable()
					moveDownBttn.Enable()
				}
				moveButtons := container.NewGridWithColumns(2, moveUpBttn, moveTopBttn, moveDownBttn, moveBotBttn)
				renameButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
					renameEntry := widget.NewEntry()
					renameEntry.PlaceHolder = "Enter new name for address"
					nameEntryWarningFrst := ""
					nameEntryWarning := binding.BindString(&nameEntryWarningFrst)
					nameEntryWarningLabel := widget.NewLabelWithData(nameEntryWarning)
					saveBttn := widget.NewButton("Save", func() {
						wallet := userAddressBook.Wallets[walletName]
						wallet.Name = renameEntry.Text
						userAddressBook.Wallets[renameEntry.Text] = wallet
						delete(userAddressBook.Wallets, walletName)
						for i, name := range userAddressBook.WalletOrder {
							if name == walletName {
								userAddressBook.WalletOrder[i] = renameEntry.Text
								break
							}
						}

						if err := saveAddressBook(userAddressBook, pwd); err != nil {
							dialog.NewError(err, mainWindowGui)
							return
						}

						adddressBookDia.Hide()
						adddressBookDia = dialog.NewInformation("Succesfully saved", fmt.Sprintf("New name saved for '%s' as '%s'", wallet.Address, renameEntry.Text), mainWindowGui)
						buildWalletButtons()
						walletScroll.Content.Refresh()
					})
					saveBttn.Disable()
					backBttn := widget.NewButton("Back", func() {
						adddressBookDia.Hide()
					})

					renameEntry.Validator = func(s string) error {
						warning, err := validateAccountInput(userAddressBook.WalletOrder, nil, s, "name", false)
						fmt.Println("rename err", err)
						nameEntryWarning.Set(warning)
						if err != nil {
							saveBttn.Disable()
						} else {
							saveBttn.Enable()
						}
						return err

					}

					buttonsContainer := container.NewGridWithColumns(2, backBttn, saveBttn)
					renameContent := container.NewVBox(renameEntry, nameEntryWarningLabel, buttonsContainer)
					renameDia := dialog.NewCustomWithoutButtons(fmt.Sprintf("Rename %s", userAddressBook.Wallets[walletName].Address), renameContent, mainWindowGui)
					adddressBookDia = renameDia
					adddressBookDia.Show()
				})

				removeBttn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {

					dialog.ShowForm("Remove Address", "Remove", "Cancel", []*widget.FormItem{

						widget.NewFormItem("Name", widget.NewLabel(wallet.Name)),
						widget.NewFormItem("Address", widget.NewLabel(wallet.Address)),
					}, func(ok bool) {
						if ok {

							delete(userAddressBook.Wallets, walletName)
							for i, name := range userAddressBook.WalletOrder {
								if name == walletName {
									userAddressBook.WalletOrder = append(userAddressBook.WalletOrder[:i], userAddressBook.WalletOrder[i+1:]...)
									break
								}
							}

							if err := saveAddressBook(userAddressBook, pwd); err != nil {
								dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
							}
							adddressBookDia = dialog.NewInformation("Address Removed", "Address removed succesfully", mainWindowGui)
							buildWalletButtons()
							adddressBookDia.Show()
							walletScroll.Content.Refresh()

						}
					}, mainWindowGui)
				})

				btnContainer := container.NewGridWithRows(1,
					renameButton,

					removeBttn,
				)
				btnContainer.Resize(fyne.NewSize(120, btnContainer.MinSize().Height))

				walletGroup := container.NewBorder(nil, widget.NewSeparator(), moveButtons, btnContainer, walletBtn)
				walletButtons.Add(walletGroup)
			}
		}

		buildWalletButtons()
		if len(userAddressBook.WalletOrder) < 1 {
			walletButtons.Add(container.NewVBox(widget.NewLabel("Please Add an dddress")))
		}
		walletScroll = container.NewVScroll(walletButtons)

		walletScroll.SetMinSize(fyne.NewSize(600, 400))

		addWallet := widget.NewButtonWithIcon("Add Address", theme.ContentAddIcon(), func() {

			walletnamefrst := ""
			walletNameBind := binding.BindString(&walletnamefrst)
			walletNameEntry := widget.NewEntryWithData(walletNameBind)
			walletNameEntry.PlaceHolder = "Enter a name for address"
			nameSuggest := fmt.Sprintf("Sparky address %v", len(userAddressBook.WalletOrder)+1)
			walletNameEntry.SetText(nameSuggest)
			warningFrst := ""
			nameEntryWarning := binding.BindString(&warningFrst)
			warningLabel := widget.NewLabelWithData(nameEntryWarning)
			warningLabel.Wrapping = fyne.TextWrapWord

			addressEntry := widget.NewEntry()
			addressEntry.PlaceHolder = "Enter an address"
			addressEntry.Validator = func(s string) error {
				result, err := validateAccountInput(nil, userAddressBook.Wallets, "", "address", true, s)
				nameEntryWarning.Set(result)
				return err

			}

			walletNameEntry.Validator = func(s string) error {

				result, err := validateAccountInput(userAddressBook.WalletOrder, nil, s, "name", false)
				nameEntryWarning.Set(result)
				return err

			}

			addForm := dialog.NewForm("Add New Address", "Save", "Cancel", []*widget.FormItem{
				widget.NewFormItem("Name", walletNameEntry),
				widget.NewFormItem("Address", addressEntry),
				widget.NewFormItem("", warningLabel),
			}, func(ok bool) {
				if ok {
					walletEntry, _ := walletNameBind.Get()

					address := addressEntry.Text

					userAddressBook.Wallets[walletEntry] = Wallet{
						Name:    walletEntry,
						Address: address,
					}
					userAddressBook.WalletOrder = append(userAddressBook.WalletOrder, walletEntry)
					if err := saveAddressBook(userAddressBook, pwd); err != nil {
						log.Println("Failed to save adress book:", err)
						dialog.ShowInformation("Error", "Failed to save adress book: "+err.Error(), mainWindowGui)
					} else {
						adddressBookDia = dialog.NewInformation("Address saved", "Address saved successfully", mainWindowGui)
						adddressBookDia.Show()
						buildWalletButtons()
						walletScroll.Content.Refresh()
					}
				}
			}, mainWindowGui)
			addForm.Resize(fyne.NewSize(600, 300))
			addForm.Show()

			addressEntry.SetValidationError(errors.New("please enter address"))

			walletNameEntry.CursorRow = len(walletNameEntry.Text)
			walletNameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
			walletNameEntry.Refresh()
			mainWindowGui.Canvas().Focus(walletNameEntry)
		})

		backButton := widget.NewButton("Back", func() {

			currentMainDialog.Hide()
		})

		accountsLayout := container.NewBorder(walletScroll, container.NewVBox(addWallet, backButton), nil, nil)
		dia := dialog.NewCustomWithoutButtons("Manage Your Address Book", accountsLayout, mainWindowGui)
		currentMainDialog = dia
		currentMainDialog.Show()
	})

}
