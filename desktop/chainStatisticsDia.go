package main

import (
	"fmt"
	"math"
	"math/big"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildAndShowChainStatistics() {

	soulSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["SOUL"].CurrentSupply)
	kcalSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["KCAL"].CurrentSupply)
	kcalBurnedSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["KCAL"].BurnedSupply)
	stakingRatio, _ := new(big.Rat).SetFrac(core.LatestChainStatisticsData.TotalStakedSoul, &soulSupplyRawBigint).Float64()
	stakingRatio *= 100.0

	crownSupply := core.StringToBigInt(core.LatestTokenData.Token["CROWN"].CurrentSupply)

	maxFullyBoostedSoulmaster := new(big.Int).Div(&crownSupply, big.NewInt(20))
	// Convert to big.Float
	maxFullyBoostedSoulmasterFloat := new(big.Float).SetInt(maxFullyBoostedSoulmaster)

	// Convert to float64
	maxFullyBoostedSoulmasterFloat64, _ := maxFullyBoostedSoulmasterFloat.Float64()

	boostedRawStakedSupply := new(big.Float).Mul(maxFullyBoostedSoulmasterFloat, new(big.Float).SetInt(core.SoulMasterThreshold))
	notBoostedRawStakedSupply := new(big.Float).Sub(new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul), boostedRawStakedSupply)

	estimatedBoostedDailyKcalGeneration := new(big.Float).Mul(new(big.Float).Add(new(big.Float).Mul(boostedRawStakedSupply, new(big.Float).SetFloat64(core.KcalProdRate*2)), new(big.Float).Mul(notBoostedRawStakedSupply, new(big.Float).SetFloat64(core.KcalProdRate))), big.NewFloat(100))

	estimatedBoostedDailyKcalGenerationInt := new(big.Int)
	estimatedBoostedDailyKcalGeneration.Int(estimatedBoostedDailyKcalGenerationInt) // Truncate to int

	boostedStakedSupplyRatio, _ := new(big.Float).Mul(new(big.Float).Quo(boostedRawStakedSupply, new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul)), big.NewFloat(100)).Float64()

	// fmt.Println("kcalSupplyRawBigint", kcalSupplyRawBigint.String())
	// fmt.Println("estimatedBoostedDailyKcalGenerationInt", estimatedBoostedDailyKcalGenerationInt.String())

	annualKcalSupplyGrowthRate := new(big.Float).Mul(
		new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(estimatedBoostedDailyKcalGenerationInt), big.NewFloat(365)),
			new(big.Float).SetInt(&kcalSupplyRawBigint),
		),
		big.NewFloat(100),
	)
	fmt.Println("Kcal Raw supply", core.FormatBalance(&kcalSupplyRawBigint, core.KcalDecimals))
	fmt.Println("Kcal Burned supply", core.FormatBalance(&kcalBurnedSupplyRawBigint, core.KcalDecimals))

	kcalSupplyAt241120241710 := new(big.Int).SetInt64(2004801680000000000)
	averageDailyKcalGenerationSince231120241710 := new(big.Int).Mul(new(big.Int).Div(new(big.Int).Sub(&kcalSupplyRawBigint, kcalSupplyAt241120241710), big.NewInt((time.Now().UTC().Unix()-1732371000)/3600)), big.NewInt(24))
	averageDailyKcalGenerationSince231120241710Str := core.FormatBalance(averageDailyKcalGenerationSince231120241710, core.KcalDecimals)

	averageDailyKcalGenSinceMainnetLaunch := new(big.Int).Div(new(big.Int).Add(&kcalBurnedSupplyRawBigint, &kcalSupplyRawBigint), new(big.Int).Div(new(big.Int).SetInt64((time.Now().UTC().Unix()-1570665600)), new(big.Int).SetInt64(86400)))
	averageDailyKcalGenSinceMainnetLaunchStr := core.FormatBalance(averageDailyKcalGenSinceMainnetLaunch, core.KcalDecimals)
	mainTokenInfo := fmt.Sprintf("Phantasma Stake\n \tCurrent Supply \t %s Soul\n\n Phantasma Energy\n \tCurrent Supply \t %s Kcal \n\tBurned Supply\t %s Kcal", core.FormatBalance(&soulSupplyRawBigint, core.SoulDecimals), core.FormatBalance(&kcalSupplyRawBigint, core.KcalDecimals), core.FormatBalance(&kcalBurnedSupplyRawBigint, core.KcalDecimals))
	sparkGenRate := math.Pow10(core.KcalDecimals) * core.KcalProdRate
	stakingInfo := fmt.Sprintf("Soul Master Reward\t%.4f\tSoul\nSpark gen per Soul:\t%.0f Spark (or %.4f Kcal)\nStaked Supply\t\t%s\tSoul\nStaking Ratio\t\t%.2f%%\nSoul Master APR\t\t%.2f%% (in Soul and generated Kcal not included)\nSoul Masters\t\t%v\nEligible Soul Masters\t%v\nCrown Supply\t\t%v\nNext Crown dist. date\t%v\n\n***Below Data Calculated Based on Stake amount in stake contract and Crown Supply***\nFully boosted Soul Masters\t %.0f\nBoosted daily Kcal generation\t %v Kcal\nBoosted Staked Supply Ratio\t %.2f%% \nAnn. Kcal Supply Growth Rate\t %.2f%%\n\nAverage daily Kcal claim since mainnet launch\t%s\n\nAverage daily Kcal claim since 24 11 2024 17 10\t%s",
		core.LatestChainStatisticsData.EstSMReward, sparkGenRate, core.KcalProdRate, core.FormatBalance(core.LatestChainStatisticsData.TotalStakedSoul, core.SoulDecimals), stakingRatio, core.LatestChainStatisticsData.SMApr, core.LatestChainStatisticsData.TotalMaster, core.LatestChainStatisticsData.EligibleMaster, core.LatestTokenData.Token["CROWN"].CurrentSupply, time.Unix(core.LatestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"), maxFullyBoostedSoulmasterFloat64, core.FormatBalance(estimatedBoostedDailyKcalGenerationInt, core.KcalDecimals), boostedStakedSupplyRatio, annualKcalSupplyGrowthRate, averageDailyKcalGenSinceMainnetLaunchStr, averageDailyKcalGenerationSince231120241710Str)

	statistics := container.NewVBox(

		widget.NewLabelWithStyle("Main Tokens info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(mainTokenInfo),
		widget.NewLabelWithStyle("Staking Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(stakingInfo),
	)
	bckBttn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { currentMainDialog.Hide() })
	chnStcsCntnt := container.NewVScroll(statistics)
	chnStcsLyt := container.NewBorder(nil, bckBttn, nil, nil, chnStcsCntnt)
	currentMainDialog = dialog.NewCustomWithoutButtons(fmt.Sprintf("Chain Statistics For %v", core.UserSettings.NetworkName), chnStcsLyt, mainWindowGui)
	currentMainDialog.Resize(fyne.NewSize(chnStcsCntnt.MinSize().Width, mainWindowGui.Canvas().Size().Height-50))
	closeUpdatingDialog()
	currentMainDialog.Show()

}
