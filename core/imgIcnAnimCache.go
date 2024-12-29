package core

import (
	"os"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
)

var (
	AnimCache      = make(map[string]fyne.Resource)
	animCacheMutex sync.Mutex
	AnimSets       = map[string][]string{
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

var BadgeImageCache = make(map[string]fyne.Resource)
var badgeImageCacheMutex sync.Mutex

var IconCache = make(map[string]fyne.Resource)
var iconCacheMutex sync.Mutex

// loadIconResource loads an icon resource with caching
func LoadIconResource(path string) fyne.Resource {
	iconCacheMutex.Lock()
	defer iconCacheMutex.Unlock()

	if resource, found := IconCache[path]; found {
		return resource
	}

	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}

	IconCache[path] = resource
	return resource
}

// loadBadgeImageResource loads a badge image resource with caching
func LoadBadgeImageResource(path string) fyne.Resource {
	badgeImageCacheMutex.Lock()
	defer badgeImageCacheMutex.Unlock()

	if resource, found := BadgeImageCache[path]; found {
		return resource
	}

	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		return nil
	}

	BadgeImageCache[path] = resource
	return resource
}

// loadImageResource loads an image resource with caching

func LoadAnimResource(path string) fyne.Resource {
	animCacheMutex.Lock()
	defer animCacheMutex.Unlock()

	if resource, found := AnimCache[path]; found {
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

	AnimCache[path] = resource
	return resource
}
