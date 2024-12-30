package main

import (
	"fmt"
	"math/big"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildBalanceButtons(creds core.Credentials) {
	var tokenBalanceBox *fyne.Container //*********
	tokenBoxes := container.NewVBox()
	nftBoxes := container.NewVBox()

	if core.LatestAccountData.NftTypes >= 1 {
		haveNftContent := widget.NewLabelWithStyle("There is some Smart NFTs that could probably teach you a thing or two.\nLet’s hope they share their secrets with you and make you a billionaire! 💸🧠", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		haveNftContent.Wrapping = fyne.TextWrapWord
		nftBoxes.Add(haveNftContent)

		for _, token := range core.LatestAccountData.SortedNftList {

			formattedBalance := core.FormatBalance(core.LatestAccountData.NonFungible[token].Amount, int(core.LatestAccountData.NonFungible[token].Decimals))
			tokenBalanceBox = createTokenBalance(core.LatestAccountData.NonFungible[token].Symbol, formattedBalance, len(core.LatestAccountData.NonFungible[token].Ids) > 0, creds, int(core.LatestAccountData.NonFungible[token].Decimals), core.LatestAccountData.NonFungible[token].Name)
			nftBoxes.Add(tokenBalanceBox)
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

	nftBtns.Content = nftBoxes
	tokenBtns.Content = tokenBoxes
	nftBtns.Refresh()
	tokenBtns.Refresh()
}

func buildAndShowAccInfo(creds core.Credentials) {
	accKcalAmount := core.LatestAccountData.FungibleTokens["KCAL"].Amount
	accSoulAmount := core.LatestAccountData.FungibleTokens["SOUL"].Amount
	if accKcalAmount == nil {
		accKcalAmount = big.NewInt(0)
	}

	if accSoulAmount == nil {
		accSoulAmount = big.NewInt(0)
	}
	accStakingInfoHeader := " "
	accountSummaryHeader := " "
	if accKcalAmount.Cmp(big.NewInt(3000000000)) < 0 && (core.LatestAccountData.TokenCount > 0 || core.LatestAccountData.NftTypes > 0) {
		accountSummaryHeader = "Looks like Sparky low on sparks! ⚡️🕹️\n Your account needs some Phantasma Energy (KCAL) to keep the ghostly gears turning. Time to add some KCAL and get that blockchain buzzing faster than a haunted hive!"
	} else {
		accountSummaryHeader = "Account Summary"
	}
	// fmt.Println("accIsStaker ", accIsStaker)
	// fmt.Println("accOnChainName ", accOnChainName)

	accSummaryInfo := fmt.Sprintf("Local Name:\t%s\nOn Chain Name:\t%v\nNick name:\t\t%v\nAddress:\t\t%s\nHave %v token(s), %v NFT(s) from %v collections",
		creds.LastSelectedWallet, core.LatestAccountData.OnChainName, core.LatestAccountData.NickName, creds.Wallets[creds.LastSelectedWallet].Address, core.LatestAccountData.TokenCount, core.LatestAccountData.TotalNft, core.LatestAccountData.NftTypes)

	if core.LatestAccountData.IsStaker {

		if core.LatestAccountData.KcalBoost > 0 && core.LatestAccountData.KcalBoost != 100 {

			accSummaryInfo = accSummaryInfo + "\n\nCrowned and cranking out sparks! You're a token-generating powerhouse! 👑⚡"
		} else if core.LatestAccountData.KcalBoost == 100 {
			accSummaryInfo = accSummaryInfo + "\n\nFully boosted and blazing with sparks! All hail the king we bow before your greatness! 👑⚡"
		}

		if !core.LatestAccountData.IsSoulMaster {
			accSummaryInfo = accSummaryInfo + "\n\nSo close yet so far! You're in the staking game, but need a bit more to level up to Soulmaster.\nKeep it up—you’re almost there! 🚀"
		}
	} else if accSoulAmount.Cmp(big.NewInt(0)) > 0 && core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 {
		accSummaryInfo = accSummaryInfo + "\n\nNo staking, no reward! Join the stakers' squad and stop lurking in the shadows.\n It's way more fun on our side! 😜✨"
		accStakingInfoHeader = "Not staking SOUL? SOUL Snoozer💤\nJust sitting on potential without staking it!💸"
	} else if (core.LatestAccountData.TokenCount > 0 || core.LatestAccountData.NftTypes > 0) && (core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 || accSoulAmount.Cmp(big.NewInt(100000000)) < 0) {

		accSummaryInfo = accSummaryInfo + "\n\nGot tokens but no SOUL? That's like having a sandwich without the filling!\nTime to get some SOUL and make it whole! 🥪💫"

	} else {
		accSummaryInfo = accSummaryInfo + "\n\nAn account without assets? It's like a party without snacks!\nTime to fill it up and join the fun! 🎉"
	}

	// fmt.Printf("creds.Wallets[creds.LastSelectedWallet].Address" + creds.Wallets[creds.LastSelectedWallet].Address)

	accStakingInfo := " "

	if core.LatestAccountData.IsStaker {
		accStakingInfoHeader = "Staking Info"
		accStakingInfo = fmt.Sprintf("Staked Soul:\t\t%v Soul\nGenerated sparks:\t%v Kcal\nNext spark gen. after:\t%v\nSpark gen. boost:\t\t%v%%\nDaily spark gen.:\t\t%v Kcal\nCollect Soul after:\t\t%v\n",
			core.FormatBalance(core.LatestAccountData.StakedBalances.Amount, core.SoulDecimals), core.FormatBalance(core.LatestAccountData.StakedBalances.Unclaimed, core.KcalDecimals), core.FormatDuration(time.Duration(core.LatestAccountData.RemainedTimeForKcalGen)*time.Second), core.LatestAccountData.KcalBoost, core.FormatBalance(core.LatestAccountData.KcalDailyProd, core.KcalDecimals), core.FormatDuration(time.Duration(core.LatestAccountData.RemainedTimeForUnstake)*time.Second))

		if accSoulAmount.Cmp(big.NewInt(100000000)) > 0 && core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) >= 0 {
			accStakingInfo = accStakingInfo + "\nSoul slacker! Go full throttle and stake it all! 🚀"
		}
	}

	accSoulMasterInfoHeader := " "
	accSoulMasterInfo := " "
	if core.LatestAccountData.IsSoulMaster {
		accSoulMasterInfoHeader = "Soul Master Info"
		accSoulMasterInfo = fmt.Sprintf("This month's reward\t%.2f Soul\nSoul Master APR\t\t%.2f%% (in Soul, generated sparks not included)\n",
			core.LatestChainStatisticsData.EstSMReward, core.LatestChainStatisticsData.SMApr)
		if !core.LatestAccountData.IsEligibleForCurrentSmReward {
			accSoulMasterInfo = accSoulMasterInfo + "Have Soulmaster status, yet no reward?\nIt’s like bringing snacks to a party and forgetting the drinks.\nContinue Hodling, almost there! 💪\n"
		} else {
			accSoulMasterInfo = accSoulMasterInfo + "Hang tight, Soulmaster! Your reward is just around the corner. Hodl a bit more! 💪💰\n"
		}

		if !core.LatestAccountData.IsEligibleForCurrentCrown {
			accSoulMasterInfo = accSoulMasterInfo + "So close to being crowned, but not quite there yet?\nYou’re like a royal jester without the hat! Keep Hodling, your crown awaits! 👑\n"
		} else {
			accSoulMasterInfo = accSoulMasterInfo + "You're on the brink of royalty.\nKeep hodling, your crown awaits in the current distribution! 👑\n"
		}
	}
	accSumHeader := widget.NewLabelWithStyle(accountSummaryHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	accSumHeader.Wrapping = fyne.TextWrapWord
	accInfo := container.NewVBox(
		accSumHeader,
		widget.NewRichTextWithText(accSummaryInfo),
		widget.NewLabelWithStyle(accStakingInfoHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(accStakingInfo),
		widget.NewLabelWithStyle(accSoulMasterInfoHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(accSoulMasterInfo),
	)
	var x, y float32
	if core.LatestAccountData.IsBalanceUpdated {
		buildBadges()
		buildBalanceButtons(creds)
		x, y = tokenBtns.Offset.Components()
		tokenBtns.SetMinSize(fyne.NewSize(0, 525))
		tokenBtns.Refresh()
		tokenBtns.Offset = fyne.NewPos(x, y)

		x, y = nftBtns.Offset.Components()
		nftBtns.SetMinSize(fyne.NewSize(0, 525))
		nftBtns.Refresh()
		nftBtns.Offset = fyne.NewPos(x, y)
		core.LatestAccountData.IsBalanceUpdated = false
	}

	showStakingPage(creds)

	x, y = accInfoTab.Offset.Components()
	accInfoTab.Content = accInfo
	accInfoTab.SetMinSize(fyne.NewSize(0, 525))
	accInfoTab.Refresh()
	accInfoTab.Offset = fyne.NewPos(x, y)

}
