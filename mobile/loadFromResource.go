package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	case "KCAL":
		return resourceKCALPng
	case "SOUL":
		return resourceSOULPng
	case "CROWN":
		return resourceCROWNPng
	case "RAA":
		return resourceRAAPng
	case "BRC":
		return resourceBRCPng
	case "GAME":
		return resourceGAMEPng
	case "GHOST":
		return resourceGHOSTPng
	case "GOATI":
		return resourceGOATIPng
	case "MKNI":
		return resourceMKNIPng
	case "SNFT":
		return resourceSNFTPng
	case "TIGER":
		return resourceTIGERPng
	case "TTRS":
		return resourceTTRSPng
	default:
		return resourcePlaceholderPng

	}

}
