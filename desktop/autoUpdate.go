package main

import (
	"fmt"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2/dialog"
)

var updateBalanceTimeOut *time.Ticker

const updateInterval = 15 // in seconds

func autoUpdate(timeout int, creds core.Credentials, rootPath string) {
	if updateBalanceTimeOut != nil {
		updateBalanceTimeOut.Stop()
	}
	buildAndShowAccInfo(creds)
	updateBalanceTimeOut = time.NewTicker(time.Duration(timeout) * time.Second)
	go func() {

		for range updateBalanceTimeOut.C {
			fmt.Println("****Auto Update Balances*****")
			err := core.DataFetch(creds, rootPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("an error happened during auto data fetch,\n %v", err), mainWindowGui)

			}

			buildAndShowAccInfo(creds)

		}
	}()
}
