package main

import (
	"log"
	"os"

	"example/sashaTelegram/internal/server"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables")
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}
	server := server.New(bot, os.Getenv("PASSWORD"))
	server.Run(tgbotapi.NewUpdate(0))
}
