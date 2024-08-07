package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"example/sashaTelegram/internal/chat"
	"example/sashaTelegram/internal/openweathermap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *server) saveHandler() {
	for {
		time.Sleep(1 * time.Hour)

		if err := s.saveGobData(); err != nil {
			slog.Error("Can not save gob data", "err", err)
			os.Exit(1)
		}
	}
}

func (server *server) eventsHandler() {
	for {
		time.Sleep(time.Minute)
		for _, event := range server.events {
			if !event.IsReady() {
				continue
			}
			if len(server.photos) >= 5 {
				msg := tgbotapi.NewMessage(event.ID(), "No place for event photo in queue, try again later")
				if _, err := server.bot.Send(msg); err != nil {
					chat.CantSendMessage(err)
				}

				slog.Info("No place for event photo in queue", "event", event)

				continue
			}
			server.photos <- event.Photo()
		}
	}
}

func (server *server) photosHandler() {
	cmd := exec.Command("./motor_driver.bin", "0", "0", "True", "0", "3", "")
	if err := cmd.Run(); err != nil {
		slog.Error("Can not initialise the camera", "err", err)
		os.Exit(1)
	}

	var currentX int
	for photo := range server.photos {
		slog.Info("Taking photo", "photo", photo)

		data, err := photo.Take(currentX)
		if err != nil {
			server.phoneInit(photo, err)
			continue
		}
		currentX = photo.X

		msg := tgbotapi.NewPhoto(photo.ID, tgbotapi.FileBytes{Name: "image.jpg", Bytes: data})
		msg.Caption = fmt.Sprintf("X: %v Y: %v", photo.X, photo.Y)
		if _, err := server.bot.Send(msg); err != nil {
			chat.CantSendMessage(err)
		}
	}
}

func (server *server) sunsetHandler(city string, owm *openweathermap.OpenWeatherMap) {
	for {
		sunset, err := owm.SunsetTime(city)
		if err != nil {
			slog.Error("Can not get sunset time", "err", err)
		}

		slog.Info("sunset time", "time", sunset)

		server.config.SunsetTime = sunset
		time.Sleep(time.Hour * 24)
	}
}
