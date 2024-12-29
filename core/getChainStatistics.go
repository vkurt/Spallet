package core

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"time"

	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

const soulMasterRewardPool = 125000

var SoulMasterThreshold = big.NewInt(5000000000000)
var MinSoulStake = big.NewInt(100000000)
var KcalProdRate = 0.002

type ChainStatisticsData struct {
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

var LatestChainStatisticsData ChainStatisticsData

func GetChainStatistics() error {
	currentTime := time.Now().Unix()
	passedTimeFromLastFetch := currentTime - LatestChainStatisticsData.DataFetchTime

	if passedTimeFromLastFetch > 3600 {
		fmt.Println("******Refreshing data from chain for chain stats*********")
		LatestChainStatisticsData.DataFetchTime = currentTime

		// getting last Master claim date
		sbGetLastMasterClaim := scriptbuilder.BeginScript()
		sbGetLastMasterClaim.CallContract("stake", "GetLastMasterClaim")
		script := sbGetLastMasterClaim.EndScript()
		encodedScript := hex.EncodeToString(script)
		response, err := Client.InvokeRawScript(UserSettings.ChainName, encodedScript)
		LatestChainStatisticsData.LastMasterClaimTimestamp = response.DecodeResult().AsNumber().Int64()
		if err != nil {
			return fmt.Errorf("error fetching last Master claim date: %v", err)
		}
		// ***

		// getting total staked soul
		account, err := Client.GetAccount("S3dP2jjf1jUG9nethZBWbnu9a6dFqB7KveTWU7znis6jpDy")
		if err != nil {
			return fmt.Errorf("getting total staked soul: %v", err)
		}
		totalStakedSoulBig := StringToBigInt(account.Balances[0].Amount)
		LatestChainStatisticsData.TotalStakedSoul = &totalStakedSoulBig
		fmt.Println("totalStakedSoul:", LatestChainStatisticsData.TotalStakedSoul)
		// ***

		// getting other data from chain with callcontract
		sb := scriptbuilder.BeginScript()
		sb.CallContract("gas", "GetLastInflationDate") // getting last inflation Timestamp
		sb.CallContract("stake", "GetMasterCount")     // getting total Soul Master count
		nextClaimDate := LatestChainStatisticsData.LastMasterClaimTimestamp + 32*86400
		sb.CallContract("stake", "GetClaimMasterCount", time.Unix(nextClaimDate, 0)) // getting eligible SoulMaster count for next claim
		script = sb.EndScript()
		encodedScript = hex.EncodeToString(script)
		response, err = Client.InvokeRawScript(UserSettings.ChainName, encodedScript)
		if err != nil {
			return fmt.Errorf("error fetching chain statistics: %v", err)
		}

		LatestChainStatisticsData.LastInflationTimeStamp = response.DecodeResults(0).AsNumber().Int64()

		LatestChainStatisticsData.TotalMaster = int16(response.DecodeResults(1).AsNumber().Int64())

		LatestChainStatisticsData.EligibleMaster = int16(response.DecodeResults(2).AsNumber().Int64())

		LatestChainStatisticsData.EstSMReward = float64(soulMasterRewardPool) / float64(LatestChainStatisticsData.EligibleMaster)

		LatestChainStatisticsData.NextInfTimeStamp = LatestChainStatisticsData.LastInflationTimeStamp + 90*86400

		if LatestChainStatisticsData.NextInfTimeStamp > int64(currentTime) {
			LatestChainStatisticsData.RemainedTimeForCrown = LatestChainStatisticsData.NextInfTimeStamp - int64(currentTime)
		} else {
			LatestChainStatisticsData.RemainedTimeForCrown = 0
		}

		// Adjust for 8 decimals: divide by 10^8
		divisor := new(big.Float).SetInt(big.NewInt(100000000))
		thresholdFloat := new(big.Float).SetInt(SoulMasterThreshold)
		thresholdFloat.Quo(thresholdFloat, divisor)

		// Calculate (estimatedSoulmasterReward * 12 * 100)
		rewardFloat := new(big.Float).SetFloat64(LatestChainStatisticsData.EstSMReward)
		rewardFloat.Mul(rewardFloat, big.NewFloat(12))
		rewardFloat.Mul(rewardFloat, big.NewFloat(100))

		// Divide by adjusted SoulMasterThreshold
		soulMasterAprBig := new(big.Float).Quo(rewardFloat, thresholdFloat)

		// Convert result to float64
		soulMasterAprFloat, _ := soulMasterAprBig.Float64()
		LatestChainStatisticsData.SMApr = soulMasterAprFloat

		// saveLatestChainData("chainstats", LatestChainStatisticsData)
		return nil

	}

	fmt.Println("*****Chain Stats*****")
	fmt.Println("eligible Soulmasters for next reward:", LatestChainStatisticsData.EligibleMaster)
	fmt.Println("estimated Soulmaster Reward:", LatestChainStatisticsData.EstSMReward)
	fmt.Println("last Inflation Time:", time.Unix(LatestChainStatisticsData.LastInflationTimeStamp, 0).Format("02-01-2006 15:04"))
	fmt.Println("last Soul Master reward Distribution:", time.Unix(LatestChainStatisticsData.LastMasterClaimTimestamp, 0).Format("02-01-2006 15:04"))
	fmt.Println("totalMasterCount:", LatestChainStatisticsData.TotalMaster)
	fmt.Println("remained time for crown", time.Duration(LatestChainStatisticsData.RemainedTimeForCrown))
	fmt.Println("next inflation date: ", time.Unix(LatestChainStatisticsData.NextInfTimeStamp, 0).Format("02-01-2006 15:04"))

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
