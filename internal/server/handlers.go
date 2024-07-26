package server

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"example/sashaTelegram/internal/openweathermap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (server *Server) eventsHandler() {
	for {
		time.Sleep(time.Minute)
		for _, event := range server.events {
			if !event.IsReady() {
				continue
			}
			if len(server.photos) >= 5 {
				msg := tgbotapi.NewMessage(event.ID(), "No place for event photo in queue, try again later")
				if _, err := server.bot.Send(msg); err != nil {
					log.Fatal("Failed to send a message:", err)
				}
				continue
			}
			server.photos <- event.Photo()
		}
	}
}

func (server *Server) photosHandler() {
	cmd := exec.Command("./motor_driver.bin", "0", "0", "True", "0", "3", "")
	if err := cmd.Run(); err != nil {
		log.Fatal("Error turning camera to 0:", err)
	}
	var currentX int
	for photo := range server.photos {
		data, err := photo.Take(currentX)
		if err != nil {
			exec.Command("./phone_init.sh").Run()
			server.photos <- photo
			log.Println("Failed to take a photo:", err)
			continue
		}
		currentX = photo.X

		msg := tgbotapi.NewPhoto(photo.ID, tgbotapi.FileBytes{Name: "image.jpg", Bytes: data})
		msg.Caption = fmt.Sprintf("X: %v Y: %v", photo.X, photo.Y)
		if _, err := server.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	}
}

func (server *Server) sunsetHandler(city string, owm *openweathermap.Config) {
	for {
		sunset, err := owm.SunsetTime(city)
		if err != nil {
			log.Fatal("Failed to get sunset time:", err)
		}
		server.config.SunsetTime = sunset
		time.Sleep(time.Hour * 24)
	}
}
