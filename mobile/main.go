package main

import (
	"path/filepath"

	"fmt"
	"spallet/core"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var mainPayload = "Spallet Mobile"
var rootPath string
var currentMainDialog dialog.Dialog

var mainWindow fyne.Window
var homeTab *container.TabItem
var swapTab *container.TabItem
var hodlTab *container.TabItem
var historyTab *container.TabItem
var mainWindowLyt *container.AppTabs
var nftTab *container.TabItem
var spallet fyne.App
var logged bool
var timeOut int64

func closeAllWindowsAndDialogs() {
	for _, window := range spallet.Driver().AllWindows() {
		if window.Title() != "Spallet Mobile" {
			window.Close()
		}
	}
	if currentMainDialog != nil {
		currentMainDialog.Hide()

	}
	if pwdDia != nil {
		pwdDia.Hide()
	}
	if updatingDialog != nil {
		updatingDialog.Hide()
	}

	if settingsDia != nil {
		settingsDia.Hide()
	}

}

func main() {

	spallet = app.New()
	mainWindow = spallet.NewWindow("Spallet Mobile")
	mainWindow.SetMaster()
	spallet.Settings().SetTheme(theme.DarkTheme())
	rootPath = spallet.Storage().RootURI().Path()
	spallet.Lifecycle().SetOnEnteredForeground(func() {
		activeTime := time.Now()
		if logged && activeTime.Unix() > timeOut {

			showExistingUserLogin()
		}
		if logoutTicker != nil {
			logoutTicker.Stop()
		}

	})
	spallet.Lifecycle().SetOnExitedForeground(func() {
		passiveTime := time.Now()
		timeOut = passiveTime.Unix() + int64(core.UserSettings.LgnTmeOut)*60
		if logged && core.UserSettings.LgnTmeOut > 0 {
			startLogoutTicker(core.UserSettings.LgnTmeOut)
		} else {
			showExistingUserLogin()
		}
	})

	fmt.Println("Root Path: ", rootPath)

	homeTab = container.NewTabItemWithIcon("", theme.HomeIcon(), widget.NewLabel("Home Content"))
	nftTab = container.NewTabItemWithIcon("", resourceNftPng, widget.NewLabel("Nft Content"))
	swapTab = container.NewTabItemWithIcon("", resourceSwapPng, widget.NewLabel("Swap Content"))
	hodlTab = container.NewTabItemWithIcon("", resourceHodlPng, widget.NewLabel("HODL Content"))
	historyTab = container.NewTabItemWithIcon("", resourceHistoryPng, widget.NewLabel("Transaction History Content"))
	mainWindowLyt = container.NewAppTabs(homeTab, nftTab, hodlTab, swapTab, historyTab)

	mainWindowLyt.SetTabLocation(container.TabLocationBottom)

	// Set content
	mainWindow.SetContent(mainWindowLyt)

	firstTime := !core.FileExists(filepath.Join(rootPath, "data/essential/credentials.spallet"))

	if firstTime {
		showWelcomePage()

	} else {

		showExistingUserLogin()

	}

	mainWindow.ShowAndRun()
}

func showExistingUserLogin() {
	closeAllWindowsAndDialogs()
	if logoutTicker != nil {
		logoutTicker.Stop()
	}
	if updateBalanceTimeOut != nil {
		updateBalanceTimeOut.Stop()
	}
	logged = false
	passwordEntry := widget.NewPasswordEntry()
	invalidPwdMessage := binding.NewString()
	invalidPasswordLabel := widget.NewLabelWithData(invalidPwdMessage)
	resetWalletBtn := widget.NewButton("Forgot Your Password ?", func() {
		var pwdForgotDia dialog.Dialog
		closeBtn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { pwdForgotDia.Hide() })
		confirmBtn := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() { showPasswordSetupPage(); pwdForgotDia.Hide() })
		forgetKeysMessage := widget.NewRichTextFromMarkdown("Uh-oh, it seems your password has gone on a moon mission without you! 🚀🔑\n\nWe can't recover it now told you to keep it safe! But if you're ready to start fresh, type **'confirm'** below, and we'll hit the reset button like a crypto market crash.\n\n**⚠️ Warning: This will wipe all your accounts cleaner than a new block on the blockchain.**\n\n ***P.S. This time, guard your password like your most precious crypto. 🛡️***")
		forgetKeysMessage.Wrapping = fyne.TextWrapWord
		confirmEntry := widget.NewEntry()
		confirmEntry.Validator = func(s string) error {
			if s == "confirm" {
				confirmBtn.Enable()
				return nil
			} else {
				confirmBtn.Disable()
				return fmt.Errorf("please write confirm")
			}
		}
		confirmEntry.SetPlaceHolder("Please Write 'confirm'")
		btns := container.NewGridWithColumns(2, closeBtn, confirmBtn)
		content := container.NewBorder(nil, btns, nil, nil, container.NewVBox(forgetKeysMessage, confirmEntry))
		pwdForgotDia = dialog.NewCustomWithoutButtons("Forgot Your Password ?", content, mainWindow)
		width := mainWindow.Content().Size().Width
		pwdForgotDia.Resize(fyne.NewSize(width, 0))
		pwdForgotDia.Show()
	})
	invalidPasswordLabel.TextStyle = fyne.TextStyle{Bold: true}
	resetWalletBtn.Hide()
	invalidPasswordLabel.Hide()

	logIn := func() {
		rawPassword := passwordEntry.Text

		creds, err := core.LoadCredentials("data/essential/credentials.spallet", rawPassword, rootPath)
		if err != nil {
			if strings.Contains(err.Error(), "cipher: message authentication failed") {
				invalidPasswordLabel.Show()
				resetWalletBtn.Show()
				invalidPwdMessage.Set("Invalid Password!")
			} else {
				invalidPwdMessage.Set(err.Error())
			}
			return
		} else if creds.Password == rawPassword {
			ldAddrBk, err := core.LoadAddressBook("data/essential/addressbook.spallet", rawPassword, rootPath)
			if err != nil {
				dialog.ShowInformation("Error", "Failed to load address book: "+err.Error(), mainWindow)
				core.UserAddressBook = ldAddrBk
			} else {
				core.UserAddressBook = ldAddrBk
			}

			showUpdatingDialog()
			core.LoadSettings("data/essential/settings.spallet", rootPath) // Load settings at startup
			core.LoadTokenCache(rootPath)

			err = core.GetChainStatistics()
			if err != nil {
				closeUpdatingDialog()
				dialog.ShowError(fmt.Errorf("an error happened during fetching data,\n %v", err), mainWindow)
				return
			} else {
				core.DataFetch(creds, rootPath)

				core.LatestTokenData.ChainTokenUpdateTime = time.Now().UTC().Unix() - 3590 // we will update data automaticaly 15 sec after login with auto update
				core.LatestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
				core.LatestTokenData.AccTokenUpdateTime = time.Now().UTC().Unix()
				// mainWindowGui()
				autoUpdate(updateInterval, creds)
				mainWindow.SetContent(mainWindowLyt)
				closeUpdatingDialog()
				logged = true

			}

		} else {
			dialog.ShowInformation("Error", "Invalid password", mainWindow)

		}

	}
	passwordEntry.OnSubmitted = func(s string) {
		logIn()
	}

	pwdHeader := widget.NewLabel("Enter your password")
	pwdHeader.Alignment = fyne.TextAlignCenter
	logInCont := container.NewVBox(
		pwdHeader,
		passwordEntry,
		widget.NewButton("Login", func() {
			logIn()
		}),
		invalidPasswordLabel,
		resetWalletBtn)
	logInCont.Resize(fyne.NewSize(400, 150))
	logInLyt := container.New(layout.NewCenterLayout())
	logInLyt.Objects = []fyne.CanvasObject{
		logInCont,
	}
	logInLyt.Resize(fyne.NewSize(400, 150))
	mainWindow.SetContent(logInLyt)
	mainWindow.Canvas().Focus(passwordEntry)

}

// Function to show the updating dialog

var updatingDialog dialog.Dialog

func showUpdatingDialog() {
	if mainWindow != nil {
		spinner := widget.NewProgressBarInfinite()
		updatingContent := container.NewVBox(
			widget.NewLabel("Please wait while data is being updated..."),
			spinner,
		)
		updatingDialog = dialog.NewCustomWithoutButtons("Updating", updatingContent, mainWindow)
		updatingDialog.Show()
	}
}

func closeUpdatingDialog() {
	if updatingDialog != nil {
		updatingDialog.Hide()
	}
}
