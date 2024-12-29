package main

import (
	"fmt"
	"spallet/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func buildBadges() *fyne.Container {
	// var enableAnimation bool = core.LatestAccountData.IsStaker
	var mainBadgePath string
	var crownBadgePath string
	var soulMasterBadgePath string
	var stakingBadgePath string
	var networkBadgePath string
	// var defaultAnSpeed = 25.0

	fmt.Println("******Building badges*****")

	switch core.LatestAccountData.BadgeName {
	case "lord":
		mainBadgePath = "img/stats/lord.png"
	case "master":
		mainBadgePath = "img/stats/master.png"
	case "acolyte":
		mainBadgePath = "img/stats/acolyte.png"
	case "wanderer":
		mainBadgePath = "img/stats/wanderer.png"
	case "snoozer":
		mainBadgePath = "img/stats/snoozer.png"
	case "apprentice":
		mainBadgePath = "img/stats/apprentice.png"
	default:
		mainBadgePath = "img/stats/UNKNOWN.png"
	}

	// if core.LatestAccountData.KcalBoost > 0 {
	// 	defaultAnSpeed = defaultAnSpeed / (float64(core.LatestAccountData.KcalBoost)/100 + 1)
	// } else {
	// 	defaultAnSpeed = 25
	// }

	mainBadge := canvas.NewImageFromResource(core.LoadBadgeImageResource(mainBadgePath))
	mainBadge.FillMode = canvas.ImageFillContain
	mainBadge.SetMinSize(fyne.NewSize(150, 150))

	switch {
	case core.LatestAccountData.IsSoulMaster && core.LatestAccountData.IsEligibleForCurrentCrown && core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN.png"
		soulMasterBadgePath = "img/stats/soul_master.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case core.LatestAccountData.IsSoulMaster && core.LatestAccountData.IsEligibleForCurrentCrown && !core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN.png"
		soulMasterBadgePath = "img/stats/soul_master_en.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case core.LatestAccountData.IsSoulMaster && !core.LatestAccountData.IsEligibleForCurrentCrown && !core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN_en.png"
		soulMasterBadgePath = "img/stats/soul_master_en.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case core.LatestAccountData.IsSoulMaster && !core.LatestAccountData.IsEligibleForCurrentCrown && core.LatestAccountData.IsEligibleForCurrentSmReward:
		crownBadgePath = "img/stats/CROWN_en.png"
		soulMasterBadgePath = "img/stats/soul_master.png"
		stakingBadgePath = "img/stats/Kcal.png"
	case core.LatestAccountData.IsStaker:
		crownBadgePath = "img/stats/CROWN_ne.png"
		soulMasterBadgePath = "img/stats/soul_master_ne.png"
		stakingBadgePath = "img/stats/Kcal.png"
	default:
		crownBadgePath = "img/stats/CROWN_ne.png"
		soulMasterBadgePath = "img/stats/soul_master_ne.png"
		stakingBadgePath = "img/stats/Kcal_ns.png"
	}

	if core.UserSettings.NetworkName == "mainnet" {
		networkBadgePath = "img/stats/mainnet.png"
	} else {
		networkBadgePath = "img/stats/testnet.png"
	}

	networkBadge := canvas.NewImageFromResource(core.LoadBadgeImageResource(networkBadgePath))
	networkBadge.FillMode = canvas.ImageFillContain
	networkBadge.SetMinSize(fyne.NewSize(26, 26))

	crownBadge := canvas.NewImageFromResource(core.LoadBadgeImageResource(crownBadgePath))
	crownBadge.FillMode = canvas.ImageFillContain
	crownBadge.SetMinSize(fyne.NewSize(26, 26))

	soulMasterBadge := canvas.NewImageFromResource(core.LoadBadgeImageResource(soulMasterBadgePath))
	soulMasterBadge.FillMode = canvas.ImageFillContain
	soulMasterBadge.SetMinSize(fyne.NewSize(26, 26))

	stakingBadge := canvas.NewImageFromResource(core.LoadBadgeImageResource(stakingBadgePath))
	stakingBadge.FillMode = canvas.ImageFillContain
	stakingBadge.SetMinSize(fyne.NewSize(26, 26))

	emptyArea := canvas.NewImageFromResource(core.LoadBadgeImageResource("img/stats/spacer.png"))
	emptyArea.FillMode = canvas.ImageFillContain
	emptyArea.SetMinSize(fyne.NewSize(7, 7))

	imageContainer := container.NewBorder(
		nil, nil, nil,
		container.NewVBox(networkBadge, emptyArea, emptyArea, crownBadge, soulMasterBadge, stakingBadge, emptyArea),
		mainBadge,
	)

	// stopBadgeAnimation()

	// if enableAnimation && !animationRunning { // liked it at first but it using unnecessary cpu power
	// 	animationRunning = true
	// 	stopAnimation = make(chan bool)
	// 	fmt.Println("Animation started")
	// 	go func() {
	// 		var scale float32 = 1.0
	// 		var increment float32 = 0.01
	// 		anSpeed := time.Duration(defaultAnSpeed) * time.Millisecond

	// 		fmt.Println("anSpeed", anSpeed)
	// 		for {
	// 			select {
	// 			case <-stopAnimation:
	// 				animationRunning = false
	// 				return
	// 			case <-time.Tick(anSpeed):
	// 				if scale >= 1.1 || scale <= 0.9 {
	// 					increment = -increment
	// 				}
	// 				scale += increment
	// 				stakingBadge.Resize(fyne.NewSize(26*scale, 26*scale))
	// 				stakingBadge.Refresh()
	// 			}
	// 		}
	// 	}()
	// } else if !enableAnimation && animationRunning {
	// 	fmt.Println("Animation stopped")
	// 	stopBadgeAnimation()
	// }
	badgesSize := imageContainer.MinSize()
	accBadges.Content = imageContainer
	accBadges.SetMinSize(badgesSize)
	accBadges.Refresh()
	return imageContainer
}
