package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

var soulMasterRewardPool = 125000

type ChainStatisticsData struct {
	SoulData                 TokenData
	KcalData                 TokenData
	CrownData                TokenData
	TotalStakedSoul          *big.Int
	EstSMReward              float64
	SMApr                    float64
	RemainedTimeForCrown     int64
	TotalMaster              int16
	EligibleMaster           int16
	LastInflationTimeStamp   int64
	LastMasterClaimTimestamp int64
	NextInfTimeStamp         int64
	DataFetchTime            int64
}

var latestChainStatisticsData ChainStatisticsData

func getChainStatistics() error {
	currentTime := time.Now().Unix()
	passedTimeFromLastFetch := currentTime - latestChainStatisticsData.DataFetchTime

	if passedTimeFromLastFetch > 9 {
		fmt.Println("******Refreshing data from chain for chain stats*********")
		latestChainStatisticsData.DataFetchTime = currentTime
		fetchChainMainTokensFromChain()

		// getting last Master claim date
		sbGetLastMasterClaim := scriptbuilder.BeginScript()
		sbGetLastMasterClaim.CallContract("stake", "GetLastMasterClaim")
		script := sbGetLastMasterClaim.EndScript()
		encodedScript := hex.EncodeToString(script)
		response, err := client.InvokeRawScript(chain, encodedScript)
		latestChainStatisticsData.LastMasterClaimTimestamp = response.DecodeResult().AsNumber().Int64()
		if err != nil {
			return fmt.Errorf("error fetching last Master claim date: %v", err)
		}
		// ***

		// getting total staked soul
		account, err := client.GetAccount("S3dP2jjf1jUG9nethZBWbnu9a6dFqB7KveTWU7znis6jpDy")
		if err != nil {
			return fmt.Errorf("getting total staked soul: %v", err)
		}
		totalStakedSoulBig := StringToBigInt(account.Balances[0].Amount)
		latestChainStatisticsData.TotalStakedSoul = &totalStakedSoulBig
		fmt.Println("totalStakedSoul:", latestChainStatisticsData.TotalStakedSoul)
		// ***

		// getting other data from chain with callcontract
		sb := scriptbuilder.BeginScript()
		sb.CallContract("gas", "GetLastInflationDate") // getting last inflation Timestamp
		sb.CallContract("stake", "GetMasterCount")     // getting total Soul Master count
		nextClaimDate := latestChainStatisticsData.LastMasterClaimTimestamp + 32*86400
		sb.CallContract("stake", "GetClaimMasterCount", time.Unix(nextClaimDate, 0)) // getting eligible SoulMaster count for next claim
		script = sb.EndScript()
		encodedScript = hex.EncodeToString(script)
		response, err = client.InvokeRawScript(chain, encodedScript)

		latestChainStatisticsData.LastInflationTimeStamp = response.DecodeResults(0).AsNumber().Int64()

		latestChainStatisticsData.TotalMaster = int16(response.DecodeResults(1).AsNumber().Int64())

		latestChainStatisticsData.EligibleMaster = int16(response.DecodeResults(2).AsNumber().Int64())

		latestChainStatisticsData.EstSMReward = float64(soulMasterRewardPool) / float64(latestChainStatisticsData.EligibleMaster)

		latestChainStatisticsData.NextInfTimeStamp = latestChainStatisticsData.LastInflationTimeStamp + 90*86400

		if latestChainStatisticsData.NextInfTimeStamp > int64(currentTime) {
			latestChainStatisticsData.RemainedTimeForCrown = latestChainStatisticsData.NextInfTimeStamp - int64(currentTime)
		} else {
			latestChainStatisticsData.RemainedTimeForCrown = 0
		}

		// Adjust for 8 decimals: divide by 10^8
		divisor := new(big.Float).SetInt(big.NewInt(100000000))
		thresholdFloat := new(big.Float).SetInt(soulMasterThreshold)
		thresholdFloat.Quo(thresholdFloat, divisor)

		// Calculate (estimatedSoulmasterReward * 12 * 100)
		rewardFloat := new(big.Float).SetFloat64(latestChainStatisticsData.EstSMReward)
		rewardFloat.Mul(rewardFloat, big.NewFloat(12))
		rewardFloat.Mul(rewardFloat, big.NewFloat(100))

		// Divide by adjusted soulMasterThreshold
		soulMasterAprBig := new(big.Float).Quo(rewardFloat, thresholdFloat)

		// Convert result to float64
		soulMasterAprFloat, _ := soulMasterAprBig.Float64()
		latestChainStatisticsData.SMApr = soulMasterAprFloat
		if err != nil {
			return fmt.Errorf("error fetching chain statistics: %v", err)
		}
		// saveLatestChainData("chainstats", latestChainStatisticsData)
		return nil

	}

	fmt.Println("*****Chain Stats*****")
	fmt.Println("eligible Soulmasters for next reward:", latestChainStatisticsData.EligibleMaster)
	fmt.Println("estimated Soulmaster Reward:", latestChainStatisticsData.EstSMReward)
	fmt.Println("last Inflation Time:", time.Unix(latestChainStatisticsData.LastInflationTimeStamp, 0).Format("02-01-2006 15:04"))
	fmt.Println("last Soul Master reward Distribution:", time.Unix(latestChainStatisticsData.LastMasterClaimTimestamp, 0).Format("02-01-2006 15:04"))
	fmt.Println("totalMasterCount:", latestChainStatisticsData.TotalMaster)
	fmt.Println("remained time for crown", time.Duration(latestChainStatisticsData.RemainedTimeForCrown))
	fmt.Println("next inflation date: ", time.Unix(latestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"))

	return nil
}

// func saveLatestChainData(filename string, chainData ChainStatisticsData) error {
// 	data, err := json.Marshal(chainData)
// 	if err != nil {
// 		return err
// 	}

// 	// fmt.Printf("Saving Chain Data: %s\n", &data)
// 	return os.WriteFile(filename, []byte(data), 0600)
// }

// func loadChainData(filename string) (ChainStatisticsData, error) {
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		return ChainStatisticsData{}, err
// 	}

// 	var savedChainData ChainStatisticsData
// 	err = json.Unmarshal(data, &savedChainData)
// 	if err != nil {
// 		return ChainStatisticsData{}, err
// 	}
// 	return savedChainData, nil
// }
