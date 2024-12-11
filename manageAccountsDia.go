package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
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

func manageAccountsDia(creds Credentials) {
	changed := false //if account order, new added,name changed,removed  it will redraw mainwindow gui i am doing this because if i didnot redraw mainwindow gui wallet list will show old names but possibly binding is better solution
	// Usage
	askPwdDia(userSettings.AskPwd, creds.Password, mainWindowGui, func(correct bool) {
		fmt.Println("result", correct)
		if !correct {
			return
		}
		// Continue with your code here
		walletButtons := container.NewVBox()
		var manageAccCurrDia dialog.Dialog
		// var maxWidth float32 = 0.0
		var buildWalletButtons func()
		// for _, walletName := range creds.WalletOrder {
		// 	wallet := creds.Wallets[walletName]
		// 	btn := widget.NewButton(wallet.Name+"\n"+wallet.Address, func() {})
		// 	btnSize := btn.MinSize()
		// 	if btnSize.Width > maxWidth {
		// 		maxWidth = btnSize.Width
		// 	}
		// }

		walletScroll := container.NewVScroll(walletButtons)

		moveUp := func(index int) {
			if index > 0 {
				creds.WalletOrder[index], creds.WalletOrder[index-1] = creds.WalletOrder[index-1], creds.WalletOrder[index]
				if err := saveCredentials(creds); err != nil {
					log.Println("Failed to save credentials:", err)
					dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
				}
				changed = true
				buildWalletButtons()
				walletScroll.Content.Refresh()

			}
		}

		moveDown := func(index int) {
			if index < len(creds.WalletOrder)-1 {
				creds.WalletOrder[index], creds.WalletOrder[index+1] = creds.WalletOrder[index+1], creds.WalletOrder[index]
				if err := saveCredentials(creds); err != nil {
					log.Println("Failed to save credentials:", err)
					dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
				}
				changed = true
				buildWalletButtons()
				walletScroll.Content.Refresh()

			}
		}

		moveTop := func(index int) {
			wallet := creds.WalletOrder[index]
			creds.WalletOrder = append(creds.WalletOrder[:index], creds.WalletOrder[index+1:]...)
			creds.WalletOrder = append([]string{wallet}, creds.WalletOrder...)
			if err := saveCredentials(creds); err != nil {
				log.Println("Failed to save credentials:", err)
				dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
			}
			changed = true
			buildWalletButtons()
			walletScroll.Content.Refresh()

		}

		moveBottom := func(index int) {
			wallet := creds.WalletOrder[index]
			creds.WalletOrder = append(creds.WalletOrder[:index], creds.WalletOrder[index+1:]...)
			creds.WalletOrder = append(creds.WalletOrder, wallet)
			if err := saveCredentials(creds); err != nil {
				log.Println("Failed to save credentials:", err)
				dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
			}
			changed = true
			buildWalletButtons()
			walletScroll.Content.Refresh()

		}

		buildWalletButtons = func() {

			walletButtons.Objects = nil
			for index, walletName := range creds.WalletOrder {
				wallet := creds.Wallets[walletName]
				walletBtn := widget.NewButton(wallet.Name+"\n"+wallet.Address, func() {
					creds.LastSelectedWallet = walletName
					showUpdatingDialog()
					err := dataFetch(creds)
					if err != nil {
						closeUpdatingDialog()
						dialog.ShowError(fmt.Errorf("an error happened during data fetch\n %s", err), mainWindowGui)
					}
					autoUpdate(updateInterval, creds)

					if err := saveCredentials(creds); err != nil {
						dialog.NewError(err, fyne.CurrentApp().Driver().AllWindows()[0]).Show()
					}

					mainWindow(creds)

					if manageAccCurrDia != nil {
						manageAccCurrDia.Hide()
					}
					if currentMainDialog != nil {
						currentMainDialog.Hide()
					}
					closeUpdatingDialog()
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
				if index == len(creds.WalletOrder)-1 {
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
					renameEntry.PlaceHolder = "Enter new name for account"
					nameEntryWarningFrst := ""
					nameEntryWarning := binding.BindString(&nameEntryWarningFrst)
					nameEntryWarningLabel := widget.NewLabelWithData(nameEntryWarning)
					saveBttn := widget.NewButton("Save", func() {
						wallet := creds.Wallets[walletName]
						wallet.Name = renameEntry.Text
						creds.Wallets[renameEntry.Text] = wallet
						delete(creds.Wallets, walletName)
						for i, name := range creds.WalletOrder {
							if name == walletName {
								creds.WalletOrder[i] = renameEntry.Text
								break
							}
						}
						if creds.LastSelectedWallet == walletName {
							creds.LastSelectedWallet = renameEntry.Text
							autoUpdate(updateInterval, creds)

						}
						if err := saveCredentials(creds); err != nil {
							dialog.NewError(err, mainWindowGui)
							return
						}
						manageAccCurrDia.Hide()
						manageAccCurrDia = dialog.NewInformation("Succesfully saved", fmt.Sprintf("New name saved for '%s' as '%s'", wallet.Address, renameEntry.Text), mainWindowGui)
						changed = true
						buildWalletButtons()
						walletScroll.Content.Refresh()

					})
					backBttn := widget.NewButton("Back", func() {
						manageAccCurrDia.Hide()
					})
					renameEntry.Validator = func(s string) error {
						if len(s) < 1 {
							nameEntryWarning.Set("Please enter at least 1 letter and max 20")
							saveBttn.Disable()
							return errors.New("not entered")
						} else if len(s) <= 20 {
							for _, savedName := range creds.WalletOrder {
								if savedName == s {
									nameEntryWarning.Set("Name already used")
									saveBttn.Disable()
									return errors.New("already used")
								}
							}
							nameEntryWarning.Set("")
							saveBttn.Enable()
							return nil
						} else {
							nameEntryWarning.Set("Please use less than 20 letters")
							saveBttn.Disable()
							return errors.New("too long")
						}
					}
					buttonsContainer := container.NewGridWithColumns(2, backBttn, saveBttn)
					renameContent := container.NewVBox(renameEntry, nameEntryWarningLabel, buttonsContainer)
					renameDia := dialog.NewCustomWithoutButtons(fmt.Sprintf("Rename %s", creds.Wallets[walletName].Address), renameContent, mainWindowGui)
					manageAccCurrDia = renameDia
					manageAccCurrDia.Show()
				})

				showKeyButton := widget.NewButtonWithIcon("", theme.WarningIcon(), func() {
					dialog.ShowCustom("Please dont share this info with anyone", "I'll be careful with this", container.NewVBox(
						widget.NewForm(widget.NewFormItem("Name", widget.NewLabel(wallet.Name)), widget.NewFormItem("Address", widget.NewLabel(wallet.Address)), widget.NewFormItem("Wif", widget.NewButtonWithIcon(creds.Wallets[walletName].WIF, theme.ContentCopyIcon(), func() {
							fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(creds.Wallets[walletName].WIF)
							dialog.ShowInformation("Copied", "Wif copied to the clipboard", mainWindowGui)
						}))),
					), mainWindowGui)
				})

				removeBttn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					if len(creds.WalletOrder) == 1 {
						dialog.ShowInformation("Not allowed", "You cannot remove the last account from your wallet.\nPlease add another account before removing this one.", mainWindowGui)
						return
					}

					dialog.ShowForm("Remove Account", "Remove", "Cancel", []*widget.FormItem{

						widget.NewFormItem("Name", widget.NewLabel(wallet.Name)),
						widget.NewFormItem("Address", widget.NewLabel(wallet.Address)),
					}, func(ok bool) {
						if ok {

							delete(creds.Wallets, walletName)
							for i, name := range creds.WalletOrder {
								if name == walletName {
									creds.WalletOrder = append(creds.WalletOrder[:i], creds.WalletOrder[i+1:]...)
									break
								}
							}
							if walletName == creds.LastSelectedWallet && len(creds.WalletOrder) > 0 {
								creds.LastSelectedWallet = creds.WalletOrder[0]
								autoUpdate(updateInterval, creds)

							} else if walletName == creds.LastSelectedWallet && len(creds.WalletOrder) == 0 {
								creds.LastSelectedWallet = ""

								latestAccountData = AccountInfoData{FungibleTokens: make(map[string]AccToken), NonFungible: make(map[string]AccToken)}
							}
							if err := saveCredentials(creds); err != nil {
								dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
							}
							manageAccCurrDia = dialog.NewInformation("Account Removed", "Account removed succesfully", mainWindowGui)
							changed = true
							buildWalletButtons()
							// manageAccCurrDia.Show()
							walletScroll.Content.Refresh()

						}
					}, mainWindowGui)
				})

				migrateBttn := widget.NewButtonWithIcon("", theme.ContentRedoIcon(), func() {
					var migToGen, migToKey, migToAcc bool
					var migrateDiaT dialog.Dialog
					var migToAccDia dialog.Dialog
					var migrateDiaI dialog.Dialog
					var migToKeyDia dialog.Dialog

					migAccFeeLimit := new(big.Int).Set(userSettings.DefaultGasLimit)

					migrateDiaTBackBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
						migrateDiaT.Hide()
						migrateDiaI.Show()

					})
					migrateTCntBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
						var migToAccConfirmDia dialog.Dialog
						var migToGenConfirmDia dialog.Dialog
						var migToKeyConfirmDia dialog.Dialog

						fmt.Println(migToAcc, migToGen, migToKey)
						if migToAcc { // migrate one of the saved accounts which is dont have staked soul
							var migToAccAddr string
							var migToAccName string
							var selectableAcc []string

							migToAccSep1 := widget.NewSeparator()
							migToAccSep2 := widget.NewSeparator()
							migToAccSep3 := widget.NewSeparator()
							migToAccSep4 := widget.NewSeparator()
							migToAccCntBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
								fromLabel := widget.NewLabelWithStyle("Departure Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
								fromAccForm := widget.NewForm(widget.NewFormItem("Name", widget.NewLabel(wallet.Name)), widget.NewFormItem("Address", widget.NewLabel(wallet.Address)))
								toLabel := widget.NewLabelWithStyle("Destination Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
								toAccForm := widget.NewForm(widget.NewFormItem("Name", widget.NewLabel(migToAccName)), widget.NewFormItem("Address", widget.NewLabel(migToAccAddr)))
								migToAccWarnig := widget.NewLabelWithStyle("After confirming Specky will start moving your assets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
								migToAccConfirmBttn := widget.NewButton("Confirm and Migrate", func() {

									keyPair, err := cryptography.FromWIF(wallet.WIF)
									if err != nil {
										fyne.CurrentApp().SendNotification(&fyne.Notification{
											Title:   "Transaction Failed",
											Content: fmt.Sprintf("Invalid WIF: %v", err),
										})
										return
									}
									creds.LastSelectedWallet = migToAccName
									if err := saveCredentials(creds); err != nil {
										log.Println("Failed to save credentials:", err)
										dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
										return
									}

									expire := time.Now().UTC().Add(time.Second * 300).Unix()
									sb := scriptbuilder.BeginScript()
									sb.AllowGas(keyPair.Address().String(), cryptography.NullAddress().String(), userSettings.GasPrice, migAccFeeLimit)
									sb.CallContract("account", "migrate", keyPair.Address().String(), migToAccAddr)
									sb.SpendGas(keyPair.Address().String())
									script := sb.EndScript()

									tx := blockchain.NewTransaction(userSettings.NetworkName, userSettings.ChainName, script, uint32(expire), []byte("test"))
									tx.Sign(keyPair)
									txHex := hex.EncodeToString(tx.Bytes())
									// fmt.Println("*****Tx: \n" + txHex)

									// Start the animation
									startAnimation("send", "Specky is delivering wait a bit...")

									// Here, you can use stopChan if needed later, for example:
									// defer close(stopChan) when you need to ensure it gets closed properly.

									// Send the transaction
									SendTransaction(txHex, creds)
									autoUpdate(updateInterval, creds)
								})

								migToAccBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
									migrateDiaT.Show()
									migToAccConfirmDia.Hide()

								})

								migToAccClsBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
									migToAccConfirmDia.Hide()
									migrateDiaT.Hide()
									migrateDiaI.Hide()
								})

								gasLimitFloat, _ := userSettings.DefaultGasLimit.Float64()
								gasSlider := widget.NewSlider(userSettings.GasLimitSliderMin, userSettings.GasLimitSliderMax)
								gasSlider.Value = gasLimitFloat

								warning := binding.NewString()
								warning.Set("You have enough Kcal to fill Specky's engines")
								warningLabel := widget.NewLabelWithData(warning)
								warningLabel.Bind(warning)
								feeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
								feeErr := checkFeeBalance(feeAmount)
								if feeErr != nil {
									warning.Set(feeErr.Error())
									migToAccConfirmBttn.Disable()
								}
								gasSlider.OnChangeEnded = func(f float64) {
									migAccFeeLimit.SetInt64(int64(f))
									adjustedFeeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
									feeErr := checkFeeBalance(adjustedFeeAmount)
									if feeErr != nil {
										warning.Set(feeErr.Error())
										migToAccConfirmBttn.Disable()

									} else {
										warning.Set("You have enough Kcal to fill Specky's engines")
										migToAccConfirmBttn.Enable()
									}
								}

								gasForm := widget.NewForm(widget.NewFormItem("Specky's energy limit", gasSlider))

								migToAccInfo := container.NewVBox(fromLabel, migToAccSep1, fromAccForm, migToAccSep2, toLabel, migToAccSep3, toAccForm, migToAccSep4, migToAccWarnig, gasForm, warningLabel)
								migtoAccBttns := container.NewGridWithColumns(3, migToAccBckBttn, migToAccClsBttn, migToAccConfirmBttn)
								migToAccConfirmDiaLyt := container.NewBorder(nil, migtoAccBttns, nil, nil, migToAccInfo)
								migToAccConfirmDia = dialog.NewCustomWithoutButtons("Please confirm below information", migToAccConfirmDiaLyt, mainWindowGui)
								migToAccConfirmDia.Resize(fyne.NewSize(600, 337))
								migToAccConfirmDia.Show()

							})
							migToAccCntBttn.Disable()
							for _, walletSel := range creds.WalletOrder {
								if wallet.Name != walletSel {
									selectableAcc = append(selectableAcc, walletSel)
								}

							}
							migToAccBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
								migrateDiaT.Show()
								migToAccDia.Hide()
							})
							migToAccLabel := widget.NewLabel(fmt.Sprintf("Please select one of your accounts you want to migrate %s", wallet.Address))
							migToAccLabel.Wrapping = fyne.TextWrapWord
							migToAccSelect := widget.NewSelect(selectableAcc, func(s string) {
								migToAccAddr = creds.Wallets[s].Address
								migToAccName = creds.Wallets[s].Name
								response, err := client.GetAccount(migToAccAddr)
								if err != nil {
									dialog.ShowError(fmt.Errorf("specky encountered an error during checking this account\n%s", err), mainWindowGui)
									migToAccCntBttn.Disable()
									return
								}
								stakedBalance := StringToBigInt(response.Stakes.Amount)
								if stakedBalance.Cmp(big.NewInt(0)) > 0 {
									dialog.ShowError(fmt.Errorf("you have staked Soul in selected account"), mainWindowGui)
									migToAccCntBttn.Disable()
									return
								} else {
									migToAccCntBttn.Enable()
								}

							})
							exitMigToAccBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
								migToAccDia.Hide()
								migrateDiaT.Hide()
								migrateDiaI.Hide()

							})
							migToAccCont := container.NewVBox(migToAccLabel, migToAccSelect)
							migToAccBttns := container.NewGridWithColumns(3, migToAccBckBttn, exitMigToAccBttn, migToAccCntBttn)
							migToAccLyt := container.NewBorder(nil, migToAccBttns, nil, nil, migToAccCont)
							migToAccDia = dialog.NewCustomWithoutButtons("Please select one of your account from below", migToAccLyt, mainWindowGui)
							migToAccDia.Resize(fyne.NewSize(600, 337))
							migToAccDia.Show()

						} else if migToGen { //genarate new account and migrate asstes there
							migToGenSep1 := widget.NewSeparator()
							migToGenSep2 := widget.NewSeparator()
							migToGenSep3 := widget.NewSeparator()
							migToGenSep4 := widget.NewSeparator()
							genKeyPair := cryptography.GeneratePhantasmaKeys()
							genPrivateKey := genKeyPair.WIF()
							genAddress := genKeyPair.Address().String()
							migToGenNameSuggest := "Dest. Account " + fmt.Sprint(len(creds.Wallets)+1)

							fromLabel := widget.NewLabelWithStyle("Departure Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
							fromAccForm := widget.NewForm(
								widget.NewFormItem("Name", widget.NewLabel(wallet.Name)),
								widget.NewFormItem("Address",
									widget.NewLabel(wallet.Address)))
							toLabel := widget.NewLabelWithStyle("Destination Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
							genAccNameBind := binding.BindString(&migToGenNameSuggest)
							genAccNameEntry := widget.NewEntryWithData(genAccNameBind)
							toGenAccForm := widget.NewForm(
								widget.NewFormItem("Name", genAccNameEntry),
								widget.NewFormItem("Address", widget.NewLabel(genAddress)),
								widget.NewFormItem("Private Key", widget.NewButtonWithIcon(genPrivateKey, theme.ContentCopyIcon(), func() {
									fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(genPrivateKey)
									dialog.ShowInformation("Copied", "Private Key (wif) copied to the clipboard", mainWindowGui)
								},
								)))

							migToGenWarnig := widget.NewLabelWithStyle("After confirming Specky will start moving your assets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

							migToGenConfirmBttn := widget.NewButton("Confirm and Migrate", func() {

								keyPair, err := cryptography.FromWIF(wallet.WIF)
								if err != nil {
									fyne.CurrentApp().SendNotification(&fyne.Notification{
										Title:   "Transaction Failed",
										Content: fmt.Sprintf("Invalid WIF: %v", err),
									})
									return
								}

								name, _ := genAccNameBind.Get()
								creds.Wallets[name] = Wallet{
									Name:    name,
									Address: genAddress,
									WIF:     genPrivateKey,
								}
								creds.WalletOrder = append(creds.WalletOrder, name)
								creds.LastSelectedWallet = name
								if err := saveCredentials(creds); err != nil {
									log.Println("Failed to save credentials:", err)
									dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
									return
								}

								expire := time.Now().UTC().Add(time.Second * 300).Unix()
								sb := scriptbuilder.BeginScript()
								sb.AllowGas(keyPair.Address().String(), cryptography.NullAddress().String(), userSettings.GasPrice, migAccFeeLimit)
								sb.CallContract("account", "migrate", keyPair.Address().String(), genAddress)
								sb.SpendGas(keyPair.Address().String())
								script := sb.EndScript()

								tx := blockchain.NewTransaction(userSettings.NetworkName, userSettings.ChainName, script, uint32(expire), []byte("test"))
								tx.Sign(keyPair)
								txHex := hex.EncodeToString(tx.Bytes())
								// fmt.Println("*****Tx: \n" + txHex)

								// Start the animation
								startAnimation("send", "Specky is delivering wait a bit...")

								// Here, you can use stopChan if needed later, for example:
								// defer close(stopChan) when you need to ensure it gets closed properly.

								// Send the transaction
								SendTransaction(txHex, creds)
								autoUpdate(updateInterval, creds)
								mainWindow(creds)
							})

							migToGenBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
								migrateDiaT.Show()
								migToGenConfirmDia.Hide()

							})

							migToGenClsBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
								migToGenConfirmDia.Hide()
								migrateDiaT.Hide()
								migrateDiaI.Hide()
							})

							gasLimitFloat, _ := migAccFeeLimit.Float64()
							gasSliderBind := binding.BindFloat(&gasLimitFloat)
							gasSlider := widget.NewSliderWithData(userSettings.GasLimitSliderMin, userSettings.GasLimitSliderMax, gasSliderBind)
							gasSlider.Value = gasLimitFloat

							warning := binding.NewString()
							warning.Set("You have enough Kcal to fill Specky's engines")
							warningLabel := widget.NewLabelWithData(warning)
							warningLabel.Bind(warning)
							feeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
							feeErr := checkFeeBalance(feeAmount)
							if feeErr != nil {
								warning.Set(feeErr.Error())
								migToGenConfirmBttn.Disable()
							}
							nameErr := false
							updateConfirmButtonStatus := func() {
								if feeErr == nil && !nameErr {
									migToGenConfirmBttn.Enable()
								} else {
									migToGenConfirmBttn.Disable()
								}
							}
							gasSlider.OnChangeEnded = func(f float64) {
								migAccFeeLimit.SetInt64(int64(f))
								adjustedFeeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
								feeErr := checkFeeBalance(adjustedFeeAmount)
								if feeErr != nil {
									warning.Set(feeErr.Error())
									updateConfirmButtonStatus()

								} else {
									warning.Set("You have enough Kcal to fill Specky's engines")
									updateConfirmButtonStatus()
								}
							}
							genAccNameEntry.Validator = func(s string) error {
								if len(s) < 1 {
									warning.Set("Please enter at least 1 letter and max 20 for name")
									nameErr = true
									updateConfirmButtonStatus()
									return errors.New("not entered")
								} else if len(s) <= 20 {
									for _, savedName := range creds.WalletOrder {
										if savedName == s {
											warning.Set("Name already used")
											nameErr = true
											updateConfirmButtonStatus()
											return errors.New("already used")
										}
									}
									warning.Set("You can continue with this name")
									nameErr = false
									updateConfirmButtonStatus()
									return nil
								} else if len(s) > 20 {
									warning.Set("Please use less than 20 letters")
									nameErr = true
									updateConfirmButtonStatus()
									return errors.New("too long")
								} else {
									warning.Set("You can continue with this name")
									nameErr = false
									updateConfirmButtonStatus()
									return nil
								}

							}

							gasForm := widget.NewForm(widget.NewFormItem("Specky's energy limit", gasSlider))

							migToGenInfo := container.NewVBox(fromLabel, migToGenSep1, fromAccForm, migToGenSep2, toLabel, migToGenSep3, toGenAccForm, migToGenSep4, migToGenWarnig, gasForm, warningLabel)
							migToGenBttns := container.NewGridWithColumns(3, migToGenBckBttn, migToGenClsBttn, migToGenConfirmBttn)
							migToGenConfirmDiaLyt := container.NewBorder(nil, migToGenBttns, nil, nil, migToGenInfo)
							migToGenConfirmDia = dialog.NewCustomWithoutButtons("Please confirm below information", migToGenConfirmDiaLyt, mainWindowGui)
							migToGenConfirmDia.Resize(fyne.NewSize(600, 337))

							migToGenConfirmDia.Show()
							genAccNameEntry.CursorColumn = len(genAccNameEntry.Text)
							mainWindowGui.Canvas().Focus(genAccNameEntry)

						} else if migToKey {

							migToKeycSep1 := widget.NewSeparator()
							migToKeycSep2 := widget.NewSeparator()
							migToKeycSep3 := widget.NewSeparator()
							migToKeycSep4 := widget.NewSeparator()
							migToKeyPrvKey := widget.NewEntry()
							migToKeyPrvKey.PlaceHolder = "Enter your wif"
							migToKeyNameFrst := ""
							migToKeyNameBind := binding.BindString(&migToKeyNameFrst)
							migToKeyNameEntry := widget.NewEntryWithData(migToKeyNameBind)
							migToKeyNameEntry.PlaceHolder = "Enter a name for account"
							migToKeyNameSuggest := fmt.Sprintf("Dest. Account %v", len(creds.WalletOrder)+1)
							migToKeyNameEntry.SetText(migToKeyNameSuggest)
							warningFrst := ""
							warning := binding.BindString(&warningFrst)
							warningLabel := widget.NewLabelWithData(warning)
							nameErr, prvKeyErr := false, false

							migToKeyCntBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {

								migToKeyPair, err := cryptography.FromWIF(migToKeyPrvKey.Text)
								if err != nil {
									dialog.ShowInformation("Error", "Invalid WIF format", mainWindowGui)
									return
								}

								migToKeyAddress := migToKeyPair.Address().String()

								response, err := client.GetAccount(migToKeyAddress)
								if err != nil {
									dialog.ShowError(fmt.Errorf("specky encountered an error during checking this account\n%s", err), mainWindowGui)
									return
								}
								stakedBalance := StringToBigInt(response.Stakes.Amount)
								if stakedBalance.Cmp(big.NewInt(0)) > 0 {
									dialog.ShowError(fmt.Errorf("you have staked Soul in entered Wif account"), mainWindowGui)
									return
								}

								fromLabel := widget.NewLabelWithStyle("Departure Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
								fromAccForm := widget.NewForm(widget.NewFormItem("Name", widget.NewLabel(wallet.Name)), widget.NewFormItem("Address", widget.NewLabel(wallet.Address)))
								toLabel := widget.NewLabelWithStyle("Destination Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
								toAccForm := widget.NewForm(widget.NewFormItem("Name", widget.NewLabel(migToKeyNameEntry.Text)), widget.NewFormItem("Address", widget.NewLabel(migToKeyAddress)))
								migToKeyWarnig := widget.NewLabelWithStyle("After confirming Specky will start moving your assets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
								migToKeyConfirmBttn := widget.NewButton("Confirm and Migrate", func() {

									keyPair, err := cryptography.FromWIF(wallet.WIF)
									if err != nil {
										fyne.CurrentApp().SendNotification(&fyne.Notification{
											Title:   "Transaction Failed",
											Content: fmt.Sprintf("Invalid WIF: %v", err),
										})
										return
									}

									creds.Wallets[migToKeyNameEntry.Text] = Wallet{
										Name:    migToKeyNameEntry.Text,
										Address: migToKeyAddress,
										WIF:     migToKeyPair.WIF(),
									}
									creds.WalletOrder = append(creds.WalletOrder, migToKeyNameEntry.Text)
									creds.LastSelectedWallet = migToKeyNameEntry.Text
									if err := saveCredentials(creds); err != nil {
										log.Println("Failed to save credentials:", err)
										dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
										return
									}

									expire := time.Now().UTC().Add(time.Second * 300).Unix()
									sb := scriptbuilder.BeginScript()
									sb.AllowGas(keyPair.Address().String(), cryptography.NullAddress().String(), userSettings.GasPrice, migAccFeeLimit)
									sb.CallContract("account", "migrate", keyPair.Address().String(), migToKeyAddress)
									sb.SpendGas(keyPair.Address().String())
									script := sb.EndScript()

									tx := blockchain.NewTransaction(userSettings.NetworkName, userSettings.ChainName, script, uint32(expire), []byte("test"))
									tx.Sign(keyPair)
									txHex := hex.EncodeToString(tx.Bytes())
									// fmt.Println("*****Tx: \n" + txHex)

									// Start the animation
									startAnimation("send", "Specky is delivering wait a bit...")

									// Here, you can use stopChan if needed later, for example:
									// defer close(stopChan) when you need to ensure it gets closed properly.

									// Send the transaction
									SendTransaction(txHex, creds)
									autoUpdate(updateInterval, creds)
									mainWindow(creds)

								})

								migToKeyBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
									migToKeyDia.Show()
									migToKeyConfirmDia.Hide()

								})

								migToKeyClsBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
									migToKeyConfirmDia.Hide()
									migrateDiaT.Hide()
									migrateDiaI.Hide()
								})

								gasLimitFloat, _ := migAccFeeLimit.Float64()
								gasSlider := widget.NewSlider(userSettings.GasLimitSliderMin, userSettings.GasLimitSliderMax)
								gasSlider.Value = gasLimitFloat

								warning := binding.NewString()
								warning.Set("You have enough Kcal to fill Specky's engines")
								warningLabel := widget.NewLabelWithData(warning)
								warningLabel.Bind(warning)
								feeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
								feeErr := checkFeeBalance(feeAmount)
								if feeErr != nil {
									warning.Set(feeErr.Error())
									migToKeyConfirmBttn.Disable()
								}
								gasSlider.OnChangeEnded = func(f float64) {
									migAccFeeLimit.SetInt64(int64(f))
									adjustedFeeAmount := new(big.Int).Mul(migAccFeeLimit, userSettings.GasPrice)
									feeErr := checkFeeBalance(adjustedFeeAmount)
									if feeErr != nil {
										warning.Set(feeErr.Error())
										migToKeyConfirmBttn.Disable()

									} else {
										warning.Set("You have enough Kcal to fill Specky's engines")
										migToKeyConfirmBttn.Enable()
									}
								}

								gasForm := widget.NewForm(widget.NewFormItem("Specky's energy limit", gasSlider))

								migToKeyInfo := container.NewVBox(fromLabel, migToKeycSep1, fromAccForm, migToKeycSep2, toLabel, migToKeycSep3, toAccForm, migToKeycSep4, migToKeyWarnig, gasForm, warningLabel)
								migToKeyBttns := container.NewGridWithColumns(3, migToKeyBckBttn, migToKeyClsBttn, migToKeyConfirmBttn)
								migToKeyConfirmDiaLyt := container.NewBorder(nil, migToKeyBttns, nil, nil, migToKeyInfo)
								migToKeyConfirmDia = dialog.NewCustomWithoutButtons("Please confirm below information", migToKeyConfirmDiaLyt, mainWindowGui)
								migToKeyConfirmDia.Resize(fyne.NewSize(600, 337))
								migToKeyConfirmDia.Show()

							})

							migToKeyCntBttn.Disable()

							updatemigToKeyCntBttnStat := func() {
								if !nameErr && !prvKeyErr {
									migToKeyCntBttn.Enable()
								} else {
									migToKeyCntBttn.Disable()
								}
							}
							migToKeyPrvKey.Validator = func(s string) error {
								noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
								matched, _ := regexp.MatchString("^[KL5][a-zA-Z0-9]{0,51}$", s)

								if len(s) < 1 {
									warning.Set("Please enter wif")
									prvKeyErr = true
									updatemigToKeyCntBttnStat()
									return errors.New("please enter wif")
								} else if len(s) < 52 {
									if !noSpaces {

										warning.Set("Wif can't contain spaces")
										prvKeyErr = true
										updatemigToKeyCntBttnStat()
										return errors.New("wif can't contain spaces")
									} else if !matched {
										prvKeyErr = true
										updatemigToKeyCntBttnStat()
										warning.Set("Wif can't contain special characters and must start with 'K', 'L', or '5'")
										return errors.New("wif can't contain special characters and must start with 'K', 'L', or '5'")
									} else {
										prvKeyErr = true
										updatemigToKeyCntBttnStat()
										warning.Set("Wif key is too short")
										return errors.New("wif key is too short")
									}
								} else if len(s) > 52 {
									prvKeyErr = true
									updatemigToKeyCntBttnStat()
									warning.Set("Wif key is too long")
									return errors.New("wif key is too long")
								}

								if noSpaces && matched && len(s) >= 51 {
									for _, wallet := range creds.Wallets {
										if wallet.WIF == migToKeyPrvKey.Text {
											prvKeyErr = true
											updatemigToKeyCntBttnStat()
											warning.Set(fmt.Sprintf("This Wif is already used with Name %s.", wallet.Name))
											return errors.New("this Wif is already used")
										}
									}
									warning.Set("")
									prvKeyErr = false
									updatemigToKeyCntBttnStat()
									return nil
								} else {
									return errors.New("wrong key format")
								}
							}

							migToKeyNameEntry.Validator = func(s string) error {
								if len(s) < 1 {
									warning.Set("Please enter at least 1 letter and max 20 for name")
									return errors.New("not entered")
								} else if len(s) <= 20 {
									for _, savedName := range creds.WalletOrder {
										if savedName == s {
											nameErr = true
											updatemigToKeyCntBttnStat()
											warning.Set("Name already used")
											return errors.New("already used")
										}
									}
									warning.Set("")
									nameErr = false
									updatemigToKeyCntBttnStat()
									return nil
								} else {
									nameErr = true
									updatemigToKeyCntBttnStat()
									warning.Set("Please use less than 20 letters")
									return errors.New("too long")
								}
							}

							migToKeyBckBttn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
								migrateDiaT.Show()
								migToKeyDia.Hide()
							})

							migToKeyAccForm := widget.NewForm(
								widget.NewFormItem("Name", migToKeyNameEntry),
								widget.NewFormItem("Private Key (Wif)", migToKeyPrvKey),
							)
							exitMigToKeyBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
								migToKeyDia.Hide()
								migrateDiaT.Hide()
								migrateDiaI.Hide()

							})
							migToKeycCont := container.NewVBox(migToKeyAccForm, warningLabel)
							migToKeycBttns := container.NewGridWithColumns(3, migToKeyBckBttn, exitMigToKeyBttn, migToKeyCntBttn)
							migToKeycLyt := container.NewBorder(nil, migToKeycBttns, nil, nil, migToKeycCont)
							migToKeyDia = dialog.NewCustomWithoutButtons("Please enter destination account details", migToKeycLyt, mainWindowGui)
							migToKeyDia.Resize(fyne.NewSize(600, 337))
							migToKeyDia.Show()

						} else {
							dialog.ShowError(fmt.Errorf("specky encounterd a problem"), mainWindowGui)
							return
						}

					})
					migrateTCntBttn.Disable()
					migrateTLabel := widget.NewLabel(fmt.Sprintf("Please choose one of the options below to migrate your account\nName:\t%s\nAddress:\t%s\n", wallet.Name, wallet.Address))
					migrateTLabel.Wrapping = fyne.TextWrapWord
					migrateOptions := widget.NewRadioGroup([]string{"Migrate to My Accounts", "Generate new account and migrate it to there", "Migrate it to another Private key"}, func(s string) {

						if s == "Generate new account and migrate it to there" {
							migToGen = true
							migToKey = false
							migToAcc = false
							migrateTCntBttn.Enable()
							// fmt.Println(migToGen, migToKey, migToAcc)
						} else if s == "Migrate it to another Private key" {
							migToGen = false
							migToKey = true
							migToAcc = false
							migrateTCntBttn.Enable()
							// fmt.Println(migToGen, migToKey, migToAcc)
						} else if s == "Migrate to My Accounts" {
							migToGen = false
							migToKey = false
							migToAcc = true
							migrateTCntBttn.Enable()
							// fmt.Println(migToGen, migToKey, migToAcc)
						} else {
							migToGen = false
							migToKey = false
							migToAcc = false
							migrateTCntBttn.Disable()
						}

					})
					exitMigTBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
						migrateDiaT.Hide()
						migrateDiaI.Hide()
					})
					migrateDiaTButtonGroup := container.NewGridWithColumns(3, migrateDiaTBackBttn, exitMigTBttn, migrateTCntBttn)
					migrateTcontent := container.NewVBox(migrateTLabel, migrateOptions)
					migrateDiaTLyt := container.NewBorder(nil, migrateDiaTButtonGroup, nil, nil, migrateTcontent)
					migrateDiaT = dialog.NewCustomWithoutButtons("Select Migration Type", migrateDiaTLyt, mainWindowGui)
					migrateDiaT.Resize(fyne.NewSize(600, 337))

					// **********
					migrateDiaICntBttn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
						migrateDiaI.Hide()
						migrateDiaT.Show()

					})
					migrationInfoLabel := widget.NewRichTextFromMarkdown("**Migration Benefits**\n\n*By migrating, you can seamlessly transfer all your assets to a new wallet without compromising on the following:*\n\n1. _Crown and Master Rewards Eligibility_\n\n2. _On-Chain Name Retention_\n\n3. _Preserved Voting Power_\n\n**⚠️ Note: The target address must not contain any staked Soul.**")
					migrationInfoLabel.Wrapping = fyne.TextWrapWord
					exitMigBttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { migrateDiaI.Hide() })
					migrateDiaIBttnGroup := container.NewGridWithColumns(2, exitMigBttn, migrateDiaICntBttn)
					migrateDiaLyt := container.NewBorder(nil, migrateDiaIBttnGroup, nil, nil, container.NewVBox(migrationInfoLabel))
					migrateDiaI = dialog.NewCustomWithoutButtons("Migration Information", migrateDiaLyt, mainWindowGui)
					migrateDiaI.Resize(fyne.NewSize(600, 337))
					migrateDiaI.Show()

				})
				btnContainer := container.NewGridWithColumns(2,
					renameButton,
					showKeyButton,
					migrateBttn,
					removeBttn,
				)
				btnContainer.Resize(fyne.NewSize(120, btnContainer.MinSize().Height))

				walletGroup := container.NewBorder(nil, widget.NewSeparator(), moveButtons, btnContainer, walletBtn)
				walletButtons.Add(walletGroup)
			}
		}

		buildWalletButtons()
		if len(creds.WalletOrder) < 1 {
			walletButtons.Add(container.NewVBox(widget.NewLabel("Please Add/Generate an account")))
		}
		walletScroll = container.NewVScroll(walletButtons)

		walletScroll.SetMinSize(fyne.NewSize(600, 400))

		addWallet := widget.NewButtonWithIcon("Add Account", theme.ContentAddIcon(), func() {
			privateKey := widget.NewEntry()
			privateKey.PlaceHolder = "Enter your wif"
			walletnamefrst := ""
			walletNameBind := binding.BindString(&walletnamefrst)
			walletNameEntry := widget.NewEntryWithData(walletNameBind)
			walletNameEntry.PlaceHolder = "Enter a name for account"
			nameSuggest := fmt.Sprintf("Sparky account %v", len(creds.WalletOrder)+1)
			walletNameEntry.SetText(nameSuggest)
			warningFrst := ""
			warning := binding.BindString(&warningFrst)
			warningLabel := widget.NewLabelWithData(warning)

			privateKey.Validator = func(s string) error {
				noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
				matched, _ := regexp.MatchString("^[KL5][a-zA-Z0-9]{0,51}$", s)

				if len(s) < 1 {
					warning.Set("Please enter wif")
					return errors.New("please enter wif")
				} else if len(s) < 52 {
					if !noSpaces {
						warning.Set("Wif can't contain spaces")
						return errors.New("wif can't contain spaces")
					} else if !matched {
						warning.Set("Wif can't contain special characters and must start with 'K', 'L', or '5'")
						return errors.New("wif can't contain special characters and must start with 'K', 'L', or '5'")
					} else {
						warning.Set("Wif key is too short")
						return errors.New("wif key is too short")
					}
				} else if len(s) > 52 {
					warning.Set("Wif key is too long")
					return errors.New("wif key is too long")
				}

				if noSpaces && matched && len(s) >= 51 {
					for _, wallet := range creds.Wallets {
						if wallet.WIF == privateKey.Text {
							warning.Set("This Wif is already used.")
							return errors.New("this Wif is already used")
						}
					}
					warning.Set("")
					return nil
				} else {
					return errors.New("wrong key format")
				}
			}

			walletNameEntry.Validator = func(s string) error {
				if len(s) < 1 {
					warning.Set("Please enter at least 1 letter and max 20 for name")
					return errors.New("not entered")
				} else if len(s) <= 20 {
					for _, savedName := range creds.WalletOrder {
						if savedName == s {
							warning.Set("Name already used")
							return errors.New("already used")
						}
					}
					warning.Set("")
					return nil
				} else {
					warning.Set("Please use less than 20 letters")
					return errors.New("too long")
				}
			}

			addForm := dialog.NewForm("Add New Wallet", "Save", "Cancel", []*widget.FormItem{
				widget.NewFormItem("Wallet Name", walletNameEntry),
				widget.NewFormItem("Private Key", privateKey),
				widget.NewFormItem("", warningLabel),
			}, func(ok bool) {
				if ok {
					walletEntry, _ := walletNameBind.Get()
					keyPair, err := cryptography.FromWIF(privateKey.Text)
					if err != nil {
						dialog.ShowInformation("Error", "Invalid WIF format", mainWindowGui)
						return
					}
					address := keyPair.Address().String()

					creds.Wallets[walletEntry] = Wallet{
						Name:    walletEntry,
						Address: address,
						WIF:     privateKey.Text,
					}
					creds.WalletOrder = append(creds.WalletOrder, walletEntry)
					if err := saveCredentials(creds); err != nil {
						log.Println("Failed to save credentials:", err)
						dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
					} else {
						manageAccCurrDia = dialog.NewInformation("Account saved", "Account saved successfully", mainWindowGui)
						manageAccCurrDia.Show()
						changed = true
						buildWalletButtons()
						walletScroll.Content.Refresh()

					}
				}
			}, mainWindowGui)
			addForm.Resize(fyne.NewSize(600, 300))
			addForm.Show()

			privateKey.SetValidationError(errors.New("please enter wif"))
			privateKey.Refresh()

			walletNameEntry.CursorRow = len(walletNameEntry.Text)
			walletNameEntry.Refresh()
			walletNameEntry.FocusGained()
		})

		generateAccount := widget.NewButtonWithIcon("Generate Account", theme.SearchReplaceIcon(), func() {
			keyPair := cryptography.GeneratePhantasmaKeys()
			privateKey := keyPair.WIF()
			address := keyPair.Address().String()
			walletNameSuggest := "Sparky Account " + fmt.Sprint(len(creds.Wallets)+1)
			nameEntryFrst := ""
			nameEntryBind := binding.BindString(&nameEntryFrst)
			nameEntry := widget.NewEntryWithData(nameEntryBind)
			nameEntryBind.Set(walletNameSuggest)
			warningFrst := ""
			warningBind := binding.BindString(&warningFrst)
			warning := widget.NewLabelWithData(warningBind)
			nameEntry.Validator = func(s string) error {
				if len(s) < 1 {
					warningBind.Set("Please enter at least 1 letter and max 20 for name")
					return errors.New("not entered")
				} else if len(s) <= 20 {
					for _, savedName := range creds.WalletOrder {
						if savedName == s {
							warningBind.Set("Name already used")
							return errors.New("already used")
						}
					}
					warningBind.Set("")
					return nil
				} else {
					warningBind.Set("Please use less than 20 letters")
					return errors.New("too long")
				}
			}

			if manageAccCurrDia != nil {
				manageAccCurrDia.Hide()
			}
			manageAccCurrDia = dialog.NewForm("New account generated", "Save", "Scrap", []*widget.FormItem{
				widget.NewFormItem("Name", nameEntry),
				widget.NewFormItem("Address", widget.NewLabel(address)),
				widget.NewFormItem("Private Key (Wif)", widget.NewButtonWithIcon(privateKey, theme.ContentCopyIcon(), func() { fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(privateKey) })),
				widget.NewFormItem("", warning),
			}, func(b bool) {
				if b {
					name, _ := nameEntryBind.Get()
					creds.Wallets[name] = Wallet{
						Name:    name,
						Address: address,
						WIF:     privateKey,
					}
					creds.WalletOrder = append(creds.WalletOrder, name)
					if err := saveCredentials(creds); err != nil {
						log.Println("Failed to save credentials:", err)
						dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
					} else {
						manageAccCurrDia = dialog.NewInformation("Account saved", "Account saved successfully", mainWindowGui)
						manageAccCurrDia.Show()
						changed = true
						buildWalletButtons()
						walletScroll.Content.Refresh()

					}
				}
			}, mainWindowGui)
			manageAccCurrDia.Show()
			nameEntry.CursorRow = len(nameEntry.Text)
			nameEntry.FocusGained()

		})

		backButton := widget.NewButton("Back", func() {
			if changed {
				mainWindow(creds)
			}

			currentMainDialog.Hide()
		})

		accountsLayout := container.NewBorder(walletScroll, container.NewVBox(addWallet, generateAccount, backButton), nil, nil)
		if currentMainDialog != nil {
			currentMainDialog.Hide()
		}
		currentMainDialog = dialog.NewCustomWithoutButtons("Manage Your Accounts", accountsLayout, mainWindowGui)

		currentMainDialog.Show()
	})

}
