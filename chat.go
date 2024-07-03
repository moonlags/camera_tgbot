package main

import (
	"fmt"
	"log"
	"math/rand"
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
		msg := tgbotapi.NewMessage(chat.id,
			"Welcome, you have entered as guest, try sending a coordinates")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.vars.guestPassword = fmt.Sprint(rand.Uint32())
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
		msg := tgbotapi.NewMessage(chat.id,
			"/help -  Get a list of commands\n/dice - Throw a dice and take a photo\n/eventcreate - Create an event\n/eventsunset - Create sunset event\n/eventdelete - Delete an event\n/sunsettime - Get sunset time\n/guestpass - Generate guest password")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case "/dice":
		x, y := 360/6*chat.diceRoll(), 90/6*chat.diceRoll()
		time.Sleep(5 * time.Second)
		msg := tgbotapi.NewMessage(chat.id, fmt.Sprintf("Taking photo on X: %v Y: %v", x, y))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.requestPhoto(x, y)
		return chat.protectedHandler
	case "/sunsettime":
		stime := chat.vars.sunsetTime
		msg := tgbotapi.NewMessage(chat.id,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case "/eventcreate":
		if _, ok := chat.events[chat.id]; ok {
			msg := tgbotapi.NewMessage(chat.id, "Please delete your existing event first")
			if _, err := chat.bot.Send(msg); err != nil {
				log.Fatal("Error sending a message:", err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.id, "Specify 'X Y Hour Minute' to create an event")
			if _, err := chat.bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			return chat.eventCreationHanlder
		}
	case "/eventsunset":
		if _, ok := chat.events[chat.id]; ok {
			msg := tgbotapi.NewMessage(chat.id, "Please delete your existing event first")
			if _, err := chat.bot.Send(msg); err != nil {
				log.Fatal("Error sending a message:", err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.id, "Specify 'X Y' to create an event")
			if _, err := chat.bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			return chat.sunsetEventCreationHanlder
		}
	case "/eventdelete":
		delete(chat.events, chat.id)
		msg := tgbotapi.NewMessage(chat.id, "Event deleted")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.protectedHandler
	case "/guestpass":
		msg := tgbotapi.NewMessage(chat.id, fmt.Sprintf("Guest password is %v", chat.vars.guestPassword))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.protectedHandler
	}
	return nil
}

func (chat *Chat) sunsetEventCreationHanlder(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, "Specify 'X Y' in valid format")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.sunsetEventCreationHanlder
	}
	if x < 0 || x > 360 || y < 0 || y > 90 {
		msg := tgbotapi.NewMessage(chat.id, "X should be in range 0 - 360, Y in range 0 - 90")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.sunsetEventCreationHanlder
	}
	chat.events[chat.id] = &SunsetEvent{X: x, Y: y, ID: chat.id, Sunset: &chat.vars.sunsetTime}
	msg := tgbotapi.NewMessage(chat.id, "Sunset event created")
	if _, err := chat.bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}
	return chat.protectedHandler
}

func (chat *Chat) eventCreationHanlder(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var x, y, hour, minute int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d %d %d", &x, &y, &hour, &minute); err != nil {
		msg := tgbotapi.NewMessage(chat.id, "Specify 'X Y Hour Minute' in valid format")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.eventCreationHanlder
	}
	if x < 0 || x > 360 || y < 0 || y > 90 || hour < 0 || hour > 24 || minute < 0 || minute > 59 {
		msg := tgbotapi.NewMessage(chat.id, "X should be in range 0 - 360, Y 0 - 90, Hour 0 - 24, Minute 0 - 59")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.eventCreationHanlder
	}
	chat.events[chat.id] = &StaticEvent{X: x, Y: y, ShotTime: time.Date(0, 0, 0, hour, minute, 0, 0, time.Local), ID: chat.id}
	msg := tgbotapi.NewMessage(chat.id, "Event created")
	if _, err := chat.bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}
	return chat.protectedHandler
}

func (chat *Chat) protectedHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.id,
			"Values should be in format of 'X Y'")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	}
	if err := chat.requestPhoto(x, y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, err.Error())
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	}
	msg := tgbotapi.NewMessage(chat.id, "Your photo is in the queue, please wait")
	if _, err := chat.bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}
	return chat.protectedHandler
}

func (chat *Chat) guestCommandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.id,
			"/help -  Get a list of commands\n/dice - Throw a dice and take a photo\n/sunsettime - Get sunset time\n")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	case "/dice":
		x, y := 360/6*chat.diceRoll(), 90/6*chat.diceRoll()
		time.Sleep(5 * time.Second)
		msg := tgbotapi.NewMessage(chat.id, fmt.Sprintf("Taking photo on X: %v Y: %v", x, y))
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.requestPhoto(x, y)
		return chat.guestHandler
	case "/sunsettime":
		stime := chat.vars.sunsetTime
		msg := tgbotapi.NewMessage(
			chat.id,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()),
		)
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
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
		msg := tgbotapi.NewMessage(chat.id,
			"Values should be in format of 'X Y'")
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	}
	if err := chat.requestPhoto(x, y); err != nil {
		msg := tgbotapi.NewMessage(chat.id, err.Error())
		if _, err := chat.bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	}
	msg := tgbotapi.NewMessage(chat.id, "Your photo is in the queue, please wait")
	if _, err := chat.bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}
	return chat.guestHandler
}

func (chat *Chat) requestPhoto(x int, y int) error {
	if x < 0 || x > 360 || y < 0 || y > 90 {
		return fmt.Errorf("X should be in range 0 - 360, Y in range 0 - 90")
	} else if len(chat.photos) >= 5 {
		return fmt.Errorf("Queue is full, try again later")
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
