package main

import (
	"fmt"
	"log"
	"net/http"
)

func setZoom(zoom int) {
	url := fmt.Sprintf("http://127.0.0.1:8080/ptz?zoom=%v", zoom)
	if _, err := http.Get(url); err != nil {
		log.Fatal("Failed to set camera zoom:", err)
	}
}
