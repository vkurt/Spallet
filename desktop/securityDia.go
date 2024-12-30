package main

import (
	"fmt"
	"log"
	"regexp"
	"spallet/core"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var logoutTicker *time.Ticker

var pwdDia dialog.Dialog

func startLogoutTicker(timeout int) {
	if logoutTicker != nil {
		logoutTicker.Stop()
	}
	timeout *= 60
	if timeout <= 0 {
		timeout = 1
	}

	logoutTicker = time.NewTicker(time.Duration(timeout) * time.Second)
	go func() {
		for range logoutTicker.C {
			w := container.NewBorder(nil, nil, nil, nil)
			mainWindowGui.SetContent(w)
			if currentMainDialog != nil {
				currentMainDialog.Hide()
			}
			currentMainDialog = dialog.NewInformation("Log in time out", "Please log in", mainWindowGui)
			currentMainDialog.Show()
			showExistingUserLogin()
		}
	}()
}

func openSecurityDia(creds core.Credentials) {
	// pwd := widget.NewPasswordEntry()
	var pwdLen int
	var askPwd = core.UserSettings.AskPwd
	askPwdDia(true, creds.Password, mainWindowGui, func(correct bool) {
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
					dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), mainWindowGui)
					return
				}
				currentMainDialog.Hide()
				dialog.ShowInformation("Settings saved", "Password not Changed\nSettings saved", mainWindowGui)
			} else {
				creds.Password = newPwdCnfrm.Text
				if err := core.SaveCredentials(creds, rootPath); err != nil {
					log.Println("Failed to save credentials:", err)
					dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
					return
				}

				if err := core.SaveAddressBook(core.UserAddressBook, newPwdCnfrm.Text, rootPath); err != nil {
					log.Println("Failed to save address book:", err)
					dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
					return
				}

				core.UserSettings.AskPwd = askPwd
				core.UserSettings.LgnTmeOut = lgnTmeOutMnt
				core.UserSettings.SendOnly = sendOnlyKnown
				if err := core.SaveSettings(rootPath); err != nil {
					log.Println("Failed to save settings:", err)
					dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), mainWindowGui)
					return
				}
				mainWindow(creds)
				currentMainDialog.Hide()
				dialog.ShowInformation("Settings saved", "Password Changed\nSettings saved", mainWindowGui)
			}
		})
		saveBttn.Disable()

		lgnTmeOutFrmItm := widget.NewFormItem("Login Time Out", container.New(layout.NewHBoxLayout(), lgnTmeOut, widget.NewLabel("Minutes (min 0 max 120)"), layout.NewSpacer()))
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
		sendOnlyKnownChck := widget.NewCheck("Send assets only known addresses", func(b bool) {
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
			widget.NewFormItem("New Password", newPwd),
			widget.NewFormItem("Confirm", newPwdCnfrm),
			widget.NewFormItem("Ask Password", PwdAskTypeChecks),
			widget.NewFormItem("", sendOnlyKnownChck),
			lgnTmeOutFrmItm,
		)

		exitBttn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
			currentMainDialog.Hide()
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

		securityFormContent := container.NewVScroll(securityForm)
		securityButtons := container.NewGridWithColumns(2, exitBttn, saveBttn)
		securityFormLayout := container.NewBorder(nil, securityButtons, nil, nil, securityFormContent)
		securityFormDia := dialog.NewCustomWithoutButtons("Security settings", securityFormLayout, mainWindowGui)
		currentMainDialog = securityFormDia
		currentMainDialog.Resize(fyne.NewSize(600, 345))
		currentMainDialog.Show()

	})
}

// // Usage
//
//	askPwdDia(askPwd, creds.Password, mainWindowGui, func(correct bool) {
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
		pwdDia.Resize(fyne.NewSize(368, 207))
		pwdDia.Show()
		mainWindow.Canvas().Focus(pwdEntry)
	} else {
		callback(true)
	}
}
