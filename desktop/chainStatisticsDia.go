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
	var mainnetLaunchDate = int64(1570665600)
	var currentTime = time.Now().UTC().Unix()
	var daysPassedSinceMainnetLAunch = (currentTime - mainnetLaunchDate) / 86400
	soulSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["SOUL"].CurrentSupply)
	kcalSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["KCAL"].CurrentSupply)
	kcalBurnedSupplyRawBigint := core.StringToBigInt(core.LatestTokenData.Token["KCAL"].BurnedSupply)
	stakingRatio, _ := new(big.Rat).SetFrac(core.LatestChainStatisticsData.TotalStakedSoul, &soulSupplyRawBigint).Float64()
	stakingRatio *= 100.0

	crownSupply := core.StringToBigInt(core.LatestTokenData.Token["CROWN"].CurrentSupply)

	// Kcal absolute maximum calculations
	maxFullyBoostedSoulmaster := new(big.Int).Div(&crownSupply, big.NewInt(20))
	maxFullyBoostedSoulmasterFloat := new(big.Float).SetInt(maxFullyBoostedSoulmaster)
	maxFullyBoostedSoulmasterFloat64, _ := maxFullyBoostedSoulmasterFloat.Float64()
	boostedRawStakedSupply := new(big.Float).Mul(maxFullyBoostedSoulmasterFloat, new(big.Float).SetInt(core.SoulMasterThreshold))
	if boostedRawStakedSupply.Cmp(new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul)) > 0 {
		boostedRawStakedSupply = new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul)
	}
	notBoostedRawStakedSupply := new(big.Float).Sub(new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul), boostedRawStakedSupply)
	estimatedBoostedDailyKcalGeneration := new(big.Float).Mul(new(big.Float).Add(new(big.Float).Mul(boostedRawStakedSupply, new(big.Float).SetFloat64(core.KcalProdRate*2)), new(big.Float).Mul(notBoostedRawStakedSupply, new(big.Float).SetFloat64(core.KcalProdRate))), big.NewFloat(100))
	estimatedBoostedDailyKcalGenerationInt := new(big.Int)
	estimatedBoostedDailyKcalGeneration.Int(estimatedBoostedDailyKcalGenerationInt) // Truncate to int
	boostedStakedSupplyRatio, _ := new(big.Float).Mul(new(big.Float).Quo(boostedRawStakedSupply, new(big.Float).SetInt(core.LatestChainStatisticsData.TotalStakedSoul)), big.NewFloat(100)).Float64()

	// fmt.Println("kcalSupplyRawBigint", kcalSupplyRawBigint.String())
	// fmt.Println("estimatedBoostedDailyKcalGenerationInt", estimatedBoostedDailyKcalGenerationInt.String())

	annualKcalSupplyGrowthRateMax := new(big.Float).Mul(
		new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(estimatedBoostedDailyKcalGenerationInt), big.NewFloat(365)),
			new(big.Float).SetInt(&kcalSupplyRawBigint),
		),
		big.NewFloat(100),
	)
	annualKcalSupplyGrowthRateMaxFloat, _ := annualKcalSupplyGrowthRateMax.Float64()

	// Kcal All time avarage calculations
	averageDailyKcalClaimsSinceMainnetLaunch := new(big.Int).Div(new(big.Int).Add(&kcalBurnedSupplyRawBigint, &kcalSupplyRawBigint), big.NewInt(daysPassedSinceMainnetLAunch))
	averageYearlyKcalSupplyGrowthRateSinceMainnetLaunch := new(big.Float).Mul(
		new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(averageDailyKcalClaimsSinceMainnetLaunch), big.NewFloat(365)),
			new(big.Float).SetInt(&kcalSupplyRawBigint),
		),
		big.NewFloat(100),
	)
	averageDailyKcalBurnSinceMainnetLaunch := new(big.Int).Div(&kcalBurnedSupplyRawBigint, big.NewInt(daysPassedSinceMainnetLAunch))
	averageYearlyKcalSupplyGrowthRateSinceMainnetLaunchFloat, _ := averageYearlyKcalSupplyGrowthRateSinceMainnetLaunch.Float64()

	// Kcal 2025 calculations
	// SOUL SUPPLY 124 129 197.66163057
	// KCAL 207 602 775.3940180069
	// BURNED	935 293.2819919151 2025 01 01 01:00 gmt+3( 1735682400 )
	kcalSupplyChangeSince2025 := new(big.Int).Sub(&kcalSupplyRawBigint, big.NewInt(2076027753940180069))
	burnedKcalSupply2025 := new(big.Int).Sub(&kcalBurnedSupplyRawBigint, big.NewInt(9352932819919151))
	daysPassedSince2025 := (currentTime - 1735682400) / 86400
	averageDailyKcalClaimSince2025 := new(big.Int).Div(new(big.Int).Add(kcalSupplyChangeSince2025, burnedKcalSupply2025), big.NewInt(daysPassedSince2025))
	averageDailyKcalBurnSince2025 := new(big.Int).Div(burnedKcalSupply2025, big.NewInt(daysPassedSince2025))

	netKcalDailySupplyChangeSince2025 := new(big.Int).Div(kcalSupplyChangeSince2025, big.NewInt(daysPassedSince2025))

	averageYearlyKcalSupplyGrowthRateSince2025 := new(big.Float).Mul(
		new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(netKcalDailySupplyChangeSince2025), big.NewFloat(365)),
			new(big.Float).SetInt(big.NewInt(2076027753940180069)),
		),
		big.NewFloat(100),
	)
	averageYearlyKcalSupplyGrowthRateSince2025Float, _ := averageYearlyKcalSupplyGrowthRateSince2025.Float64()

	// kcalSupplyAt241120241710 := new(big.Int).SetInt64(2004801680000000000)
	// averageDailyKcalGenerationSince231120241710 := new(big.Int).Mul(new(big.Int).Div(new(big.Int).Sub(&kcalSupplyRawBigint, kcalSupplyAt241120241710), big.NewInt((currentTime-1732371000)/3600)), big.NewInt(24))
	// averageDailyKcalGenerationSince231120241710Str := core.FormatBalance(averageDailyKcalGenerationSince231120241710, core.KcalDecimals)

	mainTokenInfo := fmt.Sprintf("Phantasma Stake\n \tCurrent Supply \t %s Soul\n\n Phantasma Energy\n \tCurrent Supply \t %s Kcal \n\tBurned Supply\t %s Kcal", core.FormatBalance(&soulSupplyRawBigint, core.SoulDecimals), core.FormatBalance(&kcalSupplyRawBigint, core.KcalDecimals), core.FormatBalance(&kcalBurnedSupplyRawBigint, core.KcalDecimals))
	sparkGenRate := math.Pow10(core.KcalDecimals) * core.KcalProdRate
	stakingInfo := fmt.Sprintf("Soul Master Reward\t%.4f\tSoul\nSpark gen per Soul:\t%.0f Spark (or %.4f Kcal)\nStaked Supply\t\t%s\tSoul\nStaking Ratio\t\t%.2f%%\nSoul Master APR\t\t%.2f%% (in Soul and generated Kcal not included)\nSoul Masters\t\t%v\nEligible Soul Masters\t%v\nCrown Supply\t\t%v\nNext Crown dist. date\t%v",
		core.LatestChainStatisticsData.EstSMReward, sparkGenRate, core.KcalProdRate, core.FormatBalance(core.LatestChainStatisticsData.TotalStakedSoul, core.SoulDecimals), stakingRatio, core.LatestChainStatisticsData.SMApr, core.LatestChainStatisticsData.TotalMaster, core.LatestChainStatisticsData.EligibleMaster, core.LatestTokenData.Token["CROWN"].CurrentSupply, time.Unix(core.LatestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"))

	kcalInfo := ""

	if core.UserSettings.NetworkName != "mainnet" {

		kcalInfo = fmt.Sprintf("Statistics since 2025\n\tonly for mainnet\n\nStatistics Since Mainnet Launch\n\tAverage Daily Burn\t%s Kcal\n\tAverage Daily Claims\t%s Kcal\n\tAnnual supply growth\t%.2f%%\n\nAbsolute maximum calculations\n\tFully boosted soul masters\t%.0f\n\tBoosted Kcal Generation\t%s Kcal\n\tBoosted Soul Supply Ratio\t%.2f%%\n\tAnnual supply growth\t\t%.2f%%",
			core.FormatBalance(averageDailyKcalBurnSinceMainnetLaunch, core.KcalDecimals), core.FormatBalance(averageDailyKcalClaimsSinceMainnetLaunch, core.KcalDecimals), averageYearlyKcalSupplyGrowthRateSinceMainnetLaunchFloat,
			maxFullyBoostedSoulmasterFloat64, core.FormatBalance(estimatedBoostedDailyKcalGenerationInt, core.KcalDecimals), boostedStakedSupplyRatio, annualKcalSupplyGrowthRateMaxFloat,
		)

	} else {
		kcalInfo = fmt.Sprintf("Statistics since 2025\n\tCirc. supply change\t%s Kcal\n\tBurned Suppy\t\t%s Kcal\n\tAverage Daily Burn\t%s Kcal\n\tAverage Daily Claims\t%s Kcal\n\tAnnual supply growth\t%.2f%%\n\nStatistics Since Mainnet Launch\n\tAverage Daily Burn\t%s Kcal\n\tAverage Daily Claims\t%s Kcal\n\tAnnual supply growth\t%.2f%%\n\nAbsolute maximum calculations\n\tFully boosted soul masters\t%.0f\n\tBoosted Kcal Generation\t%s Kcal\n\tBoosted Soul Supply Ratio\t%.2f%%\n\tAnnual supply growth\t\t%.2f%%",
			core.FormatBalance(kcalSupplyChangeSince2025, core.KcalDecimals), core.FormatBalance(burnedKcalSupply2025, core.KcalDecimals), core.FormatBalance(averageDailyKcalBurnSince2025, core.KcalDecimals), core.FormatBalance(averageDailyKcalClaimSince2025, core.KcalDecimals), averageYearlyKcalSupplyGrowthRateSince2025Float,
			core.FormatBalance(averageDailyKcalBurnSinceMainnetLaunch, core.KcalDecimals), core.FormatBalance(averageDailyKcalClaimsSinceMainnetLaunch, core.KcalDecimals), averageYearlyKcalSupplyGrowthRateSinceMainnetLaunchFloat,
			maxFullyBoostedSoulmasterFloat64, core.FormatBalance(estimatedBoostedDailyKcalGenerationInt, core.KcalDecimals), boostedStakedSupplyRatio, annualKcalSupplyGrowthRateMaxFloat,
		)

	}

	statistics := container.NewVBox(

		widget.NewLabelWithStyle("Main Tokens info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(mainTokenInfo),
		widget.NewLabelWithStyle("Staking Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(stakingInfo),
		widget.NewLabelWithStyle("Informations and Statistics for Kcal", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(kcalInfo),
	)

	bckBttn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { currentMainDialog.Hide() })
	chnStcsCntnt := container.NewVScroll(statistics)
	chnStcsLyt := container.NewBorder(nil, bckBttn, nil, nil, chnStcsCntnt)
	currentMainDialog = dialog.NewCustomWithoutButtons(fmt.Sprintf("Chain Statistics For %v", core.UserSettings.NetworkName), chnStcsLyt, mainWindowGui)
	currentMainDialog.Resize(fyne.NewSize(chnStcsCntnt.MinSize().Width+20, mainWindowGui.Canvas().Size().Height-50))
	closeUpdatingDialog()
	currentMainDialog.Show()

}
