package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func loadImgFromResource(imgName string, imgSize fyne.Size) *canvas.Image {
	// Create resource file name

	// Retrieve the resource by its name
	resource := getResourceByName(imgName)
	image := canvas.NewImageFromResource(resource)
	image.SetMinSize(imgSize)
	return image
}

func getResourceByName(name string) fyne.Resource {
	// Retrieve the resource from the variables
	switch name {
	case "swap":
		return resourceSwapPng
	case "hodl":
		return resourceHodlPng
	case "history":
		return resourceHistoryPng
	default:
		return theme.BrokenImageIcon()

	}

}
