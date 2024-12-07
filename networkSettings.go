package main

import (
	"math/big"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

func showNetworkSettingsPage(creds Credentials) {
	var feeInfoDia dialog.Dialog

	var saveButton *widget.Button
	settingsForSave := userSettings

	customChainName := binding.BindString(&settingsForSave.ChainName)
	customChainNetwork := binding.BindString(&settingsForSave.NetworkName)
	defaultRpc := "https://pharpc1.phantasma.info/rpc"
	customChainRpc := binding.BindString(&defaultRpc)
	customChainAccExplorerLink := binding.BindString(&settingsForSave.AccExplorerLink)
	customChainTxExplorerLink := binding.BindString(&settingsForSave.TxExplorerLink)

	customChainNameEntry := widget.NewEntryWithData(customChainName)
	customChainNetworkEntry := widget.NewEntryWithData(customChainNetwork)
	customChainRpcEntry := widget.NewEntryWithData(customChainRpc)
	customChainTxExplorerLinkEntry := widget.NewEntryWithData(customChainTxExplorerLink)
	customChainAccExplorerLinkEntry := widget.NewEntryWithData(customChainAccExplorerLink)

	customNetworkSettingsForm := widget.NewForm(
		widget.NewFormItem("Chain Name", customChainNameEntry),
		widget.NewFormItem("Network Name", customChainNetworkEntry),
		widget.NewFormItem("Rpc Address", customChainRpcEntry),
		widget.NewFormItem("Tx Explorer Link", customChainTxExplorerLinkEntry),
		widget.NewFormItem("Acc Explorer Link", customChainAccExplorerLinkEntry),
	)
	customChainNameEntry.Disable()
	customChainNetworkEntry.Disable()
	customChainRpcEntry.Disable()
	customChainAccExplorerLinkEntry.Disable()
	customChainTxExplorerLinkEntry.Disable()

	defaultFeeLimit := widget.NewEntry()
	dexBaseGasLimit := widget.NewEntry()
	feeLimitSliderMin := widget.NewEntry()
	feeLimitSliderMax := widget.NewEntry()
	dexBaseGasLimit.SetText(userSettings.DexBaseFeeLimit.String())
	feePrice := widget.NewEntry()
	defaultFeeLimit.SetText(userSettings.DefaultGasLimit.String())
	feeLimitSliderMin.SetText(strconv.FormatFloat(userSettings.GasLimitSliderMin, 'f', -1, 64))
	feeLimitSliderMax.SetText(strconv.FormatFloat(userSettings.GasLimitSliderMax, 'f', -1, 64))
	feePrice.SetText(userSettings.GasPrice.String())

	feeExplain := widget.NewRichTextFromMarkdown("When you perform transactions on the blockchain, certain fees are involved. Here's what you need to know:\n\n1- **Fee Limit:** This is the maximum amount of tokens that can be used for a transaction. In Spallet, we have set a default value for the fee limit which is default fee limit for hodling transactions and token sends. Dex base fee limit is for interacting one pool if more pools involved your fee limit will be increased.\n\n2- **Adjustable Fee Limit:** For other types of transactions, such as sending tokens, you might not need such a high fee limit. To give you flexibility, we provide a slider that allows you to adjust the fee limit before you send your transaction. You can increase or decrease it depending on the complexity of your transaction.\n\n3- **Fee Price:** This is the cost per unit of the fee tokens. In the Phantasma blockchain, the fee token is Kcal, and the settings you adjust in the wallet are done in Sparks. Note that 1 Kcal equals 10,000,000,000 Sparks. Higher fee prices mean your transaction will be processed faster, as transactions with higher fees are given higher priority when the blockchain is busy.\n\n**How It Works**\n\n**For Hodling Transactions:** The default fee limit is set to ensure smooth processing. You can adjust it too.\n\n* **For Other Transactions:** Adjust the fee limit using the slider to match the complexity of your transaction.\n\n* **Higher Fee Limit:** Use for transactions with many operations, ensuring they go through successfully.\n\n* **Lower Fee Limit:** Use for simpler transactions to save on fees.\n\nBy understanding these settings, you can better manage your transaction costs and ensure your transactions are processed efficiently.")
	feeExplain.Wrapping = fyne.TextWrapWord
	feeExplainCntnt := container.NewVScroll(feeExplain)
	feeExplainCntnt.Resize(fyne.NewSize(500, 550))

	feeInfoDia = dialog.NewCustom("Understanding Transaction Fees", "Close", feeExplainCntnt, mainWindowGui)
	feeInfoDia.Resize(fyne.NewSize(700, 500))
	feeSettingsForm := widget.NewForm(
		widget.NewFormItem("Fee Price", feePrice),
		widget.NewFormItem("Default Fee Limit", defaultFeeLimit),
		widget.NewFormItem("Dex Base Fee Limit", dexBaseGasLimit),
		widget.NewFormItem("", widget.NewLabel("Fee limit Slider Settings")),
		widget.NewFormItem("Max", feeLimitSliderMax),
		widget.NewFormItem("Min", feeLimitSliderMin),
		widget.NewFormItem("", widget.NewButtonWithIcon("", theme.HelpIcon(), func() { feeInfoDia.Show() })),
	)

	// Inner validation function
	validateEntries := func() {
		dexBaseGasLimitValue, dexFeeErr := new(big.Int).SetString(dexBaseGasLimit.Text, 10)
		defaultFeeLimitValue, defErr := new(big.Int).SetString(defaultFeeLimit.Text, 10)
		feeLimitSliderMinValue, minErr := strconv.ParseFloat(feeLimitSliderMin.Text, 64)
		feeLimitSliderMaxValue, maxErr := strconv.ParseFloat(feeLimitSliderMax.Text, 64)
		_, priceErr := new(big.Int).SetString(feePrice.Text, 10)

		valid := true

		if !defErr || minErr != nil || maxErr != nil || !priceErr || !dexFeeErr || dexBaseGasLimitValue == nil {
			valid = false
		}

		if valid && (dexBaseGasLimitValue.Cmp(big.NewInt(25000)) < 0 || defaultFeeLimitValue.Cmp(big.NewInt(int64(feeLimitSliderMinValue))) < 0 || defaultFeeLimitValue.Cmp(big.NewInt(int64(feeLimitSliderMaxValue))) > 0) {
			valid = false
		}

		if valid && feeLimitSliderMinValue > feeLimitSliderMaxValue {
			valid = false
		}

		if valid && feeLimitSliderMaxValue < feeLimitSliderMinValue {
			valid = false
		}

		if valid && (len(customChainNameEntry.Text) < 1 || len(customChainNetworkEntry.Text) < 1 || len(customChainRpcEntry.Text) < 1 || len(customChainTxExplorerLinkEntry.Text) < 1 || len(customChainAccExplorerLinkEntry.Text) < 1) {
			valid = false
		}

		if valid {
			saveButton.Enable()
		} else {
			saveButton.Disable()
		}
	}

	// Add validation to entries
	dexBaseGasLimit.OnChanged = func(string) { validateEntries() }
	defaultFeeLimit.OnChanged = func(string) { validateEntries() }
	feeLimitSliderMin.OnChanged = func(string) { validateEntries() }
	feeLimitSliderMax.OnChanged = func(string) { validateEntries() }
	feePrice.OnChanged = func(string) { validateEntries() }
	customChainNameEntry.OnChanged = func(string) { validateEntries() }
	customChainNetworkEntry.OnChanged = func(string) { validateEntries() }
	customChainRpcEntry.OnChanged = func(string) { validateEntries() }
	customChainTxExplorerLinkEntry.OnChanged = func(string) { validateEntries() }
	customChainAccExplorerLinkEntry.OnChanged = func(string) { validateEntries() }

	networkSelect := widget.NewSelect([]string{"Mainnet", "Testnet", "Custom"}, func(selected string) {
		switch selected {
		case "Mainnet":
			settingsForSave.ChainName = "main"
			customChainName.Reload()
			settingsForSave.NetworkName = "mainnet"
			customChainNetwork.Reload()
			client = rpc.NewRPCMainnet()
			settingsForSave.RpcType = "mainnet"
			settingsForSave.TxExplorerLink = "https://explorer.phantasma.info/en/transaction?id="
			customChainTxExplorerLink.Reload()
			settingsForSave.AccExplorerLink = "https://explorer.phantasma.info/en/address?id="
			customChainAccExplorerLink.Reload()
			settingsForSave.NetworkType = "Mainnet"
			customChainNameEntry.Disable()
			customChainNetworkEntry.Disable()
			customChainRpcEntry.Disable()
			customChainAccExplorerLinkEntry.Disable()
			customChainTxExplorerLinkEntry.Disable()
		case "Testnet":
			settingsForSave.ChainName = "main"
			customChainName.Reload()
			settingsForSave.NetworkName = "testnet"
			customChainNetwork.Reload()
			client = rpc.NewRPCTestnet()
			settingsForSave.RpcType = "testnet"
			settingsForSave.TxExplorerLink = "https://test-explorer.phantasma.info/en/transaction?id="
			customChainTxExplorerLink.Reload()
			settingsForSave.AccExplorerLink = "https://test-explorer.phantasma.info/en/address?id="
			customChainAccExplorerLink.Reload()
			settingsForSave.NetworkType = "Testnet"
			customChainNameEntry.Disable()
			customChainNetworkEntry.Disable()
			customChainRpcEntry.Disable()
			customChainAccExplorerLinkEntry.Disable()
			customChainTxExplorerLinkEntry.Disable()
		case "Custom":
			customChainNameEntry.Enable()
			customChainNetworkEntry.Enable()
			customChainRpcEntry.Enable()
			customChainAccExplorerLinkEntry.Enable()
			customChainTxExplorerLinkEntry.Enable()
			settingsForSave.RpcType = "custom"
			settingsForSave.NetworkType = "Custom"
		}

	})
	networkSelect.SetSelected(userSettings.NetworkType) // Set initial value and trigger OnChanged

	saveButton = widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {

		if networkSelect.Selected == "Mainnet" || networkSelect.Selected == "Testnet" {
			userSettings.AccExplorerLink = settingsForSave.AccExplorerLink
			userSettings.TxExplorerLink = settingsForSave.TxExplorerLink
			userSettings.NetworkName = settingsForSave.NetworkName
			userSettings.ChainName = settingsForSave.ChainName
			userSettings.RpcType = settingsForSave.RpcType
			userSettings.NetworkType = settingsForSave.NetworkType

		} else {
			userSettings.ChainName = customChainNameEntry.Text
			userSettings.NetworkName = customChainNetworkEntry.Text
			userSettings.CustomRpcLink = customChainRpcEntry.Text
			userSettings.AccExplorerLink = customChainAccExplorerLinkEntry.Text
			userSettings.TxExplorerLink = customChainAccExplorerLinkEntry.Text
			userSettings.RpcType = settingsForSave.RpcType
			userSettings.NetworkType = settingsForSave.NetworkType
		}
		settingsForSave.GasLimitSliderMax, _ = strconv.ParseFloat(feeLimitSliderMax.Text, 64)
		settingsForSave.GasLimitSliderMin, _ = strconv.ParseFloat(feeLimitSliderMin.Text, 64)
		settingsForSave.DefaultGasLimit, _ = new(big.Int).SetString(defaultFeeLimit.Text, 10)
		settingsForSave.GasPrice, _ = new(big.Int).SetString(feePrice.Text, 10)
		userSettings.DexBaseFeeLimit, _ = new(big.Int).SetString(dexBaseGasLimit.Text, 10)

		userSettings.GasLimitSliderMin = settingsForSave.GasLimitSliderMin
		userSettings.GasLimitSliderMax = settingsForSave.GasLimitSliderMax
		userSettings.DefaultGasLimit = settingsForSave.DefaultGasLimit
		userSettings.GasPrice = settingsForSave.GasPrice

		err := saveSettings()
		if err == nil {
			currentMainDialog.Hide()
			showUpdatingDialog()
			dataFetch(creds)
			// mainWindowGui.Canvas().Content().Refresh()
			// mainWindow(creds)
			closeUpdatingDialog()
			dialog.ShowInformation("Settings Saved", "Settings Saved successfully", mainWindowGui)
		} else {
			dialog.ShowError(err, mainWindowGui)
		}

	})

	bckbttn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { currentMainDialog.Hide() })
	bttnContainer := container.NewGridWithColumns(2, bckbttn, saveButton)
	ntwrkScroll := container.NewVBox(networkSelect, customNetworkSettingsForm, feeSettingsForm)
	networkSettingsLyt := container.NewBorder(nil, bttnContainer, nil, nil, container.NewVScroll(ntwrkScroll))

	currentMainDialog = dialog.NewCustomWithoutButtons("Network Settings", networkSettingsLyt, mainWindowGui)
	currentMainDialog.Resize(fyne.NewSize(600, mainWindowGui.Canvas().Size().Height-50))
	currentMainDialog.Show()

}
