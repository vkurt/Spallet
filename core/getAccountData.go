package core

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"

	"strconv"
	"time"

	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
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
	KcalDailyProd                *big.Int
	RemainedTimeForUnstake       int64
	LastStakeTimestamp           int64
	SoulmasterSince              int64
	BadgeName                    string
	NickName                     string
	StatCheckTime                int64
	TransactionCount             int
	Network                      string
	IsBalanceUpdated             bool
	SortedTokenList              []string
	SortedNftList                []string
}

var LatestAccountData = AccountInfoData{
	FungibleTokens: make(map[string]AccToken),
	NonFungible:    make(map[string]AccToken),
}
var lastCountdownUpdate int64

// sorts tokens in alphabetical order, based on their names, Soul and Kcal will be always on top
func sortTokensAlhabetical(slice []string) []string {
	// Check if "Soul" and "Kcal" are in the slice
	var containsSoul, containsKcal, containsCrown bool
	filteredSlice := []string{}
	for _, v := range slice {
		if v == "SOUL" {
			containsSoul = true
		}
		if v == "KCAL" {
			containsKcal = true
		}

		if v == "CROWN" {
			containsCrown = true
		}

		if v != "SOUL" && v != "KCAL" && v != "CROWN" {

			filteredSlice = append(filteredSlice, v)
		}
	}

	// Sort the remaining elements alphabetically
	slices.Sort(filteredSlice)

	// Construct the final slice with "Soul" and "Kcal" on top if they are present
	result := []string{}
	if containsSoul {
		result = append(result, "SOUL")
	}
	if containsKcal {
		result = append(result, "KCAL")
	}
	if containsCrown {
		result = append(result, "CROWN")
	}
	result = append(result, filteredSlice...)

	return result
}

func GetAccountData(walletAddress string, creds Credentials, rootPath string) error {
	currentUtcTime := time.Now().UTC()
	passedTime := currentUtcTime.Unix() - LatestAccountData.StatCheckTime
	fetchAccData := false
	var err error
	fmt.Println("******Getting acc statistics*********")
	var actualTxCount int

	if LatestAccountData.Address == walletAddress && LatestAccountData.TransactionCount > 0 {
		actualTxCount = LatestAccountData.TransactionCount
	}

	if passedTime >= 10 {
		actualTxCount, err = Client.GetAddressTransactionCount(walletAddress, UserSettings.ChainName)
		if err != nil {
			return err
		}
	}

	fmt.Println("acctxcount ", LatestAccountData.TransactionCount)
	fmt.Println("actualTxCount", actualTxCount)

	if LatestAccountData.Address == walletAddress && actualTxCount != LatestAccountData.TransactionCount && LatestAccountData.Network == UserSettings.NetworkName { // fetch acount data tx count changed, network changed, on wallet change
		fetchAccData = true
		LatestAccountData = AccountInfoData{
			FungibleTokens: make(map[string]AccToken),
			NonFungible:    make(map[string]AccToken),
		}

	} else if LatestAccountData.Network != UserSettings.NetworkName {
		fetchAccData = true

		LatestAccountData = AccountInfoData{
			FungibleTokens: make(map[string]AccToken),
			NonFungible:    make(map[string]AccToken),
		}
		LatestTokenData = UpdatedTokenData{
			AllTokenUpdateTime:   currentUtcTime.Unix() - 135,
			ChainTokenUpdateTime: currentUtcTime.Unix() - 135,
			AccTokenUpdateTime:   currentUtcTime.Unix(),
			Token:                make(map[string]TokenData),
		}
		LoadTokenCache(rootPath)
	} else if LatestAccountData.Address != walletAddress {
		fetchAccData = true
		LatestAccountData = AccountInfoData{
			FungibleTokens: make(map[string]AccToken),
			NonFungible:    make(map[string]AccToken),
		}
	}

	if LatestChainStatisticsData.DataFetchTime < currentUtcTime.Unix() {
		passed := currentUtcTime.Unix() - lastCountdownUpdate
		if passed < LatestChainStatisticsData.RemainedTimeForCrown {
			LatestChainStatisticsData.RemainedTimeForCrown -= passed

		} else {
			LatestChainStatisticsData.RemainedTimeForCrown = 0

		}
	}

	if LatestAccountData.StatCheckTime < currentUtcTime.Unix() && !fetchAccData && LatestAccountData.IsStaker { // trying to update staking countdowns without updating it from chain
		passed := currentUtcTime.Unix() - lastCountdownUpdate
		if passed < LatestAccountData.RemainedTimeForKcalGen {
			LatestAccountData.RemainedTimeForKcalGen -= passed

		} else if LatestAccountData.IsStaker {
			LatestAccountData.RemainedTimeForKcalGen = 86400 + LatestAccountData.RemainedTimeForKcalGen - passed
			LatestAccountData.StakedBalances.Unclaimed = new(big.Int).Add(LatestAccountData.KcalDailyProd, LatestAccountData.StakedBalances.Unclaimed)

		}

		if passed < LatestAccountData.RemainedTimeForUnstake {
			LatestAccountData.RemainedTimeForUnstake -= passed

		} else {
			LatestAccountData.RemainedTimeForUnstake = 0

		}
		lastCountdownUpdate = currentUtcTime.Unix()
		// buildAndShowAccInfo(creds)
		// showStakingPage(creds)

	}

	if fetchAccData {
		LatestAccountData.IsBalanceUpdated = true
		lastCountdownUpdate = currentUtcTime.Unix()
		LatestAccountData.StatCheckTime = currentUtcTime.Unix()
		LatestAccountData.Address = walletAddress
		LatestAccountData.TransactionCount = actualTxCount
		LatestAccountData.Network = UserSettings.NetworkName
		fmt.Println("******Refreshing data from chain for Account data*********")

		LatestAccountData.StatCheckTime = currentUtcTime.Unix()
		LatestAccountData.Address = walletAddress
		fmt.Printf("getting wallet balance info for: %v\n", walletAddress)
		account, err := Client.GetAccount(walletAddress)
		if err != nil {
			return err
		}
		// buildAndShowTxes(walletAddress, 1, 10)
		LatestAccountData.NftTypes = 0
		LatestAccountData.TotalNft = 0
		LatestAccountData.TokenCount = 0

		// for k := range accNftBalances {
		// 	delete(accNftBalances, k)
		// }

		// for k := range accTokenBalances {
		// 	delete(accTokenBalances, k)
		// }

		stakedSoulBalance := big.NewInt(0)
		stakedSoulBalance.SetString(account.Stakes.Amount, 10)
		if stakedSoulBalance == nil {
			stakedSoulBalance = big.NewInt(0)
		}

		if stakedSoulBalance.Cmp(SoulMasterThreshold) >= 0 {
			LatestAccountData.IsSoulMaster = true
			LatestAccountData.IsStaker = true
		} else if stakedSoulBalance.Cmp(SoulMasterThreshold) < 0 && stakedSoulBalance.Cmp(MinSoulStake) >= 0 {
			LatestAccountData.IsEligibleForCurrentSmReward = false
			LatestAccountData.IsEligibleForCurrentCrown = false
			LatestAccountData.IsSoulMaster = false
			LatestAccountData.IsStaker = true
		} else {
			LatestAccountData.IsSoulMaster = false
			LatestAccountData.IsStaker = false
			LatestAccountData.IsEligibleForCurrentCrown = false
			LatestAccountData.IsEligibleForCurrentSmReward = false
		}
		fmt.Printf("staked soul: %v \nSoul master: %v\nSoulmaster threshold: %v \n", stakedSoulBalance.String(), LatestAccountData.IsSoulMaster, SoulMasterThreshold)

		unclaimedKcal := big.NewInt(0)
		unclaimedKcal.SetString(account.Stakes.Unclaimed, 10)
		if stakedSoulBalance == nil {
			stakedSoulBalance = big.NewInt(0)
		}
		if unclaimedKcal == nil {
			unclaimedKcal = big.NewInt(0)
		}
		LatestAccountData.StakedBalances = Stake{
			Amount:    stakedSoulBalance,
			Unclaimed: unclaimedKcal,
			Time:      account.Stakes.Time,
		}

		crownAmount := 0
		for _, token := range account.Balances {

			amountBig := StringToBigInt(token.Amount)

			if len(token.Ids) == 0 {
				ftTokenData, _ := UpdateOrCheckTokenCache(token.Symbol, 3, "check", rootPath)
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

					Series:   ftTokenData.Series,
					External: ftTokenData.External,
					Price:    ftTokenData.Price,
					Amount:   &amountBig,
					Chain:    token.Chain,
					Ids:      token.Ids,
				}
				// fmt.Println("token", fungible.Symbol)
				// fmt.Println("tkontoken", token.Symbol)
				LatestAccountData.TokenCount++
				LatestAccountData.SortedTokenList = append(LatestAccountData.SortedTokenList, ftTokenData.Symbol)
				LatestAccountData.FungibleTokens[token.Symbol] = fungible
			} else {
				nftTokenData, _ := UpdateOrCheckTokenCache(token.Symbol, 3, "check", rootPath)
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

					Series:   nftTokenData.Series,
					External: nftTokenData.External,
					Price:    nftTokenData.Price,
					Amount:   &amountBig,
					Chain:    token.Chain,
					Ids:      token.Ids,
				}
				if token.Symbol == "Phantasma Crown" {
					crownAmount, _ = strconv.Atoi(token.Amount)

				}
				LatestAccountData.NftTypes++
				LatestAccountData.SortedNftList = append(LatestAccountData.SortedNftList, nftTokenData.Symbol)
				nftAmount, _ := strconv.Atoi(token.Amount)
				LatestAccountData.TotalNft += int64(nftAmount)
				LatestAccountData.NonFungible[token.Symbol] = nonFungible

			}
		}
		if LatestAccountData.TokenCount > 1 {
			LatestAccountData.SortedTokenList = sortTokensAlhabetical(LatestAccountData.SortedTokenList)
		}
		if LatestAccountData.NftTypes > 1 {
			LatestAccountData.SortedNftList = sortTokensAlhabetical(LatestAccountData.SortedNftList)
		}
		// tokenScrollgtry.Objects = []fyne.CanvasObject{tokenScrollg}
		LatestAccountData.KcalBoost = int16(crownAmount) * 5

		// **************************************************************************************************************************

		fmt.Println("getting account statistics for: " + walletAddress)
		sb := scriptbuilder.BeginScript()
		// sb2 := scriptbuilder.BeginScript()
		var encodedScript1 string
		// var encodedScript2 string
		var response1 response.ScriptResult
		// var response2 response.ScriptResult

		check := 0

		if LatestAccountData.IsSoulMaster {
			sb.CallContract("stake", "GetStakeTimestamp", walletAddress)    //returns last stake timestamp
			sb.CallContract("stake", "GetTimeBeforeUnstake", walletAddress) //returns past time from last kcal generation
			sb.CallContract("stake", "GetMasterDate", walletAddress)        // used for Phantasma Crown eligibility, returns first date of being soulmaster
			sb.CallContract("account", "LookUpAddress", walletAddress)      // returns wallet's onchain name
			script1 := sb.EndScript()
			encodedScript1 = hex.EncodeToString(script1)

			// creating another script because server not returns more than 4 results

			// sb2.CallContract("gas", "GetLastInflationDate") //returns last inflation event timestamp
			// script2 := sb2.EndScript()
			// encodedScript2 = hex.EncodeToString(script2)

			check = 1

		} else if LatestAccountData.IsStaker {
			sb.CallContract("stake", "GetStakeTimestamp", walletAddress)    //returns last stake timestamp
			sb.CallContract("stake", "GetTimeBeforeUnstake", walletAddress) //returns past time from last kcal generation
			sb.CallContract("account", "LookUpAddress", walletAddress)      // returns wallet's onchain name
			script := sb.EndScript()
			encodedScript1 = hex.EncodeToString(script)
			check = 2
		} else {
			check = 0
			fmt.Println("Adress is not a staker")
			LatestAccountData.OnChainName = "anonymous"
		}

		if check >= 1 {
			checkResponse1, err := Client.InvokeRawScript(UserSettings.ChainName, encodedScript1)
			if err != nil {
				panic("Script1 invocation failed! Error: " + err.Error())
			}
			response1 = checkResponse1

			// checkResponse2, err := Client.InvokeRawScript(chain, encodedScript2)
			// if err != nil {
			// 	panic("Script2 invocation failed! Error: " + err.Error())
			// }
			// response2 = checkResponse2
		}
		// } else if check == 2 {

		// 	checkResponse1, err := Client.InvokeRawScript(chain, encodedScript1)
		// 	if err != nil {
		// 		panic("Script1 invocation failed! Error: " + err.Error())
		// 	}
		// 	response1 = checkResponse1

		// }

		if check == 1 {

			LatestAccountData.LastStakeTimestamp = response1.DecodeResults(0).AsNumber().Int64()
			passedTimeAfterKcalGen := response1.DecodeResults(1).AsNumber().Int64()
			LatestAccountData.SoulmasterSince = response1.DecodeResults(2).AsNumber().Int64()
			LatestAccountData.OnChainName = response1.DecodeResults(3).AsString()

			// fmt.Printf("passedTimeAfterKcalGen %v \n", passedTimeAfterKcalGen)

			// fmt.Printf("accLastStakeTimestamp %v \n", accLastStakeTimestamp)

			timeBeforeUnstake := LatestAccountData.LastStakeTimestamp + 86401

			// fmt.Printf("timeBeforeUnstake %v \n", timeBeforeUnstake)

			// fmt.Printf("currentUtcTime %v \n", currentUtcTime.Unix())

			// fmt.Printf("accIsStaker %v \n", accIsStaker)

			if currentUtcTime.Unix() >= timeBeforeUnstake && LatestAccountData.IsStaker {
				LatestAccountData.RemainedTimeForUnstake = 0
			} else {
				LatestAccountData.RemainedTimeForUnstake = timeBeforeUnstake - currentUtcTime.Unix()
			}

			// fmt.Printf("accRemainedTimeForUnstake %v \n", accRemainedTimeForUnstake)

			LatestAccountData.RemainedTimeForKcalGen = 86400 - passedTimeAfterKcalGen

			// fmt.Printf("accIsSoulMaster %v \n", accIsSoulMaster)
			// fmt.Printf("lastInflationTimeStamp %v \n", LatestChainStatisticsData.LastInflationTimeStamp)
			// fmt.Printf("accSoulmasterSince %v \n", accSoulmasterSince)

			if LatestAccountData.SoulmasterSince < LatestChainStatisticsData.LastInflationTimeStamp && LatestAccountData.IsSoulMaster {
				LatestAccountData.IsEligibleForCurrentCrown = true
			} else {
				LatestAccountData.IsEligibleForCurrentCrown = false
			}

			year, month, _ := currentUtcTime.Date()
			timeZone := currentUtcTime.Location()
			firstDayOfCurrentMonth := time.Date(year, month, 1, 0, 0, 0, 0, timeZone)

			if LatestAccountData.SoulmasterSince < firstDayOfCurrentMonth.Unix() {
				LatestAccountData.IsEligibleForCurrentSmReward = true
			} else {
				LatestAccountData.IsEligibleForCurrentSmReward = false
			}

			LatestAccountData.KcalDailyProd = CalculateKcalDailyProd(LatestAccountData.KcalBoost, LatestAccountData.StakedBalances.Amount, KcalProdRate)

		} else if check == 2 {

			LatestAccountData.LastStakeTimestamp = response1.DecodeResults(0).AsNumber().Int64()
			passedTimeAfterKcalGen := response1.DecodeResults(1).AsNumber().Int64()
			LatestAccountData.OnChainName = response1.DecodeResults(2).AsString()

			LatestAccountData.KcalDailyProd = CalculateKcalDailyProd(LatestAccountData.KcalBoost, LatestAccountData.StakedBalances.Amount, KcalProdRate)

			LatestAccountData.RemainedTimeForUnstake = LatestAccountData.LastStakeTimestamp + 86401

			if currentUtcTime.Unix() >= LatestAccountData.RemainedTimeForUnstake && LatestAccountData.IsStaker {
				LatestAccountData.RemainedTimeForUnstake = 0
			} else {
				LatestAccountData.RemainedTimeForUnstake = LatestAccountData.RemainedTimeForUnstake - currentUtcTime.Unix()
			}

			LatestAccountData.RemainedTimeForKcalGen = 86400 - passedTimeAfterKcalGen

			fmt.Printf(" account Details :\n lastStakeTimeStamp %v\n TimeBeforeUnstake %v\n LookUpAddress %v\n", LatestAccountData.LastStakeTimestamp, LatestAccountData.RemainedTimeForUnstake, LatestAccountData.OnChainName)

		}

		fmt.Println("accKcalBoost accIsSoulMaster ", LatestAccountData.KcalBoost, LatestAccountData.IsSoulMaster)

		accSoulAmount := LatestAccountData.FungibleTokens["SOUL"].Amount

		if accSoulAmount == nil {
			accSoulAmount = big.NewInt(0)
		}
		if LatestAccountData.KcalBoost == 100 && LatestAccountData.IsSoulMaster {
			LatestAccountData.BadgeName = "lord"
			LatestAccountData.NickName = "Spark Lord 🔥"
		} else if LatestAccountData.KcalBoost > 0 && LatestAccountData.IsSoulMaster {
			LatestAccountData.BadgeName = "master"
			LatestAccountData.NickName = "Spark Master 💥"

		} else if LatestAccountData.IsSoulMaster {
			LatestAccountData.BadgeName = "apprentice"
			LatestAccountData.NickName = "Spark Apprentice ✨"

		} else if accSoulAmount.Cmp(big.NewInt(100000000)) > 0 && LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) >= 0 {
			LatestAccountData.BadgeName = "snoozer"
			LatestAccountData.NickName = "Soul slacker 😴"

		} else if LatestAccountData.IsStaker {
			LatestAccountData.BadgeName = "acolyte"
			LatestAccountData.NickName = "Spark Acolyte ⚡️"

		} else if accSoulAmount.Cmp(big.NewInt(100000000)) >= 0 && LatestAccountData.OnChainName == "anonymous" {
			LatestAccountData.BadgeName = "snoozer"
			LatestAccountData.NickName = "Soul snoozer💤"

		} else if LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(100000000)) < 0 && accSoulAmount.Cmp(big.NewInt(100000000)) < 0 {
			LatestAccountData.BadgeName = "wanderer"
			LatestAccountData.NickName = "Soulless wanderer 🌑"

		}

	}
	// saveLatestAccountData("accountdata", creds, LatestAccountData)
	return nil
}
func CalculateKcalDailyProd(accKcalBoost int16, stakedAmountKcalCalc *big.Int, KcalProdRate float64) *big.Int {
	// Convert accKcalBoost to *big.Float and calculate boost factor
	boostFactor := new(big.Float).SetFloat64(float64(accKcalBoost) / 100.0)
	boostFactor = boostFactor.Add(boostFactor, big.NewFloat(1.0))

	// Decimal correction for stakedAmountKcalCalc: soul 8 decimals but kcal 10
	stakedAmountKcalCalcCorrected := new(big.Int).Mul(stakedAmountKcalCalc, big.NewInt(100))

	// Convert stakedAmount to *big.Float
	stakedAmountFloat := new(big.Float).SetInt(stakedAmountKcalCalcCorrected)

	// Convert KcalProdRate to *big.Float
	kcalProdRateFloat := new(big.Float).SetFloat64(KcalProdRate)

	// Calculate daily production as *big.Float
	dailyProdFloat := new(big.Float).Mul(boostFactor, stakedAmountFloat)
	dailyProdFloat = dailyProdFloat.Mul(dailyProdFloat, kcalProdRateFloat)

	// Convert back to *big.Int
	dailyProd := new(big.Int)
	dailyProdFloat.Int(dailyProd) // Truncate to int

	return dailyProd
}
