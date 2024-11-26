package main

import (
	"fmt"
	"math/big"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type AccountInfoData struct {
	Name                         string
	Address                      string
	FungibleTokens               map[string]AccToken
	NonFungible                  map[string]AccToken
	OnChainName                  string
	StakedBalances               Stake
	IsStaker                     bool
	IsSoulMaster                 bool
	IsEligibleForCurrentCrown    bool
	IsEligibleForCurrentSmReward bool
	KcalBoost                    int16
	NftTypes                     int8
	TotalNft                     int64
	TokenCount                   int8
	RemainedTimeForKcalGen       int64
	KcalDailyProd                big.Int
	RemainedTimeForUnstake       int64
	LastStakeTimestamp           int64
	SoulmasterSince              int64
	BadgeName                    string
	NickName                     string
	StatCheckTime                int64
	TransactionCount             int
	Network                      string
}

func buildAndShowAccInfo(creds Credentials) {
	accKcalAmount := latestAccountData.FungibleTokens["KCAL"].Amount
	accSoulAmount := latestAccountData.FungibleTokens["SOUL"].Amount
	accStakingInfoHeader := " "
	accountSummaryHeader := " "
	if accKcalAmount.Cmp(big.NewInt(10000000000)) < 0 && (latestAccountData.TokenCount > 0 || latestAccountData.NftTypes > 0) {
		accountSummaryHeader = "Running low on Phantasma Energy(Kcal)?\n You’re running on empty! Charge up those sparks you can’t transact without them.\n It's the fuel of the Phantasma chain! ⚡💡"
	} else {
		accountSummaryHeader = "Account Summary"
	}
	// fmt.Println("accIsStaker ", accIsStaker)
	// fmt.Println("accOnChainName ", accOnChainName)

	fmt.Println("accKcalBoost accIsSoulMaster ", latestAccountData.KcalBoost, latestAccountData.IsSoulMaster)
	if latestAccountData.KcalBoost == 100 && latestAccountData.IsSoulMaster {
		latestAccountData.BadgeName = "lord"
		latestAccountData.NickName = "Spark Lord 🔥"
	} else if latestAccountData.KcalBoost > 0 && latestAccountData.IsSoulMaster {
		latestAccountData.BadgeName = "master"
		latestAccountData.NickName = "Spark Master 💥"

	} else if latestAccountData.IsSoulMaster {
		latestAccountData.BadgeName = "apprentice"
		latestAccountData.NickName = "Spark Apprentice ✨"

	} else if accSoulAmount.Cmp(big.NewInt(100000000)) > 0 && latestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) >= 0 {
		latestAccountData.BadgeName = "snoozer"
		latestAccountData.NickName = "Soul slacker 😴"

	} else if latestAccountData.IsStaker {
		latestAccountData.BadgeName = "acolyte"
		latestAccountData.NickName = "Spark Acolyte ⚡️"

	} else if accSoulAmount.Cmp(big.NewInt(100000000)) >= 0 && latestAccountData.OnChainName == "anonymous" {
		latestAccountData.BadgeName = "snoozer"
		latestAccountData.NickName = "Soul snoozer💤"

	} else if latestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 && accSoulAmount.Cmp(big.NewInt(100000000)) < 0 {
		latestAccountData.BadgeName = "wanderer"
		latestAccountData.NickName = "Soulless wanderer 🌑"

	}

	buildBadges()

	accSummaryInfo := fmt.Sprintf("Local Name:\t%s\nOn Chain Name:\t%v\nNick name:\t\t%v\nAddress:\t\t%s\nHave %v token(s), %v NFT(s) from %v collections",
		creds.LastSelectedWallet, latestAccountData.OnChainName, latestAccountData.NickName, creds.Wallets[creds.LastSelectedWallet].Address, latestAccountData.TokenCount, latestAccountData.TotalNft, latestAccountData.NftTypes)

	if latestAccountData.IsStaker {

		if latestAccountData.KcalBoost > 0 && latestAccountData.KcalBoost != 100 {

			accSummaryInfo = accSummaryInfo + "\n\nCrowned and cranking out sparks! You're a token-generating powerhouse! 👑⚡"
		} else if latestAccountData.KcalBoost == 100 {
			accSummaryInfo = accSummaryInfo + "\n\nFully boosted and blazing with sparks! All hail the king we bow before your greatness! 👑⚡"
		}

		if !latestAccountData.IsSoulMaster {
			accSummaryInfo = accSummaryInfo + "\n\nSo close yet so far! You're in the staking game, but need a bit more to level up to Soulmaster.\nKeep it up—you’re almost there! 🚀"
		}
	} else if accSoulAmount.Cmp(big.NewInt(0)) > 0 && latestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 {
		accSummaryInfo = accSummaryInfo + "\n\nNo staking, no reward! Join the stakers' squad and stop lurking in the shadows.\n It's way more fun on our side! 😜✨"
		accStakingInfoHeader = "Not staking SOUL? SOUL Snoozer💤\nJust sitting on potential without staking it!💸"
	} else if (latestAccountData.TokenCount > 0 || latestAccountData.NftTypes > 0) && (latestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 || accSoulAmount.Cmp(big.NewInt(100000000)) < 0) {

		accSummaryInfo = accSummaryInfo + "\n\nGot tokens but no SOUL? That's like having a sandwich without the filling!\nTime to get some SOUL and make it whole! 🥪💫"

	} else {
		accSummaryInfo = accSummaryInfo + "\n\nAn account without assets? It's like a party without snacks!\nTime to fill it up and join the fun! 🎉"
	}

	// fmt.Printf("creds.Wallets[creds.LastSelectedWallet].Address" + creds.Wallets[creds.LastSelectedWallet].Address)

	accStakingInfo := " "

	if latestAccountData.IsStaker {
		accStakingInfoHeader = "Staking Info"
		accStakingInfo = fmt.Sprintf("Staked Soul:\t\t%v Soul\nGenerated sparks:\t%v Kcal\nNext spark gen. after:\t%v\nSpark gen. boost:\t\t%v%%\nDaily spark gen.:\t\t%v Kcal\nCollect Soul after:\t\t%v\n",
			formatBalance(latestAccountData.StakedBalances.Amount, soulDecimals), formatBalance(latestAccountData.StakedBalances.Unclaimed, kcalDecimals), time.Duration(latestAccountData.RemainedTimeForKcalGen)*time.Second, latestAccountData.KcalBoost, formatBalance(latestAccountData.KcalDailyProd, kcalDecimals), time.Duration(latestAccountData.RemainedTimeForUnstake)*time.Second)

		if accSoulAmount.Cmp(big.NewInt(100000000)) > 0 && latestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) >= 0 {
			accStakingInfo = accStakingInfo + "\nSoul slacker! Go full throttle and stake it all! 🚀"
		}
	}

	accSoulMasterInfoHeader := " "
	accSoulMasterInfo := " "
	if latestAccountData.IsSoulMaster {
		accSoulMasterInfoHeader = "Soul Master Info"
		accSoulMasterInfo = fmt.Sprintf("This month's reward\t%.2f Soul\nSoul Master APR\t\t%.2f%% (in Soul, generated sparks not included)\n",
			latestChainStatisticsData.EstSMReward, latestChainStatisticsData.SMApr)
		if !latestAccountData.IsEligibleForCurrentSmReward {
			accSoulMasterInfo = accSoulMasterInfo + "Have Soulmaster status, yet no reward?\nIt’s like bringing snacks to a party and forgetting the drinks.\nContinue Hodling, almost there! 💪\n"
		} else {
			accSoulMasterInfo = accSoulMasterInfo + "Hang tight, Soulmaster! Your reward is just around the corner. Hodl a bit more! 💪💰\n"
		}

		if !latestAccountData.IsEligibleForCurrentCrown {
			accSoulMasterInfo = accSoulMasterInfo + "So close to being crowned, but not quite there yet?\nYou’re like a royal jester without the hat! Keep Hodling, your crown awaits! 👑\n"
		} else {
			accSoulMasterInfo = accSoulMasterInfo + "You're on the brink of royalty.\nKeep hodling, your crown awaits in the current distribution! 👑\n"
		}
	}
	accInfo := container.NewVBox(
		widget.NewLabelWithStyle(accountSummaryHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewRichTextWithText(accSummaryInfo),
		widget.NewLabelWithStyle(accStakingInfoHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(accStakingInfo),
		widget.NewLabelWithStyle(accSoulMasterInfoHeader, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(accSoulMasterInfo),
	)

	x, y := tokenTab.Offset.Components()
	tokenTab.SetMinSize(fyne.NewSize(0, 525))
	tokenTab.Refresh()
	tokenTab.Offset = fyne.NewPos(x, y)

	x, y = nftTab.Offset.Components()
	nftTab.SetMinSize(fyne.NewSize(0, 525))
	nftTab.Refresh()
	nftTab.Offset = fyne.NewPos(x, y)

	x, y = accInfoTab.Offset.Components()
	accInfoTab.Content = accInfo
	accInfoTab.SetMinSize(fyne.NewSize(0, 525))
	accInfoTab.Refresh()
	accInfoTab.Offset = fyne.NewPos(x, y)

}
