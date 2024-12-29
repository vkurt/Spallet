package main

import (
	"fmt"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2/dialog"
)

var updateBalanceTimeOut *time.Ticker

const updateInterval = 15 // in seconds

func autoUpdate(timeout int, creds core.Credentials) {
	if updateBalanceTimeOut != nil {
		updateBalanceTimeOut.Stop()
	}
	homeGui(creds)
	nftGui()
	hodlGui(creds)
	swapGui(creds)
	historyTabGui(creds.Wallets[creds.LastSelectedWallet].Address, 1, 25)
	updateBalanceTimeOut = time.NewTicker(time.Duration(timeout) * time.Second)
	core.LatestAccountData.IsBalanceUpdated = false
	go func() {

		for range updateBalanceTimeOut.C {
			fmt.Println("****Auto Update Balances*****")
			err := core.DataFetch(creds, rootPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("an error happened during auto data fetch,\n %v", err), mainWindow)

			}
			if core.LatestAccountData.IsBalanceUpdated {
				homeGui(creds)
				nftGui()
				hodlGui(creds)
				swapGui(creds)
				historyTabGui(creds.Wallets[creds.LastSelectedWallet].Address, 1, 25)
				core.LatestAccountData.IsBalanceUpdated = false

			}

		}
	}()
}
