package main

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot    *tgbotapi.BotAPI
	chats  map[int64]*Chat
	events map[int64]Event
	photos chan Photo
}

func NewServer(bot *tgbotapi.BotAPI) *Server {
	server := &Server{
		bot:    bot,
		chats:  make(map[int64]*Chat),
		events: make(map[int64]Event),
		photos: make(chan Photo),
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
	//	for photo := range server.photos {
	// todo
	//	}
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
	if _, ok := server.chats[update.FromChat().ID]; ok {
		chat := &Chat{
			id:     update.FromChat().ID,
			photos: server.photos,
			events: server.events,
		}
		chat.handler = chat.UnathorizedHandler
		server.chats[chat.id] = chat
	}
	server.chats[update.FromChat().ID].handle(update) // todo
}
