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
			fee, _ := new(big.Int).SetString(txResult.Fee, 10)
			feeStr := formatBalance(*fee, kcalDecimals)
			showTxResultDialog("Transaction successfully minted.", fmt.Sprintf("Tx hash:\t%s\nFee:\t\t%s Kcal", txResult.Hash[:30]+"...", feeStr), creds, txHash)

			break
		} else if txResult.StateIsFault() {
			fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction failed.", fmt.Sprintf("tx hash: %s", txResult.Hash[:30]+"..."), creds, txHash)

			break
		} else if tries > 14 {
			fmt.Println("Transaction Data fetch timed out, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog("Transaction Data fetch timed out.", fmt.Sprintf("tx hash: %s", txHash[:30]+"..."), creds, txHash)

			break
		}
		time.Sleep(250 * time.Millisecond)
		tries++
	}
}

func showTxResultDialog(header string, result string, cred Credentials, txHash string) {
	resultLabel := widget.NewLabel(result)
	headerLabel := widget.NewLabelWithStyle(header, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	resultLabel.Wrapping = fyne.TextWrapWord
	var resultDia dialog.Dialog
	dataFetch(cred)
	explorerBtn := widget.NewButton("Show on explorer", func() {
		explorerURL := fmt.Sprintf("%s%s", userSettings.TxExplorerLink, txHash)
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

	d := dialog.NewCustomWithoutButtons("Transaction Result", resultLyt, mainWindowGui)
	d.Resize(fyne.NewSize(400, 225))
	resultDia = d
	currentMainDialog.Hide()
	resultDia.Show()
}

// SendTransaction sends a transaction and handles the result
func SendTransaction(txHex string, creds Credentials, onSuccess func(string, Credentials), onFailure func(error, Credentials, string)) {
	go func() {
		txHash, err := client.SendRawTransaction(txHex)
		if err != nil || util.ErrorDetect(txHash) {
			onFailure(err, creds, txHash)
		} else {
			onSuccess(txHash, creds)
		}
	}()
}

func handleSuccess(txHash string, creds Credentials) {
	waitForTxResult(txHash, creds)
}

func handleFailure(err error, creds Credentials, txHash string) {
	showTxResultDialog("Transaction failed.", fmt.Sprintf("%v", err), creds, txHash)
}
