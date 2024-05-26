package main

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot      *tgbotapi.BotAPI
	chats    map[int64]Chat
	events   map[int64]Event
	photosCh chan Photo
}

func NewServer(bot *tgbotapi.BotAPI) *Server {
	server := &Server{
		bot:      bot,
		chats:    make(map[int64]Chat),
		events:   make(map[int64]Event),
		photosCh: make(chan Photo),
	}
	return server
}

func (server *Server) Run(config tgbotapi.UpdateConfig) {
	updates := server.bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		server.sessions
	}
}

func (server *Server) eventsHandler() {
	for {
		time.Sleep(time.Minute)
		for _, event := range server.events {
			if !event.IsReady() {
				continue
			}
		}
	}
}
