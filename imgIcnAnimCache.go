package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
)

var (
	animCache      = make(map[string]fyne.Resource)
	animCacheMutex sync.Mutex
	animSets       = map[string][]string{
		"send": {
			"img/anim/send/1.png", "img/anim/send/2.png", "img/anim/send/3.png",
			"img/anim/send/4.png", "img/anim/send/5.png", "img/anim/send/6.png",
			"img/anim/send/7.png", "img/anim/send/8.png", "img/anim/send/9.png", "img/anim/send/10.png",
		},
		"forging": {
			"img/anim/forging/1.png", "img/anim/forging/2.png", "img/anim/forging/3.png",
			"img/anim/forging/4.png", "img/anim/forging/5.png", "img/anim/forging/6.png",
			"img/anim/forging/7.png", "img/anim/forging/8.png",
		},
		"fill": {
			"img/anim/drainfill/8.png", "img/anim/drainfill/7.png", "img/anim/drainfill/6.png", "img/anim/drainfill/5.png",
			"img/anim/drainfill/4.png", "img/anim/drainfill/3.png", "img/anim/drainfill/2.png",
			"img/anim/drainfill/1.png", "img/anim/drainfill/0.png",
		},
		"drain": {
			"img/anim/drainfill/0.png", "img/anim/drainfill/1.png", "img/anim/drainfill/2.png", "img/anim/drainfill/3.png",
			"img/anim/drainfill/4.png", "img/anim/drainfill/5.png", "img/anim/drainfill/6.png",
			"img/anim/drainfill/7.png", "img/anim/drainfill/8.png",
		},
		// Add more image sets as needed
	}
)

var badgeImageCache = make(map[string]fyne.Resource)
var badgeImageCacheMutex sync.Mutex

var iconCache = make(map[string]fyne.Resource)
var iconCacheMutex sync.Mutex

// loadIconResource loads an icon resource with caching
func loadIconResource(path string) fyne.Resource {
	iconCacheMutex.Lock()
	defer iconCacheMutex.Unlock()

	if resource, found := iconCache[path]; found {
		return resource
	}

	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}

	iconCache[path] = resource
	return resource
}

// loadBadgeImageResource loads a badge image resource with caching
func loadBadgeImageResource(path string) fyne.Resource {
	badgeImageCacheMutex.Lock()
	defer badgeImageCacheMutex.Unlock()

	if resource, found := badgeImageCache[path]; found {
		return resource
	}

	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}

	badgeImageCache[path] = resource
	return resource
}

// loadImageResource loads an image resource with caching

func loadAnimResource(path string) fyne.Resource {
	animCacheMutex.Lock()
	defer animCacheMutex.Unlock()

	if resource, found := animCache[path]; found {
		return resource
	}

	imageFile, err := os.Open(path)
	if err != nil {
		return nil
	}

	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		imageFile.Close()
		return nil
	}

	runtime.SetFinalizer(imageFile, func(f *os.File) {
		f.Close()
	})

	animCache[path] = resource
	return resource
}

func startAnimation(animSetKey, dialogTitle string) chan bool {
	// Get the appropriate image set
	images, exists := animSets[animSetKey]
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
	imgResource := loadAnimResource(imagePath)
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
