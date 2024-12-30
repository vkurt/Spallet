package main

import (
	"fmt"
	"math/big"
	"net/url"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func waitForTxResult(txHash string, creds core.Credentials, txCount int) {
	tries := 0
	txResult, _ := core.Client.GetTransaction(txHash)

	for {
		if tries > 0 {
			txResult, _ = core.Client.GetTransaction(txHash)
			fmt.Println("Tx state: " + fmt.Sprint(txResult.State))
		}
		if tries < 1 {
			time.Sleep(750 * time.Millisecond)
		} else if txResult.StateIsSuccess() {
			fmt.Println("Transaction successfully minted, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()

			showTxResultDialog("Transaction successfully minted.", creds, txResult, txCount)

			break
		} else if txResult.StateIsFault() {
			fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction failed.", creds, txResult, txCount)

			break
		} else if tries > 14 {
			fmt.Println("Transaction Data fetch timed out, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction Data fetch timed out.", creds, response.TransactionResult{Hash: txHash, Fee: "0"}, txCount)

			break
		}
		time.Sleep(250 * time.Millisecond)
		tries++
	}
}

func showTxResultDialog(header string, creds core.Credentials, txResult response.TransactionResult, txCount int) {
	fee, err := new(big.Int).SetString(txResult.Fee, 10)
	if fee == nil || !err {
		fee = big.NewInt(0)
	}
	feeStr := core.FormatBalance(fee, core.KcalDecimals)

	resultMessage := fmt.Sprintf("Tx hash:\t%v\nFee:\t\t%v Kcal", txResult.Hash, feeStr)
	resultLabel := widget.NewLabel(resultMessage)
	resultLabel.Truncation = fyne.TextTruncateEllipsis

	headerLabel := widget.NewLabelWithStyle(header, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	var resultDia dialog.Dialog

	explorerBtn := widget.NewButton("Show on explorer", func() {
		explorerURL := fmt.Sprintf("%s%s", core.UserSettings.TxExplorerLink, txResult.Hash)
		if parsedURL, err := url.Parse(explorerURL); err == nil {
			fyne.CurrentApp().OpenURL(parsedURL)
		}

	})
	closeBtn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
		resultDia.Hide()
		if currentMainDialog != nil {
			currentMainDialog.Hide()
		}

	})
	btns := container.NewVBox(explorerBtn, closeBtn)
	resultLyt := container.NewBorder(headerLabel, btns, nil, nil, resultLabel)

	resultDia = dialog.NewCustomWithoutButtons("Transaction Result", resultLyt, mainWindowGui)
	resultDia.Resize(fyne.NewSize(400, 225))
	if currentMainDialog != nil {
		currentMainDialog.Hide()
	}
	resultDia.Show()
	closeUpdatingDialog()
	if txResult.StateIsSuccess() {
		for i := 0; i < 40; i++ {
			fmt.Println("Checking tx count")
			core.DataFetch(creds, rootPath)
			time.Sleep(time.Millisecond * 250)
			if txCount < core.LatestAccountData.TransactionCount {
				autoUpdate(updateInterval, creds, rootPath)
				break
			}
		}
	} else {

		time.Sleep(time.Second * 2)
		core.DataFetch(creds, rootPath)
		autoUpdate(updateInterval, creds, rootPath)
	}

}

// SendTransaction sends a transaction and handles the result
func sendTransaction(txHex string, creds core.Credentials) {
	txCount := core.LatestAccountData.TransactionCount
	go func() {
		txHash, err := core.Client.SendRawTransaction(txHex)
		if err != nil || util.ErrorDetect(txHash) {
			dialog.ShowError(fmt.Errorf("an error happened during sending transaction,\n%v", err), mainWindowGui)
			if currentMainDialog != nil {
				currentMainDialog.Hide()
			}
		} else {
			waitForTxResult(txHash, creds, txCount)
		}
	}()
}
