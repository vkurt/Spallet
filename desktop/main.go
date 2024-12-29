package main

import (
	"fmt"
	"math/big"
	"net/url"
	"spallet/core"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// var (
//
//	accTokenBalances = make(map[string]AccToken)
//	accNftBalances   = make(map[string]AccToken)
//
// )

var accInfoTab = container.NewVScroll(container.NewVBox())
var stakingTab = container.NewVScroll(container.NewVBox())
var dexTab = container.NewVScroll(container.NewVBox())

var accBadges = container.NewScroll(container.NewVBox())
var mainWindowGui fyne.Window

var tokenBtns = container.NewVScroll(container.NewVBox())
var nftBtns = container.NewVScroll(container.NewVBox())

var pageSize = uint(25)
var page = uint(1)
var mainPayload = "Spallet"
var spallet fyne.App
var rootPath = ""

func main() {

	spallet = app.New()
	mainWindowGui = spallet.NewWindow("Spallet")
	mainWindowGui.Resize(fyne.NewSize(800, 600))

	spallet.Lifecycle().SetOnEnteredForeground(func() {
		if logoutTicker != nil {
			logoutTicker.Stop()
		}

	})
	spallet.Lifecycle().SetOnExitedForeground(func() {
		startLogoutTicker(core.UserSettings.LgnTmeOut)
	})

	firstTime := !core.FileExists("data/essential/credentials.spallet")

	if firstTime {
		showWelcomePage()

	} else {

		showExistingUserLogin()

	}

	mainWindowGui.ShowAndRun()
}

// Additional helper function to create token balance display
func createTokenBalance(symbol string, balance string, isNFT bool, creds core.Credentials, decimals int, name string) *fyne.Container {
	iconPath := fmt.Sprintf("img/icons/%s.png", symbol)
	placeholderPath := "img/icons/placeholder.png"
	var icon *canvas.Image

	// Load the icon using the cache
	iconResource := core.LoadIconResource(iconPath)
	if iconResource == nil {
		// Load placeholder image using the cache if the icon is not found
		iconResource = core.LoadIconResource(placeholderPath)
	}
	icon = canvas.NewImageFromResource(iconResource)

	icon.SetMinSize(fyne.NewSize(64, 64))   // Adjust the size to fit the height of the button
	icon.FillMode = canvas.ImageFillContain // Maintain aspect ratio

	// Multi-line text
	text := widget.NewLabel(name + "\n" + balance + " " + symbol)
	text.Alignment = fyne.TextAlignLeading

	buttonContent := container.NewHBox(icon, text)

	customButton := widget.NewButton("", func() {
		{
			if isNFT {
				showSendNFTDia(symbol, creds)
			} else {
				showSendTokenDia(symbol, creds, int8(decimals))
			}
		}
	})

	// Adjust the height of the button to fit the icon
	customButton.Resize(fyne.NewSize(customButton.MinSize().Width, icon.MinSize().Height))

	// Add padding around the button content
	paddedContent := container.NewPadded(buttonContent)

	// Create a final container to hold both the button and its content
	finalContent := container.NewBorder(nil, nil, nil, nil, customButton, paddedContent)
	return container.NewVBox(finalContent, widget.NewSeparator())
}

// Function to update wallet info
func updateWalletInfo(creds core.Credentials, walletInfo *fyne.Container) {
	walletInfo.Objects = []fyne.CanvasObject{

		container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("", theme.SearchIcon(), func(Address string) func() {
			return func() {
				explorerURL := fmt.Sprintf("%s%s", core.UserSettings.AccExplorerLink, creds.Wallets[creds.LastSelectedWallet].Address)
				parsedURL, err := url.Parse(explorerURL)
				if err != nil {
					fmt.Println("Failed to parse URL:", err)
					return
				}
				err = fyne.CurrentApp().OpenURL(parsedURL)
				if err != nil {
					fmt.Println("Failed to open URL:", err)
				}
			}
		}(creds.Wallets[creds.LastSelectedWallet].Address)), widget.NewButtonWithIcon(creds.Wallets[creds.LastSelectedWallet].Address, theme.ContentCopyIcon(), func() {
			fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(creds.Wallets[creds.LastSelectedWallet].Address)
			dialog.ShowInformation("Copied", "Wallet Address copied to clipboard", fyne.CurrentApp().Driver().AllWindows()[0])
		})),
	}
	walletInfo.Refresh()
}

func showExistingUserLogin() {
	if logoutTicker != nil {
		logoutTicker.Stop()
	}
	if updateBalanceTimeOut != nil {
		updateBalanceTimeOut.Stop()
	}

	// stopBadgeAnimation()

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
		pwdForgotDia = dialog.NewCustomWithoutButtons("Forgot Your Password ?", content, mainWindowGui)
		pwdForgotDia.Resize(fyne.NewSize(720, 405))
		pwdForgotDia.Show()
	})
	invalidPasswordLabel.TextStyle = fyne.TextStyle{Bold: true}
	resetWalletBtn.Hide()
	invalidPasswordLabel.Hide()

	logIn := func() {
		rawPassword := passwordEntry.Text

		creds, err := core.LoadCredentials("data/essential/credentials.spallet", rawPassword, "")
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
				dialog.ShowInformation("Error", "Failed to load address book: "+err.Error(), mainWindowGui)
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
				dialog.ShowError(fmt.Errorf("an error happened during fetching data,\n %v", err), mainWindowGui)
				return
			} else {
				core.DataFetch(creds, rootPath)

				autoUpdate(updateInterval, creds, rootPath)
				core.LatestTokenData.ChainTokenUpdateTime = time.Now().UTC().Unix() - 3590 // we will update data automaticaly 15 sec after login with auto update
				core.LatestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
				core.LatestTokenData.AccTokenUpdateTime = time.Now().UTC().Unix()
				mainWindow(creds)
				closeUpdatingDialog()

			}
			// this will update  tokens data from cache if user dont dave cache yet it will create one with main tokens

			// var foundWalletNumber = 0
			// var listedWallets = len(creds.WalletOrder)
			// for _, found := range creds.Wallets { //check if there is a unvisible wallet we have
			// 	var notListedFounded = true
			// 	for _, listed := range creds.WalletOrder {
			// 		if found.Name == listed {
			// 			notListedFounded = false
			// 			continue

			// 		}
			// 		fmt.Println("found not listed wallet ", notListedFounded)

			// 	}

			// 	if notListedFounded {
			// 		creds.WalletOrder = append(creds.WalletOrder, found.Name)
			// 	}
			// 	foundWalletNumber++
			// 	fmt.Println("found wallet ", found.Name)
			// }
			// if foundWalletNumber != listedWallets {
			// 	if err := saveCredentials(creds); err != nil {
			// 		log.Println("Failed to save credentials:", err)
			// 		closeUpdatingDialog()
			// 		dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), fyne.CurrentApp().Driver().AllWindows()[0])
			// 		return
			// 	}
			// }

			// 			if err := core.DataFetch(creds); err != nil {
			// 				closeUpdatingDialog()
			// 				dialog.ShowInformation("Error", "Failed to get wallet balance: "+err.Error(), mainWindowGui)
			// 				return
			// 			} else {
			// }

		} else {
			dialog.ShowInformation("Error", "Invalid password", mainWindowGui)
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
	mainWindowGui.SetContent(logInLyt)
	mainWindowGui.Canvas().Focus(passwordEntry)

}

func mainWindow(creds core.Credentials) {

	walletInfo := container.NewVBox()

	historyContent := container.NewVBox()
	walletSelect := widget.NewSelect(creds.WalletOrder, nil) // Define walletSelect first
	soonContent := widget.NewRichTextWithText("If you're a true OG you already know the drill, in Phantasma everything is just around the corner (SOON). To make life easier for newcomers, here’s a handy list of questions you might avoid asking in community channels, because the answer is always the same: SOON.\n\n1-Wen moon?\n2-Wen marketing\n3-Wen commnunication?\n4-Wen new listing?\n5-Wen Kcal listing?\n6-Wen decentralisation\n7-Wen live-lite to live (this is for samf)\n8-Wen billion dollar partnership\n9-Wen i can buy Pizza with Kcal?\n\n__Some soons for Spallet__\n\n1-Better Gui\n2-Better humors(as you can see, already getting there) 😂🤣\n3-Less bugs (if you found a bug its not a bug its a FEATURE 🤡)\n...and some bla bla bla... 😉🗨️")
	soonContent.Wrapping = fyne.TextWrapWord

	tabContainer := container.NewAppTabs(
		container.NewTabItem("Info", accInfoTab),
		container.NewTabItem("Tokens", tokenBtns),
		container.NewTabItem("NFTs", nftBtns),
		container.NewTabItem("Hodling", stakingTab),
		container.NewTabItem("History", historyContent),
		container.NewTabItem("Dex", dexTab),
		container.NewTabItem("Bridge", widget.NewLabel("When Phantasma enables this\n will try to integrate it but no promises so it means SOON 😂")),
		container.NewTabItem("Soon", soonContent),
	)

	tabContainer.SetTabLocation(container.TabLocationTop)
	tabContainer.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "Hodling":
			accSoul, ok := core.LatestAccountData.FungibleTokens["SOUL"]
			if (ok && accSoul.Amount.Cmp(core.MinSoulStake) >= 0) || core.LatestAccountData.IsStaker {

				feeAmount := new(big.Int).Mul(core.UserSettings.DefaultGasLimit, core.UserSettings.GasPrice)
				err := core.CheckFeeBalance(feeAmount)
				if err != nil {
					dialog.ShowInformation("Low energy", fmt.Sprintf("This account dont have enough Kcal to fill Specky's engines\nPlease check your default fee limit/price in network settings\nor get some Kcal\n\n%s", err), mainWindowGui)
				}
			}

		case "History":
			showUpdatingDialog()
			buildAndShowTxes(creds.Wallets[creds.LastSelectedWallet].Address, page, pageSize, historyContent)
			closeUpdatingDialog()
		case "Dex":
			showUpdatingDialog()
			dexTab = createDexContent(creds)
			closeUpdatingDialog()
		}

	}

	walletDetails := container.NewVBox(walletInfo, tabContainer)

	manageAccountsButton := widget.NewButton("Manage Accounts", func() {
		manageAccountsDia(creds)
	})
	networkButton := widget.NewButton("Network", func() {
		showNetworkSettingsPage(creds)
	})
	chainStatsButton := widget.NewButton("Chain Statistics", func() {
		showUpdatingDialog()
		core.GetChainStatistics() // Assume this updates the walletDetails with chain stats
		buildAndShowChainStatistics()

	})
	refreshButton := widget.NewButton("Refresh", func() {
		showUpdatingDialog()
		err := core.DataFetch(creds, rootPath)
		if err != nil {
			dialog.ShowError(err, mainWindowGui)
		}
		closeUpdatingDialog()
	})
	AddrBkBttn := widget.NewButton("Address Book", func() {
		adddressBookDia(creds.Password)
	})

	// accBadges.Objects = nil
	// accBadges.Objects = []fyne.CanvasObject{buildBadges()} // Initialize badges
	// accBadges.Refresh()
	seperator := widget.NewSeparator()
	securityBttn := widget.NewButton("Security", func() {
		openSecurityDia(creds)
	})
	backupBttn := widget.NewButton("Rescue Point", func() { showBackupDia(creds) })
	menu := container.NewBorder(
		container.NewVBox(walletSelect, accBadges),
		container.NewVBox(refreshButton, backupBttn, AddrBkBttn, securityBttn, chainStatsButton, manageAccountsButton, networkButton),
		nil, nil, nil,
	)

	// Define walletSelect first
	walletSelect.Selected = creds.LastSelectedWallet
	walletSelect.OnChanged = func(selected string) {
		creds.LastSelectedWallet = selected

		showUpdatingDialog()
		err := core.DataFetch(creds, rootPath)
		if err != nil {
			closeUpdatingDialog()
			dialog.ShowInformation("Error", "Failed to update wallet balance: "+err.Error(), mainWindowGui)
		} else {
			updateWalletInfo(creds, walletInfo)

			core.SaveCredentials(creds, rootPath)
			autoUpdate(updateInterval, creds, rootPath)
			dexTab = createDexContent(creds) // dex page also need an update but dunno this is a godd solution
			dexTab.Refresh()

			closeUpdatingDialog()
		}
	}

	updateWalletInfo(creds, walletInfo)

	split := container.NewBorder(nil, nil, container.NewHBox(container.NewPadded(menu), seperator), nil, container.NewPadded(walletDetails))

	mainWindowGui.SetContent(split)

	// _, ok := creds.Wallets[creds.LastSelectedWallet]

	// if !ok {
	// 	if len(creds.WalletOrder) > 0 {
	// 		creds.LastSelectedWallet = creds.WalletOrder[0]
	// 	} else {
	// 		if currentMainDialog != nil {
	// 			currentMainDialog.Hide()
	// 		}
	// 		manageAccountsDia(creds)
	// 		dialog.ShowError(errors.New("cant find any wallet data\nrestart the wallet or enter your keys\nor paste backed up wallet data\nif you didnot backed up your Keys you lost access to assets"), mainWindowGui)

	// 		return

	// 	}

	// }

}
