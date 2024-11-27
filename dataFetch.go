package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

type TokenData struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Decimals      int    `json:"decimals"`
	CurrentSupply string `json:"currentSupply"`
	MaxSupply     string `json:"maxSupply"`
	BurnedSupply  string `json:"burnedSupply"`
	Address       string `json:"address"`
	Owner         string `json:"owner"`
	Flags         string `json:"flags"`

	Series   []string `json:"series"`
	External []string `json:"external"`
	Price    []string `json:"price"`
}

type AccToken struct {
	Symbol        string
	Name          string
	Decimals      int
	CurrentSupply string
	MaxSupply     string
	BurnedSupply  string
	Address       string
	Owner         string
	Flags         string
	Script        string
	Series        []string
	External      []string
	Price         []string
	Amount        big.Int
	Chain         string
	Ids           []string
}

type UpdatedTokenData struct {
	ChainTokenUpdateTime int64                `json:"chain_token_update_time"`
	AccTokenUpdateTime   int64                `json:"acc_token_update_time"`
	AllTokenUpdateTime   int64                `json:"all_token_update_time"`
	Token                map[string]TokenData `json:"token"`
}

var updateBalanceTimeOut *time.Ticker

const updateInterval = 15 // in seconds

var latestTokenData = UpdatedTokenData{Token: make(map[string]TokenData)}

func autoUpdate(timeout int, creds Credentials) {
	if updateBalanceTimeOut != nil {
		updateBalanceTimeOut.Stop()
	}
	updateBalanceTimeOut = time.NewTicker(time.Duration(timeout) * time.Second)
	go func() {

		for range updateBalanceTimeOut.C {
			fmt.Println("****Auto Update Balances*****")
			dataFetch(creds)
		}
	}()
}

func saveTokenCache() error {
	filename := "data/cache/" + userSettings.NetworkName + "token.cache"

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&latestTokenData)
	if err != nil {
		return err
	}

	return nil
}

func loadTokenCache() {
	path := "data/cache/" + userSettings.NetworkName + "token.cache"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("file not found err", err)
		// File doesn't exist, create with default settings
		updateOrCheckCache("", 3, "chain")
		err = saveTokenCache()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save default settings\n%v", err), mainWindowGui)
		}

		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&latestTokenData)
	if err != nil {
		fmt.Println("decode err", err)
		updateOrCheckCache("", 3, "chain")
		err = saveTokenCache()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load Token Cache and saved default\n%v", err), mainWindowGui)
		}
	}

}

func dataFetch(creds Credentials) error {
	_, ok := creds.Wallets[creds.LastSelectedWallet]

	if !ok {
		if len(creds.WalletOrder) > 0 {
			creds.LastSelectedWallet = creds.WalletOrder[0]
		} else {
			if currentMainDialog != nil {
				currentMainDialog.Hide()
			}
			creds.LastSelectedWallet = ""
			tokenTab.Content = container.NewVBox(widget.NewLabel("This wallet not containing any account"))
			nftTab.Content = container.NewVBox(widget.NewLabel("This wallet not containing any account"))
			latestAccountData = AccountInfoData{FungibleTokens: make(map[string]AccToken), NonFungible: make(map[string]AccToken)}
			buildAndShowAccInfo(creds) // tryinh to not crash wallet if user somehow removes all acc data
			// mainWindow(creds, regularTokens, nftTokens)
			// manageAccountsDia(creds)
			// currentMainDialog.Show()
			// dialog.ShowError(errors.New("cant find any wallet data\nrestart the wallet and enter your keys\nor paste backed up wallet data\nif you didnot backed up your Keys you lost access to assets"), mainWindowGui)

			return nil

		}

	}

	fmt.Println("******data fetch fettching data*************")
	// if fileExists("chainstats") && fileExists("accountdata") && fetchFromSaved {
	// 	fmt.Println("reading files*************")
	// 	savedChainData, err := loadChainData("chainstats")
	// 	if err != nil {
	// 		dialog.ShowInformation("Error", "Failed to load saved chain data: "+err.Error(), w)
	// 	} else {
	// 		latestChainStatisticsData = savedChainData
	// 	}

	// 	savedAccountData, err := loadAccountData("accountdata", creds.Password)
	// 	if err != nil {
	// 		dialog.ShowInformation("Error", "Failed to load saved account data: "+err.Error(), w)
	// 	} else {
	// 		latestAccountData = savedAccountData
	// 	}

	// } else {

	// }
	var err error
	var currentUtcTime = time.Now().UTC().Unix()
	if (currentUtcTime - latestTokenData.ChainTokenUpdateTime) > 3600 { //no need to update this data frequently

		_, err = updateOrCheckCache("", 3, "chain")
		if err != nil {
			return err
		}
		fmt.Println("***********reading chain*************")
		err = getChainStatistics()
		if err != nil {
			return err
		}
		latestTokenData.ChainTokenUpdateTime = currentUtcTime
	}

	err = getAccountData(creds.Wallets[creds.LastSelectedWallet].Address, creds)
	if err != nil {
		return err
	}

	return nil
}

// "accfungible" will update and save only selected accounts fungible token data
// "accnft" will update and save only selected accounts Nft token data
// "acc" will update and save only  selected accounts all token data
// "chain" will update and save only chain main tokens
// "all": will update and save all cached token data
// "check": will check if cache have this token or it will update it from chain and return data !!! this option is for retrieving token data others just for update
func updateOrCheckCache(symbol string, retries int, updateType string) (TokenData, error) {
	fmt.Println("*****Updating/checking Token cache*****")
	switch updateType {

	case "accfungible":
		fmt.Println("-Updating acc Token Data")
		for _, userToken := range latestAccountData.FungibleTokens {
			_, tokenData, _ := checkChainForTokenInfo(retries, userToken.Symbol)
			latestTokenData.Token[userToken.Symbol] = tokenData
		}
		saveTokenCache()
		return TokenData{}, nil
	case "accnft": //will update all User token data
		fmt.Println("-Updating acc nft Data")
		for _, userToken := range latestAccountData.NonFungible {
			_, tokenData, _ := checkChainForTokenInfo(retries, userToken.Symbol)
			latestTokenData.Token[userToken.Symbol] = tokenData
		}
		saveTokenCache()
		return TokenData{}, nil

	case "acc":
		fmt.Println("-Updating acc Token Data")
		for _, userToken := range latestAccountData.FungibleTokens {
			_, tokenData, _ := checkChainForTokenInfo(retries, userToken.Symbol)
			latestTokenData.Token[userToken.Symbol] = tokenData
		}

		for _, userToken := range latestAccountData.NonFungible {
			_, tokenData, _ := checkChainForTokenInfo(retries, userToken.Symbol)
			latestTokenData.Token[userToken.Symbol] = tokenData
		}
		latestTokenData.AccTokenUpdateTime = time.Now().UTC().Unix()
		saveTokenCache()
		return TokenData{}, nil

	case "chain": // will update only chain main tokens
		fmt.Println("-Updating Chain Token Data")
		chainTokens := []string{"SOUL", "KCAL", "CROWN"}
		for _, token := range chainTokens {
			_, tokenData, _ := checkChainForTokenInfo(retries, token)
			latestTokenData.Token[token] = tokenData
		}
		latestTokenData.ChainTokenUpdateTime = time.Now().UTC().Unix()
		saveTokenCache()
		return TokenData{}, nil

	case "all": //will update all cached token data
		fmt.Println("-Updating All Token Data")
		cantFindTokenData := false
		latestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
		for _, token := range latestTokenData.Token {
			cantFindTokenData, tokenData, _ := checkChainForTokenInfo(retries, token.Symbol)
			latestTokenData.Token[tokenData.Symbol] = tokenData
			if cantFindTokenData {
				break
			}

		}
		if cantFindTokenData {
			fmt.Println("!!!!!! Token cache maybe corrupted resetting to current acc's tokens!!!!!!")
			for k := range latestTokenData.Token {
				delete(latestTokenData.Token, k)
			}
			latestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
			updateOrCheckCache("", retries, "chain")
			updateOrCheckCache("", retries, "acc")

		} else {
			latestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
			saveTokenCache()
		}

		return TokenData{}, nil

	case "check": //will check if cache have this token or it will get data from chain
		fmt.Println("-checking cache for Token Data")
		token, ok := latestTokenData.Token[symbol]
		if ok {
			fmt.Println("****yay we have " + symbol + " token data in memory****")
			tokenData := TokenData{
				Symbol:        token.Symbol,
				Name:          token.Name,
				Decimals:      token.Decimals,
				CurrentSupply: token.CurrentSupply,
				MaxSupply:     token.MaxSupply,
				BurnedSupply:  token.BurnedSupply,
				Address:       token.Address,
				Owner:         token.Owner,
				Flags:         token.Flags,
				Series:        nil, // it seems sdk returning wrong types or maybe server
				External:      nil,
				Price:         nil,
			}
			return tokenData, nil
		} else {
			_, tokenData, _ := checkChainForTokenInfo(retries, symbol)
			latestTokenData.Token[symbol] = tokenData
			saveTokenCache()
		}

	}

	return TokenData{}, fmt.Errorf("failed to fetch tokens after %d attempts", retries)

}

// Fetch new data from the server with retry logic
// it will return a boolean for notifiying token data available, token data, and error
func checkChainForTokenInfo(retries int, symbol string) (bool, TokenData, error) {
	cantFindTokenData := false
	var err error
	var chainTokenData response.TokenResult
	for i := 0; i < retries; i++ {
		fmt.Println("*checking token data from chain for token "+symbol+" try ", i+1, "*")
		chainTokenData, err = client.GetToken(symbol, false)
		if err == nil {

			tokenData := TokenData{
				Symbol:        chainTokenData.Symbol,
				Name:          chainTokenData.Name,
				Decimals:      chainTokenData.Decimals,
				CurrentSupply: chainTokenData.CurrentSupply,
				MaxSupply:     chainTokenData.MaxSupply,
				BurnedSupply:  chainTokenData.BurnedSupply,
				Address:       chainTokenData.Address,
				Owner:         chainTokenData.Owner,
				Flags:         chainTokenData.Flags,

				Series:   nil, // it seems sdk returning wrong types or maybe server
				External: nil,
				Price:    nil,
			}

			fmt.Println("-found token data for", symbol)
			return cantFindTokenData, tokenData, err

		}
		// Log the error and retry after a delay
		log.Printf("Error fetching tokens (attempt %d/%d): %v", i+1, retries, err)
		time.Sleep(500 * time.Millisecond) // Delay before retrying
		if i == 2 {
			cantFindTokenData = true
			err = fmt.Errorf("can't find token data on chain after %d attempts", i+1)
			break
		}
	}
	return cantFindTokenData, TokenData{}, err
}
