package core

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

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
	Amount        *big.Int
	Chain         string
	Ids           []string
}

type UpdatedTokenData struct {
	ChainTokenUpdateTime int64                `json:"chain_token_update_time"`
	AccTokenUpdateTime   int64                `json:"acc_token_update_time"`
	AllTokenUpdateTime   int64                `json:"all_token_update_time"`
	Token                map[string]TokenData `json:"token"`
}

var LatestTokenData = UpdatedTokenData{Token: make(map[string]TokenData)}

func SaveTokenCache(rootPath string) error {
	filename := filepath.Join(rootPath, "data/cache/"+UserSettings.NetworkName+"token.cache")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&LatestTokenData)
	if err != nil {
		return err
	}

	return nil
}

func LoadTokenCache(rootPath string) error {
	path := filepath.Join(rootPath, "data/cache/"+UserSettings.NetworkName+"token.cache")
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("file not found err", err)
		// File doesn't exist, create with default settings
		UpdateOrCheckTokenCache("", 3, "chain", rootPath)
		err = SaveTokenCache(rootPath)
		if err != nil {
			return fmt.Errorf("failed to save default settings\n%v", err)
		}

		return nil
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&LatestTokenData)
	if err != nil {
		fmt.Println("decode err", err)
		UpdateOrCheckTokenCache("", 3, "chain", rootPath)
		err = SaveTokenCache(rootPath)
		if err != nil {
			return fmt.Errorf("failed to load Token Cache and saved default\n%v", err)
		}
	}
	return nil
}

// "accfungible" will update and save only selected accounts fungible token data
// "accnft" will update and save only selected accounts Nft token data
// "acc" will update and save only  selected accounts all token data
// "chain" will update and save only chain main tokens
// "all": will update and save all cached token data
// "check": will check if cache have this token or it will update it from chain and return data !!! this option is for retrieving token data others just for update
func UpdateOrCheckTokenCache(symbol string, retries int, updateType string, rootPath string) (TokenData, error) {
	fmt.Println("*****Updating/checking Token cache*****")
	switch updateType {

	case "accfungible":
		fmt.Println("-Updating acc Token Data")
		for _, userToken := range LatestAccountData.FungibleTokens {
			_, tokenData, err := checkChainForTokenInfo(retries, userToken.Symbol)
			if err == nil {
				LatestTokenData.Token[userToken.Symbol] = tokenData
			}

		}
		SaveTokenCache(rootPath)
		return TokenData{}, nil
	case "accnft": //will update all User token data
		fmt.Println("-Updating acc nft Data")
		for _, userToken := range LatestAccountData.NonFungible {
			_, tokenData, err := checkChainForTokenInfo(retries, userToken.Symbol)

			if err == nil {
				LatestTokenData.Token[userToken.Symbol] = tokenData
			}

		}
		SaveTokenCache(rootPath)
		return TokenData{}, nil

	case "acc":
		fmt.Println("-Updating acc Token Data")
		for _, userToken := range LatestAccountData.FungibleTokens {
			_, tokenData, err := checkChainForTokenInfo(retries, userToken.Symbol)
			if err == nil {
				LatestTokenData.Token[userToken.Symbol] = tokenData
			}
		}

		for _, userToken := range LatestAccountData.NonFungible {
			_, tokenData, err := checkChainForTokenInfo(retries, userToken.Symbol)
			if err == nil {
				LatestTokenData.Token[userToken.Symbol] = tokenData
			}

		}
		LatestTokenData.AccTokenUpdateTime = time.Now().UTC().Unix()
		SaveTokenCache(rootPath)
		return TokenData{}, nil

	case "chain": // will update only chain main tokens
		fmt.Println("-Updating Chain Token Data")
		chainTokens := []string{"SOUL", "KCAL", "CROWN"}
		for _, token := range chainTokens {
			_, tokenData, err := checkChainForTokenInfo(retries, token)
			if err == nil {
				LatestTokenData.Token[token] = tokenData
			}

		}
		LatestTokenData.ChainTokenUpdateTime = time.Now().UTC().Unix()
		SaveTokenCache(rootPath)
		return TokenData{}, nil

	case "all": //will update all cached token data
		fmt.Println("-Updating All Token Data")
		cantFindTokenData := false
		LatestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
		for _, token := range LatestTokenData.Token {
			cantFindTokenData, tokenData, _ := checkChainForTokenInfo(retries, token.Symbol)

			if cantFindTokenData {
				break
			} else {
				LatestTokenData.Token[tokenData.Symbol] = tokenData
			}

		}
		if cantFindTokenData {
			fmt.Println("!!!!!! Token cache maybe corrupted resetting to current acc's tokens!!!!!!")
			for k := range LatestTokenData.Token {
				delete(LatestTokenData.Token, k)
			}
			LatestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
			UpdateOrCheckTokenCache("", retries, "chain", rootPath)
			UpdateOrCheckTokenCache("", retries, "acc", rootPath)

		} else {
			LatestTokenData.AllTokenUpdateTime = time.Now().UTC().Unix()
			SaveTokenCache(rootPath)
		}

		return TokenData{}, nil

	case "check": //will check if cache have this token or it will get data from chain
		fmt.Println("-checking cache for Token Data")
		token, ok := LatestTokenData.Token[symbol]
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
			_, tokenData, err := checkChainForTokenInfo(retries, symbol)
			if err == nil {
				LatestTokenData.Token[symbol] = tokenData
				SaveTokenCache(rootPath)
			} else {
				return TokenData{}, fmt.Errorf("failed to save token cache, %v", err)
			}

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
		chainTokenData, err = Client.GetToken(symbol, false)
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
