package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"spallet/core"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	"github.com/tyler-smith/go-bip39"
)

// Function to show Welcome Page
func showWelcomePage() {

	welcomeMsg := widget.NewLabelWithStyle("Welcome to Spallet!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	humorousMsg := widget.NewLabel("So, you’ve got the soul of a crypto warrior, huh? Whether you’re riding the waves with Specky👻 or Sparky🔥, this wallet is your trusty companion in the Phantasma universe. 🐦⚡")
	humorousMsg.Wrapping = fyne.TextWrapWord // Ensure humorous message wraps correctly
	whatIsSpalletHeader := widget.NewLabelWithStyle("What is Spallet", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	whatIsSpallet := widget.NewLabel("Spallet is a community wallet developed for the Phantasma Blockchain. The name is a playful blend of Sparky, Specky (mostly Sparky), and Wallet—resulting in Spallet. I aimed for a catchy and fun name for this wallet.\n\nWith Spallet, I want to inject some fun and creativity into the world of crypto wallets by reflecting a gaming-oriented chain with small animations, humor, and more. Although I am not a highly experienced developer, my goal is to create a wallet that is engaging and enjoyable to use.\n\nI developed Spallet partly because I don't like Poltergeist's design and particularly dislike seeing that guy's name still on its license. I hope Spallet can help foster a new culture within the Phantasma community—who knows what we might achieve, right?")
	whatIsSpallet.Wrapping = fyne.TextWrapWord
	disclaimerMsgHeader := widget.NewLabelWithStyle("DisClaimer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	disclaimerMsg := widget.NewLabel("This wallet is open-sourced and developed with the guidance of AI. The creator is not a security expert and will not accept any responsibility for any potential losses. Use at your own risk!\n\nTranslation of Disclaimer: I’m not a security guru, so if you lose your moon bag, please don’t sue me.")
	disclaimerMsg.Wrapping = fyne.TextWrapWord // Ensure disclaimer message wraps correctly
	securityHeader := widget.NewLabelWithStyle("Security", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	securityHeader.Wrapping = fyne.TextWrapWord
	securityMessage := widget.NewLabel("This wallet uses AES256 to securely store your wallet data on your hard drive. However, given my limited expertise, please exercise caution and do not solely rely on this security measure.")
	securityMessage.Wrapping = fyne.TextWrapWord
	acceptButton := widget.NewButton("Accept and Continue", func() {
		featuresPage()
	})
	welcomeContent := container.NewVBox(
		welcomeMsg,
		humorousMsg,
		whatIsSpalletHeader,
		whatIsSpallet,
		disclaimerMsgHeader,
		disclaimerMsg,
		securityHeader,
		securityMessage,
	)
	scrollContent := container.NewVScroll(welcomeContent)

	welcomeLyt := container.NewBorder(nil, acceptButton, nil, nil, scrollContent)
	welcomeLyt.Resize(fyne.NewSize(800, 600))
	mainWindowGui.SetContent(
		welcomeLyt)
	mainWindowGui.Resize(fyne.NewSize(800, 600))
}

func featuresPage() {
	featuresHeader := widget.NewLabelWithStyle("Features of Spallet", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	features := widget.NewRichTextFromMarkdown("1- Bugs, it means if you found a bug its a feature \n\n2- Nicknames and badges based on Staked soul\n\n3- Account migration from manage accounts menu\n\n4- Sending assets between your accounts\n\n5- Sending assets to address book recipients\n\n6-Collecting Master rewards\n\n7- Collecting Crown rewards\n\n8-Eligibility badges\n\n9-Detailed Account information\n\n10- Showing some chain statistics\n\n11- Detailed staking information under hodling tab\n\n12- Adjustable log in time out between 3-120 min\n\n13- Send assets to only known addresses\n\n14- Wallet backup/restore from restore point menu\n\n15- Custom network settings\n\n16- Auto update balances every 15 seconds\n\n17- Saturn Dex trading with routing, in amount estimation from user's desired out amount, price impact calculation\n\nalso some other things i forget :)\n\n **What we dont have in spallet**\n\n1- Phantasma link\n\n2- Showing Nft pictures and details (go SDK limitation and my limited knowledge)\n\n3- Burning tokens\n\nsome other things i dont remember\n\n**Planned Features**\n\nI've planned some features for this wallet, but hey, I'm doing this for fun. Feel free to use it as it is. Since it's open-sourced, you can fork it and continue its development or contribute its code if you like.")
	features.Wrapping = fyne.TextWrapWord
	scrollContent := container.NewVScroll(features)
	continueBttn := widget.NewButton("Continue to wallet setup", func() {
		showPasswordSetupPage()

	})
	featuresLyt := container.NewBorder(featuresHeader, continueBttn, nil, nil, scrollContent)
	faturesContent := container.NewPadded(featuresLyt)
	mainWindowGui.SetContent(faturesContent)
}

// Function to show Password Setup Page
func showPasswordSetupPage() {

	pwdFrst := ""
	pwdBind := binding.BindString(&pwdFrst)
	passwordEntry := widget.NewEntryWithData(pwdBind)
	passwordEntry.Password = true
	confirmPasswordEntry := widget.NewPasswordEntry()

	var creds = core.Credentials{Wallets: make(map[string]core.Wallet)}

	var pwdIsValid, cnfrmIsValid bool
	pwdHeader := widget.NewLabelWithStyle("Set up a Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	pwdCnfrmForm := widget.NewForm(

		widget.NewFormItem("Password", passwordEntry),
		widget.NewFormItem("Confirm ", confirmPasswordEntry),
	)
	submitButton := widget.NewButton("Submit", func() {
		creds.Password = passwordEntry.Text // Save hashed password
		showWalletSetupPage(creds)
	})
	submitButton.Disable()
	updateSubmitBttn := func() {
		if pwdIsValid && cnfrmIsValid {
			submitButton.Enable()
		} else {
			submitButton.Disable()
		}

	}
	passwordEntry.Validator = func(s string) error {
		if len(s) < 6 {
			pwdIsValid = false
			updateSubmitBttn()
			return fmt.Errorf("min 6 characters")

		}
		pwdIsValid = true
		updateSubmitBttn()
		return nil
	}

	confirmPasswordEntry.Validator = func(s string) error {
		if len(s) < 6 {
			cnfrmIsValid = false
			updateSubmitBttn()
			return fmt.Errorf("enter your password")
		}
		pwd, _ := pwdBind.Get()
		_, err := core.PwdMatch(s, pwd)
		if err != nil {
			cnfrmIsValid = false
			updateSubmitBttn()
			return err
		} else {
			cnfrmIsValid = true
			updateSubmitBttn()
			return nil
		}
	}
	// Create a centered submit button

	passwordEntry.SetValidationError(fmt.Errorf("enter your password"))
	confirmPasswordEntry.SetValidationError(fmt.Errorf("enter your password"))
	warning := widget.NewLabelWithStyle("⚠️If you forget your password, there will be no way to recover it⚠️", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	submitFormLyt := container.NewVBox(pwdHeader, pwdCnfrmForm, submitButton, warning)

	pwdSetupLyt := container.NewCenter(
		submitFormLyt,
	)

	pwdSetupLyt.Resize(fyne.NewSize(400, 300))
	mainWindowGui.SetContent(pwdSetupLyt)
	mainWindowGui.Canvas().Focus(passwordEntry)
}

// Function to show Wallet Setup Page
func showWalletSetupPage(creds core.Credentials) {
	generateWalletButton := widget.NewButton("Generate New Account", func() {
		generateNewWalletPage(creds) // Correctly pointing to generateNewWalletPage
	})
	importWifButton := widget.NewButton("Import Account", func() {
		showImportAccountPage(creds)
	})
	restorePointBttn := widget.NewButton("Restore Point", func() {

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
								foundFiles += fmt.Sprintf("%v \n", fileName)

								switch fileName {
								case "credentials.spallet": // restoring unsaved accounts
									// fmt.Println(pwd)
									fmt.Println("**************restoring Accounts************")
									ldCreds, err := core.LoadCredentials(filePath, pwd, "")
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
											name := fmt.Sprintf("%v...%v", ldCredsWallet.Address[:8], ldCredsWallet.Address[len(ldCredsWallet.Address)-8:len(ldCredsWallet.Address)]) //if user registered same name giving it to a new name
											walletToAdd := core.Wallet{
												Name:    name,
												Address: ldCredsWallet.Address,
												WIF:     ldCredsWallet.WIF,
											}
											foundAccounts++
											creds.Wallets[name] = walletToAdd

										} else if !isSavedWallet && !isSavedName {
											creds.Wallets[ldCredsWallet.Name] = ldCredsWallet

											foundAccounts++
										}

									}
									creds.LastSelectedWallet = ldCreds.LastSelectedWallet
									creds.WalletOrder = ldCreds.WalletOrder
								case "addressbook.spallet": // restoring unsaved addresses to addressbook
									fmt.Println("**************restoring adress Book************")

									ldAdrBk, err := core.LoadAddressBook(filePath, pwd, "")
									if err != nil {
										dialog.ShowError(err, mainWindowGui)
										return
									}
									fmt.Println(len(ldAdrBk.Wallets))
									for _, ldAddrBkAddr := range ldAdrBk.Wallets {
										isSavedWallet := false
										isSavedName := false
										fmt.Println("restore adress", ldAddrBkAddr.Name)
										for _, savedAddr := range core.UserAddressBook.Wallets {

											if savedAddr.Address == ldAddrBkAddr.Address {
												isSavedWallet = true

											}
											if savedAddr.Name == ldAddrBkAddr.Name {
												isSavedName = true
											}

										}

										if !isSavedWallet && isSavedName {
											name := fmt.Sprintf("%v...%v", ldAddrBkAddr.Address[:8], ldAddrBkAddr.Address[len(ldAddrBkAddr.Address)-8:len(ldAddrBkAddr.Address)])
											walletToAdd := core.Wallet{
												Name:    name,
												Address: ldAddrBkAddr.Address,
											}
											foundAddress++
											core.UserAddressBook.Wallets[name] = walletToAdd
											core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder, name)
										} else if !isSavedWallet && !isSavedName {
											core.UserAddressBook.Wallets[ldAddrBkAddr.Name] = ldAddrBkAddr
											core.UserAddressBook.WalletOrder = append(core.UserAddressBook.WalletOrder, ldAddrBkAddr.Name)
											foundAddress++
										}

									}

								case "settings.spallet": //restoring user settings
									fmt.Println("**************restoring User Settings************")
									core.LoadSettings(filePath, "")
									core.SaveSettings("")
									settingsRestored = true

								}

							} else {

								switch fileName {
								case "credentials.spallet":

									dialog.ShowInformation("RESTORE FAILED", "cannot find critical file\nPlease make sure folder you selected includes 'credentials.spallet'", mainWindowGui)
									return

								case "settings.spallet":

									core.UserSettings = core.DefaultSettings()

									if err := core.SaveSettings(""); err != nil {
										log.Println("Failed to save Settings:", err)
										dialog.ShowInformation("Error", "Failed to save Settings: "+err.Error(), mainWindowGui)
										return
									}
								case "addressbook.spallet":
									if err := core.SaveAddressBook(core.UserAddressBook, creds.Password, ""); err != nil {
										log.Println("Failed to save Addressbook:", err)
										dialog.ShowInformation("Error", "Failed to save Addressbook: "+err.Error(), mainWindowGui)
										return
									}

								}

								notFoundFiles += fmt.Sprintf("%v \n", fileName)
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
							restoreInfo += fmt.Sprintln("Cant find any new address for Address book.")
						}

						if settingsRestored {
							restoreInfo += fmt.Sprintln("Found settings and applied")
						} else {
							restoreInfo += fmt.Sprintln("Cant find settings")
						}

						restoreInfo += fmt.Sprintf("\nFound Files\n%s\nNot Found Files\n%s", foundFiles, notFoundFiles)

						if foundAccounts > 0 || foundAddress > 0 || settingsRestored {
							fmt.Println("foundAccounts", foundAccounts)
							if err := core.SaveCredentials(creds, ""); err != nil {
								log.Println("Failed to save credentials:", err)
								dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
							}

							if err := core.SaveAddressBook(core.UserAddressBook, creds.Password, ""); err != nil {
								log.Println("Failed to save Address Book:", err)
								dialog.ShowInformation("Error", "Failed to save Address Book: "+err.Error(), mainWindowGui)
							}
							if currentMainDialog != nil {
								currentMainDialog.Hide()
							}

							restoreDia.Hide()
							showUpdatingDialog()
							err := core.GetChainStatistics()
							if err == nil {
								core.UpdateOrCheckTokenCache("", 3, "chain", rootPath)
								core.DataFetch(creds, rootPath)
								mainWindow(creds)
								closeUpdatingDialog()
								dialog.ShowInformation("Found new data", restoreInfo, mainWindowGui)

								autoUpdate(updateInterval, creds, rootPath)

							} else {
								dialog.ShowError(err, mainWindowGui)
								return
							}

						} else {

							dialog.ShowInformation("Restore Failed", fmt.Sprintf("Cant find any data please make sure file names are correct\nFound Files\n%s\nNot Found Files\n%s", foundFiles, notFoundFiles), mainWindowGui)

						}

					}
				}, mainWindowGui)
				askPwdDia.Show()
				mainWindowGui.Canvas().Focus(pwdEntry)
			}, mainWindowGui)
			openFolderDia.SetConfirmText("Restore From Here")
			openFolderDia.Resize(fyne.NewSize(600, 500))
			openFolderDia.Show()

		})

		restoreExplaination := widget.NewRichTextFromMarkdown("Fully restores data from the backup folder, please select folder contains you want to restore \n\n1- credentials.spallet file adds only unsaved accounts(private keys) to your current accounts. \n\n2- addressbook.spallet adds only unsaved addresses to your address book.\n\n3- settings.spallet overwrites your current settings.\n\n⚠️**You can delete files from back up folder if you dont want them to add**⚠️")
		restoreExplaination.Wrapping = fyne.TextWrapWord

		bttns := container.NewGridWithColumns(2, rstBckBttn, continueBttn)

		restoreOptDiaLyt := container.NewBorder(nil, bttns, nil, nil, container.NewVBox(restoreExplaination))

		restoreDia = dialog.NewCustomWithoutButtons("Restore information", restoreOptDiaLyt, mainWindowGui)
		restoreDia.Resize(fyne.NewSize(600, 340))
		restoreDia.Show()

	})

	walletSetupContent := container.NewVBox(
		widget.NewLabelWithStyle("Choose a way to add new account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		restorePointBttn,
		generateWalletButton,
		importWifButton)

	walletSetupLyt := container.NewCenter(walletSetupContent)
	mainWindowGui.SetContent(walletSetupLyt)
}

func generateNewWalletPage(creds core.Credentials) {
	// Generate a random entropy (128 bits for 12-word phrase, 256 bits for 24-word phrase)
	entropy, err := bip39.NewEntropy(core.DefaultMnemonicEntropy)
	if err != nil {
		panic(err)
	} // Generate the mnemonic phrase
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		dialog.ShowError(err, mainWindowGui)
		return
	}
	pk, err := core.MnemonicToPk(mnemonic, 0)
	if err != nil {
		dialog.ShowError(err, mainWindowGui)
		return
	}

	keyPair := cryptography.NewPhantasmaKeys(pk)
	privateKey := keyPair.WIF()
	address := keyPair.Address().String()
	nameEntry := widget.NewEntry()
	nameEntry.SetText("Sparky Account 1")
	nameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
	var isValidName, wifCopied, seedCopied bool
	okButton := widget.NewButton("Continue", func() {

		if creds.Wallets == nil {
			creds.Wallets = make(map[string]core.Wallet)
		}
		// Add wallet to credentials and mark as last used
		creds.Wallets[nameEntry.Text] = core.Wallet{
			Name:     nameEntry.Text,
			Address:  address,
			WIF:      privateKey,
			Mnemonic: mnemonic,
		}
		creds.WalletOrder = append(creds.WalletOrder, nameEntry.Text)
		creds.LastSelectedWallet = nameEntry.Text

		err := startWallet(creds)
		if err != nil {
			return
		}

	})
	okButton.Disable() // Initially disable the Continue button
	updateokBttnState := func() {
		if isValidName && wifCopied && seedCopied {
			okButton.Enable()
		} else {
			okButton.Disable()
		}
	}

	nameEntry.Validator = func(s string) error {
		names := []string{}
		_, err := core.ValidateAccountInput(names, nil, s, "name", false)

		if err != nil {
			isValidName = false
			updateokBttnState()
			return err
		} else {
			isValidName = true
			updateokBttnState()
			return nil
		}
	}
	copyWifButton := widget.NewButtonWithIcon(privateKey, theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(privateKey)
		dialog.ShowInformation("Copied", "Private Key (WIF) copied to clipboard", mainWindowGui)
		wifCopied = true // Enable the Continue button after WIF is copied
		updateokBttnState()
	})
	btnText := core.FormatMnemonic(mnemonic, 6)
	copyMnemonicButton := widget.NewButtonWithIcon(btnText, theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(privateKey)
		dialog.ShowInformation("Copied", "Seed phrase copied to clipboard", mainWindowGui)
		seedCopied = true // Enable the Continue button after Seed is copied
		updateokBttnState()
	})

	cancelButton := widget.NewButton("Cancel", func() {
		showWalletSetupPage(creds) // Go back to wallet setup page
	})

	generatedAccForm := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Address", widget.NewLabel(address)),
		widget.NewFormItem("Private Key (Wif)", copyWifButton),
		widget.NewFormItem("Seed Phrase", copyMnemonicButton),
	)
	warning := widget.NewLabelWithStyle("⚠️In order to continue please copy your Wif and Seed Phrase and store them in a safe place⚠️", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	// Use container.NewMax to cover full width
	genAccHeader := widget.NewLabelWithStyle("Generated account information", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	generateAccContent := container.NewVBox(genAccHeader, generatedAccForm, warning, container.NewGridWithColumns(2, cancelButton, okButton))
	generateAccLyt := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), container.NewHBox(generateAccContent), layout.NewSpacer()), layout.NewSpacer())

	mainWindowGui.SetContent(generateAccLyt)
	mainWindowGui.Canvas().Focus(nameEntry)
}

// Function to Show Import WIF Page
func showImportAccountPage(creds core.Credentials) {
	wifOrSeedEntry := widget.NewEntry()
	walletNameEntry := widget.NewEntry()
	walletNameEntry.SetText("Sparky Account 1")
	walletNameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
	var isValidName, isValidWif, isSeed bool
	var keyPair cryptography.PhantasmaKeys
	var mnemonic = ""
	importButton := widget.NewButton("Import", func() {
		if isSeed {
			mnemonic = wifOrSeedEntry.Text
			pk, err := core.MnemonicToPk(mnemonic, 0)
			if err != nil {
				dialog.ShowInformation("Error", fmt.Sprintf("%v", err), mainWindowGui)
				return
			}

			keyPair = cryptography.NewPhantasmaKeys(pk)

		} else {
			var err error
			keyPair, err = cryptography.FromWIF(wifOrSeedEntry.Text)
			if err != nil {
				dialog.ShowInformation("Error", "Invalid WIF format", mainWindowGui)
				return
			}
		}

		address := keyPair.Address().String()
		walletName := walletNameEntry.Text

		if creds.Wallets == nil {
			creds.Wallets = make(map[string]core.Wallet)
		}
		// Add wallet to credentials and mark as last used
		creds.Wallets[walletName] = core.Wallet{
			Name:     walletName,
			Address:  address,
			WIF:      keyPair.WIF(),
			Mnemonic: mnemonic,
		}
		creds.WalletOrder = append(creds.WalletOrder, walletName)
		creds.LastSelectedWallet = walletName

		err := startWallet(creds)
		if err != nil {
			return
		}
	})
	importButton.Disabled()
	wifOrSeedEntryForm := widget.NewForm(
		widget.NewFormItem("Name", walletNameEntry),
		widget.NewFormItem("Wif Or Seed Phrase", wifOrSeedEntry),
	)

	updateImportBttnState := func() {
		if isValidName && isValidWif {
			importButton.Enable()
		} else {
			importButton.Disable()
		}

	}

	wifOrSeedEntry.Validator = func(s string) error {
		s = strings.TrimSpace(s)
		containsSpace := strings.Contains(s, " ")
		if containsSpace {
			err := core.SeedPhraseValidator(s)
			if err != nil {
				return err
			} else {
				return nil
			}
		} else {
			isSeed = false
			_, err := core.WifValidator(s)
			if err != nil {
				isValidWif = false
				updateImportBttnState()
				return err
			}
			isValidWif = true
			updateImportBttnState()
			return nil
		}
	}

	walletNameEntry.Validator = func(s string) error {
		names := []string{}
		_, err := core.ValidateAccountInput(names, nil, s, "name", false)

		if err != nil {
			isValidName = false
			updateImportBttnState()
			return err
		} else {
			isValidName = true
			updateImportBttnState()
			return nil
		}

	}
	cancelButton := widget.NewButton("Back", func() {
		showWalletSetupPage(creds) // Go back to wallet setup page
	})

	formHeader := widget.NewLabelWithStyle("Please enter account details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	space := widget.NewLabel("\t\t\t\t\t\t\t\t\t\t\t") //  still dont understand how to control width inside an layout it shrinks to min size so tried to prevent it with this
	importWifContent := container.NewVBox(
		formHeader,
		wifOrSeedEntryForm,
		container.NewGridWithColumns(2, cancelButton, importButton),
		space,
	)

	importWifLyt := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), container.NewHBox(importWifContent), layout.NewSpacer()), layout.NewSpacer())

	mainWindowGui.SetContent(importWifLyt)
	mainWindowGui.Canvas().Focus(walletNameEntry)

}

func startWallet(creds core.Credentials) error {
	if err := core.SaveCredentials(creds, ""); err != nil {
		log.Println("Failed to save credentials:", err)
		dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
		return err
	}
	if !core.FileExists("data/essential/settings.spallet") {
		core.UserSettings = core.DefaultSettings()

		if err := core.SaveSettings(""); err != nil {
			log.Println("Failed to save Settings:", err)
			dialog.ShowInformation("Error", "Failed to save Settings: "+err.Error(), mainWindowGui)
			return err
		}

	} else {
		core.LoadSettings("data/essential/settings.spallet", "")
	}

	if err := core.SaveAddressBook(core.UserAddressBook, creds.Password, ""); err != nil {
		log.Println("Failed to save Addressbook:", err)
		dialog.ShowInformation("Error", "Failed to save Addressbook: "+err.Error(), mainWindowGui)
		return err
	}

	showUpdatingDialog()
	err := core.GetChainStatistics()
	if err != nil {
		dialog.ShowError(err, mainWindowGui)
		return err
	} else {
		core.UpdateOrCheckTokenCache("", 3, "chain", rootPath)
		core.DataFetch(creds, rootPath)
		mainWindow(creds)
		closeUpdatingDialog()

		autoUpdate(updateInterval, creds, rootPath)
	}
	return nil
}
