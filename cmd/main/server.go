package main

import (
	"log/slog"
	"os"
	"time"

	"example/sashaTelegram/internal/chat"
	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/openweathermap"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type server struct {
	bot    *tgbotapi.BotAPI
	chats  map[int64]*chat.Chat
	events map[int64]event.Event
	photos chan *photo.Photo
	config *chat.Config
}

func (server *server) run() {
	if err := server.loadGobData(); err != nil {
		slog.Warn("Can not load gob data", "err", err)
	}

	go server.saveHandler()
	go server.photosHandler()
	go server.eventsHandler()
	go server.sunsetHandler("Jurmala", openweathermap.New(os.Getenv("WEATHER_KEY")))

	updates := server.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for update := range updates {
		if update.Message == nil {
			continue
		}

		slog.Info("message", "text", update.Message.Text, "firstname", update.Message.From.FirstName)

		go server.passToChat(update)
	}
}

func (s *server) passToChat(update tgbotapi.Update) {
	if _, ok := s.chats[update.FromChat().ID]; !ok {
		slog.Info("new chat", "firstname", update.Message.From.FirstName)

		chat := chat.New(s.bot, update.FromChat().ID, s.photos, s.events, s.config)
		s.chats[chat.ID] = chat
		go s.chatDestruct(chat.ID, time.Hour*8)
	}
	s.chats[update.FromChat().ID].Handle(update)
}

func (server *server) chatDestruct(id int64, dur time.Duration) {
	time.Sleep(dur)
	delete(server.chats, id)
}
