package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	bot     *tgbotapi.BotAPI
	id      int64
	photos  chan Photo
	events  map[int64]Event
	handler handlerFn
	vars    *ServerVars
}

func (chat *Chat) handle(update tgbotapi.Update) {
	chat.handler = chat.handler(update)
}

type handlerFn func(tgbotapi.Update) handlerFn

func (chat *Chat) unauthorizedHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case chat.vars.password:
		msg := tgbotapi.NewMessage(chat.id, "Welcome, try sending a coordinates")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case chat.vars.guestPassword:
		msg := tgbotapi.NewMessage(chat.id, "Welcome, you have entered as guest, try sending a coordinates")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	}
	msg := tgbotapi.NewMessage(chat.id, "Hello, please send a password")
	if _, err := chat.bot.Send(msg); err != nil {
		log.Fatal("Error sending a message:", err)
	}
	return chat.unauthorizedHandler
}

func (chat *Chat) commandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.id, "/help -  Get a list of commands\n/dice - Throw a dice and take a photo\n/eventcreate - Create an event\n/eventsunset - Create sunset event\n/eventdelete - Delete an event\n/sunsettime - Get sunset time\n/guestpass - Generate guest password")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	case "dice":
		x, y := 360/6*chat.diceRoll(), 90/6*chat.diceRoll()
		time.Sleep(5 * time.Second)
		msg := tgbotapi.NewMessage(chat.id, fmt.Sprintf("Taking photo on X: %v Y: %v", x, y))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.requestPhoto(x, y)
	}
	return nil
}

func (chat *Chat) protectedHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, "Values should be in format of 'X Y', where X and Y are integers")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	}
	if err := chat.requestPhoto(x, y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, err.Error())
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	}
	return chat.protectedHandler
}

func (chat *Chat) guestCommandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.id, "/help -  Get a list of commands\n/dice - Throw a dice and take a photo\n/sunsettime - Get sunset time\n")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	case "dice":
		x, y := 360/6*chat.diceRoll(), 90/6*chat.diceRoll()
		time.Sleep(5 * time.Second)
		msg := tgbotapi.NewMessage(chat.id, fmt.Sprintf("Taking photo on X: %v Y: %v", x, y))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.requestPhoto(x, y)
	}
	return nil
}

func (chat *Chat) guestHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.guestCommandsHandler(update); handler != nil {
		return handler
	}
	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, "Values should be in format of 'X Y', where X and Y are integers")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	}
	if err := chat.requestPhoto(x, y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, err.Error())
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
	}
	return chat.guestHandler
}

func (chat *Chat) requestPhoto(x int, y int) error {
	if x < 0 || x > 360 || y < 0 || y > 90 {
		return fmt.Errorf("X should be in range 0 - 360, Y in range 0 - 90")
	}
	chat.photos <- Photo{x, y, chat.id}
	return nil
}

func (chat *Chat) diceRoll() int {
	dice := tgbotapi.NewDice(chat.id)
	result, err := chat.bot.Send(dice)
	if err != nil {
		log.Fatal("Error sending a dice:", err)
	}
	return result.Dice.Value
}
