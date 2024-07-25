package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"time"

	"example/sashaTelegram/internal/chat"
	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot    *tgbotapi.BotAPI
	chats  map[int64]*chat.Chat
	events map[int64]event.Event
	photos chan *photo.Photo
	config *chat.Config
}

func New(bot *tgbotapi.BotAPI, password string) *Server {
	server := &Server{
		bot:    bot,
		chats:  make(map[int64]*chat.Chat),
		events: make(map[int64]event.Event),
		photos: make(chan *photo.Photo, 5),
		config: &chat.Config{
			Password:      password,
			GuestPassword: fmt.Sprint(rand.Uint32()),
		},
	}
	return server
}

func (server *Server) Run(config tgbotapi.UpdateConfig) {
	go server.photosHandler()
	go server.eventsHandler()
	go server.sunsetHandler()

	updates := server.bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		go server.passToChat(update)
	}
}

func (server *Server) sunsetHandler() {
	for {
		url := "https://api.sunrisesunset.io/json?lat=56.968&lng=23.77038&timezone=UTC"
		response, err := http.Get(url)
		if err != nil {
			log.Fatal("Failed to make http request:", err)
		}
		var respStruct struct {
			Results struct {
				Date   string `json:"date"`
				Sunset string `json:"sunset"`
			} `json:"results"`
		}
		if err := json.NewDecoder(response.Body).Decode(&respStruct); err != nil {
			log.Fatal("Failed to decode json:", err)
		}
		response.Body.Close()
		sunsetTime, err := time.Parse("2006-01-02 3:04:05 PM", respStruct.Results.Date+" "+respStruct.Results.Sunset)
		if err != nil {
			log.Fatal("Failed to parse sunset time:", err)
		}
		server.config.SunsetTime = sunsetTime.Local()
		time.Sleep(time.Hour * 24)
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

func (s *Server) passToChat(update tgbotapi.Update) {
	if _, ok := s.chats[update.FromChat().ID]; !ok {
		chat := chat.New(s.bot, update.FromChat().ID, s.photos, s.events, s.config)
		s.chats[chat.ID] = chat
		go s.chatDestruct(chat.ID, time.Hour*8)
	}
	s.chats[update.FromChat().ID].Handle(update)
}

func (server *Server) chatDestruct(id int64, dur time.Duration) {
	time.Sleep(dur)
	delete(server.chats, id)
}
