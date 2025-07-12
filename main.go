package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"time"
	"unsafe"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	PhotoError uint8 = iota
	PhotoReady
)

var PASSWORD = os.Getenv("CAMERA_PASSWORD")

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("failed to initialize bot", err)
	}

	chats := make(map[int64]*Chat)
	events := make(map[int64]*[]Event)
	sunsetTime := time.Now()
	guestPassword := fmt.Sprint(rand.Uint32())

	photoRequests := make(chan Photo)
	go tcpHandler(bot, photoRequests)
	go eventsHandler(bot, events, photoRequests)
	go sunsetHandler("Jurmala", &sunsetTime)

	u := tgbotapi.NewUpdate(-1)
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("recieved message %s from %s\n", update.Message.Text, update.Message.From.FirstName)

		chatid := update.Message.Chat.ID
		if _, ok := chats[chatid]; !ok {
			log.Println("creating new chat for", update.Message.From.FirstName)

			var chatEvents []Event
			if _, ok := events[chatid]; ok {
				chatEvents = *events[chatid]
			} else {
				chatEvents = make([]Event, 0)
			}

			chat := newChat(bot, &chatEvents, &sunsetTime, &guestPassword)
			chats[chatid] = &chat

			events[chatid] = &chatEvents

			go expireChat(time.After(time.Hour*8), chatid, chats)
		}

		go chats[chatid].handleUpdate(update)
	}
}

func expireChat(timech <-chan time.Time, chatID int64, chats map[int64]*Chat) {
	<-timech
	delete(chats, chatID)
}

func tcpHandler(bot *tgbotapi.BotAPI, photoRequests chan Photo) {
	address := os.Getenv("TCP_ADDRESS")
	if address == "" {
		log.Fatal("TCP_ADDRESS variable is not set")
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("failed to connect to camera", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for photo := range photoRequests {
		buf := make([]byte, len(PASSWORD)+int(unsafe.Sizeof(PhotoConfig{})))
		binary.Encode(buf, binary.BigEndian, PASSWORD)

		if _, err := binary.Encode(buf, binary.BigEndian, photo.toConfig()); err != nil {
			log.Println("failed to encode photo config", err)
			continue
		}

		var code uint8
		if err := binary.Read(reader, binary.BigEndian, &code); err != nil {
			log.Println("failed to read message code", err)
			break
		}

		if code == PhotoReady {
			var lenght int32
			if err := binary.Read(reader, binary.BigEndian, &lenght); err != nil {
				log.Println("failed to read photo lenght", err)
				break
			}

			photoData := make([]byte, lenght)
			if _, err := io.ReadFull(reader, photoData); err != nil {
				log.Println("failed to read photo data", err)
			}

			msg := tgbotapi.NewPhoto(photo.reciever, tgbotapi.FileBytes{Name: "photoaf.jpg", Bytes: photoData})
			msg.Caption = fmt.Sprintf("X: %d Y: %d ZOOM: %d MODE: %d", photo.x, photo.y, photo.zoom, photo.mode)

			if _, err := bot.Send(msg); err != nil {
				log.Println("failed to send photo", err)
			}
		}

	}
}

func eventsHandler(bot *tgbotapi.BotAPI, events map[int64]*[]Event, photoRequests chan Photo) {
	for {
		time.Sleep(time.Minute)
		for _, chatEvents := range events {
			for _, event := range *chatEvents {
				if !event.isReady() {
					continue
				}
				photo := event.eventPhoto()
				log.Println("event is ready", event.eventPhoto())

				if err := cam.queuePhoto(photo); err != nil {
					msg := tgbotapi.NewMessage(photo.reciever, err.Error())
					if _, err := bot.Send(msg); err != nil {
						log.Println("failed to send photo", err)
					}
				}
			}
		}
	}
}

func sunsetHandler(city string, sunsetTime *time.Time) {
	apiKey := os.Getenv("OWM_KEY")
	if apiKey == "" {
		log.Fatal("OWM_KEY variable is not set")
	}

	for {
		sunset, err := getSunsetTime(apiKey, city)
		if err != nil {
			log.Println("failed to get sunset time", err)
			time.Sleep(time.Hour * 24)
			continue
		}

		log.Println("sunset time", sunset)
		*sunsetTime = sunset

		time.Sleep(time.Hour * 24)
	}
}
