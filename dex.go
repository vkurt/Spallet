package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"sync"
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

type PoolReserve struct {
	Symbol  string
	Decimal int
	Amount  *big.Int
}

type Pool struct {
	Reserve1 PoolReserve
	Reserve2 PoolReserve
}

type TransactionDataForDex struct {
	AmountIn  *big.Int
	AmountOut *big.Int
	TokenIn   string
	TokenOut  string
	Pool      string
}

type CachedPoolData struct {
	Pool    Pool
	Updated time.Time
}

var (
	poolCache     = make(map[string]CachedPoolData)
	cacheDuration = 15 * time.Second
	// poolUpdateInterval = 15 * time.Second
	mu sync.Mutex
)

type DexPools struct {
	PoolKeyCount int      `json:"pool_count"`
	PoolList     []string `json:"pool_list"`
}

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
	tokenSet := make(map[string]bool)
	var toList []string

	for _, pool := range pools {
		tokens := strings.Split(pool, "_")
		if tokens[0] != fromToken {
			if !tokenSet[tokens[0]] {
				tokenSet[tokens[0]] = true
				toList = append(toList, tokens[0])
			}
		}

		if tokens[1] != fromToken {
			if !tokenSet[tokens[1]] {
				tokenSet[tokens[1]] = true
				toList = append(toList, tokens[1])
			}
		}
	}

	return toList
}

// func selectedPool(inToken, outToken string) string {
// 	for _, pool := range latestDexPools.PoolList {
// 		if strings.Contains(pool, inToken) {
// 			if strings.Contains(pool, outToken) {
// 				return pool
// 			}
// 		}
// 	}
// 	return ""
// }

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
	mu.Lock()
	defer mu.Unlock()

	// Check if the pool data is in the cache and is up-to-date
	if cachedData, found := poolCache[pool]; found && time.Since(cachedData.Updated) < cacheDuration {
		return cachedData.Pool
	}

	// Your existing implementation for fetching pool reserves
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

	poolData := Pool{
		Reserve1: PoolReserve{
			Symbol:  poolReserveTokens[0],
			Decimal: token1Data.Decimals,
			Amount:  reserve1,
		},
		Reserve2: PoolReserve{
			Symbol:  poolReserveTokens[1],
			Decimal: token2Data.Decimals,
			Amount:  reserve2,
		},
	}
	fmt.Printf("%v reserve: %v\n%v reserve: %v\n", poolReserveTokens[0], formatBalance(*poolData.Reserve1.Amount, poolData.Reserve1.Decimal), poolReserveTokens[1], formatBalance(*poolData.Reserve2.Amount, poolData.Reserve2.Decimal))

	// Store the fetched pool data in the cache
	poolCache[pool] = CachedPoolData{Pool: poolData, Updated: time.Now()}

	return poolData
}

/**
 * findAllSwapRoutes finds all possible swap routes between two tokens.
 *
 * @param pools           List of pool pairs in the format "TOKEN1_TOKEN2".
 * @param fromToken       The starting token for the swap.
 * @param toToken         The destination token for the swap.
 * @return allRoutes      All possible routes as a list of lists of pools.
 */
func findAllSwapRoutes(pools []string, fromToken, toToken string) ([][]string, error) {
	var allRoutes [][]string

	// Build a map of pools to easily find connections between tokens
	poolMap := make(map[string][]string)
	for _, pool := range pools {
		tokens := strings.Split(pool, "_")
		poolMap[tokens[0]] = append(poolMap[tokens[0]], tokens[1])
		poolMap[tokens[1]] = append(poolMap[tokens[1]], tokens[0])
	}

	// Depth-First Search (DFS) to find all possible routes
	var dfs func(currentToken string, visited map[string]bool, currentRoute []string)
	dfs = func(currentToken string, visited map[string]bool, currentRoute []string) {
		// Base case: If we reach the destination token, add the current route to allRoutes
		if currentToken == toToken {
			allRoutes = append(allRoutes, append([]string(nil), currentRoute...))
			return
		}

		// Recursive case: Explore connected tokens (pools)
		for _, nextToken := range poolMap[currentToken] {
			if !visited[nextToken] && nextToken != "" {
				visited[nextToken] = true
				currentRoute = append(currentRoute, currentToken+"_"+nextToken)
				dfs(nextToken, visited, currentRoute)
				currentRoute = currentRoute[:len(currentRoute)-1]
				visited[nextToken] = false
			}
		}
	}

	// Initialize visited map and start DFS from the fromToken
	visited := make(map[string]bool)
	visited[fromToken] = true
	dfs(fromToken, visited, []string{})

	// Ensure we do not return incomplete routes
	filteredRoutes := [][]string{}
	for _, route := range allRoutes {
		valid := true
		for _, step := range route {
			tokens := strings.Split(step, "_")
			if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
				valid = false
				break
			}
		}
		if valid {
			fmt.Println("found valid route", route)
			filteredRoutes = append(filteredRoutes, route)
		}
	}
	if len(filteredRoutes) == 0 {
		return nil, errors.New("no valid route found")
	}

	return filteredRoutes, nil
}

/**
 * reverseCalculateInputAmounts calculates the input amounts required to achieve a specific output amount along a given route.
 *
 * @param route            The best route as a list of pools.
 * @param desiredOutAmount The desired output amount.
 * @param pools            List of pool pairs in the format "TOKEN1_TOKEN2".
 * @return calculatedInAmount The calculated input amount to achieve the desired output amount.
 * @return error           Any error encountered during the computation.
 */
func reverseCalculateInputAmounts(route []string, desiredOutAmount *big.Int, pools []string) (*big.Int, error) {
	currentOutAmount := new(big.Int).Set(desiredOutAmount)

	// Function to get correct pool reserves by considering the order of tokens
	getCorrectPoolReserves := func(token1, token2 string) (string, Pool) {
		poolKey := token1 + "_" + token2
		reverseKey := token2 + "_" + token1
		pool := Pool{}
		if contains(pools, poolKey) {
			pool = getPoolReserves(poolKey)
		} else if contains(pools, reverseKey) {
			pool = getPoolReserves(reverseKey)
			pool.Reserve1, pool.Reserve2 = pool.Reserve2, pool.Reserve1 // Swap reserves
			poolKey = reverseKey
		}
		return poolKey, pool
	}

	// Iterate through the route in reverse to calculate required input amounts
	for i := len(route) - 1; i >= 0; i-- {
		pool := route[i]
		tokens := strings.Split(pool, "_")
		tokenIn, tokenOut := tokens[0], tokens[1]

		_, poolReserves := getCorrectPoolReserves(tokenIn, tokenOut)

		// Calculate the required input amount using calculateSwapIn
		inAmount, err := calculateSwapIn(currentOutAmount, poolReserves.Reserve1.Amount, poolReserves.Reserve2.Amount)
		if err != nil {
			return nil, err
		}

		// Update the current out amount for the previous swap step
		currentOutAmount = inAmount
	}

	return currentOutAmount, nil
}

/**
 * evaluateRoutes finds the best swap route among all possible routes by calculating price impacts or maximizing output amount.
 *
 * @param routes               All possible routes as a list of lists of pools.
 * @param fromToken            The starting token for the swap.
 * @param pools                List of pool pairs in the format "TOKEN1_TOKEN2".
 * @param inAmount             The amount of the starting token to swap.
 * @param slippageTolerance    Maximum allowable price impact for selecting the best priced pool.
 * @param selectionMethod      Selection method for the best route: "highestOutput", "lowestImpact", "auto".
 * @return bestRoute           The best route as a list of pools.
 * @return bestTransactionData The transaction data for the best route.
 * @return lowestPriceImpact   The lowest price impact for the best route.
 * @return bestRouteString     The best route as a string in the format "token1=>token2=>token3".
 * @return numPools            The number of pools used in the best route.
 * @return finalOutAmount      The final output amount for the best route.
 * @return error               Any error encountered during the computation.
 */
func evaluateRoutes(routes [][]string, fromToken string, pools []string, inAmount *big.Int, slippageTolerance float64, selectionMethod string) ([]string, []TransactionDataForDex, float64, string, int, *big.Int, error) {
	var bestRoute []string                                       // Stores the best route as a list of pools
	var lowestPriceImpactRoute []string                          // Stores the route with the lowest price impact
	var highestOutputRoute []string                              // Stores the route with the highest output amount
	var lowestPriceImpactTransactionData []TransactionDataForDex // Stores the transaction data for the lowest price impact route
	var highestOutputTransactionData []TransactionDataForDex     // Stores the transaction data for the highest output route
	var lowestPriceImpact float64 = -1                           // Tracks the lowest price impact found
	var highestOutputPriceImpact float64 = -1                    // Tracks the price impact of the highest output route
	var lowestPriceImpactRouteString string                      // Stores the best route as a formatted string for the lowest price impact route
	var highestOutputRouteString string                          // Stores the best route as a formatted string for the highest output route
	var numPools int                                             // Stores the number of pools used in the best route
	var finalOutAmount *big.Int                                  // Stores the final output amount for the best route
	var highestOutputAmount *big.Int                             // Stores the final output amount for the highest output route
	var bestRouteString string
	var bestTransactionData []TransactionDataForDex
	fmt.Println("*****evaluating routes******")

	// Function to get correct pool reserves by considering the order of tokens
	getCorrectPoolReserves := func(token1, token2 string) (string, Pool) {
		poolKey := token1 + "_" + token2
		reverseKey := token2 + "_" + token1
		pool := Pool{}
		if contains(pools, poolKey) {
			pool = getPoolReserves(poolKey)
		} else if contains(pools, reverseKey) {
			pool = getPoolReserves(reverseKey)
			pool.Reserve1, pool.Reserve2 = pool.Reserve2, pool.Reserve1 // Swap reserves
			poolKey = reverseKey
		}
		return poolKey, pool
	}

	// Iterate through all routes to find the lowest price impact route and highest output route
	for _, route := range routes {
		var currentAmount *big.Int = new(big.Int).Set(inAmount)
		var currentTransactionData []TransactionDataForDex
		var currentRouteString []string = []string{fromToken}
		var totalPriceImpact float64 = 0
		var priceImpactFactor float64 = 1.0

		for _, pool := range route {
			tokens := strings.Split(pool, "_")
			currentToken, nextToken := tokens[0], tokens[1]
			fmt.Println("trying to get correct pools for", currentToken, nextToken)
			poolKey, poolReserves := getCorrectPoolReserves(currentToken, nextToken)
			priceImpact, outAmount, _, _ := calculateSwapAndPriceImpact(currentAmount, nil, poolReserves.Reserve1.Amount, poolReserves.Reserve2.Amount, "swapOut")

			// If outAmount is nil, skip this route
			if outAmount == nil || outAmount.Cmp(big.NewInt(0)) <= 0 {
				totalPriceImpact = -1
				break
			}

			// Calculate the compounding effect of price impact
			priceImpactFactor *= (1 - (priceImpact / 100))
			totalPriceImpact = (1 - priceImpactFactor) * 100

			currentTransactionData = append(currentTransactionData, TransactionDataForDex{
				AmountIn:  new(big.Int).Set(currentAmount),
				AmountOut: new(big.Int).Set(outAmount),
				TokenIn:   currentToken,
				TokenOut:  nextToken,
				Pool:      poolKey,
			})
			currentAmount = outAmount
			currentRouteString = append(currentRouteString, nextToken)
		}

		// Compare routes to find the lowest price impact route
		if totalPriceImpact != -1 && (lowestPriceImpact == -1 || totalPriceImpact < lowestPriceImpact) {
			lowestPriceImpact = totalPriceImpact
			lowestPriceImpactRoute = route
			lowestPriceImpactTransactionData = currentTransactionData
			lowestPriceImpactRouteString = strings.Join(currentRouteString, " > ")
		}

		// Compare routes to find the highest output route
		if totalPriceImpact != -1 && (highestOutputAmount == nil || (currentAmount.Cmp(highestOutputAmount) > 0)) {
			highestOutputAmount = currentAmount
			highestOutputPriceImpact = totalPriceImpact
			highestOutputRoute = route
			highestOutputTransactionData = currentTransactionData
			highestOutputRouteString = strings.Join(currentRouteString, " > ")
		}
	}

	// Check if a valid route was found
	if len(highestOutputRoute) == 0 || len(lowestPriceImpactRoute) == 0 {
		return nil, nil, 0, "", 0, nil, errors.New("no valid route found")
	}

	// Select the best route based on the selection method
	switch selectionMethod {
	case "highestOutput":
		bestRoute = highestOutputRoute
		bestTransactionData = highestOutputTransactionData
		bestRouteString = highestOutputRouteString
		finalOutAmount = highestOutputAmount
		lowestPriceImpact = highestOutputPriceImpact // Store the corresponding price impact

	case "lowestImpact":
		bestRoute = lowestPriceImpactRoute
		bestTransactionData = lowestPriceImpactTransactionData
		bestRouteString = lowestPriceImpactRouteString
		finalOutAmount = lowestPriceImpactTransactionData[len(lowestPriceImpactTransactionData)-1].AmountOut

	case "auto":
		if highestOutputPriceImpact != -1 && highestOutputPriceImpact <= slippageTolerance {
			fmt.Println("Auto selected highest output route")
			bestRoute = highestOutputRoute
			bestTransactionData = highestOutputTransactionData
			bestRouteString = highestOutputRouteString
			finalOutAmount = highestOutputAmount
			lowestPriceImpact = highestOutputPriceImpact // Store the corresponding price impact
		} else {
			fmt.Println("Auto select lowest impact route")
			bestRoute = lowestPriceImpactRoute
			bestTransactionData = lowestPriceImpactTransactionData
			bestRouteString = lowestPriceImpactRouteString
			finalOutAmount = lowestPriceImpactTransactionData[len(lowestPriceImpactTransactionData)-1].AmountOut
		}

	default:
		return nil, nil, 0, "", 0, nil, errors.New("invalid selection method")
	}

	numPools = len(bestRoute)

	return bestRoute, bestTransactionData, lowestPriceImpact, bestRouteString, numPools, finalOutAmount, nil
}

// Helper function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// func calculateTotalPriceImpact(route []string, inAmount *big.Int) float64 {
// 	totalPriceImpact := 0.0
// 	currentAmount := inAmount

// 	for _, pool := range route {
// 		fmt.Println("calculating price impact for", pool)
// 		poolReserves := getPoolReserves(pool)
// 		priceImpact, outAmount, _, _ := calculateSwapAndPriceImpact(currentAmount, nil, poolReserves.Reserve1.Amount, poolReserves.Reserve2.Amount, "swapOut")
// 		totalPriceImpact += priceImpact
// 		currentAmount = outAmount
// 	}

// 	return totalPriceImpact
// }

func calculateSwapOut(inAmount, inReserves, outReserves *big.Int) (*big.Int, error) {

	fmt.Printf("calculating Swap Out\nin amount %v\nin reserves %v\nout reserves %v\n", inAmount, inReserves, outReserves)
	if inAmount == nil || inReserves == nil || outReserves == nil {
		return nil, errors.New("cant calculate swap out, nil variables")
	}

	if inAmount.Cmp(big.NewInt(10000)) < 0 { // saturn dex cant process less than 5 decimals
		return nil, fmt.Errorf("in amount is too small")
	}
	outAmount := big.NewInt(0)
	// commonDecimal := outReserveDecimal - inReserveDecimal
	// var normalizedInReserves *big.Int // added this to find a bug but it is unnecessary because it turned out i made a stupid decimal error
	// var normalizedOutReserves *big.Int
	// var normalizedInAmount *big.Int

	// if commonDecimal > 0 {
	// 	normalizedInReserves = new(big.Int).Mul(inReserves, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(commonDecimal)), nil))
	// 	normalizedInAmount = new(big.Int).Mul(inAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(commonDecimal)), nil))
	// 	normalizedOutReserves = outReserves
	// } else if commonDecimal < 0 {
	// 	normalizedOutReserves = new(big.Int).Mul(outReserves, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-commonDecimal)), nil))
	// 	normalizedInReserves = inReserves
	// 	normalizedInAmount = inAmount

	// } else {
	// 	normalizedInReserves = inReserves
	// 	normalizedOutReserves = outReserves
	// 	normalizedInAmount = inAmount
	// }

	// fmt.Printf("-Normalized amounts\nin %v\nin reserve %v\nout reserve %v\n", normalizedInAmount, normalizedInReserves, normalizedOutReserves)

	inAmountMul := new(big.Int).Mul(inAmount, big.NewInt(997))
	inAmountDiv := new(big.Int).Div(inAmountMul, big.NewInt(1000))
	inAmountPlusReserves := new(big.Int).Add(inAmountDiv, inReserves)
	inReservesMulOut := new(big.Int).Mul(inReserves, outReserves)
	outAmount.Sub(outReserves, new(big.Int).Div(inReservesMulOut, inAmountPlusReserves))

	fmt.Printf("-Calculation variables\ninAmountMul %v\ninAmountDiv %v\ninAmountPlusReserves %v\ninReservesMulOut %v\n", inAmountMul, inAmountDiv, inAmountPlusReserves, inReservesMulOut)

	// // Renormalize the outAmount to the scale of inDecimals
	// if commonDecimal < 0 {
	// 	outAmount.Div(outAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-commonDecimal)), nil))
	// }

	fmt.Printf("-Calculated Swap out Amount: %s\n", outAmount.String())

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
	if outAmount == nil || inReserves == nil || outReserves == nil {
		return nil, errors.New("cant calculate swap in, nil variables")
	}

	if outAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("est out amount is zero")
	}

	if outAmount.Cmp(outReserves) >= 0 {
		return nil, errors.New("unsufficient liquidity")
	}

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

	if inAmount.Cmp(big.NewInt(10000)) < 0 { // saturn dex cant process less than 5 decimals
		return nil, fmt.Errorf("in amount is too low")
	}

	return inAmount, nil
}

/**
 * calculateSwapAndPriceImpact calculates the price impact of a swap on a DEX and returns the estimated output and input amounts.
 *
 * @param inAmount          The amount of the input token you want to swap.
 * @param outAmount         The desired amount of the output token you want to receive.
 * @param inReserves        The current reserves of the input token in the liquidity pool.
 * @param outReserves       The current reserves of the output token in the liquidity pool.
 * @param swapType          A string indicating the type of swap, either "swapOut" (input to output) or "swapIn" (output to input).
 * @return float64          The price impact as a percentage.
 * @return *big.Int         The estimated output amount for the swapOut type.
 * @return *big.Int         The estimated input amount for the swapIn type.
 * @return error            Any error encountered during the calculation.
 */
func calculateSwapAndPriceImpact(inAmount, outAmount, inReserves, outReserves *big.Int, swapType string) (float64, *big.Int, *big.Int, error) {
	fmt.Printf("*****Calculate Swap and price impact*****\nIn amount %v\noutAmount %v\nIn Reserves %v\nOut Reserves %v\nSwap Type %v\n", inAmount, outAmount, inReserves, outReserves, swapType)
	if (inAmount == nil && outAmount == nil) || inReserves == nil || outReserves == nil {
		return 0, nil, nil, errors.New("cant calculate swap and price impact")
	}
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
	currentDexFeeLimit, _ := new(big.Int).SetString(userSettings.DexBaseFeeLimit.String(), 10)
	fee := new(big.Int).Mul(currentDexFeeLimit, userSettings.GasPrice)
	err := checkFeeBalance(fee)
	var priceImpact float64
	var currentDexSlippage = userSettings.DexSlippage
	var inTokenSelected bool
	var outTokenSelected bool
	var inAmountEntryCorrect bool
	var outAmountEntryCorrect bool
	var slippageEntryCorrect bool
	var priceImpactIsNotHigh bool
	var dexTransaction []TransactionDataForDex
	var dexPayload string
	var bestRoute []string
	if err == nil {
		loadDexPools()

		slippageBinding := binding.NewString()
		slippageBinding.Set(fmt.Sprintf("%f", currentDexSlippage)) // Default slippage
		amountEntry := widget.NewEntry()
		amountEntry.SetPlaceHolder("Amount")
		amountEntry.Disable()
		routeMesssageBinding := binding.NewString()
		routeMessage := widget.NewLabelWithData(routeMesssageBinding)
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
			outAmountEntry.Disable()
			tokenOutSelect.Disable()
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
			gasFee := new(big.Int).Mul(userSettings.GasPrice, currentDexFeeLimit)
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
					err = executeSwap(dexTransaction, currentDexSlippage, creds, currentDexFeeLimit, dexPayload)
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
				inAmount, _ := convertUserInputToBigInt(amountEntry.Text, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
				fmt.Println("finding best route for", inAmount.String(), tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := findAllSwapRoutes(latestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}
				foundBestRoute, txData, impact, route, poolCount, estOutAmount, err := evaluateRoutes(swapRoutes, tokenInSelect.Selected, latestDexPools.PoolList, inAmount, currentDexSlippage, "auto")
				routeMesssageBinding.Set(route)
				bestRoute = foundBestRoute
				dexPayload = "Spallet Swap " + route
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), userSettings.DexBaseFeeLimit)
				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					max := new(big.Int).Add(inAmount, fee)
					err := checkFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					err := checkFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				}
				dexTransaction = txData
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}
				outTokendata, _ := updateOrCheckCache(tokenOutSelect.Selected, 3, "check")
				priceImpact = impact
				if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
					estOutAmountStr := formatBalance(*estOutAmount, outTokendata.Decimals)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntry.Refresh()
					checkSwapBtnState()
					return
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}
				estOutAmountStr := formatBalance(*estOutAmount, outTokendata.Decimals)
				outAmountEntry.Text = estOutAmountStr
				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
				outAmountEntry.Refresh()

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

			if tokenOutSelect.Selected != "" {
				var userSlippage float64

				userSlippage, err = strconv.ParseFloat(slippageEntry.Text, 64)
				if err != nil {
					warningMessageBinding.Set("Check your slippage")
					return fmt.Errorf("check your slippage")

				}

				outTokenSelected = true
				checkSwapBtnState()
				outAmountEntry.Enable()
				// inAmount, _ := convertUserInputToBigInt(amountEntry.Text, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
				fmt.Println("finding best route for", input, tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := findAllSwapRoutes(latestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return err
				}
				foundBestRoute, txData, impact, route, poolCount, estOutAmount, err := evaluateRoutes(swapRoutes, tokenInSelect.Selected, latestDexPools.PoolList, input, currentDexSlippage, "auto")
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return err

				}
				routeMesssageBinding.Set(route)
				bestRoute = foundBestRoute
				dexPayload = "Spallet Swap " + route
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), userSettings.DexBaseFeeLimit)

				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					max := new(big.Int).Add(input, fee)
					err := checkFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return err

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					err := checkFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return err

					}

				}

				dexTransaction = txData

				outTokendata, _ := updateOrCheckCache(tokenOutSelect.Selected, 3, "check")
				priceImpact = impact
				if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
					estOutAmountStr := formatBalance(*estOutAmount, outTokendata.Decimals)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntry.Refresh()
					checkSwapBtnState()
					return fmt.Errorf("price impact is too high")
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}
				estOutAmountStr := formatBalance(*estOutAmount, outTokendata.Decimals)
				outAmountEntry.Text = estOutAmountStr
				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
				outAmountEntry.Refresh()

			}

			if input.Cmp(big.NewInt(0)) <= 0 {
				warningMessageBinding.Set("Please enter amount")

				inAmountEntryCorrect = false
				checkSwapBtnState()
				return err
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
					slippageEntry.Validate()
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
				slippageSaveBtn.Refresh()
				checkSwapBtnState()
				return err
			} else if slippage <= 0 {
				slippageEntryCorrect = false
				slippageSaveBtn.Disable()
				slippageSaveBtn.Refresh()
				checkSwapBtnState()
				return fmt.Errorf("invalid slippage (must be greater than 0 )")
			} else if slippage > 100 {
				slippageEntryCorrect = false
				slippageSaveBtn.Disable()
				slippageSaveBtn.Refresh()
				checkSwapBtnState()
				return fmt.Errorf("invalid slippage (must be less than 100 )")
			} else {
				slippageEntryCorrect = true

				if userSettings.DexSlippage != slippage {
					currentDexSlippage = slippage
					fmt.Println("Slipage save btn enabled")
					slippageSaveBtn.Enable()

				} else {
					fmt.Println("Slipage save btn disabled")
					slippageSaveBtn.Disable()

				}
				fmt.Printf("price impact %v,current slippage %v, swapbutton state %v\n ", priceImpact, currentDexSlippage, swapBtn.Disabled())
				if priceImpact < currentDexSlippage && swapBtn.Disabled() { //for enabling swap button after user changes slippage
					amountEntry.Validate()
					outAmountEntry.Validate()
					priceImpactIsNotHigh = true

				} else if priceImpact > currentDexSlippage && !swapBtn.Disabled() {
					amountEntry.Validate()
					outAmountEntry.Validate()
					priceImpactIsNotHigh = false
				}
				slippageSaveBtn.Refresh()
				checkSwapBtnState()
				return nil
			}

		}
		swapIcon := widget.NewLabelWithStyle("🢃", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		maxBttn := widget.NewButton("Max", func() {
			if tokenInSelect.Selected == "KCAL" {
				fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
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

		// amountEntry.OnChanged = func(s string) {

		// }

		// outAmountEntry.Validator = func(s string) error {

		// }
		outAmountEntry.OnChanged = func(s string) {
			fmt.Println("Out amount Changed to ", s)
			outTokenData, _ := updateOrCheckCache(tokenOutSelect.Selected, 3, "check")
			outAmount, err := convertUserInputToBigInt(s, outTokenData.Decimals)
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
			outAmountEntryCorrect = true
			if tokenOutSelect.Selected != "" {

				inAmount, err := reverseCalculateInputAmounts(bestRoute, outAmount, latestDexPools.PoolList)
				if err != nil {
					warningMessageBinding.Set(err.Error())
					outAmountEntryCorrect = false
					checkSwapBtnState()
					return
				}

				inTokenData, _ := updateOrCheckCache(tokenInSelect.Selected, 3, "check")
				amountEntry.Text = formatBalance(*inAmount, inTokenData.Decimals)
				amountEntry.Refresh()

				fmt.Println("finding best route for", inAmount, tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := findAllSwapRoutes(latestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}
				foundBestRoute, txData, impact, route, poolCount, _, err := evaluateRoutes(swapRoutes, tokenInSelect.Selected, latestDexPools.PoolList, inAmount, currentDexSlippage, "auto")
				routeMesssageBinding.Set(route)
				bestRoute = foundBestRoute
				dexPayload = "Spallet Swap " + route
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), userSettings.DexBaseFeeLimit)

				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					max := new(big.Int).Add(inAmount, fee)
					err := checkFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, defaultSettings.GasPrice)
					err := checkFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				}
				tokenBalance := latestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
				if inAmount.Cmp(&tokenBalance) > 0 {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set("Dont have enough balance")
					return
				}

				dexTransaction = txData
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}

				priceImpact = impact
				var userSlippage float64

				userSlippage, err = strconv.ParseFloat(slippageEntry.Text, 64)
				if err != nil {
					warningMessageBinding.Set("Check your slippage")
					return

				}

				if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))

					checkSwapBtnState()
					return
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}

				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))

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
			routeMessage,
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

func executeSwap(route []TransactionDataForDex, slippageTolerance float64, creds Credentials, dexGasLimit *big.Int, payload string) error {
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

	// Convert slippage to basis points (multiply by 100 to get integer)
	slippageBasisPoints := new(big.Int).SetInt64(int64(slippageTolerance * 100))

	swapPayload := []byte(payload)
	sb := scriptbuilder.BeginScript()
	sb.AllowGas(wallet.Address, cryptography.NullAddress().String(), userSettings.GasPrice, dexGasLimit)

	for _, tx := range route {
		// Debug print script parameters
		fmt.Printf("\nConstructing SATRN.swap parameters for pool %s:\n", tx.Pool)
		fmt.Printf("1. from: %s\n", wallet.Address)
		fmt.Printf("2. amountIn: %s (%s)\n", tx.AmountIn.String(), tx.TokenIn)
		fmt.Printf("3. tokenIn: %s\n", tx.TokenIn)
		fmt.Printf("4. tokenOut: %s\n", tx.TokenOut)
		fmt.Printf("5. slippageTolerance: %d basis points\n", slippageBasisPoints)

		sb.CallContract("SATRN", "swap",
			wallet.Address,      // from
			tx.AmountIn,         // amountIn
			tx.TokenIn,          // tokenIn
			tx.TokenOut,         // tokenOut
			slippageBasisPoints, // slippageTolerance
		)
	}

	sb.SpendGas(wallet.Address)
	script := sb.EndScript()

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
