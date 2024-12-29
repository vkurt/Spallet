package main

import (
	"errors"
	"fmt"
	"log"
	"spallet/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showAdddressBookWin(pwd string) {
	addrBkWin := spallet.NewWindow("Manage your address book")
	// Usage
	askPwdDia(core.UserSettings.AskPwd, pwd, mainWindow, func(correct bool) {
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
				core.UserAddressBook.WalletOrder[index], core.UserAddressBook.WalletOrder[index-1] = core.UserAddressBook.WalletOrder[index-1], core.UserAddressBook.WalletOrder[index]
				if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), addrBkWin)
				}
				buildWalletButtons()
				walletScroll.Content.Refresh()
			}
		}

		moveDown := func(index int) {
			if index < len(core.UserAddressBook.WalletOrder)-1 {
				core.UserAddressBook.WalletOrder[index], core.UserAddressBook.WalletOrder[index+1] = core.UserAddressBook.WalletOrder[index+1], core.UserAddressBook.WalletOrder[index]
				if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), addrBkWin)
				}
				buildWalletButtons()
				walletScroll.Content.Refresh()
			}
		}

		moveTop := func(index int) {
			wallet := core.UserAddressBook.WalletOrder[index]
			core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder[:index], core.UserAddressBook.WalletOrder[index+1:]...)
			core.UserAddressBook.WalletOrder = append([]string{wallet}, core.UserAddressBook.WalletOrder...)
			if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
				log.Println("Failed to save address book:", err)
				dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), addrBkWin)
			}
			buildWalletButtons()
			walletScroll.Content.Refresh()
		}

		moveBottom := func(index int) {
			wallet := core.UserAddressBook.WalletOrder[index]
			core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder[:index], core.UserAddressBook.WalletOrder[index+1:]...)
			core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder, wallet)
			if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
				log.Println("Failed to save address book:", err)
				dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), addrBkWin)
			}
			buildWalletButtons()
			walletScroll.Content.Refresh()
		}

		buildWalletButtons = func() {
			walletButtons.Objects = nil
			for index, walletName := range core.UserAddressBook.WalletOrder {
				wallet := core.UserAddressBook.Wallets[walletName]
				shortAddr := wallet.Address[:8] + "..." + wallet.Address[len(wallet.Address)-9:]
				walletBtn := widget.NewButton(wallet.Name+"\n"+shortAddr, func() {

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
				if index == len(core.UserAddressBook.WalletOrder)-1 {
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
						wallet := core.UserAddressBook.Wallets[walletName]
						wallet.Name = renameEntry.Text
						core.UserAddressBook.Wallets[renameEntry.Text] = wallet
						delete(core.UserAddressBook.Wallets, walletName)
						for i, name := range core.UserAddressBook.WalletOrder {
							if name == walletName {
								core.UserAddressBook.WalletOrder[i] = renameEntry.Text
								break
							}
						}

						if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
							dialog.NewError(err, addrBkWin)
							return
						}

						adddressBookDia.Hide()
						adddressBookDia = dialog.NewInformation("Succesfully saved", fmt.Sprintf("New name saved for '%s' as '%s'", wallet.Address, renameEntry.Text), addrBkWin)
						buildWalletButtons()
						walletScroll.Content.Refresh()
					})
					saveBttn.Disable()
					backBttn := widget.NewButton("Back", func() {
						adddressBookDia.Hide()
					})

					renameEntry.Validator = func(s string) error {
						warning, err := core.ValidateAccountInput(core.UserAddressBook.WalletOrder, nil, s, "name", false)
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
					renameDia := dialog.NewCustomWithoutButtons(fmt.Sprintf("Rename %s", core.UserAddressBook.Wallets[walletName].Address), renameContent, addrBkWin)
					adddressBookDia = renameDia
					adddressBookDia.Show()
				})

				removeBttn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {

					dialog.ShowForm("Remove Address", "Remove", "Cancel", []*widget.FormItem{

						widget.NewFormItem("Name", widget.NewLabel(wallet.Name)),
						widget.NewFormItem("Address", widget.NewLabel(wallet.Address)),
					}, func(ok bool) {
						if ok {

							delete(core.UserAddressBook.Wallets, walletName)
							for i, name := range core.UserAddressBook.WalletOrder {
								if name == walletName {
									core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder[:i], core.UserAddressBook.WalletOrder[i+1:]...)
									break
								}
							}

							if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
								dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
							}
							adddressBookDia = dialog.NewInformation("Address Removed", "Address removed succesfully", addrBkWin)
							buildWalletButtons()
							adddressBookDia.Show()
							walletScroll.Content.Refresh()

						}
					}, addrBkWin)
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
		if len(core.UserAddressBook.WalletOrder) < 1 {
			walletButtons.Add(container.NewVBox(widget.NewLabel("Please Add an dddress")))
		}
		walletScroll = container.NewVScroll(walletButtons)

		walletScroll.SetMinSize(fyne.NewSize(600, 400))

		addWallet := widget.NewButtonWithIcon("Add Address", theme.ContentAddIcon(), func() {

			walletnamefrst := ""
			walletNameBind := binding.BindString(&walletnamefrst)
			walletNameEntry := widget.NewEntryWithData(walletNameBind)
			walletNameEntry.PlaceHolder = "Enter a name for address"
			nameSuggest := fmt.Sprintf("Sparky address %v", len(core.UserAddressBook.WalletOrder)+1)
			walletNameEntry.SetText(nameSuggest)
			warningFrst := ""
			nameEntryWarning := binding.BindString(&warningFrst)
			warningLabel := widget.NewLabelWithData(nameEntryWarning)
			warningLabel.Wrapping = fyne.TextWrapWord

			addressEntry := widget.NewEntry()
			addressEntry.PlaceHolder = "Enter an address"
			addressEntry.Validator = func(s string) error {
				result, err := core.ValidateAccountInput(nil, core.UserAddressBook.Wallets, "", "address", true, s)
				nameEntryWarning.Set(result)
				return err

			}

			walletNameEntry.Validator = func(s string) error {

				result, err := core.ValidateAccountInput(core.UserAddressBook.WalletOrder, nil, s, "name", false)
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

					core.UserAddressBook.Wallets[walletEntry] = core.Wallet{
						Name:    walletEntry,
						Address: address,
					}
					core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder, walletEntry)
					if err := core.SaveAddressBook(core.UserAddressBook, pwd, rootPath); err != nil {
						log.Println("Failed to save adress book:", err)
						dialog.ShowInformation("Error", "Failed to save adress book: "+err.Error(), addrBkWin)
					} else {
						adddressBookDia = dialog.NewInformation("Address saved", "Address saved successfully", addrBkWin)
						adddressBookDia.Show()
						buildWalletButtons()
						walletScroll.Content.Refresh()
					}
				}
			}, addrBkWin)
			addForm.Resize(fyne.NewSize(600, 300))
			addForm.Show()

			addressEntry.SetValidationError(errors.New("please enter address"))

			walletNameEntry.CursorRow = len(walletNameEntry.Text)
			walletNameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
			walletNameEntry.Refresh()
			addrBkWin.Canvas().Focus(walletNameEntry)
		})

		backButton := widget.NewButton("Back", func() {

			addrBkWin.Hide()
		})

		accountsLayout := container.NewBorder(walletScroll, container.NewVBox(addWallet, backButton), nil, nil)
		addrBkWin.SetContent(accountsLayout)
		// dia := dialog.NewCustomWithoutButtons("Manage Your Address Book", accountsLayout, mainWindowGui)
		// currentMainDialog = dia
		addrBkWin.Show()
	})

}
