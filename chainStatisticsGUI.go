package main

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func buildAndShowChainStatistics(walletDetails *fyne.Container) {

	soulSupplyRawBigint := StringToBigInt(latestChainStatisticsData.SoulData.CurrentSupply)
	kcalSupplyRawBigint := StringToBigInt(latestChainStatisticsData.KcalData.CurrentSupply)
	kcalBurnedSupplyRawBigint := StringToBigInt(latestChainStatisticsData.KcalData.BurnedSupply)
	stakingRatio, _ := new(big.Rat).SetFrac(latestChainStatisticsData.TotalStakedSoul, &soulSupplyRawBigint).Float64()
	stakingRatio *= 100.0

	crownSupply := StringToBigInt(latestChainStatisticsData.CrownData.CurrentSupply)

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

	mainTokenInfo := fmt.Sprintf("Phantasma Stake\n \tCurrent Supply \t %s Soul\n\n Phantasma Energy\n \tCurrent Supply \t %s Kcal \n\tBurned Supply\t %s Kcal", formatBalance(soulSupplyRawBigint, soulDecimals), formatBalance(kcalSupplyRawBigint, kcalDecimals), formatBalance(kcalBurnedSupplyRawBigint, kcalDecimals))
	sparkGenRate := math.Pow10(kcalDecimals) * kcalProdRate
	stakingInfo := fmt.Sprintf("Soul Master Reward\t%.4f\tSoul\nSpark gen per Soul:\t%.0f Spark (or %.4f Kcal)\nStaked Supply\t\t%s\tSoul\nStaking Ratio\t\t%.2f%%\nSoul Master APR\t\t%.2f%% (in Soul and generated Kcal not included)\nSoul Masters\t\t%v\nEligible Soul Masters\t%v\nCrown Supply\t\t%v\nNext Crown dist. date\t%v\n\n***Below Data Calculated Based on Stake amount in stake contract and Crown Supply***\nFully boosted Soul Masters\t %.0f\nBoosted daily Kcal generation\t %v Kcal\nBoosted Staked Supply Ratio\t %.2f%% \nAnn. Kcal Supply Growth Rate\t %.2f%%",
		latestChainStatisticsData.EstSMReward, sparkGenRate, kcalProdRate, formatBalance(*latestChainStatisticsData.TotalStakedSoul, soulDecimals), stakingRatio, latestChainStatisticsData.SMApr, latestChainStatisticsData.TotalMaster, latestChainStatisticsData.EligibleMaster, latestChainStatisticsData.CrownData.CurrentSupply, time.Unix(latestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"), maxFullyBoostedSoulmasterFloat64, formatBalance(*estimatedBoostedDailyKcalGenerationInt, kcalDecimals), boostedStakedSupplyRatio, annualKcalSupplyGrowthRate)

	statistics := container.NewVBox(
		widget.NewLabelWithStyle("Chain Statistics For "+network, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Main Tokens info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(mainTokenInfo),
		widget.NewLabelWithStyle("Staking Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(stakingInfo),
	)

	walletDetails.Objects = []fyne.CanvasObject{statistics}
	walletDetails.Refresh()
}
