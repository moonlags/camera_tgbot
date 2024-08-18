package chat

import (
	"fmt"
	"math/rand"

	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (chat *Chat) commandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.ID,
			"/help -  Get a list of commands\n/random - Take a random photo\n/zoom - set zoom\n/mode - set photo mode\n/eventcreate - Create an event\n/eventsunset - Create sunset event\n/eventdelete - Delete an event\n/sunsettime - Get sunset time\n/guestpass - Generate guest password")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	case "/mode":
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Current mode is %d:\n0 - none\n1 - monochrome\n2 - negative\n3 - sepia\n4 - aqua\n5 - whiteboard\n6 - blackboard\n7 - nashville\n8 - hefe\n9 - valencia\n10 -xproll\n11 - lofi\n12 - sierra\n13 - walden", chat.Mode))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.modeSelectionHandler
	case "/random":
		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %v Y: %v Zoom: %v Mode: %d", x, y, zoom, mode))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}

		photo, _ := photo.New(x, y, zoom, mode, chat.ID)
		chat.QueuePhoto(update, photo)

		return chat.protectedHandler
	case "/zoom":
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Your current zoom is %v, type a number between 0 and 10", chat.Zoom))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.zoomHandler
	case "/sunsettime":
		stime := chat.Config.SunsetTime
		msg := tgbotapi.NewMessage(chat.ID,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	case "/eventcreate":
		if _, ok := chat.Events[chat.ID]; ok {
			msg := tgbotapi.NewMessage(chat.ID, "Please delete your existing event first")
			if _, err := chat.Bot.Send(msg); err != nil {
				CantSendMessage(err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y Hour Minute' to create an event")
			if _, err := chat.Bot.Send(msg); err != nil {
				CantSendMessage(err)
			}
			return chat.eventCreationHandler
		}
	case "/eventsunset":
		if _, ok := chat.Events[chat.ID]; ok {
			msg := tgbotapi.NewMessage(chat.ID, "Please delete your existing event first")
			if _, err := chat.Bot.Send(msg); err != nil {
				CantSendMessage(err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y' to create an event")
			if _, err := chat.Bot.Send(msg); err != nil {
				CantSendMessage(err)
			}
			return chat.sunsetEventCreationHanlder
		}
	case "/eventdelete":
		delete(chat.Events, chat.ID)
		msg := tgbotapi.NewMessage(chat.ID, "Event deleted")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	case "/guestpass":
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Guest password is %v", chat.Config.GuestPassword))
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	}
	return nil
}

func (c *Chat) modeSelectionHandler(update tgbotapi.Update) handlerFn {
	if handler := c.commandsHandler(update); handler != nil {
		return handler
	}
	var mode int
	if _, err := fmt.Sscanf(update.Message.Text, "%d", &mode); err != nil || mode < 0 || mode > 13 {
		msg := tgbotapi.NewMessage(c.ID, "Mode should be an integer between 0 and 13")
		if _, err := c.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return c.modeSelectionHandler
	}
	c.Mode = mode
	msg := tgbotapi.NewMessage(c.ID, "New mode is set")
	if _, err := c.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}
	return c.protectedHandler
}

func (chat *Chat) zoomHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var zoom int
	if _, err := fmt.Sscanf(update.Message.Text, "%d", &zoom); err != nil || zoom < 0 || zoom > 10 {
		msg := tgbotapi.NewMessage(chat.ID, "Zoom should be an integer between 0 and 10")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.zoomHandler
	}
	chat.Zoom = zoom
	msg := tgbotapi.NewMessage(chat.ID, "New zoom is set")
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}
	return chat.protectedHandler
}

func (chat *Chat) sunsetEventCreationHanlder(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}

	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y' in valid format")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.sunsetEventCreationHanlder
	}

	event, err := event.NewSunset(&chat.Config.SunsetTime, x, y, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.sunsetEventCreationHanlder
	}

	chat.Events[chat.ID] = event

	msg := tgbotapi.NewMessage(chat.ID, "Sunset event created")
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}

	return chat.protectedHandler
}

func (chat *Chat) eventCreationHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}

	var x, y, hour, minute int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d %d %d", &x, &y, &hour, &minute); err != nil {
		msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y Hour Minute' in valid format")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.eventCreationHandler
	}

	event, err := event.NewStatic(hour, minute, x, y, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.eventCreationHandler
	}

	chat.Events[chat.ID] = event

	msg := tgbotapi.NewMessage(chat.ID, "Event created")
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}

	return chat.protectedHandler
}

func (chat *Chat) protectedHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}

	var x, y int
	if _, err := fmt.Sscanf(update.Message.Text, "%d %d", &x, &y); err != nil {
		msg := tgbotapi.NewMessage(chat.ID,
			"Values should be in format of 'X Y'")
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	}

	photo, err := photo.New(x, y, chat.Zoom, chat.Mode, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, "Error: "+err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}
		return chat.protectedHandler
	}

	msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %d Y: %d Zoom: %d Mode: %d", x, y, chat.Zoom, chat.Mode))
	if _, err := chat.Bot.Send(msg); err != nil {
		CantSendMessage(err)
	}

	chat.QueuePhoto(update, photo)

	return chat.protectedHandler
}
