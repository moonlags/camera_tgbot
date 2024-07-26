package server

import (
	"fmt"
	"math/rand"
	"time"

	"example/sashaTelegram/internal/chat"
	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/openweathermap"
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

func (server *Server) Run(config tgbotapi.UpdateConfig, owm *openweathermap.Config) {
	go server.photosHandler()
	go server.eventsHandler()
	go server.sunsetHandler("Jurmala", owm)

	updates := server.bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		go server.passToChat(update)
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
