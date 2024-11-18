package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showNetworkSettingsPage(walletDetails *fyne.Container, creds Credentials) {
	customRpcEntry := widget.NewEntry()
	customRpcEntry.SetPlaceHolder("Enter custom RPC URL")
	customRpcEntry.Hide()

	networkSelect := widget.NewSelect([]string{"Mainnet", "Testnet", "Custom"}, func(selected string) {
		userSettings.Network = selected
		fmt.Println("Selected Network:", userSettings.Network) // Debug print
		if selected == "Custom" {
			customRpcEntry.Show()
		} else {
			customRpcEntry.Hide()
		}
	})
	networkSelect.SetSelected(userSettings.Network) // Set initial value and trigger OnChanged

	chainEntry := widget.NewEntry()
	chainEntry.SetPlaceHolder("Enter chain name")
	chainEntry.Text = userSettings.Chain // Set initial value
	chainEntry.OnChanged = func(chainName string) {
		userSettings.Chain = chainName
	}

	saveButton := widget.NewButton("Save", func() {
		if userSettings.Network == "Custom" && customRpcEntry.Text == "" {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Custom RPC URL cannot be empty",
			})
			return
		}

		userSettings.CustomRPC = customRpcEntry.Text
		err := saveSettings()
		if err != nil {
			dialog.ShowInformation("Error", "Failed to save settings: "+err.Error(), fyne.CurrentApp().Driver().AllWindows()[0])
		} else {
			dialog.ShowInformation("Success", "Settings saved successfully!", fyne.CurrentApp().Driver().AllWindows()[0])
			fmt.Println("Saved Network:", userSettings.Network) // Debug print
		}
	})

	// Update walletDetails content to show settings
	walletDetails.Objects = []fyne.CanvasObject{
		networkSelect,
		customRpcEntry,
		widget.NewLabel("Chain Name:"),
		chainEntry,
		saveButton,
		widget.NewButton("Back", func() {
			mainWindow(creds, regularTokens, nftTokens)
		}),
	}
	walletDetails.Refresh()
}
