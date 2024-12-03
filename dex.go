package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

var dexGasLimit = big.NewInt(30000)

type DexPools struct {
	PoolKeyCount int      `json:"pool_count"`
	PoolList     []string `json:"pool_list"`
}

type PoolReserve struct {
	Symbol  string
	Amount  *big.Int
	Decimal int
}

type Pool struct {
	Reserve1 PoolReserve
	Reserve2 PoolReserve
}

var selectedPoolData Pool
var latestDexPools DexPools

func generateFromList(userTokens []string, pools []string) []string {
	fromList := make(map[string]bool)
	fmt.Println("Generating from list")
	for _, token := range userTokens {
		for _, pool := range pools {
			fmt.Println("Pool", pool)
			tokens := strings.Split(pool, "_")
			if tokens[0] == token || tokens[1] == token {
				fromList[token] = true
				break
			}
		}
	}

	// Convert map keys to a slice
	var fromListSlice []string
	for token := range fromList {
		fromListSlice = append(fromListSlice, token)
	}

	return fromListSlice
}
func updatePools() {

	currentPoolCount := getCountOfTokenPairsAndReserveKeys()

	if latestDexPools.PoolKeyCount < currentPoolCount {
		checkFrom := latestDexPools.PoolKeyCount
		latestDexPools.PoolKeyCount = currentPoolCount

		for i := checkFrom; i < currentPoolCount; i++ {
			var maxCheck int
			remainedChecks := currentPoolCount - i
			if remainedChecks > 4 { // it seems it only returns 2 pool key in one response it is returning 4 results for other cases like getting account details dunno
				maxCheck = 4
			} else {
				maxCheck = remainedChecks
			}
			fmt.Printf("Checking from %v, max check %v\n", checkFrom, maxCheck)
			maxCheckIndex := i + maxCheck
			sb := scriptbuilder.BeginScript()
			for checkIndex := i; checkIndex < maxCheckIndex; checkIndex += 2 {
				i += 2
				checkFrom += 2
				sb.CallContract("SATRN", "getTokenPairAndReserveKeysOnList", checkIndex)
				remainedChecks -= 2
				fmt.Printf("Checked index %v remained checks %v\n", checkIndex, remainedChecks)
			}

			scriptPairKey := sb.EndScript()
			encodedScript := hex.EncodeToString(scriptPairKey)
			responsePairKeys, err := client.InvokeRawScript(userSettings.ChainName, encodedScript)
			if err != nil {
				dialog.ShowError(fmt.Errorf("an error happened during updating pools!\n%v ", err.Error()), mainWindowGui)
				return
			}

			fmt.Println("Result count ", len(responsePairKeys.Results))
			for i := range responsePairKeys.Results {
				poolWithKey := responsePairKeys.DecodeResults(i).AsString()
				pool := removeKey(poolWithKey)
				latestDexPools.PoolList = append(latestDexPools.PoolList, pool)
				fmt.Printf("added pool: %v to list\n", pool)
			}

			if remainedChecks > 1 {
				i--
			}

		}

		saveDexPools()
	}

	// for _, pool := range poolList {
	// 	poolTokens := strings.Split(pool, "_")

	// 	dexPools[pool] = Pool{
	// 		Reserve1: PoolReserve{Symbol: poolTokens[0]},
	// 		Reserve2: PoolReserve{Symbol: poolTokens[1]},
	// 	}

	// }

}

func removeKey(poolWithKey string) string {
	var result string
	i := strings.LastIndex(poolWithKey, "_")
	result = poolWithKey[:i]
	return result
}

func generateToList(fromToken string, pools []string) []string {
	var toList []string

	for _, pool := range pools {
		tokens := strings.Split(pool, "_")
		if tokens[0] == fromToken {
			toList = append(toList, tokens[1])
		} else if tokens[1] == fromToken {
			toList = append(toList, tokens[0])
		}
	}

	return toList
}

func selectedPool(inToken, outToken string) string {
	for _, pool := range latestDexPools.PoolList {
		if strings.Contains(pool, inToken) {
			if strings.Contains(pool, outToken) {
				return pool
			}
		}
	}
	return ""
}

func getCountOfTokenPairsAndReserveKeys() int {
	sb := scriptbuilder.BeginScript()
	sb.CallContract("SATRN", "getCountOfTokenPairsAndReserveKeysOnList")
	script := sb.EndScript()
	encodedScript := hex.EncodeToString(script)

	response, err := client.InvokeRawScript(userSettings.ChainName, encodedScript)
	if err != nil {
		panic("Script1 invocation failed! Error: " + err.Error())
	}

	count := response.DecodeResult().AsNumber().Int64()

	fmt.Printf("Total token pairs and reserve keys listed: %v\n", count)
	return int(count)
}

// gets pool reserves from pool name
func getPoolReserves(pool string) Pool {
	poolReserveTokens := strings.Split(pool, "_")
	poolReserve1 := pool + "_" + poolReserveTokens[0]
	poolReserve2 := pool + "_" + poolReserveTokens[1]
	fmt.Printf("*****Checking pool %v reserves*****\n", pool)
	sb := scriptbuilder.BeginScript()
	sb.CallContract("SATRN", "getTokenPairAndReserveKeysOnListVALUE", poolReserve1)
	sb.CallContract("SATRN", "getTokenPairAndReserveKeysOnListVALUE", poolReserve2)
	scriptValue := sb.EndScript()
	encodedScript := hex.EncodeToString(scriptValue)
	responseValue, err := client.InvokeRawScript(userSettings.ChainName, encodedScript)
	if err != nil {
		panic("Script1 invocation failed! Error: " + err.Error())
	}
	reserve1 := responseValue.DecodeResults(0).AsNumber()
	reserve2 := responseValue.DecodeResults(1).AsNumber()
	token1Data, _ := updateOrCheckCache(poolReserveTokens[0], 3, "check")
	token2Data, _ := updateOrCheckCache(poolReserveTokens[1], 3, "check")

	poolData := Pool{Reserve1: PoolReserve{
		Symbol:  poolReserveTokens[0],
		Decimal: token1Data.Decimals,
		Amount:  reserve1,
	}, Reserve2: PoolReserve{
		Symbol:  poolReserveTokens[1],
		Decimal: token2Data.Decimals,
		Amount:  reserve2,
	},
	}
	fmt.Printf("%v reserve: %v\n%v reserve: %v\n", poolReserveTokens[0], formatBalance(*poolData.Reserve1.Amount, poolData.Reserve1.Decimal), poolReserveTokens[1], formatBalance(*poolData.Reserve2.Amount, poolData.Reserve2.Decimal))
	return poolData

}

func calculateSwapOut(inAmount, inReserves, outReserves *big.Int) (*big.Int, error) {
	outAmount := big.NewInt(0)

	inAmountMul := new(big.Int).Mul(inAmount, big.NewInt(997))
	inAmountDiv := new(big.Int).Div(inAmountMul, big.NewInt(1000))
	inAmountPlusReserves := new(big.Int).Add(inAmountDiv, inReserves)
	inReservesMulOut := new(big.Int).Mul(inReserves, outReserves)
	outAmount.Sub(outReserves, new(big.Int).Div(inReservesMulOut, inAmountPlusReserves))

	fmt.Printf("Calculate Swap Amount: %s\n", outAmount.String())

	if outAmount.Cmp(outReserves) >= 0 {
		return nil, fmt.Errorf("unsufficient liquidity")
	}
	if outAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("in amount is too small")
	}

	return outAmount, nil
}
func calculateSwapIn(outAmount, inReserves, outReserves *big.Int) (*big.Int, error) {
	// Define constants for fee calculations
	const feeNumerator = 997
	const feeDenominator = 1000

	// Calculate the numerator: outAmount * inReserves * feeDenominator
	numerator := new(big.Int).Mul(outAmount, inReserves)
	numerator = new(big.Int).Mul(numerator, big.NewInt(feeDenominator))

	// Calculate the denominator: (outReserves - outAmount) * feeNumerator
	denominator := new(big.Int).Sub(outReserves, outAmount)
	denominator = new(big.Int).Mul(denominator, big.NewInt(feeNumerator))

	// Calculate the input amount
	inAmount := new(big.Int).Div(numerator, denominator)

	// Adding one to handle rounding issues
	inAmount = inAmount.Add(inAmount, big.NewInt(1))

	fmt.Printf("Calculate Swap In Amount: %s\n", inAmount.String())
	if inAmount.Cmp(big.NewInt(0)) <= 0 {

		return nil, fmt.Errorf("unsufficient liquidity")
	}

	return inAmount, nil
}

// calculateSwapAndPriceImpact calculates the price impact of a swap on a DEX and returns the estimated output and input amounts.
// Parameters:
// - inAmount: The amount of the input token you want to swap.
// - outAmount: The desired amount of the output token you want to receive.
// - inReserves: The current reserves of the input token in the liquidity pool.
// - outReserves: The current reserves of the output token in the liquidity pool.
// - swapType: A string indicating the type of swap, either "swapOut" (input to output) or "swapIn" (output to input).
// Returns:
// - float64: The price impact as a percentage.
// - *big.Int: The estimated output amount for the swapOut type.
// - *big.Int: The estimated input amount for the swapIn type.
// - error: Any error encountered during the calculation.
func calculateSwapAndPriceImpact(inAmount, outAmount, inReserves, outReserves *big.Int, swapType string) (float64, *big.Int, *big.Int, error) {
	initialPrice := new(big.Float).Quo(new(big.Float).SetInt(outReserves), new(big.Float).SetInt(inReserves))

	var finalInReserves, finalOutReserves *big.Int
	var estimatedOutAmount, estimatedInAmount *big.Int
	var err error

	if swapType == "swapOut" {
		finalInReserves = new(big.Int).Add(inReserves, inAmount)
		estimatedOutAmount, err = calculateSwapOut(inAmount, inReserves, outReserves)
		if err != nil {
			return 0, nil, nil, err
		}
		finalOutReserves = new(big.Int).Sub(outReserves, estimatedOutAmount)
	} else if swapType == "swapIn" {
		finalOutReserves = new(big.Int).Sub(outReserves, outAmount)
		estimatedInAmount, err = calculateSwapIn(outAmount, inReserves, outReserves)
		if err != nil {
			return 0, nil, nil, err
		}
		finalInReserves = new(big.Int).Add(inReserves, estimatedInAmount)
	} else {
		return 0, nil, nil, fmt.Errorf("invalid swap type")
	}

	finalPrice := new(big.Float).Quo(new(big.Float).SetInt(finalOutReserves), new(big.Float).SetInt(finalInReserves))

	priceImpact := new(big.Float).Quo(new(big.Float).Sub(initialPrice, finalPrice), initialPrice)
	priceImpact.Mul(priceImpact, big.NewFloat(100)) // Convert to percentage

	priceImpactFloat64, _ := priceImpact.Float64() // Convert big.Float to float64

	return priceImpactFloat64, estimatedOutAmount, estimatedInAmount, nil
}

func createDexContent(creds Credentials) *container.Scroll {
	fee := new(big.Int).Mul(dexGasLimit, userSettings.GasPrice)
	err := checkFeeBalance(fee)
	var priceImpact float64
	var currentDexSlippage = userSettings.DexSlippage
	var inTokenSelected bool
	var outTokenSelected bool
	var inAmountEntryCorrect bool
	var outAmountEntryCorrect bool
	var slippageEntryCorrect bool
	var priceImpactIsNotHigh bool
	if err == nil {
		loadDexPools()
		amountInBinding := binding.NewString()
		slippageBinding := binding.NewString()
		slippageBinding.Set(fmt.Sprintf("%f", currentDexSlippage)) // Default slippage
		amountEntry := widget.NewEntryWithData(amountInBinding)
		amountEntry.SetPlaceHolder("Amount")
		amountEntry.Disable()
		warningMessageBinding := binding.NewString()
		warningMessageBinding.Set("Please select in token")
		warningMessage := widget.NewLabelWithData(warningMessageBinding)
		outAmountEntry := widget.NewEntry()
		var userTokens []string
		for _, token := range latestAccountData.FungibleTokens {
			userTokens = append(userTokens, token.Symbol)

		}
		updatePools()

		fmt.Println("user token count, pool count", len(userTokens), len(latestDexPools.PoolList))
		fromList := generateFromList(userTokens, latestDexPools.PoolList)
		tokenOutSelect := widget.NewSelect([]string{}, nil)
		tokenOutSelect.Disable()
		outAmountEntry.Disable()
		tokenInSelect := widget.NewSelect(fromList, nil)
		tokenInSelect.OnChanged = func(s string) {
			inTokenSelected = true
			toList := generateToList(s, latestDexPools.PoolList)
			tokenOutSelect.ClearSelected()
			amountEntry.SetText("")
			tokenOutSelect.SetOptions(toList)
			amountEntry.Enable()
			outAmountEntry.SetText("")
			mainWindowGui.Canvas().Focus(amountEntry)
			warningMessageBinding.Set("Please enter amount")
		}

		slippageEntry := widget.NewEntryWithData(slippageBinding)
		slippageEntry.SetPlaceHolder("Slippage Tolerance %")

		swapBtn := widget.NewButton("Swap Tokens", func() {
			if tokenInSelect.Selected == "" || tokenOutSelect.Selected == "" {
				dialog.ShowError(fmt.Errorf("please select tokens"), mainWindowGui)
				return
			}

			amountStr := amountEntry.Text
			slippageStr, _ := slippageBinding.Get()

			// Parse and validate amount
			amount, err := convertUserInputToBigInt(amountStr, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid amount: %v", err), mainWindowGui)
				return
			}

			// Verify sufficient balance
			token := latestAccountData.FungibleTokens[tokenInSelect.Selected]
			if amount.Cmp(&token.Amount) > 0 {
				dialog.ShowError(fmt.Errorf("insufficient %s balance", tokenInSelect.Selected), mainWindowGui)
				return
			}

			slippage, err := strconv.ParseFloat(slippageStr, 64)
			if err != nil || slippage <= 0 || slippage > 100 {
				dialog.ShowError(fmt.Errorf("invalid slippage (must be between 0 and 100)"), mainWindowGui)
				return
			}

			// Check KCAL for gas
			gasFee := new(big.Int).Mul(userSettings.GasPrice, dexGasLimit)
			if err := checkFeeBalance(gasFee); err != nil {
				dialog.ShowError(err, mainWindowGui)
				return
			}

			// Confirm swap
			confirmMessage := fmt.Sprintf("Swap %s %s for estimated %s %s\nPrice Impact %.2f%% (or price increase for %s)\nSlippage: %.1f%%\nGas Fee: %s KCAL",
				formatBalance(*amount, token.Decimals),
				tokenInSelect.Selected,
				outAmountEntry.Text,
				tokenOutSelect.Selected,
				priceImpact,
				tokenOutSelect.Selected,
				slippage,
				formatBalance(*gasFee, kcalDecimals))

			dialog.ShowConfirm("Confirm Swap", confirmMessage, func(confirmed bool) {
				if confirmed {
					err = executeSwap(tokenInSelect.Selected, tokenOutSelect.Selected, amount, slippage, creds)
					if err != nil {
						dialog.ShowError(err, mainWindowGui)
					}
				}
			}, mainWindowGui)
		})
		swapBtn.Disable()
		checkSwapBtnState := func() {
			fmt.Println("Swap button State", inTokenSelected, outTokenSelected, inAmountEntryCorrect, outAmountEntryCorrect, slippageEntryCorrect, priceImpactIsNotHigh)
			if inTokenSelected && outTokenSelected && inAmountEntryCorrect && outAmountEntryCorrect && slippageEntryCorrect && priceImpactIsNotHigh {
				swapBtn.Enable()
			} else {
				swapBtn.Disable()
			}

		}
		tokenOutSelect.OnChanged = func(s string) {

			if s != "" {
				var userSlippage float64
				if slippageEntry.Validate() == nil {
					userSlippage, _ = strconv.ParseFloat(slippageEntry.Text, 64)
				} else {
					warningMessageBinding.Set("Check your slippage")
					return
				}

				outTokenSelected = true
				checkSwapBtnState()
				outAmountEntry.Enable()

				pool := selectedPool(tokenInSelect.Selected, s)
				selectedPoolData = getPoolReserves(pool)
				inAmount, _ := convertUserInputToBigInt(amountEntry.Text, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
				if selectedPoolData.Reserve1.Symbol == tokenInSelect.Selected {
					inReserve := selectedPoolData.Reserve1.Amount
					outReserve := selectedPoolData.Reserve2.Amount
					impact, estOutAmount, _, err := calculateSwapAndPriceImpact(inAmount, nil, inReserve, outReserve, "swapOut") //(inAmount, inReserve, outReserve)
					if err != nil {
						outAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set(err.Error())
						return
					}
					priceImpact = impact
					if priceImpact > userSlippage {
						priceImpactIsNotHigh = false
						warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
						estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
						outAmountEntry.Text = (estOutAmountStr)
						outAmountEntry.Refresh()
						checkSwapBtnState()
						return
					} else {
						priceImpactIsNotHigh = true
						inAmountEntryCorrect = true
						checkSwapBtnState()
					}
					estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
					outAmountEntry.Text = estOutAmountStr
					outAmountEntryCorrect = true
					checkSwapBtnState()
					warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
					outAmountEntry.Refresh()

				} else {
					inReserve := selectedPoolData.Reserve2.Amount
					outReserve := selectedPoolData.Reserve1.Amount
					impact, estOutAmount, _, err := calculateSwapAndPriceImpact(inAmount, nil, inReserve, outReserve, "swapOut")
					if err != nil {
						outAmountEntryCorrect = false
						warningMessageBinding.Set(err.Error())
						checkSwapBtnState()
						return
					}
					priceImpact = impact
					if priceImpact > userSlippage {
						priceImpactIsNotHigh = false
						warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
						estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
						outAmountEntry.Text = (estOutAmountStr)
						outAmountEntry.Refresh()
						checkSwapBtnState()
						return
					} else {
						priceImpactIsNotHigh = true
						inAmountEntryCorrect = true
						checkSwapBtnState()
					}
					estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve1.Decimal)
					outAmountEntry.Text = estOutAmountStr
					outAmountEntryCorrect = true
					checkSwapBtnState()
					warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
					outAmountEntry.Refresh()
				}

			}
		}
		amountEntry.Validator = func(s string) error {
			input, err := convertUserInputToBigInt(s, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
			if err != nil {
				warningMessageBinding.Set(err.Error())

				inAmountEntryCorrect = false
				checkSwapBtnState()
				return err
			}
			if input.Cmp(big.NewInt(0)) <= 0 {
				warningMessageBinding.Set("Please enter amount")

				inAmountEntryCorrect = false
				checkSwapBtnState()
				return err
			}
			if tokenInSelect.Selected == "KCAL" {
				fee := new(big.Int).Mul(dexGasLimit, defaultSettings.GasPrice)
				max := new(big.Int).Add(input, fee)
				err := checkFeeBalance(max)
				if err != nil {
					inAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set("Not enough Kcal")
					return err

				}

			}
			balance := latestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
			if input.Cmp(&balance) <= 0 {
				if tokenOutSelect.Selected == "" {
					warningMessageBinding.Set("Please select out token")
					inAmountEntryCorrect = false

				} else {
					// warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s", tokenInSelect.Selected, tokenOutSelect.Selected))
					inAmountEntryCorrect = true
				}
				checkSwapBtnState()
				tokenOutSelect.Enable()
				return nil
			} else {

				inAmountEntryCorrect = false
				checkSwapBtnState()
				warningMessageBinding.Set("Balance is not sufficent")
				return fmt.Errorf("balance is not sufficent")
			}
		}

		slippageSaveBtn := widget.NewButton("Save as default", func() {
			if slippageEntry.Validate() == nil {
				slippage, _ := strconv.ParseFloat(slippageEntry.Text, 64)
				userSettings.DexSlippage = slippage
				err := saveSettings()
				if err == nil {

					dialog.ShowInformation("Saved successfully", fmt.Sprintf("Your default slippage is set to %.2f%%", slippage), mainWindowGui)

				} else {
					dialog.ShowError(err, mainWindowGui)
				}

			}
		})
		slippageSaveBtn.Disable()
		slippageEntry.Validator = func(s string) error {
			slippage, err := strconv.ParseFloat(s, 64)

			if err != nil {
				slippageEntryCorrect = false
				slippageSaveBtn.Disable()
				checkSwapBtnState()
				return err
			} else if slippage <= 0 {
				slippageEntryCorrect = false
				slippageSaveBtn.Disable()
				checkSwapBtnState()
				return fmt.Errorf("invalid slippage (must be greater than 0 )")
			} else if slippage > 100 {
				slippageEntryCorrect = false
				slippageSaveBtn.Disable()
				checkSwapBtnState()
				return fmt.Errorf("invalid slippage (must be less than 100 )")
			} else {
				slippageEntryCorrect = true

				if userSettings.DexSlippage != slippage {
					currentDexSlippage = slippage
					fmt.Println("Slipage save btn enabled")
					slippageSaveBtn.Enable()
					slippageSaveBtn.Refresh()
				} else {
					fmt.Println("Slipage save btn disabled")
					slippageSaveBtn.Disable()
					slippageSaveBtn.Refresh()
				}
				if priceImpact < currentDexSlippage && swapBtn.Disabled() { //for enabling swap button after user changes slippage
					amountEntry.Validate()
					outAmountEntry.Validate()
					priceImpactIsNotHigh = true
				}
				checkSwapBtnState()
				return nil
			}

		}
		swapIcon := widget.NewLabelWithStyle("🢃", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		maxBttn := widget.NewButton("Max", func() {
			if tokenInSelect.Selected == "KCAL" {
				fee := new(big.Int).Mul(dexGasLimit, defaultSettings.GasPrice)
				kcalbal := latestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
				max := new(big.Int).Sub(&kcalbal, fee)
				if max.Cmp(big.NewInt(0)) >= 0 {
					amountEntry.SetText(formatBalance(*max, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals))
					amountEntry.Validate()
					amountEntry.Refresh()
				} else {
					amountEntry.SetText("Not enough Kcal")
					amountEntry.Validate()
					amountEntry.Refresh()
				}

			} else {
				amountEntry.SetText(formatBalance(latestAccountData.FungibleTokens[tokenInSelect.Selected].Amount, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals))
				amountEntry.Validate()
				amountEntry.Refresh()
			}
		})
		inTokenSelect := container.NewHBox(widget.NewLabel("From\t"), tokenInSelect)
		inTokenLyt := container.NewBorder(nil, nil, inTokenSelect, maxBttn, amountEntry)

		amountEntry.OnChanged = func(s string) {
			if tokenOutSelect.Selected != "" {
				var userSlippage float64
				if slippageEntry.Validate() == nil {

					userSlippage, _ = strconv.ParseFloat(slippageEntry.Text, 64)
				} else {
					warningMessageBinding.Set("Check your slippage")
					return
				}
				inAmount, err := convertUserInputToBigInt(amountEntry.Text, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
				if err != nil {
					return
				}
				if selectedPoolData.Reserve1.Symbol == tokenInSelect.Selected {
					inReserve := selectedPoolData.Reserve1.Amount
					outReserve := selectedPoolData.Reserve2.Amount
					impact, estOutAmount, _, err := calculateSwapAndPriceImpact(inAmount, nil, inReserve, outReserve, "swapOut") //(inAmount, inReserve, outReserve)
					if err != nil {
						warningMessageBinding.Set(err.Error())
						inAmountEntryCorrect = false
						checkSwapBtnState()
						return
					}
					priceImpact = impact
					if priceImpact > userSlippage {
						priceImpactIsNotHigh = false
						warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
						estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
						outAmountEntry.Text = (estOutAmountStr)
						outAmountEntry.Refresh()
						checkSwapBtnState()
						return
					} else {
						priceImpactIsNotHigh = true
						checkSwapBtnState()
					}
					estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntryCorrect = true
					checkSwapBtnState()
					warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
					outAmountEntry.Refresh()

				} else {
					inReserve := selectedPoolData.Reserve2.Amount
					outReserve := selectedPoolData.Reserve1.Amount
					impact, estOutAmount, _, err := calculateSwapAndPriceImpact(inAmount, nil, inReserve, outReserve, "swapOut")
					if err != nil {
						warningMessageBinding.Set(err.Error())
						outAmountEntryCorrect = false
						checkSwapBtnState()
						return
					}
					priceImpact = impact
					if priceImpact > userSlippage {
						priceImpactIsNotHigh = false
						warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
						estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve2.Decimal)
						outAmountEntry.Text = (estOutAmountStr)
						outAmountEntry.Refresh()
						checkSwapBtnState()
						return
					} else {
						priceImpactIsNotHigh = true
						checkSwapBtnState()
					}
					estOutAmountStr := formatBalance(*estOutAmount, selectedPoolData.Reserve1.Decimal)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntryCorrect = true
					checkSwapBtnState()
					warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
					outAmountEntry.Refresh()
				}

			}
		}

		outAmountEntry.OnChanged = func(s string) {
			tokenData, _ := updateOrCheckCache(tokenOutSelect.Selected, 3, "check")
			outAmount, err := convertUserInputToBigInt(s, tokenData.Decimals)
			if err != nil {
				warningMessageBinding.Set(err.Error())
				outAmountEntryCorrect = false
				checkSwapBtnState()
				return
			}
			if outAmount.Cmp(big.NewInt(0)) <= 0 {
				warningMessageBinding.Set("In amount is too small")
				outAmountEntryCorrect = false
				checkSwapBtnState()
				return
			}
			if tokenOutSelect.Selected == selectedPoolData.Reserve1.Symbol {

				impact, _, estInAmount, err := calculateSwapAndPriceImpact(nil, outAmount, selectedPoolData.Reserve2.Amount, selectedPoolData.Reserve1.Amount, "swapIn") //(inAmount, inReserve, outReserve)

				if err != nil {
					warningMessageBinding.Set(err.Error())
					return
				}
				inAmountStr := formatBalance(*estInAmount, selectedPoolData.Reserve2.Decimal)
				amountEntry.Text = inAmountStr
				amountEntry.Refresh()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
				outAmountEntryCorrect = true
				checkSwapBtnState()
				amountEntry.Validate()
			} else {
				impact, _, estInAmount, err := calculateSwapAndPriceImpact(nil, outAmount, selectedPoolData.Reserve1.Amount, selectedPoolData.Reserve2.Amount, "swapIn") //(inAmount, inReserve, outReserve)
				if err != nil {
					warningMessageBinding.Set(err.Error())
					return
				}
				inAmountStr := formatBalance(*estInAmount, selectedPoolData.Reserve1.Decimal)
				amountEntry.Text = inAmountStr
				amountEntry.Refresh()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))

				amountEntry.Validate()
				outAmountEntryCorrect = true
				checkSwapBtnState()
			}

		}
		// outAmountEntry.Disable()

		slippageLyt := container.NewBorder(nil, nil, nil, slippageSaveBtn, slippageEntry)
		outAmountEntry.SetPlaceHolder("Estimated out amount")
		outTokenSelect := container.NewHBox(widget.NewLabel("To\t"), tokenOutSelect)
		outTokenLyt := container.NewBorder(nil, nil, outTokenSelect, nil, outAmountEntry)
		form := container.NewVBox(
			inTokenLyt,

			swapIcon,
			outTokenLyt,
			widget.NewLabel("Slippage Tolerance (%):"),
			slippageLyt,
			warningMessage,
			swapBtn,
			widget.NewRichTextFromMarkdown("Powered by [Saturn Dex](https://saturn.stellargate.io/)"),
		)
		dexTab.Content = container.NewPadded(form)
		return dexTab
	} else {
		noKcalMessage := widget.NewLabelWithStyle(fmt.Sprintf("Looks like Sparky low on sparks! ⚡️🕹️\n Your swap needs some Phantasma Energy (KCAL) to keep the ghostly gears turning. Time to add some KCAL and get that blockchain buzzing faster than a haunted hive!\n You need at least %v Kcal", formatBalance(*fee, kcalDecimals)), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		noKcalMessage.Wrapping = fyne.TextWrapWord
		dexTab.Content = container.NewVBox(noKcalMessage)
		return dexTab
	}

}

func executeSwap(tokenIn, tokenOut string, amountIn *big.Int, slippageTolerance float64, creds Credentials) error {
	showUpdatingDialog()
	defer closeUpdatingDialog()

	if creds.LastSelectedWallet == "" {
		return fmt.Errorf("no wallet selected")
	}

	wallet := creds.Wallets[creds.LastSelectedWallet]
	fmt.Printf("Using wallet: %s\n", wallet.Address)

	keyPair, err := cryptography.FromWIF(wallet.WIF)
	if err != nil {
		return fmt.Errorf("invalid wallet key: %v", err)
	}

	// Format amounts
	fmt.Printf("Input amount (raw): %s\n", amountIn.String())
	fmt.Printf("Input amount (formatted): %s\n", formatBalance(*amountIn, latestAccountData.FungibleTokens[tokenIn].Decimals))

	// Convert slippage to basis points (multiply by 100 to get integer)
	slippageBasisPoints := new(big.Int).SetInt64(int64(slippageTolerance * 100))

	// Debug print script parameters
	fmt.Printf("\nConstructing SATRN.swap parameters:\n")
	fmt.Printf("1. from: %s\n", wallet.Address)
	fmt.Printf("2. amountIn: %s (%s %s)\n", amountIn.String(),
		formatBalance(*amountIn, latestAccountData.FungibleTokens[tokenIn].Decimals), tokenIn)
	fmt.Printf("3. tokenIn: %s\n", tokenIn)
	fmt.Printf("4. tokenOut: %s\n", tokenOut)
	fmt.Printf("5. slippageTolerance: %d basis points\n", slippageBasisPoints)

	// Check if we have enough balance
	balance := latestAccountData.FungibleTokens[tokenIn].Amount
	fmt.Printf("Current balance: %s %s\n",
		formatBalance(balance, latestAccountData.FungibleTokens[tokenIn].Decimals), tokenIn)

	// Set increased gas limit specifically for swap operations

	fmt.Printf("\nGas settings:\n")
	fmt.Printf("Price: %s\n", userSettings.GasPrice.String())
	fmt.Printf("Limit: %s (increased for swap)\n", dexGasLimit.String())

	swapPayload := []byte("Spallet Swap")

	sb := scriptbuilder.BeginScript()
	script := sb.AllowGas(wallet.Address, cryptography.NullAddress().String(), userSettings.GasPrice, dexGasLimit).
		CallContract("SATRN", "swap",
			wallet.Address,      // from
			amountIn,            // amountIn
			tokenIn,             // tokenIn
			tokenOut,            // tokenOut
			slippageBasisPoints, // slippageTolerance
		).
		SpendGas(wallet.Address).
		EndScript()

	fmt.Printf("\nGenerated script hex: %x\n", script)

	// Build and sign transaction
	expire := time.Now().UTC().Add(time.Second * 300).Unix()
	fmt.Printf("Transaction expiration: %v\n", time.Unix(expire, 0))

	tx := blockchain.NewTransaction(userSettings.NetworkName, userSettings.ChainName, script, uint32(expire), swapPayload) // Using custom payload
	tx.Sign(keyPair)

	txHex := hex.EncodeToString(tx.Bytes())
	fmt.Printf("Complete transaction hex: %s\n", txHex)

	txHash, err := client.SendRawTransaction(txHex)
	if err != nil {
		fmt.Printf("Failed to send transaction: %v\n", err)
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent with hash: %s\n", txHash)
	go monitorSwapTransaction(txHash)

	return nil
}

func monitorSwapTransaction(txHash string) {
	maxRetries := 30
	retryCount := 0
	retryDelay := time.Millisecond * 500

	fmt.Printf("Starting transaction monitoring for hash: %s\n", txHash)

	for {
		if retryCount >= maxRetries {
			fmt.Printf("Transaction monitoring timed out after %d retries\n", maxRetries)
			dialog.ShowError(fmt.Errorf("transaction monitoring timed out. Transaction hash: %s\nPlease check the explorer manually", txHash), mainWindowGui)
			explorerURL := fmt.Sprintf("%s%s", userSettings.TxExplorerLink, txHash)
			if parsedURL, err := url.Parse(explorerURL); err == nil {
				fyne.CurrentApp().OpenURL(parsedURL)
			}
			return
		}

		fmt.Printf("Checking transaction status (attempt %d/%d)\n", retryCount+1, maxRetries)
		txResult, err := client.GetTransaction(txHash)
		if err != nil {
			fmt.Printf("Error getting transaction status: %v\n", err)
			if strings.Contains(err.Error(), "could not decode body") ||
				strings.Contains(err.Error(), "rpc call") {
				retryCount++
				time.Sleep(retryDelay)
				continue
			}

			dialog.ShowError(fmt.Errorf("failed to get transaction status: %v", err), mainWindowGui)
			return
		}

		if txResult.StateIsSuccess() {
			fmt.Printf("Transaction successful\n")
			dialog.ShowInformation("Success", fmt.Sprintf("Swap completed successfully\nTransaction: %s", txHash), mainWindowGui)
			return
		}
		if txResult.StateIsFault() {
			fmt.Printf("Transaction failed\n")
			dialog.ShowError(fmt.Errorf("swap failed\nTransaction: %s", txHash), mainWindowGui)
			return
		}

		fmt.Printf("Transaction pending, state: %s\n", txResult.State)
		retryCount++
		time.Sleep(retryDelay)
	}
}
