package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
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
	securityMessage := widget.NewLabel("This wallet uses SHA256 to securely store your wallet data on your hard drive. However, given my limited expertise, please exercise caution and do not solely rely on this security measure.")
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
	features := widget.NewRichTextFromMarkdown("1- Bugs, it means if you found a bug its a feature \n\n2- Nicknames and badges based on Staked soul\n\n3- Account migration from manage accounts menu\n\n4- Sending assets between your accounts\n\n5- Sending assets to address book recipients\n\n6-Collecting Master rewards\n\n7- Collecting Crown rewards\n\n8-Eligibility badges\n\n9-Detailed Account information\n\n10- Showing some chain statistics\n\n11- Detailed staking information under hodling tab\n\n12- 15 minute log in time out\n\n also some other things i forget :)\n\n **What we dont have in spallet**\n\n1- Phantasma link\n\n2- Showing Nft pictures and details\n\n3-Burning tokens\n\nsome other things i dont remember\n\n**Planned Features**\n\nI've planned some features for this wallet, like integrating Saturn Dex, but hey, I'm doing this for fun. Feel free to use it as it is. Since it's open-sourced, you can fork it and continue its development or contribute its code if you like.")
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

	var creds Credentials

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
		_, err := pwdMatch(s, pwd)
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
}

// Function to show Wallet Setup Page
func showWalletSetupPage(creds Credentials) {
	generateWalletButton := widget.NewButton("Generate New Wallet", func() {
		generateNewWalletPage(creds) // Correctly pointing to generateNewWalletPage
	})
	importWifButton := widget.NewButton("Import WIF", func() {
		showImportWifPage(creds)
	})

	walletSetupContent := container.NewVBox(
		widget.NewLabelWithStyle("Choose a way to add new account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		generateWalletButton,
		importWifButton)

	walletSetupLyt := container.NewCenter(walletSetupContent)
	mainWindowGui.SetContent(walletSetupLyt)
}

func generateNewWalletPage(creds Credentials) {
	keyPair := cryptography.GeneratePhantasmaKeys()
	privateKey := keyPair.WIF()
	address := keyPair.Address().String()
	nameEntry := widget.NewEntry()
	nameEntry.SetText("Sparky Account 1")
	nameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
	var isValidName, wifCopied bool
	okButton := widget.NewButton("Continue", func() {

		if creds.Wallets == nil {
			creds.Wallets = make(map[string]Wallet)
		}
		// Add wallet to credentials and mark as last used
		creds.Wallets[nameEntry.Text] = Wallet{
			Name:    nameEntry.Text,
			Address: address,
			WIF:     privateKey,
		}
		creds.WalletOrder = append(creds.WalletOrder, nameEntry.Text)
		creds.LastSelectedWallet = nameEntry.Text
		if err := saveCredentials(creds); err != nil {
			log.Println("Failed to save credentials:", err)
			dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
			return
		} else {
			showUpdatingDialog()
			dataFetch(creds)
			mainWindow(creds, regularTokens, nftTokens)
			closeUpdatingDialog()
		}

	})
	okButton.Disable() // Initially disable the Continue button
	updateokBttnState := func() {
		if isValidName && wifCopied {
			okButton.Enable()
		} else {
			okButton.Disable()
		}
	}

	nameEntry.Validator = func(s string) error {
		names := []string{}
		_, err := validateAccountInput(names, nil, s, "name", false)

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

	cancelButton := widget.NewButton("Cancel", func() {
		showWalletSetupPage(creds) // Go back to wallet setup page
	})

	generatedAccForm := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Address", widget.NewLabel(address)),
		widget.NewFormItem("Private Key (Wif)", copyWifButton),
	)
	warning := widget.NewLabelWithStyle("⚠️In order to continue please copy your Wif and store it in a safe place⚠️", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	// Use container.NewMax to cover full width
	genAccHeader := widget.NewLabelWithStyle("Generated account information", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	generateAccContent := container.NewVBox(genAccHeader, generatedAccForm, warning, container.NewGridWithColumns(2, cancelButton, okButton))
	generateAccLyt := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), container.NewHBox(generateAccContent), layout.NewSpacer()), layout.NewSpacer())

	mainWindowGui.SetContent(generateAccLyt)
	mainWindowGui.Canvas().Focus(nameEntry)
}

// Function to Show Import WIF Page
func showImportWifPage(creds Credentials) {
	wifEntry := widget.NewEntry()
	walletNameEntry := widget.NewEntry()
	walletNameEntry.SetText("Sparky Account 1")
	walletNameEntry.TypedShortcut(&fyne.ShortcutSelectAll{})
	importButton := widget.NewButton("Import", func() {
		keyPair, err := cryptography.FromWIF(wifEntry.Text)
		if err != nil {
			dialog.ShowInformation("Error", "Invalid WIF format", mainWindowGui)
			return
		}
		address := keyPair.Address().String()
		walletName := walletNameEntry.Text

		if creds.Wallets == nil {
			creds.Wallets = make(map[string]Wallet)
		}
		// Add wallet to credentials and mark as last used
		creds.Wallets[walletName] = Wallet{
			Name:    walletName,
			Address: address,
			WIF:     wifEntry.Text,
		}
		creds.WalletOrder = append(creds.WalletOrder, walletName)
		creds.LastSelectedWallet = walletName
		if err := saveCredentials(creds); err != nil {
			log.Println("Failed to save credentials:", err)
			dialog.ShowInformation("Error", "Failed to save credentials: "+err.Error(), mainWindowGui)
			return
		} else {
			showUpdatingDialog()
			dataFetch(creds)
			mainWindow(creds, regularTokens, nftTokens)
			closeUpdatingDialog()
		}
	})
	importButton.Disabled()
	wifEntryForm := widget.NewForm(
		widget.NewFormItem("Name", walletNameEntry),
		widget.NewFormItem("Wif", wifEntry),
	)

	var isValidName, isValidWif bool
	updateImportBttnState := func() {
		if isValidName && isValidWif {
			importButton.Enable()
		} else {
			importButton.Disable()
		}

	}
	wifEntry.Validator = func(s string) error {
		_, err := wifValidator(s)
		if err != nil {
			isValidWif = false
			updateImportBttnState()
			return err
		} else {
			isValidWif = true
			updateImportBttnState()
			return nil
		}
	}
	walletNameEntry.Validator = func(s string) error {
		names := []string{}
		_, err := validateAccountInput(names, nil, s, "name", false)

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
		wifEntryForm,
		container.NewGridWithColumns(2, cancelButton, importButton),
		space,
	)

	importWifLyt := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), container.NewHBox(importWifContent), layout.NewSpacer()), layout.NewSpacer())

	mainWindowGui.SetContent(importWifLyt)
	mainWindowGui.Canvas().Focus(walletNameEntry)

}
