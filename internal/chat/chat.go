package chat

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"example/sashaTelegram/internal/event"
	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	ID      int64
	Zoom    int
	Photos  chan *photo.Photo
	Events  map[int64]event.Event
	Config  *Config
	Bot     *tgbotapi.BotAPI
	Handler handlerFn
}

type Config struct {
	SunsetTime    time.Time
	Password      string
	GuestPassword string
}

type handlerFn func(tgbotapi.Update) handlerFn

func New(bot *tgbotapi.BotAPI, id int64, photos chan *photo.Photo, events map[int64]event.Event, c *Config) *Chat {
	chat := &Chat{
		Bot:    bot,
		ID:     id,
		Photos: photos,
		Events: events,
		Config: c,
	}
	chat.Handler = chat.unauthorizedHandler
	return chat
}

func (chat *Chat) Handle(update tgbotapi.Update) {
	chat.Handler = chat.Handler(update)
}

func (chat *Chat) unauthorizedHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case chat.Config.Password:
		msg := tgbotapi.NewMessage(chat.ID, "Welcome, try sending a coordinates")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case chat.Config.GuestPassword:
		msg := tgbotapi.NewMessage(chat.ID, "Welcome, you have entered as guest, try sending a coordinates")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		chat.Config.GuestPassword = fmt.Sprint(rand.Uint32())
		return chat.guestHandler
	}
	msg := tgbotapi.NewMessage(chat.ID, "Hello, please send a password")
	if _, err := chat.Bot.Send(msg); err != nil {
		log.Fatal("Error sending a message:", err)
	}
	return chat.unauthorizedHandler
}

func (chat *Chat) commandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.ID,
			"/help -  Get a list of commands\n/random - Take a random photo\n/zoom - set zoom\n/eventcreate - Create an event\n/eventsunset - Create sunset event\n/eventdelete - Delete an event\n/sunsettime - Get sunset time\n/guestpass - Generate guest password")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case "/random":
		x, y, zoom := rand.Intn(361), rand.Intn(91), rand.Intn(11)
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %v Y: %v with zoom %v", x, y, zoom))
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		photo, _ := photo.New(x, y, zoom, chat.ID)
		chat.Photos <- photo
		return chat.protectedHandler
	case "/zoom":
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Your current zoom is %v, type a number between 0 and 10", chat.Zoom))
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.zoomHandler
	case "/sunsettime":
		stime := chat.Config.SunsetTime
		msg := tgbotapi.NewMessage(chat.ID,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()))
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	case "/eventcreate":
		if _, ok := chat.Events[chat.ID]; ok {
			msg := tgbotapi.NewMessage(chat.ID, "Please delete your existing event first")
			if _, err := chat.Bot.Send(msg); err != nil {
				log.Fatal("Error sending a message:", err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y Hour Minute' to create an event")
			if _, err := chat.Bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			return chat.eventCreationHandler
		}
	case "/eventsunset":
		if _, ok := chat.Events[chat.ID]; ok {
			msg := tgbotapi.NewMessage(chat.ID, "Please delete your existing event first")
			if _, err := chat.Bot.Send(msg); err != nil {
				log.Fatal("Error sending a message:", err)
			}
			return chat.protectedHandler
		} else {
			msg := tgbotapi.NewMessage(chat.ID, "Specify 'X Y' to create an event")
			if _, err := chat.Bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			return chat.sunsetEventCreationHanlder
		}
	case "/eventdelete":
		delete(chat.Events, chat.ID)
		msg := tgbotapi.NewMessage(chat.ID, "Event deleted")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.protectedHandler
	case "/guestpass":
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Guest password is %v", chat.Config.GuestPassword))
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.protectedHandler
	}
	return nil
}

func (chat *Chat) zoomHandler(update tgbotapi.Update) handlerFn {
	if handler := chat.commandsHandler(update); handler != nil {
		return handler
	}
	var zoom int
	if _, err := fmt.Sscanf(update.Message.Text, "%d", &zoom); err != nil || zoom < 0 || zoom > 10 {
		msg := tgbotapi.NewMessage(chat.ID, "Zoom should be an integer between 0 and 10")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.zoomHandler
	}
	chat.Zoom = zoom
	msg := tgbotapi.NewMessage(chat.ID, "New zoom is set")
	if _, err := chat.Bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
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
			log.Fatal("Failed to send a message:", err)
		}
		return chat.sunsetEventCreationHanlder
	}

	event, err := event.NewSunset(&chat.Config.SunsetTime, x, y, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.sunsetEventCreationHanlder
	}

	chat.Events[chat.ID] = event

	msg := tgbotapi.NewMessage(chat.ID, "Sunset event created")
	if _, err := chat.Bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
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
			log.Fatal("Failed to send a message:", err)
		}
		return chat.eventCreationHandler
	}

	event, err := event.NewStatic(hour, minute, x, y, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
		return chat.eventCreationHandler
	}

	chat.Events[chat.ID] = event

	msg := tgbotapi.NewMessage(chat.ID, "Event created")
	if _, err := chat.Bot.Send(msg); err != nil {
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
		msg := tgbotapi.NewMessage(chat.ID,
			"Values should be in format of 'X Y'")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	}

	photo, err := photo.New(x, y, chat.Zoom, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.protectedHandler
	}
	chat.Photos <- photo

	msg := tgbotapi.NewMessage(chat.ID, "Your photo is in the queue, please wait")
	if _, err := chat.Bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}

	return chat.protectedHandler
}

func (chat *Chat) guestCommandsHandler(update tgbotapi.Update) handlerFn {
	switch update.Message.Text {
	case "/help":
		msg := tgbotapi.NewMessage(chat.ID,
			"/help -  Get a list of commands\n/random - Take a random photo\n/sunsettime - Get sunset time\n")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	case "/random":
		x, y, zoom := rand.Intn(361), rand.Intn(91), rand.Intn(11)
		msg := tgbotapi.NewMessage(chat.ID, fmt.Sprintf("Taking photo on X: %v Y: %v with zoom %v", x, y, zoom))
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		photo, _ := photo.New(x, y, zoom, chat.ID)
		chat.Photos <- photo
		return chat.guestHandler
	case "/sunsettime":
		stime := chat.Config.SunsetTime
		msg := tgbotapi.NewMessage(
			chat.ID,
			fmt.Sprintf("Today you can see sunset in Jurmala at %02d:%02d", stime.Hour(), stime.Minute()),
		)
		if _, err := chat.Bot.Send(msg); err != nil {
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
		msg := tgbotapi.NewMessage(chat.ID,
			"Values should be in format of 'X Y'")
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	}

	photo, err := photo.New(x, y, chat.Zoom, chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(chat.ID, err.Error())
		if _, err := chat.Bot.Send(msg); err != nil {
			log.Fatal("Error sending a message:", err)
		}
		return chat.guestHandler
	}
	chat.Photos <- photo

	msg := tgbotapi.NewMessage(chat.ID, "Your photo is in the queue, please wait")
	if _, err := chat.Bot.Send(msg); err != nil {
		log.Fatal("Failed to send a message:", err)
	}

	return chat.guestHandler
}
