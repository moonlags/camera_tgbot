package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"os"

	"example/sashaTelegram/internal/chat"
	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func init() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables")
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	server := server{
		bot:    bot,
		chats:  make(map[int64]*chat.Chat),
		events: make(map[int64]event.Event),
		photos: make(chan *photo.Photo, 5),
		config: &chat.Config{
			Password:      os.Getenv("PASSWORD"),
			GuestPassword: fmt.Sprint(rand.Uint32()),
		},
	}

	server.run()
}
