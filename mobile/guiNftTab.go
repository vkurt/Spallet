package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func nftGui() {
	nftTabHeader := widget.NewLabelWithStyle("Your Smart Nft's", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	NftContent := container.NewVScroll(nftBtns)
	nftTab.Content = container.NewBorder(nftTabHeader, nil, nil, nil, NftContent)
	nftTab.Content.Refresh()
}
