package main

import (
	"fmt"
	"net/http"
)

func setNightVisionGain(gain int) {
	url := fmt.Sprintf("http://127.0.0.1:8080/settings/night_vision_gain?set=%v",gain)
	if _, err := http.Get(url); err != nil {
		fmt.Println("Failed to set night vision gain:", err)
	}
}

func setZoom(zoom int) {
	url := fmt.Sprintf("http://127.0.0.1:8080/ptz?zoom=%v", zoom)
	if _, err := http.Get(url); err != nil {
		fmt.Println("Failed to set zoom:", err)
	}
}

func enableNightVision() {
	url := "http://127.0.0.1:8080/settings/night_vision?set=on"
	if _, err := http.Get(url); err != nil {
		fmt.Println("Failed to enable night vision:", err)
	}
}

func disableNightVision() {
	url := "http://127.0.0.1:8080/settings/night_vision?set=off"
	if _, err := http.Get(url); err != nil {
		fmt.Println("Failed to disable night vision:", err)
	}
}
