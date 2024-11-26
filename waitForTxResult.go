package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
			showTxResultDialog(fmt.Sprintf("Transaction successfully minted, tx hash: %s", txResult.Hash), creds)

			break
		} else if txResult.StateIsFault() {
			fmt.Println("Transaction failed, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog(fmt.Sprintf("Transaction failed, tx hash: %s", txResult.Hash), creds)

			break
		} else if tries > 14 {
			fmt.Println("Transaction Data fetch timed out, tx hash: " + fmt.Sprint(txResult.Hash))
			currentMainDialog.Hide()
			showTxResultDialog(fmt.Sprintf("Transaction Data fetch timed out, tx hash: %s", txHash), creds)

			break
		}
		time.Sleep(250 * time.Millisecond)
		tries++
	}
}

func showTxResultDialog(result string, cred Credentials) {
	resultLabel := widget.NewLabel(result)
	resultLabel.Wrapping = fyne.TextWrapWord
	var resultDia dialog.Dialog
	updateButton := widget.NewButton("Update Balance & Return to Main", func() {
		// mainWindowGui.SetContent(resultLabel)
		dataFetch(cred)

		currentMainDialog.Hide()
		resultDia.Hide()
	})
	resultDialog := container.NewVBox(resultLabel, updateButton)
	d := dialog.NewCustomWithoutButtons("Transaction Result", resultDialog, mainWindowGui)
	d.Resize(fyne.NewSize(400, 300))
	resultDia = d
	currentMainDialog.Hide()
	resultDia.Show()
}

// SendTransaction sends a transaction and handles the result
func SendTransaction(txHex string, creds Credentials, onSuccess func(string, Credentials), onFailure func(error, Credentials)) {
	go func() {
		txHash, err := client.SendRawTransaction(txHex)
		if err != nil || util.ErrorDetect(txHash) {
			onFailure(err, creds)
		} else {
			onSuccess(txHash, creds)
		}
	}()
}

func handleSuccess(txHash string, creds Credentials) {
	waitForTxResult(txHash, creds)
}

func handleFailure(err error, creds Credentials) {
	showTxResultDialog(fmt.Sprintf("Transaction failed: %v", err), creds)
}
