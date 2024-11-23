package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

func createDexContent(creds Credentials) *fyne.Container {
	amountInBinding := binding.NewString()
	slippageBinding := binding.NewString()
	slippageBinding.Set("0.5") // Default slippage

	tokenInSelect := widget.NewSelect([]string{"SOUL", "KCAL"}, nil)
	tokenOutSelect := widget.NewSelect([]string{"SOUL", "KCAL"}, nil)

	amountEntry := widget.NewEntryWithData(amountInBinding)
	amountEntry.SetPlaceHolder("Amount")

	slippageEntry := widget.NewEntryWithData(slippageBinding)
	slippageEntry.SetPlaceHolder("Slippage Tolerance %")

	swapBtn := widget.NewButton("Swap Tokens", func() {
		if tokenInSelect.Selected == "" || tokenOutSelect.Selected == "" {
			dialog.ShowError(fmt.Errorf("please select tokens"), mainWindowGui)
			return
		}

		amountStr, _ := amountInBinding.Get()
		slippageStr, _ := slippageBinding.Get()

		// Parse and validate amount
		amount, err := convertUserInputToBigInt(amountStr, latestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid amount: %v", err), mainWindowGui)
			return
		}

		// Verify sufficient balance
		token := latestAccountData.FungibleTokens[tokenInSelect.Selected]
		if amount.Cmp(&token.Amount) > 0 {
			dialog.ShowError(fmt.Errorf("insufficient %s balance", tokenInSelect.Selected), mainWindowGui)
			return
		}

		slippage, err := strconv.ParseFloat(slippageStr, 64)
		if err != nil || slippage <= 0 || slippage > 100 {
			dialog.ShowError(fmt.Errorf("invalid slippage (must be between 0 and 100)"), mainWindowGui)
			return
		}

		// Check KCAL for gas
		gasFee := new(big.Int).Mul(userSettings.GasPrice, userSettings.DefaultGasLimit)
		if err := checkFeeBalance(gasFee); err != nil {
			dialog.ShowError(err, mainWindowGui)
			return
		}

		// Confirm swap
		confirmMessage := fmt.Sprintf("Swap %s %s for %s\nSlippage: %.1f%%\nGas Fee: %s KCAL",
			formatBalance(*amount, token.Decimals),
			tokenInSelect.Selected,
			tokenOutSelect.Selected,
			slippage,
			formatBalance(*gasFee, kcalDecimals))

		dialog.ShowConfirm("Confirm Swap", confirmMessage, func(confirmed bool) {
			if confirmed {
				err = executeSwap(tokenInSelect.Selected, tokenOutSelect.Selected, amount, slippage, creds)
				if err != nil {
					dialog.ShowError(err, mainWindowGui)
				}
			}
		}, mainWindowGui)
	})

	switchBtn := widget.NewButton("⇅", func() {
		tokenIn := tokenInSelect.Selected
		tokenOut := tokenOutSelect.Selected
		tokenInSelect.SetSelected(tokenOut)
		tokenOutSelect.SetSelected(tokenIn)
	})
	inTokenSelect := container.NewHBox(widget.NewLabel("From\t"), tokenInSelect)
	inTokenLyt := container.NewBorder(nil, nil, inTokenSelect, nil, amountEntry)

	outAmount := widget.NewEntry()
	outAmount.Disable()
	outAmount.SetPlaceHolder("in development(will show out amount)")
	outTokenSelect := container.NewHBox(widget.NewLabel("To\t"), tokenOutSelect)
	outTokenLyt := container.NewBorder(nil, nil, outTokenSelect, nil, outAmount)
	form := container.NewVBox(
		inTokenLyt,

		switchBtn,
		outTokenLyt,
		widget.NewLabel("Slippage Tolerance (%):"),
		slippageEntry,
		swapBtn,
		widget.NewRichTextFromMarkdown("Powered by [Saturn Dex](https://saturn.stellargate.io/)"),
	)

	return container.NewPadded(form)
}

func executeSwap(tokenIn, tokenOut string, amountIn *big.Int, slippageTolerance float64, creds Credentials) error {
	showUpdatingDialog()
	defer closeUpdatingDialog()

	if creds.LastSelectedWallet == "" {
		return fmt.Errorf("no wallet selected")
	}

	wallet := creds.Wallets[creds.LastSelectedWallet]
	fmt.Printf("Using wallet: %s\n", wallet.Address)

	keyPair, err := cryptography.FromWIF(wallet.WIF)
	if err != nil {
		return fmt.Errorf("invalid wallet key: %v", err)
	}

	// Format amounts
	fmt.Printf("Input amount (raw): %s\n", amountIn.String())
	fmt.Printf("Input amount (formatted): %s\n", formatBalance(*amountIn, latestAccountData.FungibleTokens[tokenIn].Decimals))

	// Convert slippage to basis points (multiply by 100 to get integer)
	slippageBasisPoints := new(big.Int).SetInt64(int64(slippageTolerance * 100))

	// Debug print script parameters
	fmt.Printf("\nConstructing SATRN.swap parameters:\n")
	fmt.Printf("1. from: %s\n", wallet.Address)
	fmt.Printf("2. amountIn: %s (%s %s)\n", amountIn.String(),
		formatBalance(*amountIn, latestAccountData.FungibleTokens[tokenIn].Decimals), tokenIn)
	fmt.Printf("3. tokenIn: %s\n", tokenIn)
	fmt.Printf("4. tokenOut: %s\n", tokenOut)
	fmt.Printf("5. slippageTolerance: %d basis points\n", slippageBasisPoints)

	// Check if we have enough balance
	balance := latestAccountData.FungibleTokens[tokenIn].Amount
	fmt.Printf("Current balance: %s %s\n",
		formatBalance(balance, latestAccountData.FungibleTokens[tokenIn].Decimals), tokenIn)

	// Set increased gas limit specifically for swap operations
	gasLimit := big.NewInt(30000)
	fmt.Printf("\nGas settings:\n")
	fmt.Printf("Price: %s\n", userSettings.GasPrice.String())
	fmt.Printf("Limit: %s (increased for swap)\n", gasLimit.String())

	swapPayload := []byte("Spallet Swap")

	sb := scriptbuilder.BeginScript()
	script := sb.AllowGas(wallet.Address, cryptography.NullAddress().String(), userSettings.GasPrice, gasLimit).
		CallContract("SATRN", "swap",
			wallet.Address,      // from
			amountIn,            // amountIn
			tokenIn,             // tokenIn
			tokenOut,            // tokenOut
			slippageBasisPoints, // slippageTolerance
		).
		SpendGas(wallet.Address).
		EndScript()

	fmt.Printf("\nGenerated script hex: %x\n", script)

	// Build and sign transaction
	expire := time.Now().UTC().Add(time.Second * 300).Unix()
	fmt.Printf("Transaction expiration: %v\n", time.Unix(expire, 0))

	tx := blockchain.NewTransaction(userSettings.NetworkName, userSettings.ChainName, script, uint32(expire), swapPayload) // Using custom payload
	tx.Sign(keyPair)

	txHex := hex.EncodeToString(tx.Bytes())
	fmt.Printf("Complete transaction hex: %s\n", txHex)

	txHash, err := client.SendRawTransaction(txHex)
	if err != nil {
		fmt.Printf("Failed to send transaction: %v\n", err)
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent with hash: %s\n", txHash)
	go monitorSwapTransaction(txHash)

	return nil
}

func monitorSwapTransaction(txHash string) {
	maxRetries := 30
	retryCount := 0
	retryDelay := time.Second * 5

	fmt.Printf("Starting transaction monitoring for hash: %s\n", txHash)

	for {
		if retryCount >= maxRetries {
			fmt.Printf("Transaction monitoring timed out after %d retries\n", maxRetries)
			dialog.ShowError(fmt.Errorf("transaction monitoring timed out. Transaction hash: %s\nPlease check the explorer manually", txHash), mainWindowGui)
			explorerURL := fmt.Sprintf("%s%s", userSettings.TxExplorerLink, txHash)
			if parsedURL, err := url.Parse(explorerURL); err == nil {
				fyne.CurrentApp().OpenURL(parsedURL)
			}
			return
		}

		fmt.Printf("Checking transaction status (attempt %d/%d)\n", retryCount+1, maxRetries)
		txResult, err := client.GetTransaction(txHash)
		if err != nil {
			fmt.Printf("Error getting transaction status: %v\n", err)
			if strings.Contains(err.Error(), "could not decode body") ||
				strings.Contains(err.Error(), "rpc call") {
				retryCount++
				time.Sleep(retryDelay)
				continue
			}

			dialog.ShowError(fmt.Errorf("failed to get transaction status: %v", err), mainWindowGui)
			return
		}

		if txResult.StateIsSuccess() {
			fmt.Printf("Transaction successful\n")
			dialog.ShowInformation("Success", fmt.Sprintf("Swap completed successfully\nTransaction: %s", txHash), mainWindowGui)
			return
		}
		if txResult.StateIsFault() {
			fmt.Printf("Transaction failed\n")
			dialog.ShowError(fmt.Errorf("swap failed\nTransaction: %s", txHash), mainWindowGui)
			return
		}

		fmt.Printf("Transaction pending, state: %s\n", txResult.State)
		retryCount++
		time.Sleep(retryDelay)
	}
}
