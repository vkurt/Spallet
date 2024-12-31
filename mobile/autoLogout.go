package main

import (
	"time"
)

var logoutTicker *time.Ticker

func startLogoutTicker(timeout int) {
	if logoutTicker != nil {
		logoutTicker.Stop()
	}
	timeout *= 60
	if timeout <= 0 {
		timeout = 1
	}

	logoutTicker = time.NewTicker(time.Duration(timeout) * time.Second)
	go func() {
		for range logoutTicker.C {
			showExistingUserLogin()
		}
	}()
}
