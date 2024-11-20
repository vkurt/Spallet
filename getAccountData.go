package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

var latestAccountData = AccountInfoData{
	FungibleTokens: make(map[string]AccToken),
	NonFungible:    make(map[string]AccToken),
}

func getAccountData(walletAddress string, creds Credentials) error {
	currentUtcTime := time.Now().UTC()
	passedTime := currentUtcTime.Unix() - latestAccountData.StatCheckTime
	fetchAccData := false
	var err error
	fmt.Println("******Getting acc statistics*********")
	var actualTxCount int
	if latestAccountData.Address == walletAddress && latestAccountData.TransactionCount > 0 {
		actualTxCount = latestAccountData.TransactionCount
	}

	if passedTime > 9 {
		actualTxCount, err = client.GetAddressTransactionCount(walletAddress, userSettings.ChainName)
		if err != nil {
			return err
		}
	}

	fmt.Println("acctxcount ", latestAccountData.TransactionCount)
	fmt.Println("actualTxCount", actualTxCount)

	if latestAccountData.Address == walletAddress && actualTxCount != latestAccountData.TransactionCount { // trying to prevent fetch data too often
		fetchAccData = true
		latestAccountData = AccountInfoData{
			FungibleTokens: make(map[string]AccToken),
			NonFungible:    make(map[string]AccToken),
		}

	} else if latestAccountData.Address != walletAddress || latestAccountData.Network != userSettings.NetworkName {
		fetchAccData = true
		latestAccountData = AccountInfoData{
			FungibleTokens: make(map[string]AccToken),
			NonFungible:    make(map[string]AccToken),
		}
	}

	if fetchAccData {
		latestAccountData.StatCheckTime = currentUtcTime.Unix()
		latestAccountData.Address = walletAddress
		latestAccountData.TransactionCount = actualTxCount
		latestAccountData.Network = userSettings.NetworkName
		fmt.Println("******Refreshing data from chain for Account data*********")
		accountSummary = nil

		latestAccountData.StatCheckTime = currentUtcTime.Unix()
		latestAccountData.Address = walletAddress
		fmt.Printf("getting wallet balance info for: %v\n", walletAddress)
		account, err := client.GetAccount(walletAddress)
		if err != nil {
			return err
		}
		// buildAndShowTxes(walletAddress, 1, 10)
		latestAccountData.NftTypes = 0
		latestAccountData.TotalNft = 0
		latestAccountData.TokenCount = 0

		regularTokens = []fyne.CanvasObject{}
		nftTokens = []fyne.CanvasObject{}
		// for k := range accNftBalances {
		// 	delete(accNftBalances, k)
		// }

		// for k := range accTokenBalances {
		// 	delete(accTokenBalances, k)
		// }

		stakedSoulBalance := new(big.Int)
		stakedSoulBalance.SetString(account.Stakes.Amount, 10)

		if stakedSoulBalance.Cmp(soulMasterThreshold) >= 0 {
			latestAccountData.IsSoulMaster = true
			latestAccountData.IsStaker = true
		} else if stakedSoulBalance.Cmp(soulMasterThreshold) < 0 && stakedSoulBalance.Cmp(minSoulStake) >= 0 {
			latestAccountData.IsEligibleForCurrentSmReward = false
			latestAccountData.IsEligibleForCurrentCrown = false
			latestAccountData.IsSoulMaster = false
			latestAccountData.IsStaker = true
		} else {
			latestAccountData.IsSoulMaster = false
			latestAccountData.IsStaker = false
			latestAccountData.IsEligibleForCurrentCrown = false
			latestAccountData.IsEligibleForCurrentSmReward = false
		}
		fmt.Printf("staked soul: %v \nSoul master: %v\nSoulmaster threshold: %v \n", stakedSoulBalance.String(), latestAccountData.IsSoulMaster, soulMasterThreshold)

		unclaimedKcal := new(big.Int)
		unclaimedKcal.SetString(account.Stakes.Unclaimed, 10)
		latestAccountData.StakedBalances = Stake{
			Amount:    *stakedSoulBalance,
			Unclaimed: *unclaimedKcal,
			Time:      account.Stakes.Time,
		}

		crownAmount := 0
		for _, token := range account.Balances {
			tokenBalance := StringToBigInt(token.Amount)
			formattedBalance := formatBalance(tokenBalance, int(token.Decimals))
			var tokenBalanceBox *fyne.Container //*********
			amountBig := StringToBigInt(token.Amount)

			if len(token.Ids) == 0 {
				ftTokenData, _ := fetchUserTokensInfoFromChain(token.Symbol, 3)
				fungible := AccToken{
					Symbol:        token.Symbol,
					Name:          ftTokenData.Name,
					Decimals:      ftTokenData.Decimals,
					CurrentSupply: ftTokenData.CurrentSupply,
					MaxSupply:     ftTokenData.MaxSupply,
					BurnedSupply:  ftTokenData.BurnedSupply,
					Address:       ftTokenData.Address,
					Owner:         ftTokenData.Owner,
					Flags:         ftTokenData.Flags,
					Script:        ftTokenData.Script,
					Series:        ftTokenData.Series,
					External:      ftTokenData.External,
					Price:         ftTokenData.Price,
					Amount:        amountBig,
					Chain:         token.Chain,
					Ids:           token.Ids,
				}
				// fmt.Println("token", fungible.Symbol)
				// fmt.Println("tkontoken", token.Symbol)
				latestAccountData.TokenCount++
				tokenBalanceBox = createTokenBalance(token.Symbol, formattedBalance, len(token.Ids) > 0, creds, int(token.Decimals), fungible.Name)
				if latestAccountData.TokenCount == 1 {
					haveTokenContent := widget.NewLabelWithStyle("Ohhh, you’ve got a moon bag! But the million-dollar question is, 'Wen moon?' 🚀🌕", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
					haveTokenContent.Wrapping = fyne.TextWrapWord
					regularTokens = append(regularTokens, haveTokenContent)
				}
				regularTokens = append(regularTokens, tokenBalanceBox) //*******
				latestAccountData.FungibleTokens[token.Symbol] = fungible
			} else {
				nftTokenData, _ := fetchUserTokensInfoFromChain(token.Symbol, 3)
				nonFungible := AccToken{
					Symbol:        token.Symbol,
					Name:          nftTokenData.Name,
					Decimals:      nftTokenData.Decimals,
					CurrentSupply: nftTokenData.CurrentSupply,
					MaxSupply:     nftTokenData.CurrentSupply,
					BurnedSupply:  nftTokenData.BurnedSupply,
					Address:       nftTokenData.Address,
					Owner:         nftTokenData.Owner,
					Flags:         nftTokenData.Flags,
					Script:        nftTokenData.Script,
					Series:        nftTokenData.Series,
					External:      nftTokenData.External,
					Price:         nftTokenData.Price,
					Amount:        amountBig,
					Chain:         token.Chain,
					Ids:           token.Ids,
				}
				if token.Symbol == "CROWN" {
					crownAmount, _ = strconv.Atoi(token.Amount)

				}
				latestAccountData.NftTypes++
				if latestAccountData.NftTypes == 1 {
					haveNftContent := widget.NewLabelWithStyle("There is some Smart NFTs that could probably teach you a thing or two.\nLet’s hope they share their secrets with you and make you a billionaire! 💸🧠", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
					haveNftContent.Wrapping = fyne.TextWrapWord
					nftTokens = append(nftTokens, haveNftContent)
				}
				nftAmount, _ := strconv.Atoi(token.Amount)
				latestAccountData.TotalNft += int64(nftAmount)
				latestAccountData.NonFungible[token.Symbol] = nonFungible
				tokenBalanceBox = createTokenBalance(token.Symbol, formattedBalance, len(token.Ids) > 0, creds, int(token.Decimals), nonFungible.Name)
				nftTokens = append(nftTokens, tokenBalanceBox) //******
			}
		}

		if latestAccountData.NftTypes < 1 {
			noNFTContent := widget.NewLabelWithStyle("No Smart NFTs in your wallet? It’s like being a gamer without a high score! Time to level up and let the games begin. 🎮✨", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			noNFTContent.Wrapping = fyne.TextWrapWord
			nftTokens = append(nftTokens, noNFTContent)
		}
		if latestAccountData.TokenCount < 1 {
			noTokenContent := widget.NewLabelWithStyle("Your wallet is so empty, even the crypto memes are feeling sorry for you. \nNo shittokens to be found here!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			noTokenContent.Wrapping = fyne.TextWrapWord
			regularTokens = append(regularTokens, noTokenContent)
		}
		latestAccountData.KcalBoost = int16(crownAmount) * 5

		// **************************************************************************************************************************

		fmt.Println("getting account statistics for: " + walletAddress)
		sb := scriptbuilder.BeginScript()
		// sb2 := scriptbuilder.BeginScript()
		var encodedScript1 string
		// var encodedScript2 string
		var response1 response.ScriptResult
		// var response2 response.ScriptResult

		check := 0

		if latestAccountData.IsSoulMaster {
			sb.CallContract("stake", "GetStakeTimestamp", walletAddress)    //returns last stake timestamp
			sb.CallContract("stake", "GetTimeBeforeUnstake", walletAddress) //returns past time from last kcal generation
			sb.CallContract("stake", "GetMasterDate", walletAddress)        // used for crown eligibility, returns first date of being soulmaster
			sb.CallContract("account", "LookUpAddress", walletAddress)      // returns wallet's onchain name
			script1 := sb.EndScript()
			encodedScript1 = hex.EncodeToString(script1)

			// creating another script because server not returns more than 4 results

			// sb2.CallContract("gas", "GetLastInflationDate") //returns last inflation event timestamp
			// script2 := sb2.EndScript()
			// encodedScript2 = hex.EncodeToString(script2)

			check = 1

		} else if latestAccountData.IsStaker {
			sb.CallContract("stake", "GetStakeTimestamp", walletAddress)    //returns last stake timestamp
			sb.CallContract("stake", "GetTimeBeforeUnstake", walletAddress) //returns past time from last kcal generation
			sb.CallContract("account", "LookUpAddress", walletAddress)      // returns wallet's onchain name
			script := sb.EndScript()
			encodedScript1 = hex.EncodeToString(script)
			check = 2
		} else {
			check = 0
			fmt.Println("Adress is not a staker")
			latestAccountData.OnChainName = "anonymous"
		}

		if check >= 1 {
			checkResponse1, err := client.InvokeRawScript(userSettings.ChainName, encodedScript1)
			if err != nil {
				panic("Script1 invocation failed! Error: " + err.Error())
			}
			response1 = checkResponse1

			// checkResponse2, err := client.InvokeRawScript(chain, encodedScript2)
			// if err != nil {
			// 	panic("Script2 invocation failed! Error: " + err.Error())
			// }
			// response2 = checkResponse2
		}
		// } else if check == 2 {

		// 	checkResponse1, err := client.InvokeRawScript(chain, encodedScript1)
		// 	if err != nil {
		// 		panic("Script1 invocation failed! Error: " + err.Error())
		// 	}
		// 	response1 = checkResponse1

		// }

		if check == 1 {

			latestAccountData.LastStakeTimestamp = response1.DecodeResults(0).AsNumber().Int64()
			passedTimeAfterKcalGen := response1.DecodeResults(1).AsNumber().Int64()
			latestAccountData.SoulmasterSince = response1.DecodeResults(2).AsNumber().Int64()
			latestAccountData.OnChainName = response1.DecodeResults(3).AsString()

			// fmt.Printf("passedTimeAfterKcalGen %v \n", passedTimeAfterKcalGen)

			// fmt.Printf("accLastStakeTimestamp %v \n", accLastStakeTimestamp)

			timeBeforeUnstake := latestAccountData.LastStakeTimestamp + 86401

			// fmt.Printf("timeBeforeUnstake %v \n", timeBeforeUnstake)

			// fmt.Printf("currentUtcTime %v \n", currentUtcTime.Unix())

			// fmt.Printf("accIsStaker %v \n", accIsStaker)

			if currentUtcTime.Unix() >= timeBeforeUnstake && latestAccountData.IsStaker {
				latestAccountData.RemainedTimeForUnstake = 0
			} else {
				latestAccountData.RemainedTimeForUnstake = timeBeforeUnstake - currentUtcTime.Unix()
			}

			// fmt.Printf("accRemainedTimeForUnstake %v \n", accRemainedTimeForUnstake)

			latestAccountData.RemainedTimeForKcalGen = 86400 - passedTimeAfterKcalGen

			// fmt.Printf("accIsSoulMaster %v \n", accIsSoulMaster)
			// fmt.Printf("lastInflationTimeStamp %v \n", latestChainStatisticsData.LastInflationTimeStamp)
			// fmt.Printf("accSoulmasterSince %v \n", accSoulmasterSince)

			if latestAccountData.SoulmasterSince < latestChainStatisticsData.LastInflationTimeStamp && latestAccountData.IsSoulMaster {
				latestAccountData.IsEligibleForCurrentCrown = true
			} else {
				latestAccountData.IsEligibleForCurrentCrown = false
			}

			year, month, _ := currentUtcTime.Date()
			timeZone := currentUtcTime.Location()
			firstDayOfCurrentMonth := time.Date(year, month, 1, 0, 0, 0, 0, timeZone)

			if latestAccountData.SoulmasterSince < firstDayOfCurrentMonth.Unix() {
				latestAccountData.IsEligibleForCurrentSmReward = true
			} else {
				latestAccountData.IsEligibleForCurrentSmReward = false
			}

			latestAccountData.KcalDailyProd = *calculateKcalDailyProd(latestAccountData.KcalBoost, latestAccountData.StakedBalances.Amount, kcalProdRate)

		} else if check == 2 {

			latestAccountData.LastStakeTimestamp = response1.DecodeResults(0).AsNumber().Int64()
			passedTimeAfterKcalGen := response1.DecodeResults(1).AsNumber().Int64()
			latestAccountData.OnChainName = response1.DecodeResults(2).AsString()

			latestAccountData.KcalDailyProd = *calculateKcalDailyProd(latestAccountData.KcalBoost, latestAccountData.StakedBalances.Amount, kcalProdRate)

			latestAccountData.RemainedTimeForUnstake = latestAccountData.LastStakeTimestamp + 86401

			if currentUtcTime.Unix() >= latestAccountData.RemainedTimeForUnstake && latestAccountData.IsStaker {
				latestAccountData.RemainedTimeForUnstake = 0
			} else {
				latestAccountData.RemainedTimeForUnstake = latestAccountData.RemainedTimeForUnstake - currentUtcTime.Unix()
			}

			latestAccountData.RemainedTimeForKcalGen = 86400 - passedTimeAfterKcalGen

			fmt.Printf(" account Details :\n lastStakeTimeStamp %v\n TimeBeforeUnstake %v\n LookUpAddress %v\n", latestAccountData.LastStakeTimestamp, latestAccountData.RemainedTimeForUnstake, latestAccountData.OnChainName)

		}

		buildAndShowAccInfo(creds)

	}
	// saveLatestAccountData("accountdata", creds, latestAccountData)
	return nil
}
func calculateKcalDailyProd(accKcalBoost int16, stakedAmountKcalCalc big.Int, kcalProdRate float64) *big.Int {
	// Convert accKcalBoost to *big.Float and calculate boost factor
	boostFactor := new(big.Float).SetFloat64(float64(accKcalBoost) / 100.0)
	boostFactor = boostFactor.Add(boostFactor, big.NewFloat(1.0))

	// Decimal correction for stakedAmountKcalCalc: soul 8 decimals but kcal 10
	stakedAmountKcalCalcCorrected := new(big.Int).Mul(&stakedAmountKcalCalc, big.NewInt(100))

	// Convert stakedAmount to *big.Float
	stakedAmountFloat := new(big.Float).SetInt(stakedAmountKcalCalcCorrected)

	// Convert kcalProdRate to *big.Float
	kcalProdRateFloat := new(big.Float).SetFloat64(kcalProdRate)

	// Calculate daily production as *big.Float
	dailyProdFloat := new(big.Float).Mul(boostFactor, stakedAmountFloat)
	dailyProdFloat = dailyProdFloat.Mul(dailyProdFloat, kcalProdRateFloat)

	// Convert back to *big.Int
	dailyProd := new(big.Int)
	dailyProdFloat.Int(dailyProd) // Truncate to int

	return dailyProd
}

// func saveLatestAccountData(filename string, creds Credentials, accData AccountInfoData) error {
// 	data, err := json.Marshal(accData)
// 	if err != nil {
// 		return err
// 	}
// 	encryptedData, err := encrypt(data, creds.Password) // Use password for encryption
// 	if err != nil {
// 		return err
// 	}
// 	// fmt.Printf("Saving Encrypted Data: %s\n", encryptedData)
// 	return os.WriteFile(filename, []byte(encryptedData), 0600)
// }
// func loadAccountData(filename string, rawPassword string) (AccountInfoData, error) {
// 	encryptedData, err := os.ReadFile(filename)
// 	if err != nil {
// 		return AccountInfoData{}, err
// 	}
// 	decryptedData, err := decrypt(string(encryptedData), rawPassword)
// 	if err != nil {
// 		return AccountInfoData{}, err
// 	}
// 	var savedAccInfo AccountInfoData
// 	err = json.Unmarshal(decryptedData, &savedAccInfo)
// 	if err != nil {
// 		return AccountInfoData{}, err
// 	}
// 	return savedAccInfo, nil
// }

var animationRunning bool = false
var stopAnimation chan bool

func buildBadges() *fyne.Container {
	var enableAnimation bool = latestAccountData.IsStaker
	var mainBadgePath string
	var crownBadgePath string
	var soulMasterBadgePath string
	var stakingBadgePath string
	var networkBadgePath string
	var defaultAnSpeed = 25.0

	fmt.Println("******Building badges*****")

	switch latestAccountData.BadgeName {
	case "lord":
		mainBadgePath = "img/stats/lord.png"
	case "master":
		mainBadgePath = "img/stats/master.png"
	case "acolyte":
		mainBadgePath = "img/stats/acolyte.png"
	case "wanderer":
		mainBadgePath = "img/stats/wanderer.png"
	case "snoozer":
		mainBadgePath = "img/stats/snoozer.png"
	case "apprentice":
		mainBadgePath = "img/stats/apprentice.png"
	default:
		mainBadgePath = "img/stats/UNKNOWN.png"
	}

	if latestAccountData.KcalBoost > 0 {
		defaultAnSpeed = defaultAnSpeed / (float64(latestAccountData.KcalBoost)/100 + 1)
	} else {
		defaultAnSpeed = 25
	}

	mainBadge := canvas.NewImageFromResource(loadBadgeImageResource(mainBadgePath))
	mainBadge.FillMode = canvas.ImageFillContain
	mainBadge.SetMinSize(fyne.NewSize(150, 150))

	switch {
	case latestAccountData.IsSoulMaster && latestAccountData.IsEligibleForCurrentCrown && latestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN.png"
		soulMasterBadgePath = "img/stats/soul_master.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case latestAccountData.IsSoulMaster && latestAccountData.IsEligibleForCurrentCrown && !latestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN.png"
		soulMasterBadgePath = "img/stats/soul_master_en.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case latestAccountData.IsSoulMaster && !latestAccountData.IsEligibleForCurrentCrown && !latestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN_en.png"
		soulMasterBadgePath = "img/stats/soul_master_en.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case latestAccountData.IsSoulMaster && !latestAccountData.IsEligibleForCurrentCrown && latestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN_en.png"
		soulMasterBadgePath = "img/stats/soul_master.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case latestAccountData.IsStaker:
		crownBadgePath = "img/stats/CROWN_ne.png"
		soulMasterBadgePath = "img/stats/soul_master_ne.png"
		stakingBadgePath = "img/stats/Kcal.png"
	default:
		crownBadgePath = "img/stats/CROWN_ne.png"
		soulMasterBadgePath = "img/stats/soul_master_ne.png"
		stakingBadgePath = "img/stats/Kcal_ns.png"
	}

	if userSettings.NetworkName == "mainnet" {
		networkBadgePath = "img/stats/mainnet.png"
	} else {
		networkBadgePath = "img/stats/testnet.png"
	}

	networkBadge := canvas.NewImageFromResource(loadBadgeImageResource(networkBadgePath))
	networkBadge.FillMode = canvas.ImageFillContain
	networkBadge.SetMinSize(fyne.NewSize(26, 26))

	crownBadge := canvas.NewImageFromResource(loadBadgeImageResource(crownBadgePath))
	crownBadge.FillMode = canvas.ImageFillContain
	crownBadge.SetMinSize(fyne.NewSize(26, 26))

	soulMasterBadge := canvas.NewImageFromResource(loadBadgeImageResource(soulMasterBadgePath))
	soulMasterBadge.FillMode = canvas.ImageFillContain
	soulMasterBadge.SetMinSize(fyne.NewSize(26, 26))

	stakingBadge := canvas.NewImageFromResource(loadBadgeImageResource(stakingBadgePath))
	stakingBadge.FillMode = canvas.ImageFillContain
	stakingBadge.SetMinSize(fyne.NewSize(26, 26))

	emptyArea := canvas.NewImageFromResource(loadBadgeImageResource("img/stats/spacer.png"))
	emptyArea.FillMode = canvas.ImageFillContain
	emptyArea.SetMinSize(fyne.NewSize(7, 7))

	imageContainer := container.NewBorder(
		nil, nil, nil,
		container.NewVBox(networkBadge, emptyArea, emptyArea, crownBadge, soulMasterBadge, stakingBadge, emptyArea),
		mainBadge,
	)

	stopBadgeAnimation()

	if enableAnimation && !animationRunning {
		animationRunning = true
		stopAnimation = make(chan bool)
		fmt.Println("Animation started")
		go func() {
			var scale float32 = 1.0
			var increment float32 = 0.01
			anSpeed := time.Duration(defaultAnSpeed) * time.Millisecond

			fmt.Println("anSpeed", anSpeed)
			for {
				select {
				case <-stopAnimation:
					animationRunning = false
					return
				case <-time.Tick(anSpeed):
					if scale >= 1.1 || scale <= 0.9 {
						increment = -increment
					}
					scale += increment
					stakingBadge.Resize(fyne.NewSize(26*scale, 26*scale))
					stakingBadge.Refresh()
				}
			}
		}()
	} else if !enableAnimation && animationRunning {
		fmt.Println("Animation stopped")
		stopBadgeAnimation()
	}

	return imageContainer
}

func stopBadgeAnimation() {
	if animationRunning {
		stopAnimation <- true
		close(stopAnimation)
		animationRunning = false
	}
}
