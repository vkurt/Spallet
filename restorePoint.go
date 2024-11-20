package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showBackupDia(creds Credentials) {
	backupInfo := widget.NewRichTextFromMarkdown("Safeguard your valuable data with these easy-to-use options.\n\n1 **Secure Backup:** Save all wallet data securely using your current password inside a folder named 'spallet backup ddmmyyyyhhmm'\n\n2 **Data Restore:** Restore your data from a backup folder.")
	backupInfo.Wrapping = fyne.TextWrapWord
	backBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { currentMainDialog.Hide() })

	var backup, restore bool
	continueBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		if backup {
			bckupDia := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil {
					dialog.ShowError(err, mainWindowGui)
					return
				}
				if uri == nil { // User clicked "Cancel"
					return
				}

				destFolder := uri.Path()

				sourceFolder := "data/essential" // Replace with your folder path
				timestamp := time.Now().Format("020120061504")
				newFolder := filepath.Join(destFolder, fmt.Sprintf("spallet backup %s", timestamp))

				err = os.MkdirAll(newFolder, os.ModePerm)
				if err != nil {
					dialog.ShowError(err, mainWindowGui)
					return
				}

				err = backupCopyFolder(sourceFolder, newFolder)
				if err != nil {
					dialog.ShowError(err, mainWindowGui)
				} else {
					dialog.ShowInformation("Success", "All data saved successfully!", mainWindowGui)
				}
			}, mainWindowGui)
			bckupDia.Resize(fyne.NewSize(mainWindowGui.Canvas().Size().Width-50, mainWindowGui.Canvas().Size().Height-50))
			bckupDia.SetConfirmText("Save Here")
			bckupDia.SetDismissText("Cancel")
			bckupDia.Show()

		} else if restore {
			var restoreDia dialog.Dialog

			// var fullRestoreDia dialog.Dialog
			var openFolderDia *dialog.FileDialog
			pwd := ""
			rstBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
				restoreDia.Hide()
			})
			continueBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {

				openFolderDia = dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, mainWindowGui)
						return
					}
					if uri == nil {
						return
					}

					directory := uri.Path()

					expectedFiles := []string{"addressbook.spallet", "credentials.spallet", "settings.spallet"}
					notFoundFiles := ""
					foundFiles := ""
					pwdEntry := widget.NewPasswordEntry()
					pwdEntry.OnChanged = func(s string) {
						pwd = s
					}
					pwdEntryFrmItm := widget.NewFormItem("Password", pwdEntry)
					askPwdDia := dialog.NewForm("Enter Password for this data", "Continue", "Cancel", []*widget.FormItem{
						pwdEntryFrmItm,
					}, func(b bool) {
						if b {
							foundAccounts := 0
							foundAddress := 0
							settingsRestored := false
							for _, fileName := range expectedFiles {
								filePath := filepath.Join(directory, fileName)
								if _, err := os.Stat(filePath); err == nil {
									foundFiles += fmt.Sprintf(fileName + "\n")

									switch fileName {
									case "credentials.spallet": // restoring unsaved accounts
										// fmt.Println(pwd)
										ldCreds, err := loadCredentials(filePath, pwd)
										if err != nil {
											dialog.ShowError(err, mainWindowGui)
											return
										}

										for _, ldCredsWallet := range ldCreds.Wallets {
											isSavedWallet := false
											isSavedName := false
											// fmt.Println("restore wallet", ldCredsWallet.Name)
											for _, savedWallet := range creds.Wallets {

												if savedWallet.WIF == ldCredsWallet.WIF {
													isSavedWallet = true

												}
												if savedWallet.Name == ldCredsWallet.Name {
													isSavedName = true
												}

											}

											if !isSavedWallet && isSavedName {
												name := fmt.Sprintf("%v...%v", ldCredsWallet.Name[:8], ldCredsWallet.Name[len(ldCredsWallet.Name)-8:len(ldCredsWallet.Name)]) //if user registered same name giving it to a new name
												walletToAdd := Wallet{
													Name:    name,
													Address: ldCredsWallet.Address,
													WIF:     ldCredsWallet.WIF,
												}
												foundAccounts++
												creds.Wallets[name] = walletToAdd
												creds.WalletOrder = append(creds.WalletOrder, name)
											} else if !isSavedWallet && !isSavedName {
												creds.Wallets[ldCredsWallet.Name] = ldCredsWallet
												creds.WalletOrder = append(creds.WalletOrder, ldCredsWallet.Name)
												foundAccounts++
											}

										}
									case "addressbook.spallet": // restoring unsaved addresses to addressbook
										ldAdrBk, err := loadAddressBook(filePath, pwd)
										if err != nil {
											dialog.ShowError(err, mainWindowGui)
											return
										}

										for _, ldAddrBkAddr := range ldAdrBk.Wallets {
											isSavedWallet := false
											isSavedName := false
											fmt.Println("restore adress", ldAddrBkAddr.Name)
											for _, savedAddr := range userAddressBook.Wallets {

												if savedAddr.Address == ldAddrBkAddr.Address {
													isSavedWallet = true

												}
												if savedAddr.Name == ldAddrBkAddr.Name {
													isSavedName = true
												}

											}

											if !isSavedWallet && isSavedName {
												name := fmt.Sprintf("%v...%v", ldAddrBkAddr.Address[:8], ldAddrBkAddr.Address[len(ldAddrBkAddr.Address)-8:len(ldAddrBkAddr.Address)])
												walletToAdd := Wallet{
													Name:    name,
													Address: ldAddrBkAddr.Address,
												}
												foundAddress++
												userAddressBook.Wallets[name] = walletToAdd
												userAddressBook.WalletOrder = append(userAddressBook.WalletOrder, name)
											} else if !isSavedWallet && !isSavedName {
												userAddressBook.Wallets[ldAddrBkAddr.Name] = ldAddrBkAddr
												userAddressBook.WalletOrder = append(userAddressBook.WalletOrder, ldAddrBkAddr.Name)
												foundAddress++
											}

										}

									case "settings.spallet": //restoring user settings
										loadSettings(filePath)
										saveSettings()
										settingsRestored = true

									}

								} else {
									notFoundFiles += fmt.Sprintf(fileName + "\n")
								}

							}

							restoreInfo := ""
							if foundAccounts > 0 {
								restoreInfo += fmt.Sprintf("Found %v new accounts and added them to your wallet data\n", foundAccounts)
							} else {
								restoreInfo += fmt.Sprintln("Cant find any new account")
							}

							if foundAddress > 0 {
								restoreInfo += fmt.Sprintf("Found %v new addresses and added them into your address book\n", foundAddress)

							} else {
								restoreInfo += fmt.Sprintln("Cant find any new address.")
							}

							if settingsRestored {
								restoreInfo += fmt.Sprintln("Found settings and applied")
							} else {
								restoreInfo += fmt.Sprintln("Cant find settings")
							}

							restoreInfo += fmt.Sprintf("\nFound Files\n%s\nNot Found Files\n%s", foundFiles, notFoundFiles)

							if foundAccounts > 0 || foundAddress > 0 || settingsRestored {
								fmt.Println("foundAccounts", foundAccounts)
								if err := saveCredentials(creds); err != nil {
									log.Println("Failed to save credentials:", err)
									dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
								}

								if err := saveAddressBook(userAddressBook, creds.Password); err != nil {
									log.Println("Failed to save Address Book:", err)
									dialog.ShowInformation("Error", "Failed to save Address Book: "+err.Error(), mainWindowGui)
								}
								currentMainDialog.Hide()
								restoreDia.Hide()
								dialog.ShowInformation("Found new data", restoreInfo, mainWindowGui)
								mainWindow(creds, regularTokens, nftTokens)

							} else {

								dialog.ShowInformation("Restore Failed", fmt.Sprintf("Cant find any new data please make sure file names are correct or your wallet data same with backup data\nFound Files\n%s\nNot Found Files\n%s", foundFiles, notFoundFiles), mainWindowGui)

							}
						}
					}, mainWindowGui)
					askPwdDia.Show()
					mainWindowGui.Canvas().Focus(pwdEntry)
				}, mainWindowGui)
				openFolderDia.SetConfirmText("Restore From Here")
				openFolderDia.Resize(fyne.NewSize(mainWindowGui.Canvas().Size().Width-50, mainWindowGui.Canvas().Size().Height-50))
				openFolderDia.Show()

			})

			restoreExplaination := widget.NewRichTextFromMarkdown("Fully restores data from the backup folder, please select folder contains you want to restore \n\n1- credentials.spallet file adds only unsaved accounts(private keys) to your current accounts. \n\n2- addressbook.spallet adds only unsaved addresses to your address book.\n\n3- settings.spallet overwrites your current settings.\n\n⚠️**You can delete files from back up folder if you dont want them to add**⚠️")
			restoreExplaination.Wrapping = fyne.TextWrapWord

			bttns := container.NewGridWithColumns(2, rstBckBttn, continueBttn)

			restoreOptDiaLyt := container.NewBorder(nil, bttns, nil, nil, container.NewVBox(restoreExplaination))

			restoreDia = dialog.NewCustomWithoutButtons("Restore information", restoreOptDiaLyt, mainWindowGui)
			restoreDia.Resize(fyne.NewSize(600, 340))
			restoreDia.Show()
		}
	})
	continueBttn.Disable()
	backupButtons := container.NewGridWithColumns(2, backBttn, continueBttn)
	backupOptions := widget.NewRadioGroup([]string{"Secure Backup", "Data Restore"}, func(s string) {
		switch s {
		case "Secure Backup":
			backup = true
			restore = false
			continueBttn.Enable()
		case "Data Restore":
			backup = false
			restore = true
			continueBttn.Enable()
		default:
			backup = false
			restore = false
			continueBttn.Disable()
		}
	})
	backupInfoLyt := container.NewBorder(backupInfo, backupButtons, nil, nil, backupOptions)
	backupDia := dialog.NewCustomWithoutButtons("Rescue Point", backupInfoLyt, mainWindowGui)
	backupDia.Resize(fyne.NewSize(600, 340))
	currentMainDialog = backupDia
	currentMainDialog.Show()
}
