package chat

import (
	"fmt"
	"math/rand"

	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (chat *Chat) guestCommandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.ID,
			"/help -  Get a list of commands\n/random - Take a random photo\n/sunsettime - Get sunset time\n")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.guestHandler
	case "/random":
		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %v Y: %v Zoom: %v Mode: %d", x, y, zoom, mode))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}

		photo, _ := photo.New(x, y, zoom, mode, chat.ID)
		chat.QueuePhoto(update, photo)

		return chat.guestHandler
	case "/sunsettime":
		stime := chat.Config.SunsetTime
		msg := tgbotapi.NewMessage(
			chat.ID,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()),
		)
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.guestHandler
	}
	return nil
}

func (chat *Chat) guestHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.guestCommandsHandler(update); handler != nil {
		return handler
	}

	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.ID,
			"Values should be in format of 'X Y'")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.guestHandler
	}

	photo, err := photo.New(x, y, chat.Zoom, chat.Mode, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, "Error: "+err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.guestHandler
	}

	msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %d Y: %d", x, y))
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}

	chat.QueuePhoto(update, photo)

	return chat.guestHandler
}
