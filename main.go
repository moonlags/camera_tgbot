package main

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	PhotoError uint8 = iota
	PhotoReady
)

var PASSWORD = os.Getenv("CAMERA_PASSWORD")

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatalln("failed to open database", err)
	}
	defer db.Close()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("failed to initialize bot", err)
	}

	chats := make(map[int64]*Chat)
	sunsetTime := time.Now()
	guestPassword := fmt.Sprint(rand.Uint32())

	events, err := getEventsFromDB(db, &sunsetTime)
	if err != nil {
		log.Fatalln("failed to get events from database", err)
	}

	photoRequests := make(chan Photo)
	go tcpHandler(bot, photoRequests)
	go sunsetHandler("Jurmala", &sunsetTime)
	go eventsHandler(events, photoRequests)

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

			if _, ok := events[chatid]; !ok {
				*events[chatid] = make([]Event, 0)
			}
			chatEvents := *events[chatid]

			chat := newChat(bot, &chatEvents, photoRequests, db, &sunsetTime, &guestPassword)
			chats[chatid] = &chat

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

	for {
		log.Println("connection to", address)

		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Println("failed to connect to camera", err)
			time.Sleep(time.Minute)
		}

		photo := handleRequests(bot, conn, photoRequests)
		queuePhoto(photoRequests, photo)
	}
}

func handleRequests(bot *tgbotapi.BotAPI, conn net.Conn, photoRequests chan Photo) Photo {
	for photo := range photoRequests {
		if _, err := conn.Write([]byte(PASSWORD)); err != nil {
			log.Println("failed to write password to connection:", err)
			return photo
		}

		if err := binary.Write(conn, binary.BigEndian, photo.toConfig()); err != nil {
			log.Println("failed to write photo config", err)
			return photo
		}

		var code uint8
		if err := binary.Read(conn, binary.BigEndian, &code); err != nil {
			log.Println("failed to read message code", err)
			return photo
		}
		log.Println("recieved code", code)

		if code == PhotoReady {
			var lenght int32
			if err := binary.Read(conn, binary.BigEndian, &lenght); err != nil {
				log.Println("failed to read photo lenght", err)
				return photo
			}
			log.Println("recieved lenght", lenght)

			photoData := make([]byte, lenght)
			if _, err := io.ReadFull(conn, photoData); err != nil {
				log.Println("failed to read photo data", err)
				return photo
			}

			msg := tgbotapi.NewPhoto(photo.reciever, tgbotapi.FileBytes{Name: "photoaf.jpg", Bytes: photoData})
			msg.Caption = fmt.Sprintf("X: %d Y: %d ZOOM: %d MODE: %d", photo.x, photo.y, photo.zoom, photo.mode)
			if _, err := bot.Send(msg); err != nil {
				log.Println("failed to send message", err)
			}
		} else {
			msg := tgbotapi.NewMessage(photo.reciever, "Try again later")
			if _, err := bot.Send(msg); err != nil {
				log.Println("failed to send message", err)
			}
		}
	}
	return Photo{} // unreachable
}

func eventsHandler(events map[int64]*[]Event, photoRequests chan Photo) {
	for {
		time.Sleep(time.Minute)
		for _, chatEvents := range events {
			for _, event := range *chatEvents {
				if !event.isReady() {
					continue
				}
				photo := event.eventPhoto()
				log.Println("event is ready", event.eventPhoto())

				go queuePhoto(photoRequests, photo)
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

func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./events.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS events (userid INTEGER, x INTEGER, y INTEGER, zoom INTEGER, mode INTEGER, hour INTEGER, minute INTEGER, sunset BOOL)")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEventsFromDB(db *sql.DB, sunsetTime *time.Time) (map[int64]*[]Event, error) {
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		return nil, err
	}

	events := make(map[int64]*[]Event)
	for rows.Next() {
		var event EventDB
		if err := rows.Scan(&event.Userid, &event.X, &event.Y, &event.Zoom, &event.Mode, &event.Hour, &event.Minute, &event.Sunset); err != nil {
			log.Println("failed to scan db row", err)
			continue
		}
		log.Println("read from db", event)

		photo, err := newPhoto(event.Userid, event.X, event.Y, event.Zoom, event.Mode)
		if err != nil {
			log.Println("db has invalid values", err)
			continue
		}

		if _, ok := events[event.Userid]; !ok {
			*events[event.Userid] = make([]Event, 0)
		}

		if event.Sunset {
			ev := newSunsetEvent(photo, sunsetTime)
			*events[event.Userid] = append(*events[event.Userid], &ev)
		} else {
			ev, err := newStaticEvent(photo, event.Hour, event.Minute)
			if err != nil {
				log.Println("db has invalid values for hour and minute", err)
				continue
			}
			*events[event.Userid] = append(*events[event.Userid], &ev)
		}
	}
	return events, nil
}
