package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	token         string
	password      string
	owmKey        string
	guestPassword string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file:", err)
	}
}

func main() {
	config := config{
		token:         os.Getenv("BOT_TOKEN"),
		password:      os.Getenv("BOT_PASSWORD"),
		owmKey:        os.Getenv("OWM_KEY"),
		guestPassword: fmt.Sprint(rand.Uint32()),
	}

	server, err := newServer(config)
	if err != nil {
		log.Fatal("failed to initialize server:", err)
	}

	server.run()
}
