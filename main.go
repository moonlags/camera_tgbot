package main

import (
	"log"
	"os"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load env variables")
	}
}

func main() {
	defer func() {
		recover()
		debug.PrintStack()
	}()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}
	server := NewServer(bot)
	server.Run(tgbotapi.NewUpdate(0))
}
