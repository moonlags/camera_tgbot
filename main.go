package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var PASSWORD = os.Getenv("CAMERA_PASSWORD")

func main() {
	camera, err := newCamera()
	if err != nil {
		log.Fatal("failed to initialize camera", err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("failed to initialize bot", err)
	}

	chats := make(map[int64]Chat)
	events := make(map[int64]*[]Event)
	sunsetTime := time.Now()
	guestPassword := fmt.Sprint(rand.Uint32())

	go photoHandler(bot, camera)
	go eventsHandler(bot, events, camera)
	go sunsetHandler("Jurmala", &sunsetTime)

	u := tgbotapi.NewUpdate(-1)
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("recieved message %s from %s\n", update.Message.Text, update.Message.From.FirstName)

		chatid := update.Message.Chat.ID
		chat, ok := chats[chatid]
		if !ok {
			chatEvents := make([]Event, 0)

			chat = newChat(bot, camera, &chatEvents, &sunsetTime, &guestPassword)
			chats[chatid] = chat

			events[chatid] = &chatEvents

			expireChat(time.After(time.Hour*8), chatid, chats)
		}
	}
}

func expireChat(timech <-chan time.Time, chatID int64, chats map[int64]Chat) {
	<-timech
	delete(chats, chatID)
}

func photoHandler(bot *tgbotapi.BotAPI, cam *Camera) {
	for {
		photo, err := cam.take()
		if err != nil {
			log.Println("failed to take photo", err)

			if err := cam.phoneInit(); err != nil {
				log.Println("failed to initialize phone", err)
			}

			if !photo.retry {
				photo.retry = true
				cam.queuePhoto(photo)
			}
			continue
		}

		msg := tgbotapi.NewPhoto(photo.reciever, tgbotapi.FilePath("photoaf.jpg"))
		msg.Caption = fmt.Sprintf("X: %d Y:%d ZOOM:%d MODE: %d", photo.x, photo.y, photo.zoom, photo.mode)

		if _, err := bot.Send(msg); err != nil {
			log.Println("failed to send photo", err)
		}
	}
}

func eventsHandler(bot *tgbotapi.BotAPI, events map[int64]*[]Event, cam *Camera) {
	for {
		time.Sleep(time.Minute)
		for _, chatEvents := range events {
			for _, event := range *chatEvents {
				if !event.isReady() {
					continue
				}
				photo := event.eventPhoto()
				log.Println("event is ready", event.eventPhoto())

				if err := cam.queuePhoto(photo); err != nil {
					msg := tgbotapi.NewMessage(photo.reciever, err.Error())
					if _, err := bot.Send(msg); err != nil {
						log.Println("failed to send photo", err)
					}
				}
			}
		}
	}
}

func sunsetHandler(city string, sunsetTime *time.Time) {
	apiKey := os.Getenv("OWM_KEY")
	if apiKey == "" {
		log.Fatal("OWM_KEY variable is not set")
	}

	for {
		sunset, err := getSunsetTime(apiKey, city)
		if err != nil {
			slog.Error("failed to get sunset time", "err", err)
			time.Sleep(time.Hour * 24)
			continue
		}

		log.Print("sunset time", sunset)
		*sunsetTime = sunset

		time.Sleep(time.Hour * 24)
	}
}
