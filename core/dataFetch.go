package core

import (
	"fmt"
	"time"
)

func DataFetch(creds Credentials, rootPath string) error {
	_, ok := creds.Wallets[creds.LastSelectedWallet]

	if !ok {
		if len(creds.WalletOrder) > 0 {
			creds.LastSelectedWallet = creds.WalletOrder[0]
		} else {
			return fmt.Errorf("cant find any account")
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
	if (currentUtcTime - LatestTokenData.ChainTokenUpdateTime) > 3600 { //no need to update this data frequently

		_, err = UpdateOrCheckTokenCache("", 3, "chain", rootPath)
		if err != nil {
			return err
		}
		fmt.Println("-reading chain for main tokens")

		LatestTokenData.ChainTokenUpdateTime = currentUtcTime
	}

	err = GetAccountData(creds.Wallets[creds.LastSelectedWallet].Address, creds, rootPath)
	if err != nil {
		return err
	}

	return nil
}
