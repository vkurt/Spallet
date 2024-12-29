package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

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
	PoolCache     = make(map[string]CachedPoolData)
	cacheDuration = 15 * time.Second
	// poolUpdateInterval = 15 * time.Second
	mu sync.Mutex
)

type DexPools struct {
	Network      string   `json:"network"`
	Chain        string   `json:"chain"`
	PoolKeyCount int      `json:"pool_key_count"`
	PoolList     []string `json:"pool_list"`
}

var LatestDexPools DexPools

func GenerateFromList(userTokens []string, pools []string) []string {
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
func UpdatePools(rootPath string) error {

	currentPoolCount, err := getCountOfTokenPairsAndReserveKeys()
	if err != nil {
		return err
	}

	if LatestDexPools.PoolKeyCount < currentPoolCount || LatestDexPools.Chain != UserSettings.ChainName || LatestDexPools.Network != UserSettings.NetworkName {
		checkFrom := LatestDexPools.PoolKeyCount
		LatestDexPools.PoolKeyCount = currentPoolCount
		if LatestDexPools.Chain != UserSettings.ChainName || LatestDexPools.Network != UserSettings.NetworkName {
			checkFrom = 0
			LatestDexPools.PoolList = []string{}
			LatestDexPools.Chain = UserSettings.ChainName
			LatestDexPools.Network = UserSettings.NetworkName
		}

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
			responsePairKeys, err := Client.InvokeRawScript(UserSettings.ChainName, encodedScript)
			if err != nil {

				return err
			}

			fmt.Println("Result count ", len(responsePairKeys.Results))
			for i := range responsePairKeys.Results {
				poolWithKey := responsePairKeys.DecodeResults(i).AsString()
				pool := removeKey(poolWithKey)
				LatestDexPools.PoolList = append(LatestDexPools.PoolList, pool)
				fmt.Printf("added pool: %v to list\n", pool)
			}

			if remainedChecks > 1 {
				i--
			}

		}

		SaveDexPools(rootPath)
	}
	return nil
}

func removeKey(poolWithKey string) string {
	var result string
	i := strings.LastIndex(poolWithKey, "_")
	result = poolWithKey[:i]
	return result
}

func GenerateToList(fromToken string, pools []string) []string {
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

func getCountOfTokenPairsAndReserveKeys() (int, error) {
	sb := scriptbuilder.BeginScript()
	sb.CallContract("SATRN", "getCountOfTokenPairsAndReserveKeysOnList")
	script := sb.EndScript()
	encodedScript := hex.EncodeToString(script)

	response, err := Client.InvokeRawScript(UserSettings.ChainName, encodedScript)
	if err != nil {
		return 0, err
	}

	count := response.DecodeResult().AsNumber().Int64()

	fmt.Printf("Total token pairs and reserve keys listed: %v\n", count)
	return int(count), nil
}

// gets pool reserves from pool name
func GetPoolReserves(pool string, rootPath string) (Pool, error) {
	mu.Lock()
	defer mu.Unlock()

	// Check if the pool data is in the cache and is up-to-date
	if cachedData, found := PoolCache[pool]; found && time.Since(cachedData.Updated) < cacheDuration {
		return cachedData.Pool, nil
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
	responseValue, err := Client.InvokeRawScript(UserSettings.ChainName, encodedScript)
	if err != nil {
		return Pool{}, err
	}
	reserve1 := responseValue.DecodeResults(0).AsNumber()
	reserve2 := responseValue.DecodeResults(1).AsNumber()
	token1Data, _ := UpdateOrCheckTokenCache(poolReserveTokens[0], 3, "check", rootPath)
	token2Data, _ := UpdateOrCheckTokenCache(poolReserveTokens[1], 3, "check", rootPath)

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
	fmt.Printf("%v reserve: %v\n%v reserve: %v\n", poolReserveTokens[0], FormatBalance(poolData.Reserve1.Amount, poolData.Reserve1.Decimal), poolReserveTokens[1], FormatBalance(poolData.Reserve2.Amount, poolData.Reserve2.Decimal))

	// Store the fetched pool data in the cache
	PoolCache[pool] = CachedPoolData{Pool: poolData, Updated: time.Now()}

	return poolData, nil
}

/**
 * findAllSwapRoutes finds all possible swap routes between two tokens.
 *
 * @param pools           List of pool pairs in the format "TOKEN1_TOKEN2".
 * @param fromToken       The starting token for the swap.
 * @param toToken         The destination token for the swap.
 * @param directRoutesOnly Boolean flag indicating whether to find only direct routes.
 * @return allRoutes      All possible routes as a list of lists of pools.
 */
func FindAllSwapRoutes(pools []string, fromToken, toToken string, directRoutesOnly bool) ([][]string, error) {
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

		// If directRoutesOnly is true, stop searching further
		if directRoutesOnly && len(currentRoute) > 0 {
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
func ReverseCalculateInputAmounts(route []string, desiredOutAmount *big.Int, pools []string, rootPath string) (*big.Int, error) {
	currentOutAmount := new(big.Int).Set(desiredOutAmount)

	// Function to get correct pool reserves by considering the order of tokens
	getCorrectPoolReserves := func(token1, token2 string) (string, Pool, error) {
		poolKey := token1 + "_" + token2
		reverseKey := token2 + "_" + token1
		pool := Pool{}
		var err error
		if contains(pools, poolKey) {
			pool, err = GetPoolReserves(poolKey, rootPath)
		} else if contains(pools, reverseKey) {
			pool, err = GetPoolReserves(reverseKey, rootPath)
			pool.Reserve1, pool.Reserve2 = pool.Reserve2, pool.Reserve1 // Swap reserves
			poolKey = reverseKey
		}
		return poolKey, pool, err
	}

	// Iterate through the route in reverse to calculate required input amounts
	for i := len(route) - 1; i >= 0; i-- {
		pool := route[i]
		tokens := strings.Split(pool, "_")
		tokenIn, tokenOut := tokens[0], tokens[1]

		_, poolReserves, err := getCorrectPoolReserves(tokenIn, tokenOut)
		if err != nil {
			return nil, err
		}

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
func EvaluateRoutes(rootPath string, routes [][]string, fromToken string, pools []string, inAmount *big.Int, slippageTolerance float64, selectionMethod string) ([]string, []TransactionDataForDex, float64, string, int, *big.Int, error) {
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
	var routeErrors []error // Holds errors for routes
	fmt.Println("*****evaluating routes******")

	// Function to get correct pool reserves by considering the order of tokens
	getCorrectPoolReserves := func(token1, token2 string) (string, Pool, error) {
		poolKey := token1 + "_" + token2
		reverseKey := token2 + "_" + token1
		pool := Pool{}
		var err error
		if contains(pools, poolKey) {
			pool, err = GetPoolReserves(poolKey, rootPath)
			if err != nil {
				return "", Pool{}, err
			}
		} else if contains(pools, reverseKey) {
			pool, err = GetPoolReserves(reverseKey, rootPath)
			if err != nil {
				return "", Pool{}, err
			}
			pool.Reserve1, pool.Reserve2 = pool.Reserve2, pool.Reserve1 // Swap reserves
			poolKey = reverseKey
		} else {
			return "", Pool{}, fmt.Errorf("pool not found for tokens: %s, %s", token1, token2)
		}
		return poolKey, pool, nil
	}

	// Iterate through all routes to find the lowest price impact route and highest output route
	for _, route := range routes {
		var currentAmount *big.Int = new(big.Int).Set(inAmount)
		var currentTransactionData []TransactionDataForDex
		var currentRouteString []string = []string{fromToken}
		var totalPriceImpact float64 = 0
		var priceImpactFactor float64 = 1.0
		var routeError error

		for _, pool := range route {
			tokens := strings.Split(pool, "_")
			currentToken, nextToken := tokens[0], tokens[1]
			fmt.Println("trying to get correct pools for", currentToken, nextToken)
			poolKey, poolReserves, err := getCorrectPoolReserves(currentToken, nextToken)
			if err != nil {
				routeError = err
				break
			}
			priceImpact, outAmount, _, err := CalculateSwapAndPriceImpact(currentAmount, nil, poolReserves.Reserve1.Amount, poolReserves.Reserve2.Amount, "swapOut")
			if err != nil {
				routeError = err
				break
			}

			// If outAmount is nil or zero, skip this route
			if outAmount == nil || outAmount.Cmp(big.NewInt(0)) <= 0 {
				routeError = fmt.Errorf("invalid output amount for pool: %s", poolKey)
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

		// If route encountered an error, store it and skip further processing
		if routeError != nil {
			routeErrors = append(routeErrors, routeError)
			continue
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
		if len(routeErrors) > 0 {
			return nil, nil, 0, "", 0, nil, fmt.Errorf("no valid route found, errors: %v", routeErrors)
		}
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
			fmt.Println("Auto selected lowest impact route")
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

func calculateSwapOut(inAmount, inReserves, outReserves *big.Int) (*big.Int, error) {

	fmt.Printf("calculating Swap Out\nin amount %v\nin reserves %v\nout reserves %v\n", inAmount, inReserves, outReserves)
	if inAmount == nil || inReserves == nil || outReserves == nil {
		return nil, errors.New("cant calculate swap out, nil variables")
	}

	if inAmount.Cmp(big.NewInt(10000)) < 0 { // saturn dex cant process less than 5 decimals
		return nil, fmt.Errorf("in amount is too small")
	}

	// doing this for preventing rounding errors,
	// it is causing failing route swaps because in amount for next pool can be bigger than user's balance,

	pInAmount := new(big.Int).Mul(inAmount, big.NewInt(100))
	pInReserves := new(big.Int).Mul(inReserves, big.NewInt(100))
	pOutReserves := new(big.Int).Mul(outReserves, big.NewInt(100))

	outAmount := big.NewInt(0)

	inAmountMul := new(big.Int).Mul(pInAmount, big.NewInt(997))
	inAmountDiv := new(big.Int).Div(inAmountMul, big.NewInt(1000))
	inAmountPlusReserves := new(big.Int).Add(inAmountDiv, pInReserves)
	inReservesMulOut := new(big.Int).Mul(pInReserves, pOutReserves)
	outAmount.Sub(pOutReserves, new(big.Int).Div(inReservesMulOut, inAmountPlusReserves))
	outAmount.Div(outAmount, big.NewInt(100)) // for returning normal out
	fmt.Printf("-Calculation variables\ninAmountMul %v\ninAmountDiv %v\ninAmountPlusReserves %v\ninReservesMulOut %v\n", inAmountMul, inAmountDiv, inAmountPlusReserves, inReservesMulOut)

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

	pOutAmount := new(big.Int).Mul(outAmount, big.NewInt(100))
	pInReserves := new(big.Int).Mul(inReserves, big.NewInt(100))
	pOutReserves := new(big.Int).Mul(outReserves, big.NewInt(100))

	// Calculate the numerator: outAmount * inReserves * feeDenominator
	numerator := new(big.Int).Mul(pOutAmount, pInReserves)
	numerator = new(big.Int).Mul(numerator, big.NewInt(feeDenominator))

	// Calculate the denominator: (outReserves - outAmount) * feeNumerator
	denominator := new(big.Int).Sub(pOutReserves, pOutAmount)
	denominator = new(big.Int).Mul(denominator, big.NewInt(feeNumerator))

	// Calculate the input amount
	inAmount := new(big.Int).Div(numerator, denominator)

	// Adding one to handle rounding issues
	//
	inAmount.Div(inAmount, big.NewInt(100))
	inAmount = inAmount.Add(inAmount, big.NewInt(2))
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
func CalculateSwapAndPriceImpact(inAmount, outAmount, inReserves, outReserves *big.Int, swapType string) (float64, *big.Int, *big.Int, error) {
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

func ExecuteSwap(route []TransactionDataForDex, slippageTolerance float64, creds Credentials, dexGasLimit *big.Int, payload string) (string, error) {

	if creds.LastSelectedWallet == "" {
		return "", fmt.Errorf("no wallet selected")
	}

	wallet := creds.Wallets[creds.LastSelectedWallet]
	fmt.Printf("Using wallet: %s\n", wallet.Address)

	keyPair, err := cryptography.FromWIF(wallet.WIF)
	if err != nil {
		return "", fmt.Errorf("invalid wallet key: %v", err)
	}

	// Convert slippage to basis points (multiply by 100 to get integer)
	slippageBasisPoints := new(big.Int).SetInt64(int64(slippageTolerance))

	swapPayload := []byte(payload)
	sb := scriptbuilder.BeginScript()
	sb.AllowGas(wallet.Address, cryptography.NullAddress().String(), UserSettings.GasPrice, dexGasLimit)

	for _, tx := range route {
		// Debug print script parameters
		fmt.Printf("\nConstructing SATRN.swap parameters for pool %s:\n", tx.Pool)
		fmt.Printf("1. from: %s\n", wallet.Address)
		fmt.Printf("2. amountIn: %s (%s)\n", tx.AmountIn.String(), tx.TokenIn)
		fmt.Printf("3. tokenIn: %s\n", tx.TokenIn)
		fmt.Printf("4. tokenOut: %s\n", tx.TokenOut)
		fmt.Printf("5. amountOut: %s (%s)\n", tx.AmountOut.String(), tx.TokenOut)
		fmt.Printf("6. slippageTolerance: %d basis points\n", slippageBasisPoints)

		sb.CallContract("SATRN", "swap",
			wallet.Address,               // from
			tx.AmountIn,                  // amountIn
			tx.TokenIn,                   // tokenIn
			tx.TokenOut,                  // tokenOut
			slippageBasisPoints.String(), // slippageTolerance
		)
	}

	sb.SpendGas(wallet.Address)
	script := sb.EndScript()

	fmt.Printf("\nGenerated script hex: %x\n", script)

	// Build and sign transaction
	expire := time.Now().UTC().Add(time.Second * 300).Unix()
	fmt.Printf("Transaction expiration: %v\n", time.Unix(expire, 0))

	tx := blockchain.NewTransaction(UserSettings.NetworkName, UserSettings.ChainName, script, uint32(expire), swapPayload) // Using custom payload
	tx.Sign(keyPair)

	txHex := hex.EncodeToString(tx.Bytes())
	fmt.Printf("Complete transaction hex: %s\n", txHex)

	txHash, err := Client.SendRawTransaction(txHex)
	if err != nil {
		fmt.Printf("Failed to send transaction: %v\n", err)
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent with hash: %s\n", txHash)

	return txHash, nil
}
