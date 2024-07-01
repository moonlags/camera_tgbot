package main

import (
	"fmt"
	"log"
	"math/rand"
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
	password      string
	guestPassword string
}

func NewServer(bot *tgbotapi.BotAPI) *Server {
	server := &Server{
		bot:    bot,
		chats:  make(map[int64]*Chat),
		events: make(map[int64]Event),
		photos: make(chan Photo),
		vars: &ServerVars{
			password:      os.Getenv("PASSWORD"),
			guestPassword: fmt.Sprint(rand.Uint64()),
		},
	}
	return server
}

func (server *Server) Run(config tgbotapi.UpdateConfig) {
	go server.photosHandler()
	go server.eventsHandler()

	updates := server.bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		server.handleOrCreateChat(update)
	}
}

func (server *Server) photosHandler() {
	cmd := exec.Command("./motor_driver.bin", "0", "0", "True", "0", "3", "")
	if err := cmd.Run(); err != nil {
		log.Fatal("Error turning camera to 0:", err)
	}
	var currentX int
	for photo := range server.photos {
		cmd := exec.Command("./motor_driver.bin", fmt.Sprint(photo.x), fmt.Sprint(photo.y), "False", fmt.Sprint(currentX), "3", "wget -N -P . http://127.0.0.1:8080/photoaf.jpg")
		if err := cmd.Run(); err != nil {
			log.Fatal("Error taking a shot:", err)
		}
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
