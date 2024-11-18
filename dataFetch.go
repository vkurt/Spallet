package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"fyne.io/fyne/v2"
)

type TokenData struct {
	Symbol        string   `json:"symbol"`
	Name          string   `json:"name"`
	Decimals      int      `json:"decimals"`
	CurrentSupply string   `json:"currentSupply"`
	MaxSupply     string   `json:"maxSupply"`
	BurnedSupply  string   `json:"burnedSupply"`
	Address       string   `json:"address"`
	Owner         string   `json:"owner"`
	Flags         string   `json:"flags"`
	Script        string   `json:"script"`
	Series        []string `json:"series"`
	External      []string `json:"external"`
	Price         []string `json:"price"`
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
			regularTokens = []fyne.CanvasObject{}
			nftTokens = []fyne.CanvasObject{}
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

	fmt.Println("***********reading chain*************")
	err := getChainStatistics()
	if err != nil {
		return err
	}
	err = getAccountData(creds.Wallets[creds.LastSelectedWallet].Address, creds)
	if err != nil {
		return err
	}
	return nil
}

// Fetch new data from the server with retry logic
func fetchUserTokensInfoFromChain(symbol string, retries int) (TokenData, error) {

	ft, ok := latestAccountData.FungibleTokens[symbol]
	if ok {
		fmt.Println("****yay we have " + symbol + " token data in memory****")
		tokenData := TokenData{
			Symbol:        ft.Symbol,
			Name:          ft.Name,
			Decimals:      ft.Decimals,
			CurrentSupply: ft.CurrentSupply,
			MaxSupply:     ft.MaxSupply,
			BurnedSupply:  ft.BurnedSupply,
			Address:       ft.Address,
			Owner:         ft.Owner,
			Flags:         ft.Flags,
			Script:        ft.Script,
			Series:        nil, // it seems sdk returning wrong types or maybe server
			External:      nil,
			Price:         nil,
		}
		return tokenData, nil
	}

	nft, ok := latestAccountData.FungibleTokens[symbol]
	if ok {
		fmt.Println("****yay we have " + symbol + " token data in memory****")
		tokenData := TokenData{
			Symbol:        nft.Symbol,
			Name:          nft.Name,
			Decimals:      nft.Decimals,
			CurrentSupply: nft.CurrentSupply,
			MaxSupply:     nft.MaxSupply,
			BurnedSupply:  nft.BurnedSupply,
			Address:       nft.Address,
			Owner:         nft.Owner,
			Flags:         nft.Flags,
			Script:        nft.Script,
			Series:        nil, // it seems sdk returning wrong types or maybe server
			External:      nil,
			Price:         nil,
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
				Script:        chainTokenData.Script,
				Series:        nil, // it seems sdk returning wrong types or maybe server
				External:      nil,
				Price:         nil,
			}
			return tokenData, err
		}

		// Log the error and retry after a delay
		log.Printf("Error fetching tokens (attempt %d/%d): %v", i+1, retries, err)
		time.Sleep(500 * time.Millisecond) // Delay before retrying
	}
	return TokenData{}, fmt.Errorf("failed to fetch tokens after %d attempts", retries)
}

func fetchChainMainTokensFromChain() {
	crownData, _ := client.GetToken("CROWN", false)
	latestChainStatisticsData.CrownData.Address = crownData.Address
	latestChainStatisticsData.CrownData.BurnedSupply = crownData.BurnedSupply
	latestChainStatisticsData.CrownData.CurrentSupply = crownData.CurrentSupply
	latestChainStatisticsData.CrownData.Decimals = crownData.Decimals
	latestChainStatisticsData.CrownData.External = nil
	latestChainStatisticsData.CrownData.Flags = crownData.Flags
	latestChainStatisticsData.CrownData.MaxSupply = crownData.MaxSupply
	latestChainStatisticsData.CrownData.Name = crownData.Name
	latestChainStatisticsData.CrownData.Owner = crownData.Owner
	latestChainStatisticsData.CrownData.Price = nil
	latestChainStatisticsData.CrownData.Script = crownData.Script
	latestChainStatisticsData.CrownData.Series = nil
	latestChainStatisticsData.CrownData.Symbol = crownData.Symbol

	soulData, _ := client.GetToken("SOUL", false)
	latestChainStatisticsData.SoulData.Address = soulData.Address
	latestChainStatisticsData.SoulData.BurnedSupply = soulData.BurnedSupply
	latestChainStatisticsData.SoulData.CurrentSupply = soulData.CurrentSupply
	latestChainStatisticsData.SoulData.Decimals = soulData.Decimals
	latestChainStatisticsData.SoulData.External = nil
	latestChainStatisticsData.SoulData.Flags = soulData.Flags
	latestChainStatisticsData.SoulData.MaxSupply = soulData.MaxSupply
	latestChainStatisticsData.SoulData.Name = soulData.Name
	latestChainStatisticsData.SoulData.Owner = soulData.Owner
	latestChainStatisticsData.SoulData.Price = nil
	latestChainStatisticsData.SoulData.Script = soulData.Script
	latestChainStatisticsData.SoulData.Series = nil
	latestChainStatisticsData.SoulData.Symbol = soulData.Symbol

	kcalData, _ := client.GetToken("KCAL", false)
	latestChainStatisticsData.KcalData.Address = kcalData.Address
	latestChainStatisticsData.KcalData.BurnedSupply = kcalData.BurnedSupply
	latestChainStatisticsData.KcalData.CurrentSupply = kcalData.CurrentSupply
	latestChainStatisticsData.KcalData.Decimals = kcalData.Decimals
	latestChainStatisticsData.KcalData.External = nil
	latestChainStatisticsData.KcalData.Flags = kcalData.Flags
	latestChainStatisticsData.KcalData.MaxSupply = kcalData.MaxSupply
	latestChainStatisticsData.KcalData.Name = kcalData.Name
	latestChainStatisticsData.KcalData.Owner = kcalData.Owner
	latestChainStatisticsData.KcalData.Price = nil
	latestChainStatisticsData.KcalData.Script = kcalData.Script
	latestChainStatisticsData.KcalData.Series = nil
	latestChainStatisticsData.KcalData.Symbol = kcalData.Symbol
}
