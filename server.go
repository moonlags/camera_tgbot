package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NicoNex/echotron/v3"
	"github.com/moonlags/camera_tgbot/internal/openweathermap"
)

const CITY = "Jurmala"

type server struct {
	dsp    *echotron.Dispatcher
	config config
	events map[int64]event
	sunset *time.Time
	camera *camera
	echotron.API
}

func newServer(c config) (server, error) {
	camera, err := newCamera()
	if err != nil {
		return server{}, err
	}

	server := server{
		config: c,
		events: make(map[int64]event),
		camera: camera,
		sunset: &time.Time{},
		API:    echotron.NewAPI(c.token),
	}

	dsp := echotron.NewDispatcher(c.token, server.newBot)
	server.dsp = dsp

	return server, nil
}

func (s *server) newBot(chatID int64) echotron.Bot {
	slog.Info("new bot", "chatID", chatID)

	bot := &bot{
		chatID: chatID,
		config: botConfig{
			password:      s.config.password,
			guestPassword: &s.config.guestPassword,
			sunset:        s.sunset,
		},
		events: s.events,
		camera: s.camera,
		API:    echotron.NewAPI(s.config.token),
	}
	bot.state = bot.handleMessage
	go s.destructBot(time.After(time.Hour*8), chatID)

	return bot
}

func (s *server) destructBot(timech <-chan time.Time, chatID int64) {
	<-timech
	slog.Info("destructing bot", "chatID", chatID)
	s.dsp.DelSession(chatID)
}

func (s *server) run() {
	go s.photoHandler()
	go s.eventsHandler()
	go s.sunsetHandler(CITY)

	for {
		slog.Error("server returned an error", "err", s.dsp.Poll())
		time.Sleep(time.Minute * 10)
	}
}

func (s *server) photoHandler() {
	for {
		photo, err := s.camera.take()
		if err != nil {
			slog.Error("failed to take photo", "id", photo.reciever, "err", err)
			if err := s.camera.phoneInit(); err != nil {
				slog.Error("failed to initialize phone:", "err", err)
			}

			if !photo.retry {
				photo.retry = true
				s.camera.queuePhoto(photo)
			}
			continue
		}

		if _, err := s.SendPhoto(echotron.NewInputFilePath("photoaf.jpg"), photo.reciever, &echotron.PhotoOptions{
			Caption: fmt.Sprintf("x: %d y: %d", photo.x, photo.y),
		}); err != nil {
			slog.Error("failed to send photo", "id", photo.reciever, "err", err)
			continue
		}
	}
}

func (s *server) eventsHandler() {
	for {
		time.Sleep(time.Minute)
		for _, event := range s.events {
			if !event.isReady() {
				continue
			}

			photo := event.eventPhoto()
			slog.Info("event is ready", "photo", event.eventPhoto())

			if err := s.camera.queuePhoto(photo); err != nil {
				if _, err := s.SendMessage(err.Error(), photo.reciever, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}
		}
	}
}

func (s *server) sunsetHandler(city string) {
	owm := openweathermap.New(s.config.owmKey)

	for {
		sunset, err := owm.SunsetTime(city)
		if err != nil {
			slog.Error("failed to get sunset time", "err", err)
			time.Sleep(time.Hour * 24)
			continue
		}

		slog.Info("sunset", "time", sunset)
		*s.sunset = sunset

		time.Sleep(time.Hour * 24)
	}
}
