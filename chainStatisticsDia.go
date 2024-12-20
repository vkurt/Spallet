package main

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildAndShowChainStatistics() {

	soulSupplyRawBigint := StringToBigInt(latestTokenData.Token["SOUL"].CurrentSupply)
	kcalSupplyRawBigint := StringToBigInt(latestTokenData.Token["KCAL"].CurrentSupply)
	kcalBurnedSupplyRawBigint := StringToBigInt(latestTokenData.Token["KCAL"].BurnedSupply)
	stakingRatio, _ := new(big.Rat).SetFrac(latestChainStatisticsData.TotalStakedSoul, &soulSupplyRawBigint).Float64()
	stakingRatio *= 100.0

	crownSupply := StringToBigInt(latestTokenData.Token["CROWN"].CurrentSupply)

	maxFullyBoostedSoulmaster := new(big.Int).Div(&crownSupply, big.NewInt(20))
	// Convert to big.Float
	maxFullyBoostedSoulmasterFloat := new(big.Float).SetInt(maxFullyBoostedSoulmaster)

	// Convert to float64
	maxFullyBoostedSoulmasterFloat64, _ := maxFullyBoostedSoulmasterFloat.Float64()

	boostedRawStakedSupply := new(big.Float).Mul(maxFullyBoostedSoulmasterFloat, new(big.Float).SetInt(soulMasterThreshold))
	notBoostedRawStakedSupply := new(big.Float).Sub(new(big.Float).SetInt(latestChainStatisticsData.TotalStakedSoul), boostedRawStakedSupply)

	estimatedBoostedDailyKcalGeneration := new(big.Float).Mul(new(big.Float).Add(new(big.Float).Mul(boostedRawStakedSupply, new(big.Float).SetFloat64(kcalProdRate*2)), new(big.Float).Mul(notBoostedRawStakedSupply, new(big.Float).SetFloat64(kcalProdRate))), big.NewFloat(100))

	estimatedBoostedDailyKcalGenerationInt := new(big.Int)
	estimatedBoostedDailyKcalGeneration.Int(estimatedBoostedDailyKcalGenerationInt) // Truncate to int

	boostedStakedSupplyRatio, _ := new(big.Float).Mul(new(big.Float).Quo(boostedRawStakedSupply, new(big.Float).SetInt(latestChainStatisticsData.TotalStakedSoul)), big.NewFloat(100)).Float64()

	// fmt.Println("kcalSupplyRawBigint", kcalSupplyRawBigint.String())
	// fmt.Println("estimatedBoostedDailyKcalGenerationInt", estimatedBoostedDailyKcalGenerationInt.String())

	annualKcalSupplyGrowthRate := new(big.Float).Mul(
		new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(estimatedBoostedDailyKcalGenerationInt), big.NewFloat(365)),
			new(big.Float).SetInt(&kcalSupplyRawBigint),
		),
		big.NewFloat(100),
	)
	fmt.Println("Kcal Raw supply", formatBalance(kcalSupplyRawBigint, kcalDecimals))
	fmt.Println("Kcal Burned supply", formatBalance(kcalBurnedSupplyRawBigint, kcalDecimals))

	kcalSupplyAt241120241710 := new(big.Int).SetInt64(2004801680000000000)
	averageDailyKcalGenerationSince231120241710 := new(big.Int).Mul(new(big.Int).Div(new(big.Int).Sub(&kcalSupplyRawBigint, kcalSupplyAt241120241710), big.NewInt((time.Now().UTC().Unix()-1732371000)/3600)), big.NewInt(24))
	averageDailyKcalGenerationSince231120241710Str := formatBalance(*averageDailyKcalGenerationSince231120241710, kcalDecimals)

	averageDailyKcalGenSinceMainnetLaunch := new(big.Int).Div(new(big.Int).Add(&kcalBurnedSupplyRawBigint, &kcalSupplyRawBigint), new(big.Int).Div(new(big.Int).SetInt64((time.Now().UTC().Unix()-1570665600)), new(big.Int).SetInt64(86400)))
	averageDailyKcalGenSinceMainnetLaunchStr := formatBalance(*averageDailyKcalGenSinceMainnetLaunch, kcalDecimals)
	mainTokenInfo := fmt.Sprintf("Phantasma Stake\n \tCurrent Supply \t %s Soul\n\n Phantasma Energy\n \tCurrent Supply \t %s Kcal \n\tBurned Supply\t %s Kcal", formatBalance(soulSupplyRawBigint, soulDecimals), formatBalance(kcalSupplyRawBigint, kcalDecimals), formatBalance(kcalBurnedSupplyRawBigint, kcalDecimals))
	sparkGenRate := math.Pow10(kcalDecimals) * kcalProdRate
	stakingInfo := fmt.Sprintf("Soul Master Reward\t%.4f\tSoul\nSpark gen per Soul:\t%.0f Spark (or %.4f Kcal)\nStaked Supply\t\t%s\tSoul\nStaking Ratio\t\t%.2f%%\nSoul Master APR\t\t%.2f%% (in Soul and generated Kcal not included)\nSoul Masters\t\t%v\nEligible Soul Masters\t%v\nCrown Supply\t\t%v\nNext Crown dist. date\t%v\n\n***Below Data Calculated Based on Stake amount in stake contract and Crown Supply***\nFully boosted Soul Masters\t %.0f\nBoosted daily Kcal generation\t %v Kcal\nBoosted Staked Supply Ratio\t %.2f%% \nAnn. Kcal Supply Growth Rate\t %.2f%%\n\nAverage daily Kcal claim since mainnet launch\t%s\n\nAverage daily Kcal claim since 24 11 2024 17 10\t%s",
		latestChainStatisticsData.EstSMReward, sparkGenRate, kcalProdRate, formatBalance(*latestChainStatisticsData.TotalStakedSoul, soulDecimals), stakingRatio, latestChainStatisticsData.SMApr, latestChainStatisticsData.TotalMaster, latestChainStatisticsData.EligibleMaster, latestTokenData.Token["CROWN"].CurrentSupply, time.Unix(latestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"), maxFullyBoostedSoulmasterFloat64, formatBalance(*estimatedBoostedDailyKcalGenerationInt, kcalDecimals), boostedStakedSupplyRatio, annualKcalSupplyGrowthRate, averageDailyKcalGenSinceMainnetLaunchStr, averageDailyKcalGenerationSince231120241710Str)

	statistics := container.NewVBox(

		widget.NewLabelWithStyle("Main Tokens info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(mainTokenInfo),
		widget.NewLabelWithStyle("Staking Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(stakingInfo),
	)
	bckBttn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { currentMainDialog.Hide() })
	chnStcsCntnt := container.NewVScroll(statistics)
	chnStcsLyt := container.NewBorder(nil, bckBttn, nil, nil, chnStcsCntnt)
	currentMainDialog = dialog.NewCustomWithoutButtons(fmt.Sprintf("Chain Statistics For %v", userSettings.NetworkName), chnStcsLyt, mainWindowGui)
	currentMainDialog.Resize(fyne.NewSize(chnStcsCntnt.MinSize().Width, mainWindowGui.Canvas().Size().Height-50))
	closeUpdatingDialog()
	currentMainDialog.Show()

}
