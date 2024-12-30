package main

import (
	"fmt"
	"net/url"
	"spallet/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var tokenBtns = container.NewVBox(container.NewVBox())
var nftBtns = container.NewVBox(container.NewVBox())
var settingsDia dialog.Dialog

func homeGui(creds core.Credentials) {
	var onChainNameText string
	walletSelectButtonName := creds.LastSelectedWallet
	walletAddress := creds.Wallets[creds.LastSelectedWallet].Address
	addressBookBtn := widget.NewButtonWithIcon("", resourceAddressbookPng, func() {
		showAdddressBookWin(creds.Password)
	})

	if core.LatestAccountData.OnChainName == "anonymous" && core.LatestAccountData.IsStaker {
		onChainNameText = "You can forge a name from hodling tab"
	} else if core.LatestAccountData.OnChainName != "anonymous" {
		onChainNameText = core.LatestAccountData.OnChainName
	} else {
		onChainNameText = "Stakers can forge a name."
	}
	nameMessage := widget.NewLabelWithStyle(fmt.Sprintf("%s\n%s", core.LatestAccountData.NickName, onChainNameText), fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		network := widget.NewButton("Network", func() {
			showNetworkSettingsWin(creds)
		})
		security := widget.NewButton("Security", func() {
			showSecurityWin(creds)
		})
		content := container.NewVBox(network, security)
		settingsDia = dialog.NewCustom("Settings", "Close", content, mainWindow)
		settingsDia.Show()
	})

	chainStatsBtn := widget.NewButtonWithIcon("", resourceChainstatsPng, func() {
		buildAndShowChainStatistics()
	})

	buttonsContainer := container.NewGridWithColumns(3, addressBookBtn, chainStatsBtn, settingsBtn)
	addressCopyButtonName := walletAddress[:5] + "..." + walletAddress[len(walletAddress)-5:]
	addressCopyBtn := widget.NewButtonWithIcon(addressCopyButtonName, theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(creds.Wallets[creds.LastSelectedWallet].Address)
		dialog.ShowInformation("Copied", "Address copied to the clipboard", mainWindow)
	})
	walletSelect := widget.NewButtonWithIcon(walletSelectButtonName, theme.MenuDropDownIcon(), func() {
		manageAccountsDia(creds)
	})
	buildBalanceButtons(creds)
	badges := buildBadges()
	walletExpBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		explorerURL := fmt.Sprintf("%s%s", core.UserSettings.AccExplorerLink, walletAddress)
		parsedURL, err := url.Parse(explorerURL)
		if err != nil {
			fmt.Println("Failed to parse URL:", err)
			return
		}
		err = fyne.CurrentApp().OpenURL(parsedURL)
		if err != nil {
			fmt.Println("Failed to open URL:", err)
		}
	})
	walletInfoGroupLyt := container.NewBorder(nil, nil, walletSelect, walletExpBtn, addressCopyBtn)
	homeContent := container.NewVScroll(container.NewVBox(nameMessage, walletInfoGroupLyt, buttonsContainer, badges, tokenBtns))
	// homeContent.SetMinSize()
	homeTab.Content = homeContent
	homeTab.Content.Refresh()

}
func buildBalanceButtons(creds core.Credentials) {
	var tokenBalanceBox *fyne.Container //*********
	var nftBalanceBox *fyne.Container
	tokenBoxes := container.NewVBox()
	nftBoxes := container.NewVBox()

	if core.LatestAccountData.NftTypes >= 1 {
		haveNftContent := widget.NewLabelWithStyle("There is some Smart NFTs that could probably teach you a thing or two.\nLet’s hope they share their secrets with you and make you a billionaire! 💸🧠", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		haveNftContent.Wrapping = fyne.TextWrapWord
		nftBoxes.Add(haveNftContent)

		for _, token := range core.LatestAccountData.SortedNftList {

			formattedBalance := core.FormatBalance(core.LatestAccountData.NonFungible[token].Amount, int(core.LatestAccountData.NonFungible[token].Decimals))
			nftBalanceBox = createTokenBalance(core.LatestAccountData.NonFungible[token].Symbol, formattedBalance, len(core.LatestAccountData.NonFungible[token].Ids) > 0, creds, int(core.LatestAccountData.NonFungible[token].Decimals), core.LatestAccountData.NonFungible[token].Name)
			nftBoxes.Add(nftBalanceBox)
		}

	} else {
		noNFTContent := widget.NewLabelWithStyle("No Smart NFTs in your wallet? It’s like being a gamer without a high score! Time to level up and let the games begin. 🎮✨", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		noNFTContent.Wrapping = fyne.TextWrapWord
		nftBoxes.Add(noNFTContent)
	}

	if core.LatestAccountData.TokenCount >= 1 {
		haveTokenContent := widget.NewLabelWithStyle("Ohhh, you’ve got a moon bag! But the million-dollar question is, 'Wen moon?' 🚀🌕", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		haveTokenContent.Wrapping = fyne.TextWrapWord
		tokenBoxes.Add(haveTokenContent)

		for _, token := range core.LatestAccountData.SortedTokenList {

			formattedBalance := core.FormatBalance(core.LatestAccountData.FungibleTokens[token].Amount, int(core.LatestAccountData.FungibleTokens[token].Decimals))
			tokenBalanceBox = createTokenBalance(core.LatestAccountData.FungibleTokens[token].Symbol, formattedBalance, len(core.LatestAccountData.FungibleTokens[token].Ids) > 0, creds, int(core.LatestAccountData.FungibleTokens[token].Decimals), core.LatestAccountData.FungibleTokens[token].Name)

			tokenBoxes.Add(tokenBalanceBox)
		}

	} else {
		noTokenContent := widget.NewLabelWithStyle("Your wallet is so empty, even the crypto memes are feeling sorry for you. \nNo shittokens to be found here!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		noTokenContent.Wrapping = fyne.TextWrapWord
		tokenBoxes.Add(noTokenContent)

	}

	nftBtns.Objects = nftBoxes.Objects
	tokenBtns.Objects = tokenBoxes.Objects
	nftBtns.Refresh()
	tokenBtns.Refresh()
}

func createTokenBalance(symbol string, balance string, isNFT bool, creds core.Credentials, decimals int, name string) *fyne.Container {

	icon := loadImgFromResource(symbol, fyne.NewSize(64, 64))

	icon.FillMode = canvas.ImageFillContain // Maintain aspect ratio

	// Multi-line text
	text := widget.NewLabel(name + "\n" + balance + " " + symbol)
	text.Alignment = fyne.TextAlignLeading

	buttonContent := container.NewHBox(icon, text)

	customButton := widget.NewButton("", func() {
		{
			askPwdDia(core.UserSettings.AskPwd, creds.Password, mainWindow, func(correct bool) {
				fmt.Println("result", correct)
				if !correct {
					return
				}
				if isNFT {
					showSendNFTDia(symbol, creds)
				} else {
					showSendTokenDia(symbol, creds, int8(decimals))
				}
			})

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

func buildBadges() *fyne.Container {

	var crownBadge *fyne.StaticResource
	var soulMasterBadge *fyne.StaticResource
	var stakingBadge *fyne.StaticResource
	var networkBadge *fyne.StaticResource

	fmt.Println("******Building badges*****")

	switch {
	case core.LatestAccountData.IsSoulMaster && core.LatestAccountData.IsEligibleForCurrentCrown && core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadge = resourceCROWNPng
		soulMasterBadge = resourceSoulmasterPng
		stakingBadge = resourceKCALPng
	case core.LatestAccountData.IsSoulMaster && core.LatestAccountData.IsEligibleForCurrentCrown && !core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadge = resourceCROWNPng
		soulMasterBadge = resourceSoulmasterenPng
		stakingBadge = resourceKCALPng
	case core.LatestAccountData.IsSoulMaster && !core.LatestAccountData.IsEligibleForCurrentCrown && !core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadge = resourceCROWNenPng
		soulMasterBadge = resourceSoulmasterenPng
		stakingBadge = resourceKCALPng
	case core.LatestAccountData.IsSoulMaster && !core.LatestAccountData.IsEligibleForCurrentCrown && core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadge = resourceCROWNenPng
		soulMasterBadge = resourceSoulmasterPng
		stakingBadge = resourceKCALPng
	case core.LatestAccountData.IsStaker:
		crownBadge = resourceCROWNnePng
		soulMasterBadge = resourceSoulmasternePng
		stakingBadge = resourceKCALPng
	default:
		crownBadge = resourceCROWNnePng
		soulMasterBadge = resourceSoulmasternePng
		stakingBadge = resourceKcalnsPng
	}

	if core.UserSettings.NetworkName == "mainnet" {
		networkBadge = resourceMainnetPng
	} else {
		networkBadge = resourceTestnetPng
	}

	networkBadgeImg := canvas.NewImageFromResource(networkBadge)
	networkBadgeImg.FillMode = canvas.ImageFillContain
	networkBadgeImg.SetMinSize(fyne.NewSize(32, 32))

	crownBadgeImg := canvas.NewImageFromResource(crownBadge)
	crownBadgeImg.FillMode = canvas.ImageFillContain
	// crownBadge.SetMinSize(fyne.NewSize(26, 26))

	soulMasterBadgeImg := canvas.NewImageFromResource(soulMasterBadge)
	soulMasterBadgeImg.FillMode = canvas.ImageFillContain
	// soulMasterBadge.SetMinSize(fyne.NewSize(26, 26))

	stakingBadgeImg := canvas.NewImageFromResource(stakingBadge)
	stakingBadgeImg.FillMode = canvas.ImageFillContain
	// stakingBadge.SetMinSize(fyne.NewSize(26, 26))

	imageContainer := container.NewGridWithColumns(4, stakingBadgeImg, soulMasterBadgeImg, crownBadgeImg, networkBadgeImg)

	return imageContainer
}
