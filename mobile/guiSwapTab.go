package main

import (
	"fmt"
	"math/big"
	"spallet/core"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

func swapGui(creds core.Credentials) {
	currentDexBaseFeeLimit := new(big.Int).Set(core.UserSettings.DexBaseFeeLimit)
	currentDexFeeLimit := new(big.Int).Set(currentDexBaseFeeLimit)
	fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
	err := core.CheckFeeBalance(fee)
	var priceImpact float64
	var currentDexSlippage = core.UserSettings.DexSlippage
	var inTokenSelected bool
	var outTokenSelected bool
	var inAmountEntryCorrect bool
	var outAmountEntryCorrect bool
	var priceImpactIsNotHigh bool
	var dexTransaction []core.TransactionDataForDex
	var dexPayload string
	var bestRoute []string
	var currentDexRouteEvaluation = core.UserSettings.DexRouteEvaluation
	var currentOnlyDirectRoute = core.UserSettings.DexDirectRoute
	if err == nil {
		core.LoadDexPools(rootPath)

		slippageBinding := binding.NewString()
		slippageBinding.Set(fmt.Sprintf("%.2f", currentDexSlippage)) // Default slippage
		amountEntry := widget.NewEntry()
		amountEntry.SetPlaceHolder("Amount")
		amountEntry.Disable()
		routeMesssageBinding := binding.NewString()
		routeMessage := widget.NewLabelWithData(routeMesssageBinding)
		warningMessageBinding := binding.NewString()
		warningMessageBinding.Set("Please select in token")
		warningMessage := widget.NewLabelWithData(warningMessageBinding)
		warningMessage.Truncation = fyne.TextTruncateEllipsis
		outAmountEntry := widget.NewEntry()
		var userTokens []string
		for _, token := range core.LatestAccountData.FungibleTokens {
			userTokens = append(userTokens, token.Symbol)

		}
		err := core.UpdatePools(rootPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("an error happened,%v", err), mainWindow)
			errorMessage := widget.NewLabel(fmt.Sprintf("An error happened\n%v", err))
			errorMessage.Wrapping = fyne.TextWrapWord
			swapTab.Content = errorMessage
		}

		fmt.Println("user token count, pool count", len(userTokens), len(core.LatestDexPools.PoolList))
		fromList := core.GenerateFromList(userTokens, core.LatestDexPools.PoolList)
		tokenOutSelect := widget.NewSelect([]string{}, nil)
		tokenOutSelect.Disable()
		outAmountEntry.Disable()
		tokenInSelect := widget.NewSelect(fromList, nil)
		tokenInSelect.OnChanged = func(s string) {
			inTokenSelected = true
			toList := core.GenerateToList(s, core.LatestDexPools.PoolList)
			tokenOutSelect.ClearSelected()
			amountEntry.SetText("")
			tokenOutSelect.SetOptions(toList)
			amountEntry.Enable()
			outAmountEntry.SetText("")
			mainWindow.Canvas().Focus(amountEntry)
			warningMessageBinding.Set("Please enter amount")
			outAmountEntry.Disable()
			tokenOutSelect.Disable()
		}

		swapBtn := widget.NewButton("Swap Tokens", func() {
			if tokenInSelect.Selected == "" || tokenOutSelect.Selected == "" {
				dialog.ShowError(fmt.Errorf("please select tokens"), mainWindow)
				return
			}

			amountStr := amountEntry.Text
			slippageStr, _ := slippageBinding.Get()

			// Parse and validate amount
			amount, err := core.ConvertUserInputToBigInt(amountStr, core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid amount: %v", err), mainWindow)
				return
			}

			// Verify sufficient balance
			token := core.LatestAccountData.FungibleTokens[tokenInSelect.Selected]
			if amount.Cmp(token.Amount) > 0 {
				dialog.ShowError(fmt.Errorf("insufficient %s balance", tokenInSelect.Selected), mainWindow)
				return
			}

			slippage, err := strconv.ParseFloat(slippageStr, 64)
			if err != nil || slippage <= 0 || slippage > 100 {
				dialog.ShowError(fmt.Errorf("invalid slippage (must be between 0 and 100)"), mainWindow)
				return
			}

			// Check KCAL for gas
			gasFee := new(big.Int).Mul(core.UserSettings.GasPrice, currentDexFeeLimit)
			if err := core.CheckFeeBalance(gasFee); err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}
			routeWarning := ""
			if len(dexTransaction) > 1 {
				routeWarning = "⚠️During this swap, Spallet Routing is used.⚠️\n⚠️You might have some leftover tokens from the route pools.⚠️\n☣️If your swap fails try tweaking amount or get some routing tokens.☣️\n\n"
			}
			// Confirm swap
			confirmMessage := fmt.Sprintf("%sSwap %s %s for estimated %s %s\nPrice Impact %.2f%% (or price increase for %s)\nSlippage: %.1f%%\nGas Fee: %s KCAL",
				routeWarning,
				core.FormatBalance(amount, token.Decimals),
				tokenInSelect.Selected,
				outAmountEntry.Text,
				tokenOutSelect.Selected,
				priceImpact,
				tokenOutSelect.Selected,
				slippage,
				core.FormatBalance(gasFee, core.KcalDecimals))

			dialog.ShowConfirm("Confirm Swap", confirmMessage, func(confirmed bool) {
				if confirmed {
					txHash, err := core.ExecuteSwap(dexTransaction, currentDexSlippage, creds, currentDexFeeLimit, dexPayload)
					if err != nil {
						dialog.ShowError(err, mainWindow)
					} else {
						go monitorSwapTransaction(txHash, creds)
					}
				}
			}, mainWindow)
		})
		swapBtn.Disable()
		checkSwapBtnState := func() {
			fmt.Println("Swap button State", inTokenSelected, outTokenSelected, inAmountEntryCorrect, outAmountEntryCorrect, priceImpactIsNotHigh)
			if inTokenSelected && outTokenSelected && inAmountEntryCorrect && outAmountEntryCorrect && priceImpactIsNotHigh {
				swapBtn.Enable()
			} else {
				swapBtn.Disable()
			}

		}
		tokenOutSelect.OnChanged = func(s string) {

			if s != "" {

				slippageStr, _ := slippageBinding.Get()
				userSlippage, _ := strconv.ParseFloat(slippageStr, 64)

				outTokenSelected = true
				checkSwapBtnState()
				outAmountEntry.Enable()
				inAmount, _ := core.ConvertUserInputToBigInt(amountEntry.Text, core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
				fmt.Println("finding best route for", inAmount.String(), tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := core.FindAllSwapRoutes(core.LatestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected, currentOnlyDirectRoute)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					routeMesssageBinding.Set("")
					return
				}
				foundBestRoute, txData, impact, route, poolCount, estOutAmount, err := core.EvaluateRoutes(rootPath, swapRoutes, tokenInSelect.Selected, core.LatestDexPools.PoolList, inAmount, currentDexSlippage, currentDexRouteEvaluation)

				bestRoute = foundBestRoute
				dexPayload = mainPayload + " Swap " + route
				var routeFee *big.Int
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), currentDexBaseFeeLimit)
				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					max := new(big.Int).Add(inAmount, fee)
					err := core.CheckFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					err := core.CheckFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				}
				routeMesssageBinding.Set(fmt.Sprintf("%v, required Kcal for this route %s", route, core.FormatBalance(routeFee, core.KcalDecimals)))
				dexTransaction = txData
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}
				outTokendata, _ := core.UpdateOrCheckTokenCache(tokenOutSelect.Selected, 3, "check", rootPath)
				priceImpact = impact
				if userSlippage == 99 {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}
				if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
					estOutAmountStr := core.FormatBalance(estOutAmount, outTokendata.Decimals)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntry.Refresh()
					checkSwapBtnState()
					return
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}
				estOutAmountStr := core.FormatBalance(estOutAmount, outTokendata.Decimals)
				outAmountEntry.Text = estOutAmountStr
				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
				outAmountEntry.Refresh()

			}
		}
		amountEntry.Validator = func(s string) error {
			input, err := core.ConvertUserInputToBigInt(s, core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals)
			if err != nil {
				warningMessageBinding.Set(err.Error())

				inAmountEntryCorrect = false
				checkSwapBtnState()
				return err
			}

			if tokenOutSelect.Selected != "" {
				slippageStr, _ := slippageBinding.Get()
				userSlippage, _ := strconv.ParseFloat(slippageStr, 64)

				outTokenSelected = true
				checkSwapBtnState()
				outAmountEntry.Enable()
				fmt.Println("finding best route for", input, tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := core.FindAllSwapRoutes(core.LatestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected, currentOnlyDirectRoute)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					routeMesssageBinding.Set("")
					return err
				}
				foundBestRoute, txData, impact, route, poolCount, estOutAmount, err := core.EvaluateRoutes(rootPath, swapRoutes, tokenInSelect.Selected, core.LatestDexPools.PoolList, input, currentDexSlippage, currentDexRouteEvaluation)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return err

				}

				var routeFee *big.Int
				bestRoute = foundBestRoute
				dexPayload = mainPayload + " Swap " + route
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), currentDexBaseFeeLimit)

				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					max := new(big.Int).Add(input, fee)
					err := core.CheckFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return err

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					err := core.CheckFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return err

					}

				}

				routeMesssageBinding.Set(fmt.Sprintf("%v, required Kcal for this route %s", route, core.FormatBalance(routeFee, core.KcalDecimals)))

				dexTransaction = txData

				outTokendata, _ := core.UpdateOrCheckTokenCache(tokenOutSelect.Selected, 3, "check", rootPath)
				priceImpact = impact

				if userSlippage == 99 {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				} else if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))
					estOutAmountStr := core.FormatBalance(estOutAmount, outTokendata.Decimals)
					outAmountEntry.Text = (estOutAmountStr)
					outAmountEntry.Refresh()
					checkSwapBtnState()
					return fmt.Errorf("price impact is too high")
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}
				estOutAmountStr := core.FormatBalance(estOutAmount, outTokendata.Decimals)
				outAmountEntry.Text = estOutAmountStr
				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))
				outAmountEntry.Refresh()

			}

			if input.Cmp(big.NewInt(0)) <= 0 {
				warningMessageBinding.Set("Please enter amount")

				inAmountEntryCorrect = false
				checkSwapBtnState()
				return err
			}

			balance := core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
			if input.Cmp(balance) <= 0 {
				if tokenOutSelect.Selected == "" {
					warningMessageBinding.Set("Please select out token")
					inAmountEntryCorrect = false

				} else {
					inAmountEntryCorrect = true
				}
				checkSwapBtnState()
				tokenOutSelect.Enable()
				return nil
			} else {

				inAmountEntryCorrect = false
				checkSwapBtnState()
				warningMessageBinding.Set("Balance is not sufficent")
				return fmt.Errorf("balance is not sufficent")
			}
		}
		swapIconContainer := canvas.NewImageFromResource(theme.MoveDownIcon())
		swapIconContainer.SetMinSize(fyne.NewSize(32, 32))
		swapIcon := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), container.NewHBox(swapIconContainer), layout.NewSpacer())
		maxBttn := widget.NewButton("Max", func() {
			if tokenInSelect.Selected == "KCAL" {
				fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
				kcalbal := core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
				max := new(big.Int).Sub(kcalbal, fee)
				if max.Cmp(big.NewInt(0)) >= 0 {
					amountEntry.SetText(core.FormatBalance(max, core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals))
					amountEntry.Validate()
					amountEntry.Refresh()
				} else {
					amountEntry.SetText("Not enough Kcal")
					amountEntry.Validate()
					amountEntry.Refresh()
				}

			} else {
				amountEntry.SetText(core.FormatBalance(core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Amount, core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Decimals))
				amountEntry.Validate()
				amountEntry.Refresh()
			}
		})
		inTokenSelect := container.NewHBox(widget.NewLabel("From\t"), tokenInSelect)
		inTokenLyt := container.NewBorder(nil, nil, inTokenSelect, maxBttn, amountEntry)

		outAmountEntry.OnChanged = func(s string) {
			fmt.Println("Out amount Changed to ", s)
			outTokenData, _ := core.UpdateOrCheckTokenCache(tokenOutSelect.Selected, 3, "check", rootPath)
			outAmount, err := core.ConvertUserInputToBigInt(s, outTokenData.Decimals)
			if err != nil {
				warningMessageBinding.Set(err.Error())
				outAmountEntryCorrect = false
				checkSwapBtnState()
				return
			}
			if outAmount.Cmp(big.NewInt(0)) <= 0 {
				warningMessageBinding.Set("In amount is too small")
				outAmountEntryCorrect = false
				checkSwapBtnState()
				return
			}
			outAmountEntryCorrect = true
			if tokenOutSelect.Selected != "" {

				inAmount, err := core.ReverseCalculateInputAmounts(bestRoute, outAmount, core.LatestDexPools.PoolList, rootPath)
				if err != nil {
					warningMessageBinding.Set(err.Error())
					outAmountEntryCorrect = false
					checkSwapBtnState()
					return
				}

				inTokenData, _ := core.UpdateOrCheckTokenCache(tokenInSelect.Selected, 3, "check", rootPath)
				amountEntry.Text = core.FormatBalance(inAmount, inTokenData.Decimals)
				amountEntry.Refresh()

				fmt.Println("finding best route for", inAmount, tokenInSelect.Selected, tokenOutSelect.Selected)
				swapRoutes, err := core.FindAllSwapRoutes(core.LatestDexPools.PoolList, tokenInSelect.Selected, tokenOutSelect.Selected, currentOnlyDirectRoute)
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					routeMesssageBinding.Set("")
					return
				}
				foundBestRoute, txData, impact, route, poolCount, _, err := core.EvaluateRoutes(rootPath, swapRoutes, tokenInSelect.Selected, core.LatestDexPools.PoolList, inAmount, currentDexSlippage, currentDexRouteEvaluation)

				var routeFee *big.Int
				bestRoute = foundBestRoute
				dexPayload = mainPayload + " Swap " + route
				currentDexFeeLimit.Mul(big.NewInt(int64(poolCount)), currentDexBaseFeeLimit)

				if tokenInSelect.Selected == "KCAL" {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					max := new(big.Int).Add(inAmount, fee)
					err := core.CheckFeeBalance(max)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				} else {
					fee := new(big.Int).Mul(currentDexFeeLimit, core.UserSettings.GasPrice)
					routeFee = fee
					err := core.CheckFeeBalance(fee)
					if err != nil {
						inAmountEntryCorrect = false
						checkSwapBtnState()
						warningMessageBinding.Set("Not enough Kcal")
						return

					}

				}

				routeMesssageBinding.Set(fmt.Sprintf("%v, required Kcal for this route %s", route, core.FormatBalance(routeFee, core.KcalDecimals)))

				tokenBalance := core.LatestAccountData.FungibleTokens[tokenInSelect.Selected].Amount
				if inAmount.Cmp(tokenBalance) > 0 {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set("Dont have enough balance")
					return
				}

				dexTransaction = txData
				if err != nil {
					outAmountEntryCorrect = false
					checkSwapBtnState()
					warningMessageBinding.Set(err.Error())
					return
				}

				priceImpact = impact
				slippageStr, _ := slippageBinding.Get()
				userSlippage, _ := strconv.ParseFloat(slippageStr, 64)

				if userSlippage == 99 {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				} else if priceImpact > userSlippage {
					priceImpactIsNotHigh = false
					warningMessageBinding.Set(fmt.Sprintf("Price impact is too high, current price impact %.2f%%", priceImpact))

					checkSwapBtnState()
					return
				} else {
					priceImpactIsNotHigh = true
					inAmountEntryCorrect = true
					checkSwapBtnState()
				}

				outAmountEntryCorrect = true
				checkSwapBtnState()
				warningMessageBinding.Set(fmt.Sprintf("You can swap from %s to %s with price impact %.2f%%", tokenInSelect.Selected, tokenOutSelect.Selected, impact))

			}

		}

		outAmountEntry.SetPlaceHolder("Estimated out amount")
		outTokenSelect := container.NewHBox(widget.NewLabel("To\t"), tokenOutSelect)
		outTokenLyt := container.NewBorder(nil, nil, outTokenSelect, nil, outAmountEntry)
		settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			var settingsDia dialog.Dialog
			var selectedRouteEvaluation = currentDexRouteEvaluation
			var selectedOnlyDirectRoute = currentOnlyDirectRoute
			var enteredDexBaseFeeLimit = new(big.Int).Set(currentDexBaseFeeLimit)
			var enteredSlippage = currentDexSlippage
			dexBaseFeeLimitEntry := widget.NewEntry()
			dexBaseFeeLimitEntry.SetText(currentDexBaseFeeLimit.String())
			slippageEntry := widget.NewEntryWithData(slippageBinding)
			slippageEntry.SetPlaceHolder("Slippage Tolerance %")

			routingOptions := widget.NewRadioGroup([]string{"Auto", "Direct routes only", "Lowest price impact", "Lowest price"}, func(s string) {

			})
			routingOptions.Required = true
			switch currentDexRouteEvaluation {
			case "auto":
				if currentOnlyDirectRoute {
					routingOptions.SetSelected("Direct routes only")
				} else {
					routingOptions.SetSelected("Auto")
				}
			case "lowestImpact":
				routingOptions.SetSelected("Lowest price impact")
			case "highestOutput":
				routingOptions.SetSelected("Lowest price")
			}

			settingsForm := widget.NewForm(
				widget.NewFormItem("Dex Base Fee Limit", dexBaseFeeLimitEntry),
				widget.NewFormItem("Slippage Tolerance (%)", slippageEntry),
				widget.NewFormItem("Routing Options", routingOptions),
			)
			var applyBtn *widget.Button
			closeBtn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { settingsDia.Hide() })
			applyBtn = widget.NewButton("Apply", func() {
				currentDexBaseFeeLimit = enteredDexBaseFeeLimit
				currentDexRouteEvaluation = selectedRouteEvaluation
				currentOnlyDirectRoute = selectedOnlyDirectRoute
				amountEntry.Validate()
				outAmountEntry.Validate()
				applyBtn.Disable()
			})
			var applySave *widget.Button
			applySave = widget.NewButton("Apply & Save", func() {
				currentDexBaseFeeLimit = enteredDexBaseFeeLimit
				currentDexRouteEvaluation = selectedRouteEvaluation
				currentOnlyDirectRoute = selectedOnlyDirectRoute
				amountEntry.Validate()
				outAmountEntry.Validate()
				core.UserSettings.DexRouteEvaluation = selectedRouteEvaluation
				core.UserSettings.DexDirectRoute = selectedOnlyDirectRoute
				core.UserSettings.DexBaseFeeLimit = new(big.Int).Set(enteredDexBaseFeeLimit)
				err := core.SaveSettings(rootPath)
				if err != nil {
					dialog.ShowError(err, mainWindow)
				}
				applySave.Disable()
				applyBtn.Disable()

			})
			applyBtn.Disable()
			applySave.Disable()

			settingsForm.SetOnValidationChanged(func(err error) {
				if err != nil {
					applyBtn.Disable()
					applySave.Disable()
				} else if selectedOnlyDirectRoute != core.UserSettings.DexDirectRoute || selectedRouteEvaluation != core.UserSettings.DexRouteEvaluation || enteredDexBaseFeeLimit.Cmp(core.UserSettings.DexBaseFeeLimit) != 0 {
					applyBtn.Enable()
					applySave.Enable()
				}
			})
			slippageEntry.Validator = func(s string) error {
				// Split the input string at the decimal point
				parts := strings.Split(s, ".")
				// Check if there are more than 2 decimals
				if len(parts) > 1 && len(parts[1]) > 2 {
					return fmt.Errorf("invalid slippage (must not have more than 2 decimal places)")
				}
				slippage, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				} else if slippage < 0 {
					return fmt.Errorf("invalid slippage (min 0)")
				} else if slippage > 99 {
					return fmt.Errorf("invalid slippage (max 99)")
				} else {
					enteredSlippage = slippage
					return nil
				}

			}
			dexBaseFeeLimitEntry.Validator = func(s string) error {
				entry, err := new(big.Int).SetString(s, 10)
				if !err || entry == nil {
					return fmt.Errorf("only integer values")
				} else if entry.Cmp(big.NewInt(10000)) <= 0 {
					return fmt.Errorf("bigger than 10000")
				}

				enteredDexBaseFeeLimit = entry
				return nil
			}
			dexBaseFeeLimitEntry.OnChanged = func(s string) {
				if dexBaseFeeLimitEntry.Validate() != nil || slippageEntry.Validate() != nil {
					return
				}
				if enteredDexBaseFeeLimit.Cmp(core.UserSettings.DexBaseFeeLimit) != 0 || enteredSlippage != core.UserSettings.DexSlippage {
					applyBtn.Enable()
					applySave.Enable()

				} else {
					applyBtn.Disable()
					applySave.Disable()
				}

			}

			slippageEntry.OnChanged = func(s string) {
				if dexBaseFeeLimitEntry.Validate() != nil || slippageEntry.Validate() != nil {
					return
				}
				if enteredDexBaseFeeLimit.Cmp(core.UserSettings.DexBaseFeeLimit) != 0 || enteredSlippage != core.UserSettings.DexSlippage {
					applyBtn.Enable()
					applySave.Enable()

				} else {
					applyBtn.Disable()
					applySave.Disable()
				}
			}

			routingOptions.OnChanged = func(s string) {
				if dexBaseFeeLimitEntry.Validate() != nil || slippageEntry.Validate() != nil {
					return
				}

				switch s {
				case "Auto":
					selectedRouteEvaluation = "auto"
					selectedOnlyDirectRoute = false
				case "Direct routes only":
					selectedRouteEvaluation = "auto"
					selectedOnlyDirectRoute = true
				case "Lowest price impact":
					selectedRouteEvaluation = "lowestImpact"
					selectedOnlyDirectRoute = false
				case "Lowest price":
					selectedRouteEvaluation = "highestOutput"
					selectedOnlyDirectRoute = false
				}

				if enteredDexBaseFeeLimit.Cmp(core.UserSettings.DexBaseFeeLimit) != 0 || enteredSlippage != core.UserSettings.DexSlippage || selectedRouteEvaluation != core.UserSettings.DexRouteEvaluation || selectedOnlyDirectRoute != core.UserSettings.DexDirectRoute {
					applyBtn.Enable()
					applySave.Enable()

				} else {
					applyBtn.Disable()
					applySave.Disable()
				}
			}

			bttns := container.NewGridWithColumns(3, closeBtn, applyBtn, applySave)
			sttgnsLyt := container.NewBorder(nil, bttns, nil, settingsForm)

			settingsDia = dialog.NewCustomWithoutButtons("Dex Settings", sttgnsLyt, mainWindow)
			settingsDia.Resize(fyne.NewSize(mainWindowLyt.Selected().Content.Size().Width, 0))
			settingsDia.Show()
		})
		swapHeader := widget.NewLabelWithStyle("Spallet Swap", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		form := container.NewVBox(

			inTokenLyt,
			swapIcon,
			outTokenLyt,
			routeMessage,
			warningMessage,
			swapBtn,
			container.NewBorder(nil, nil, widget.NewRichTextFromMarkdown("Powered by [Saturn Dex](https://saturn.stellargate.io/)"), settingsBtn),
		)

		swapTab.Content = container.NewBorder(swapHeader, nil, nil, nil, container.New(layout.NewVBoxLayout(), layout.NewSpacer(), container.NewPadded(form), layout.NewSpacer()))

	} else {
		noKcalMessage := widget.NewLabelWithStyle(fmt.Sprintf("Looks like Sparky low on sparks! ⚡️🕹️\n Your swap needs some Phantasma Energy (KCAL) to keep the ghostly gears turning. Time to add some KCAL and get that blockchain buzzing faster than a haunted hive!\n You need at least %v Kcal", core.FormatBalance(fee, core.KcalDecimals)), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		noKcalMessage.Wrapping = fyne.TextWrapWord
		swapTab.Content = container.NewVBox(noKcalMessage)

	}

}

func monitorSwapTransaction(txHash string, creds core.Credentials) {
	maxRetries := 12 // 2 secs of block time so we are waiting for 3 block and that is enough
	retryCount := 0
	retryDelay := time.Millisecond * 500
	showUpdatingDialog()
	fmt.Printf("Starting transaction monitoring for hash: %s\n", txHash)

	for {
		if retryCount >= maxRetries {
			fmt.Printf("Transaction monitoring timed out after %d retries\n", maxRetries)

			showTxResultDialog("Transaction monitoring timed out.", creds, response.TransactionResult{Hash: txHash, Fee: "0"})
			swapGui(creds)
			swapTab.Content.Refresh()

			return
		}

		fmt.Printf("Checking transaction status (attempt %d/%d)\n", retryCount+1, maxRetries)
		txResult, err := core.Client.GetTransaction(txHash)
		if err != nil {
			fmt.Printf("Error getting transaction status: %v\n", err)
			if strings.Contains(err.Error(), "could not decode body") ||
				strings.Contains(err.Error(), "rpc call") {
				retryCount++
				time.Sleep(retryDelay)
				continue
			}

			showTxResultDialog("Failed to get transaction status.", creds, response.TransactionResult{Hash: txHash, Fee: "0"})

			swapGui(creds)
			swapTab.Content.Refresh()

			return
		}

		if txResult.StateIsSuccess() {
			fmt.Printf("Transaction successful\n")

			showTxResultDialog("Swap completed successfully.", creds, txResult)

			swapGui(creds)
			swapTab.Content.Refresh()

			return
		}
		if txResult.StateIsFault() {
			fmt.Printf("Transaction failed\n")
			showTxResultDialog("Swap failed.", creds, txResult)

			swapGui(creds)
			swapTab.Content.Refresh()

			return
		}

		fmt.Printf("Transaction pending, state: %s\n", txResult.State)
		retryCount++
		time.Sleep(retryDelay)
	}
}
