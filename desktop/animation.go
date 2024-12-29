package main

import (
	"fmt"
	"spallet/core"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
)

func startAnimation(animSetKey, dialogTitle string) chan bool {
	// Get the appropriate image set
	images, exists := core.AnimSets[animSetKey]
	if !exists {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: fmt.Sprintf("Image set '%s' not found", animSetKey),
		})
		return nil
	}

	// Create a single image object to be updated
	animationContainer := container.New(layout.NewCenterLayout())
	imageCanvas := canvas.NewImageFromResource(nil)
	imageCanvas.FillMode = canvas.ImageFillContain
	imageCanvas.SetMinSize(fyne.NewSize(400, 300))
	animationContainer.Objects = []fyne.CanvasObject{imageCanvas}

	imageIndex := 0

	updateAnimationImage := func() {
		updateAnimImageResource(imageCanvas, images[imageIndex])
		imageIndex = (imageIndex + 1) % len(images)
	}

	// Load and show the first image
	updateAnimationImage()

	animationBox := container.NewVBox(animationContainer)
	animationBox.Resize(fyne.NewSize(400, 300)) // Set the desired size

	currentMainDialog.Hide()
	d := dialog.NewCustomWithoutButtons(dialogTitle, animationBox, mainWindowGui)
	currentMainDialog = d

	// Start the animation
	stopChan := make(chan bool)
	StartAnimation(125*time.Millisecond, updateAnimationImage, stopChan)
	d.Show()

	var closeOnce sync.Once
	currentMainDialog.SetOnClosed(func() {
		closeOnce.Do(func() {
			close(stopChan)
		})
	})

	return stopChan
}

// StartAnimation starts a ticker-based animation for updating images
func StartAnimation(tickerDuration time.Duration, updateImageFunc func(), stopChan chan bool) {
	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				updateImageFunc()
			}
		}
	}()
}

func updateAnimImageResource(imageCanvas *canvas.Image, imagePath string) {
	imgResource := core.LoadAnimResource(imagePath)
	imageCanvas.Resource = imgResource
	imageCanvas.Refresh()
}

// func loadImage(path string) image.Image {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil
// 	}
// 	defer file.Close()
// 	img, err := imaging.Decode(file)
// 	if err != nil {
// 		return nil
// 	}
// 	return img
// }
