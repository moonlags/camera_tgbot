package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}
	bot.Debug = true

	server := NewServer(bot)
	server.Run(tgbotapi.NewUpdate(0))
}
