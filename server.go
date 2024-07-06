package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot    *tgbotapi.BotAPI
	chats  map[int64]*Chat
	events map[int64]Event
	photos chan Photo
	vars   *ServerVars
}

type ServerVars struct {
	sunsetTime    time.Time
	sunriseTime   time.Time
	password      string
	guestPassword string
}

func NewServer(bot *tgbotapi.BotAPI) *Server {
	server := &Server{
		bot:    bot,
		chats:  make(map[int64]*Chat),
		events: make(map[int64]Event),
		photos: make(chan Photo, 5),
		vars: &ServerVars{
			password:      os.Getenv("PASSWORD"),
			guestPassword: fmt.Sprint(rand.Uint32()),
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
		go server.handleOrCreateChat(update)
	}
}

func (server *Server) sunsetHandler() {
	for {
		url := "https://api.sunrisesunset.io/json?lat=56.968&lng=23.77038&timezone=UTC"
		response, err := http.Get(url)
		if err != nil {
			log.Fatal("Failed to make http request:", err)
		}
		defer response.Body.Close()
		var respStruct struct {
			Results struct {
				Date    string `json:"date"`
				Sunset  string `json:"sunset"`
				Sunrise string `json:"sunrise"`
			} `json:"results"`
		}
		if err := json.NewDecoder(response.Body).Decode(&respStruct); err != nil {
			log.Fatal("Failed to decode json:", err)
		}
		sunsetTime, err := time.Parse("2006-01-02 3:04:05 PM", respStruct.Results.Date+" "+respStruct.Results.Sunset)
		if err != nil {
			log.Fatal("Failed to parse sunset time:", err)
		}
		sunriseTime, err := time.Parse("2006-01-02 3:04:05 PM", respStruct.Results.Date+" "+respStruct.Results.Sunrise)
		if err != nil {
			log.Fatal("Failed to parse sunrise time:", err)
		}
		server.vars.sunriseTime = sunriseTime.Local()
		server.vars.sunsetTime = sunsetTime.Local()
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
		setZoom(photo.zoom)
		cmd := exec.Command("./motor_driver.bin", fmt.Sprint(photo.x), fmt.Sprint(photo.y), "False", fmt.Sprint(currentX), "3", "wget -N -P . http://127.0.0.1:8080/photoaf.jpg")
		if err := cmd.Run(); err != nil {
			log.Fatal("Error taking a shot:", err)
		}
		setZoom(0)
		currentX = photo.x
		file, err := os.Open("photoaf.jpg")
		if err != nil {
			exec.Command("./phone_init.sh").Run()
			msg := tgbotapi.NewMessage(photo.id, "Error occured, please try again")
			if _, err := server.bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			continue
		}
		msg := tgbotapi.NewPhoto(photo.id, tgbotapi.FileReader{Name: "image.jpg", Reader: file})
		msg.Caption = fmt.Sprintf("X: %v Y: %v", photo.x, photo.y)
		if _, err := server.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		os.Remove("photoaf.jpg")
	}
}

func (server *Server) eventsHandler() {
	for {
		time.Sleep(time.Minute)
		for _, event := range server.events {
			if !event.IsReady() {
				continue
			}
			x, y := event.GetPos()
			if len(server.photos) >= 5 {
				msg := tgbotapi.NewMessage(event.GetID(), "No place for event photo in queue, try again later")
				if _, err := server.bot.Send(msg); err != nil {
					log.Fatal("Failed to send a message:", err)
				}
				continue
			}
			server.photos <- Photo{x: x, y: y, id: event.GetID()}
		}
	}
}

func (server *Server) handleOrCreateChat(update tgbotapi.Update) {
	if _, ok := server.chats[update.FromChat().ID]; !ok {
		chat := &Chat{
			bot:    server.bot,
			id:     update.FromChat().ID,
			photos: server.photos,
			events: server.events,
			vars:   server.vars,
		}
		chat.handler = chat.unauthorizedHandler
		server.chats[chat.id] = chat
		go server.chatDestruct(chat.id, time.Hour*8)
	}
	server.chats[update.FromChat().ID].handle(update)
}

func (server *Server) chatDestruct(id int64, dur time.Duration) {
	time.Sleep(dur)
	delete(server.chats, id)
}
