package main

import (
	"fmt"
	"log"
	"regexp"
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

var lgnTmeOutMnt = 15
var askPwd = true // will try to use it to ask password before every transaction
var logoutTicker *time.Ticker

func startLogoutTicker(timeout int) {
	if logoutTicker != nil {
		logoutTicker.Stop()
	}
	logoutTicker = time.NewTicker(time.Duration(timeout) * time.Minute)
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

func openSecurityDia(creds Credentials) {
	pwd := widget.NewPasswordEntry()
	errorLabel := widget.NewLabel("")
	security := dialog.NewForm("Dangerous area!", "Confirm", "Cancel", []*widget.FormItem{
		widget.NewFormItem("", widget.NewLabelWithStyle("Please enter your password", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		widget.NewFormItem("", pwd),
		widget.NewFormItem("", errorLabel),
	}, func(b bool) {
		if b && creds.Password == pwd.Text {
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
			lgnTmeOutMntStr := strconv.Itoa(lgnTmeOutMnt)
			lgnTmeOut.SetText(lgnTmeOutMntStr)
			var settingsChanged func()
			sendOnlyKnown := userSettings.SendOnly
			sendOnlyKnownChck := widget.NewCheck("Send assets only known addresses", func(b bool) {
				if b {
					sendOnlyKnown = true
					settingsChanged()
				} else {
					sendOnlyKnown = false
					settingsChanged()
				}

			})
			sendOnlyKnownChck.Checked = userSettings.SendOnly

			var securityForm *widget.Form
			pwdAskAll := askPwd
			pwdAskOnly := !askPwd
			// fmt.Println("askPwd", askPwd)
			var tmeOutValid, pwdValid bool
			saveBttn := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {
				lgnTmeOutMnt, _ = strconv.Atoi(lgnTmeOut.Text)
				if len(newPwdCnfrm.Text) < 6 {
					userSettings.AskPwd = askPwd
					userSettings.LgnTmeOut = lgnTmeOutMnt
					userSettings.SendOnly = sendOnlyKnown
					if err := saveSettings(); err != nil {
						log.Println("Failed to save settings:", err)
						dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), mainWindowGui)
						return
					}
					currentMainDialog.Hide()
					dialog.ShowInformation("Settings saved", "Password not updated and settings saved", mainWindowGui)
				} else {
					creds.Password = newPwdCnfrm.Text
					if err := saveCredentials(creds); err != nil {
						log.Println("Failed to save credentials:", err)
						dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
						return
					}

					if err := saveAddressBook(userAddressBook, newPwdCnfrm.Text); err != nil {
						log.Println("Failed to save address book:", err)
						dialog.ShowInformation("Error", "Failed to save address book: "+err.Error(), mainWindowGui)
						return
					}

					userSettings.AskPwd = askPwd
					userSettings.LgnTmeOut = lgnTmeOutMnt
					userSettings.SendOnly = sendOnlyKnown
					if err := saveSettings(); err != nil {
						log.Println("Failed to save settings:", err)
						dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), mainWindowGui)
						return
					}
					mainWindow(creds, regularTokens, nftTokens)
					currentMainDialog.Hide()
					dialog.ShowInformation("Settings saved", "Password updated and settings saved", mainWindowGui)
				}
			})
			saveBttn.Disable()

			formValid := func() {
				if tmeOutValid && pwdValid {
					saveBttn.Enable()
				} else {
					saveBttn.Disable()
				}
			}

			lgnTmeOutFrmItm := widget.NewFormItem("Login Time Out", container.New(layout.NewHBoxLayout(), lgnTmeOut, widget.NewLabel("Minutes (min 3 max 120)"), layout.NewSpacer()))
			lgnTmeOutFrmItm.HintText = "."

			lgnTmeOut.Validator = func(s string) error {
				noSpaces := !regexp.MustCompile(`\s`).MatchString(s)
				matched, _ := regexp.MatchString("[0-9]", s)

				if !noSpaces {
					tmeOutValid = false
					formValid()
					lgnTmeOutFrmItm.HintText = "contains space"
					securityForm.Refresh()
					return fmt.Errorf("contains space")
				} else if !matched {
					tmeOutValid = false
					formValid()
					lgnTmeOutFrmItm.HintText = "only numbers"
					securityForm.Refresh()
					return fmt.Errorf("only numbers")
				}

				value, _ := strconv.Atoi(s)

				if value < 3 {
					tmeOutValid = false
					formValid()
					lgnTmeOutFrmItm.HintText = "min 3"
					securityForm.Refresh()
					return fmt.Errorf("min 3")
				} else if value > 120 {
					tmeOutValid = false
					formValid()
					lgnTmeOutFrmItm.HintText = "max 120"
					securityForm.Refresh()
					return fmt.Errorf("max 120")
				} else {
					tmeOutValid = true
					formValid()
					lgnTmeOutFrmItm.HintText = ""
					securityForm.Refresh()
					settingsChanged() // Call when valid
					return nil
				}
			}

			newPwd.Validator = func(s string) error {
				if len(s) < 1 {
					newPwdCnfrm.SetValidationError(nil)
					return nil
				} else if len(s) < 6 {
					return fmt.Errorf("min 6 characters")
				} else {
					err := newPwdCnfrm.Validate()
					if err != nil {
						newPwdCnfrm.SetValidationError(fmt.Errorf("please confirm"))
						return fmt.Errorf("confirm password")
					}
					return nil
				}
			}

			newPwdCnfrm.Validator = func(s string) error {
				newPwdEntry, _ := newPwdBind.Get()
				newPwdCnfrmEntry, _ := newPwdCnfrmBind.Get()
				fmt.Println("newPwdCnfrm validation", newPwdEntry, newPwdCnfrmEntry)
				if len(newPwdEntry) < 1 {
					return nil
				} else {
					_, err := pwdMatch(newPwdEntry, newPwdCnfrmEntry)
					if err != nil {
						return err
					} else {
						newPwd.SetValidationError(nil)
						settingsChanged() // Call when valid
						return nil
					}
				}
			}

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

			settingsChanged = func() {
				if lgnTmeOutMntStr == lgnTmeOut.Text && askPwd == userSettings.AskPwd && pwdValid && userSettings.SendOnly == sendOnlyKnown {
					saveBttn.Disable()
				} else {
					saveBttn.Enable()
				}
			}

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

			securityForm.SetOnValidationChanged(func(err error) {
				if err == nil {
					pwdValid = true
					formValid()
				} else {
					pwdValid = false
					formValid()
				}
				settingsChanged()
			})

			securityFormContent := container.NewVScroll(securityForm)
			securityButtons := container.NewGridWithColumns(2, exitBttn, saveBttn)
			securityFormLayout := container.NewBorder(nil, securityButtons, nil, nil, securityFormContent)
			securityFormDia := dialog.NewCustomWithoutButtons("Security settings", securityFormLayout, mainWindowGui)
			currentMainDialog = securityFormDia
			currentMainDialog.Resize(fyne.NewSize(600, 345))
			currentMainDialog.Show()

		} else if b && creds.Password != pwd.Text {
			currentMainDialog.Hide()
			errorLabel.SetText("Password is invalid!")
			currentMainDialog.Show()
		}

	}, mainWindowGui)
	security.Resize(fyne.NewSize(368, 207))
	currentMainDialog = security
	currentMainDialog.Show()
	mainWindowGui.Canvas().Focus(pwd)
}

func askPwdDia(askPwd bool, pwd string, mainWindow fyne.Window, callback func(bool)) {
	if askPwd {
		pwdEntry := widget.NewPasswordEntry()
		errorLabel := widget.NewLabel("")

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
						errorLabel.SetText("Password is invalid!")
						currentMainDialog.Show()

					}
				}
				callback(false)
			}, mainWindow)
		currentMainDialog = form
		currentMainDialog.Resize(fyne.NewSize(368, 207))
		currentMainDialog.Show()
		mainWindow.Canvas().Focus(pwdEntry)
	} else {
		callback(true)
	}
}

// // Usage
// askPwdDia(askPwd, creds.Password, mainWindowGui, func(correct bool) {
// 	 fmt.Println("result", correct)
// 	 if !correct {
// 		return
// 		}
// 	 // Continue with your code here })
// 	})
