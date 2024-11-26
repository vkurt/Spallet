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
	LastUpdateTime int64                `json:"last_update_time"`
	Token          map[string]TokenData `json:"token"`
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

func loadTokenCache(creds Credentials) {
	path := "data/cache/" + userSettings.NetworkName + "token.cache"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("file not found err", err)
		// File doesn't exist, create with default settings
		latestTokenData.Token["SOUL"] = TokenData{Symbol: "SOUL"} //ensuring we always have main token data
		latestTokenData.Token["KCAL"] = TokenData{Symbol: "KCAL"}
		latestTokenData.Token["CROWN"] = TokenData{Symbol: "CROWN"}
		fetchUserTokensInfoFromChain("", 3, true, creds)
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
		latestTokenData.Token["SOUL"] = TokenData{Symbol: "SOUL"} //ensuring we always have main token data
		latestTokenData.Token["KCAL"] = TokenData{Symbol: "KCAL"}
		latestTokenData.Token["CROWN"] = TokenData{Symbol: "CROWN"}
		fetchUserTokensInfoFromChain("", 3, true, creds)
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

	if (time.Now().UTC().Unix() - latestTokenData.LastUpdateTime) > 150 { //no need to update this data frequently updating it before min logout timeout
		fmt.Println("*****Updating Token Data*****")
		_, err = fetchUserTokensInfoFromChain("", 3, true, creds)
		if err != nil {
			return err
		}
		fmt.Println("***********reading chain*************")
		err = getChainStatistics()
		if err != nil {
			return err
		}
		saveTokenCache()

	}

	err = getAccountData(creds.Wallets[creds.LastSelectedWallet].Address, creds, false)
	if err != nil {
		return err
	}

	return nil
}

// Fetch new data from the server with retry logic
// if update is true it will only update token data in memory
func fetchUserTokensInfoFromChain(symbol string, retries int, update bool, creds Credentials) (TokenData, error) {
	switch update {
	case true:
		cantFindTokenData := false
		latestTokenData.LastUpdateTime = time.Now().UTC().Unix()
		for _, token := range latestTokenData.Token {
			for i := 0; i < retries; i++ {
				fmt.Println("****updating token data from chain for token "+token.Symbol+" try ", i+1, "******")
				chainTokenData, err := client.GetToken(token.Symbol, false)
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
					latestTokenData.Token[tokenData.Symbol] = tokenData
					fmt.Println("updated token", token.Symbol)
					break
				}
				// Log the error and retry after a delay
				log.Printf("Error fetching tokens (attempt %d/%d): %v", i+1, retries, err)
				time.Sleep(500 * time.Millisecond) // Delay before retrying
				if i == 2 {
					cantFindTokenData = true
					break
				}
			}
			if cantFindTokenData {
				break
			}

		}
		if cantFindTokenData {
			fmt.Println("!!!!!! Token cache maybe corrupted resetting !!!!!!")
			for k := range latestTokenData.Token {
				delete(latestTokenData.Token, k)
			}
			latestTokenData.Token["SOUL"] = TokenData{Symbol: "SOUL"} //ensuring we always have main token data
			latestTokenData.Token["KCAL"] = TokenData{Symbol: "KCAL"}
			latestTokenData.Token["CROWN"] = TokenData{Symbol: "CROWN"}
			getAccountData(creds.Wallets[creds.LastSelectedWallet].Address, creds, true)
			saveTokenCache()
		}

		return TokenData{}, nil

	case false:
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

				Series:   nil, // it seems sdk returning wrong types or maybe server
				External: nil,
				Price:    nil,
			}
			return tokenData, nil
		}

		for i := 0; i < retries; i++ {
			fmt.Println("****shit checking chain for token " + symbol + " ******")
			chainTokenData, err := client.GetToken(symbol, false)
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
				latestTokenData.Token[tokenData.Symbol] = tokenData
				return tokenData, err
			}

			// Log the error and retry after a delay
			log.Printf("Error fetching tokens (attempt %d/%d): %v", i+1, retries, err)
			time.Sleep(500 * time.Millisecond) // Delay before retrying
		}

	}

	return TokenData{}, fmt.Errorf("failed to fetch tokens after %d attempts", retries)

}
