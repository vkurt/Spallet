package main

import (
	"fmt"
	"math/big"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
	"github.com/phantasma-io/phantasma-go/pkg/util"
)

func waitForTxResult(txHash string, creds Credentials) {
	tries := 0
	txResult, _ := client.GetTransaction(txHash)

	for {
		if tries > 0 {
			txResult, _ = client.GetTransaction(txHash)
			fmt.Println("Tx state: " + fmt.Sprint(txResult.State))
		}
		if tries < 1 {
			time.Sleep(750 * time.Millisecond)
		} else if txResult.StateIsSuccess() {
			fmt.Println("Transaction successfully minted, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()

			showTxResultDialog("Transaction successfully minted.", creds, txResult)

			break
		} else if txResult.StateIsFault() {
			fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction failed.", creds, txResult)

			break
		} else if tries > 14 {
			fmt.Println("Transaction Data fetch timed out, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction Data fetch timed out.", creds, response.TransactionResult{Hash: txHash, Fee: "0"})

			break
		}
		time.Sleep(250 * time.Millisecond)
		tries++
	}
}

func showTxResultDialog(header string, cred Credentials, txResult response.TransactionResult) {
	fee, err := new(big.Int).SetString(txResult.Fee, 10)
	if fee == nil || !err {
		fee = big.NewInt(0)
	}
	feeStr := formatBalance(*fee, kcalDecimals)

	resultMessage := fmt.Sprintf("Tx hash:\t%v\nFee:\t\t%v Kcal", txResult.Hash, feeStr)
	resultLabel := widget.NewLabel(resultMessage)
	resultLabel.Truncation = fyne.TextTruncateEllipsis

	headerLabel := widget.NewLabelWithStyle(header, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	var resultDia dialog.Dialog
	dataFetch(cred)
	explorerBtn := widget.NewButton("Show on explorer", func() {
		explorerURL := fmt.Sprintf("%s%s", userSettings.TxExplorerLink, txResult.Hash)
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
}

// SendTransaction sends a transaction and handles the result
func SendTransaction(txHex string, creds Credentials) {
	go func() {
		txHash, err := client.SendRawTransaction(txHex)
		if err != nil || util.ErrorDetect(txHash) {
			dialog.ShowError(fmt.Errorf("an error happened during sending transaction,\n%v", err), mainWindowGui)
			if currentMainDialog != nil {
				currentMainDialog.Hide()
			}
		} else {
			waitForTxResult(txHash, creds)
		}
	}()
}
