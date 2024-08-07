package chat

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	ID      int64
	Config  *Config
	Events  map[int64]event.Event
	Handler handlerFn
	Bot     *tgbotapi.BotAPI
	Photos  chan *photo.Photo
	Zoom    int
	Mode    int
}

type Config struct {
	SunsetTime    time.Time
	Password      string
	GuestPassword string
}

type handlerFn func(tgbotapi.Update) handlerFn

func New(bot *tgbotapi.BotAPI, id int64, photos chan *photo.Photo, events map[int64]event.Event, c *Config) *Chat {
	chat := &Chat{
		Bot:    bot,
		ID:     id,
		Photos: photos,
		Events: events,
		Config: c,
	}
	chat.Handler = chat.unauthorizedHandler
	return chat
}

func (chat *Chat) Handle(update tgbotapi.Update) {
	chat.Handler = chat.Handler(update)
}

func (chat *Chat) unauthorizedHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case chat.Config.Password:
		msg := tgbotapi.NewMessage(chat.ID, "Welcome, try sending a coordinates")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	case chat.Config.GuestPassword:
		msg := tgbotapi.NewMessage(chat.ID, "Welcome, you have entered as guest, try sending a coordinates")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		chat.Config.GuestPassword = fmt.Sprint(rand.Uint32())

		slog.Info("guest pass", "password", chat.Config.GuestPassword)

		return chat.guestHandler
	}
	msg := tgbotapi.NewMessage(chat.ID, "Hello, please send a password")
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}
	return chat.unauthorizedHandler
}
