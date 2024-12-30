package main

import (
	"fmt"
	"log"
	"regexp"
	"spallet/core"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var pwdDia dialog.Dialog

// func startLogoutTicker(timeout int) {
// 	if logoutTicker != nil {
// 		logoutTicker.Stop()
// 	}
// 	logoutTicker = time.NewTicker(time.Duration(timeout) * time.Minute)
// 	go func() {
// 		for range logoutTicker.C {
// 			w := container.NewBorder(nil, nil, nil, nil)
// 			scrtyStgsWin.SetContent(w)
// 			if currentMainDialog != nil {
// 				currentMainDialog.Hide()
// 			}
// 			currentMainDialog = dialog.NewInformation("Log in time out", "Please log in", scrtyStgsWin)
// 			currentMainDialog.Show()
// 			showExistingUserLogin()
// 		}
// 	}()
// }

func showSecurityWin(creds core.Credentials) {
	// pwd := widget.NewPasswordEntry()
	scrtyStgsWin := spallet.NewWindow("Security Settings")
	var pwdLen int
	var askPwd = core.UserSettings.AskPwd
	askPwdDia(true, creds.Password, mainWindow, func(correct bool) {
		fmt.Println("result", correct)
		if !correct {
			return
		}
		newPwdBindFrst := ""
		newPwdBind := binding.BindString(&newPwdBindFrst)
		newPwd := widget.NewEntryWithData(newPwdBind)
		newPwd.PlaceHolder = "Leave it empty if you don't want to change"
		newPwd.Password = true
		newPwdCnfrmfrst := ""
		newPwdCnfrmBind := binding.BindString(&newPwdCnfrmfrst)
		newPwdCnfrm := widget.NewEntryWithData(newPwdCnfrmBind)
		newPwdCnfrm.Password = true
		lgnTmeOut := widget.NewEntry()
		lgnTmeOutMnt := core.UserSettings.LgnTmeOut
		lgnTmeOutMntStr := strconv.Itoa(lgnTmeOutMnt)
		lgnTmeOut.SetText(lgnTmeOutMntStr)
		var settingsChanged func()
		sendOnlyKnown := core.UserSettings.SendOnly

		var securityForm *widget.Form
		var securityFormContent = container.NewVScroll(widget.NewLabel(""))
		// fmt.Println("askPwd", askPwd)
		var tmeOutValid, pwdValid, pwdCnfmValid bool
		saveBttn := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
			lgnTmeOutMnt, _ = strconv.Atoi(lgnTmeOut.Text)
			if len(newPwdCnfrm.Text) < 6 || len(newPwd.Text) < 6 || newPwdCnfrm.Text != newPwd.Text {
				core.UserSettings.AskPwd = askPwd
				core.UserSettings.LgnTmeOut = lgnTmeOutMnt
				core.UserSettings.SendOnly = sendOnlyKnown
				if err := core.SaveSettings(rootPath); err != nil {
					log.Println("Failed to save settings:", err)
					dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), scrtyStgsWin)
					return
				}

				dialog.ShowInformation("Settings saved", "Password not Changed\nSettings saved", scrtyStgsWin)
			} else {
				creds.Password = newPwdCnfrm.Text
				if err := core.SaveCredentials(creds, rootPath); err != nil {
					log.Println("Failed to save credentials:", err)
					dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), scrtyStgsWin)
					return
				}

				if err := core.SaveAddressBook(core.UserAddressBook, newPwdCnfrm.Text, rootPath); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), scrtyStgsWin)
					return
				}

				core.UserSettings.AskPwd = askPwd
				core.UserSettings.LgnTmeOut = lgnTmeOutMnt
				core.UserSettings.SendOnly = sendOnlyKnown
				if err := core.SaveSettings(rootPath); err != nil {
					log.Println("Failed to save settings:", err)
					dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), scrtyStgsWin)
					return
				}

				dialog.ShowInformation("Settings saved", "Password Changed\nSettings saved", scrtyStgsWin)
			}
		})
		saveBttn.Disable()

		lgnTmeOutFrmItm := widget.NewFormItem("", container.New(layout.NewHBoxLayout(), lgnTmeOut, widget.NewLabel("Minutes (min 0 max 120)"), layout.NewSpacer()))
		lgnTmeOutFrmItm.HintText = "."

		settingsChanged = func() {
			if lgnTmeOutMntStr == lgnTmeOut.Text && askPwd == core.UserSettings.AskPwd && core.UserSettings.SendOnly == sendOnlyKnown && pwdLen < 6 {
				fmt.Println("Settings not changed")
				saveBttn.Disable()
			} else if !pwdValid || !tmeOutValid || !pwdCnfmValid {
				fmt.Println("Something is wrong", pwdValid, tmeOutValid, pwdCnfmValid)
				saveBttn.Disable()
			} else if pwdValid && tmeOutValid && pwdCnfmValid {
				saveBttn.Enable()
			}
		}
		sendOnlyKnownChck := widget.NewCheck("Send assets to only known addresses", func(b bool) {
			if b {
				sendOnlyKnown = true
				settingsChanged()
			} else {
				sendOnlyKnown = false
				settingsChanged()
			}

		})
		sendOnlyKnownChck.Checked = core.UserSettings.SendOnly

		lgnTmeOut.Validator = func(s string) error {
			noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
			matched, _ := regexp.MatchString("[0-9]", s)
			pos := lgnTmeOut.Position()
			securityFormContent.Offset = pos
			if !noSpaces {
				tmeOutValid = false
				settingsChanged()
				lgnTmeOutFrmItm.HintText = "contains space"
				securityForm.Refresh()
				return fmt.Errorf("contains space")
			} else if !matched {
				tmeOutValid = false
				settingsChanged()
				lgnTmeOutFrmItm.HintText = "only numbers"
				securityForm.Refresh()
				return fmt.Errorf("only numbers")
			}

			value, _ := strconv.Atoi(s)

			if value < 0 {
				tmeOutValid = false
				settingsChanged()
				lgnTmeOutFrmItm.HintText = "min 0"
				securityForm.Refresh()
				return fmt.Errorf("min 0")
			} else if value > 120 {
				tmeOutValid = false
				settingsChanged()
				lgnTmeOutFrmItm.HintText = "max 120"
				securityForm.Refresh()
				return fmt.Errorf("max 120")
			} else {
				tmeOutValid = true

				lgnTmeOutFrmItm.HintText = ""
				securityForm.Refresh()
				settingsChanged() // Call when valid
				return nil
			}
		}

		newPwd.Validator = func(s string) error {
			pwdLen = len(s)
			if len(s) < 1 && len(newPwdCnfrm.Text) < 1 {
				newPwdCnfrm.SetValidationError(nil)
				pwdCnfmValid = true
				pwdValid = true
				settingsChanged()
				return nil
			} else if len(s) < 6 {
				pwdValid = false
				settingsChanged()
				return fmt.Errorf("min 6 characters")
			} else {
				err := newPwdCnfrm.Validate()
				if err != nil {
					pwdValid = false
					newPwdCnfrm.SetValidationError(fmt.Errorf("please confirm"))
					settingsChanged()
					return fmt.Errorf("confirm password")
				}
				pwdValid = true
				settingsChanged()
				return nil
			}
		}

		newPwdCnfrm.Validator = func(s string) error {
			newPwdEntry, _ := newPwdBind.Get()
			newPwdCnfrmEntry, _ := newPwdCnfrmBind.Get()
			// fmt.Println("newPwdCnfrm validation", newPwdEntry, newPwdCnfrmEntry)
			if len(newPwdEntry) < 1 && len(s) < 1 {
				pwdCnfmValid = true
				pwdValid = true
				newPwd.SetValidationError(nil)
				settingsChanged()
				return nil
			} else if len(s) < 6 {
				pwdCnfmValid = false
				settingsChanged()
				return fmt.Errorf("min 6 characters")
			} else {
				_, err := core.PwdMatch(newPwdEntry, newPwdCnfrmEntry)
				if err != nil {
					pwdCnfmValid = false
					settingsChanged()
					return err
				} else {
					newPwd.SetValidationError(nil)
					pwdCnfmValid = true
					pwdValid = true
					settingsChanged()
					return nil
				}
			}
		}
		pwdAskAll := askPwd
		pwdAskOnly := !askPwd
		pwdAskAllBind := binding.BindBool(&pwdAskAll)
		pwdAskAllCheck := widget.NewCheckWithData("for everything", pwdAskAllBind)
		pwdaskLgnBind := binding.BindBool(&pwdAskOnly)
		pwdAskLgnCheck := widget.NewCheckWithData("only on login", pwdaskLgnBind)

		pwdAskLgnCheck.OnChanged = func(b bool) {
			if b {
				pwdAskAll = false
				pwdAskOnly = true
				pwdAskAllCheck.Checked = false
				pwdAskAllCheck.Refresh()
				askPwd = false
				settingsChanged() // Call when changed
			} else {
				pwdAskAll = true
				pwdAskOnly = false
				pwdAskAllCheck.Checked = true
				pwdAskAllCheck.Refresh()
				askPwd = true
				settingsChanged() // Call when changed
			}
		}
		pwdAskAllCheck.OnChanged = func(b bool) {
			if b {
				pwdAskAll = true
				pwdAskOnly = false
				pwdAskLgnCheck.Checked = false
				pwdAskLgnCheck.Refresh()
				askPwd = true
				settingsChanged() // Call when changed
			} else {
				pwdAskAll = false
				pwdAskOnly = true
				pwdAskLgnCheck.Checked = true
				pwdAskLgnCheck.Refresh()
				askPwd = false
				settingsChanged() // Call when changed
			}
		}

		PwdAskTypeChecks := container.New(layout.NewHBoxLayout(), pwdAskAllCheck, pwdAskLgnCheck, layout.NewSpacer())

		securityForm = widget.NewForm(
			widget.NewFormItem("", widget.NewLabelWithStyle("New Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			widget.NewFormItem("", newPwd),
			widget.NewFormItem("", widget.NewLabelWithStyle("Confirm New Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			widget.NewFormItem("", newPwdCnfrm),
			widget.NewFormItem("", widget.NewLabelWithStyle("Ask Password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			widget.NewFormItem("", PwdAskTypeChecks),
			widget.NewFormItem("", widget.NewLabelWithStyle("Asset Sending", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			widget.NewFormItem("", sendOnlyKnownChck),
			widget.NewFormItem("", widget.NewLabelWithStyle("Login Time Out", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			lgnTmeOutFrmItm,
		)

		exitBttn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
			scrtyStgsWin.Close()
		})

		// securityForm.SetOnValidationChanged(func(err error) {
		// 	if err == nil {
		// 		pwdValid = true
		// 		settingsChanged()
		// 	} else {
		// 		pwdValid = false
		// 		settingsChanged()
		// 	}

		// })
		securityButtons := container.NewGridWithColumns(2, exitBttn, saveBttn)
		securityFormContent.Content = container.NewVBox(securityForm, securityButtons)

		scrtyStgsWin.SetContent(securityFormContent)
		scrtyStgsWin.Resize(mainWindow.Canvas().Content().Size())
		scrtyStgsWin.Show()

	})
}

// // Usage
//
//	askPwdDia(askPwd, creds.Password, scrtyStgsWin, func(correct bool) {
//		 fmt.Println("result", correct)
//		 if !correct {
//			return
//			}
//		 // Continue with your code here })
//		})
func askPwdDia(askPwd bool, pwd string, mainWindow fyne.Window, callback func(bool)) {
	if askPwd {
		pwdEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")

		pwdEntry.OnSubmitted = func(s string) {
			if s == pwd {
				pwdDia.Hide()
				callback(true)
				return
			} else {
				pwdDia.Hide()
				errorLabel.SetText("Password is invalid!")
				pwdDia.Show()
				mainWindow.Canvas().Focus(pwdEntry)

			}
		}

		form := dialog.NewCustomConfirm("Dangerous area!", "Confirm", "Cancel",
			container.NewVBox(
				widget.NewLabelWithStyle("Please enter your password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				pwdEntry,
				errorLabel,
			), func(b bool) {
				if b {
					if pwdEntry.Text == pwd {
						callback(true)
						return
					} else {
						pwdDia.Hide()
						errorLabel.SetText("Password is invalid!")
						pwdDia.Show()
						mainWindow.Canvas().Focus(pwdEntry)

					}
				}
				callback(false)
			}, mainWindow)
		pwdDia = form
		pwdDia.Resize(fyne.NewSize(mainWindow.Canvas().Size().Width, 0))
		pwdDia.Show()
		mainWindow.Canvas().Focus(pwdEntry)
	} else {
		callback(true)
	}
}
