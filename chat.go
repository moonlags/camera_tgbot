package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type stateFn func(tgbotapi.Update) stateFn

type Chat struct {
	bot    *tgbotapi.BotAPI
	camera *Camera
	events *[]Event

	sunsetTime    *time.Time
	guestPassword *string

	state stateFn
}

func newChat(bot *tgbotapi.BotAPI, cam *Camera, events *[]Event, st *time.Time, gp *string) Chat {
	chat := Chat{
		bot:           bot,
		camera:        cam,
		events:        events,
		sunsetTime:    st,
		guestPassword: gp,
	}

	chat.state = chat.handleMessage
	return chat
}

func (c *Chat) handleUpdate(update tgbotapi.Update) {
	c.state = c.state(update)
}

func (c *Chat) handleMessage(update tgbotapi.Update) stateFn {
	id := update.Message.Chat.ID

	switch update.Message.Text {
	case PASSWORD:
		log.Printf("%s logged in as owner\n", update.Message.From.FirstName)
		msg := tgbotapi.NewMessage(id, "welcome, see /help for commands")
		if _, err := c.bot.Send(msg); err != nil {
			log.Println("failed to send message", err)
		}
		return c.handleOwner
	case *c.guestPassword:
		log.Printf("%s logged in as guest\n", update.Message.From.FirstName)
		msg := tgbotapi.NewMessage(id, "welcome, you have entered as a guest, see /help for commands")
		if _, err := c.bot.Send(msg); err != nil {
			log.Println("failed to send message", err)
		}

		*c.guestPassword = fmt.Sprint(rand.Uint32())

		return c.handleGuest
	default:
		msg := tgbotapi.NewMessage(id, "please send your password")
		if _, err := c.bot.Send(msg); err != nil {
			log.Println("failed to send message", err)
		}

		return c.handleMessage
	}
}

func (c *Chat) handleOwner(update tgbotapi.Update) stateFn {
	cmd, params, found := strings.Cut(update.Message.Text, " ")
	log.Printf("command cut %s params %s\n", cmd, params)

	id := update.Message.Chat.ID

	var msg tgbotapi.MessageConfig
	switch cmd {
	case "/help", "help":
		msg = tgbotapi.NewMessage(id, "1. help - display list of commands\n2. photo X Y [ZOOM] [MODE] - take a photo, ZOOM and MODE are optional\n3. modes - list available modes\n4. random - take random photo\n5. event create X Y HOUR MINUTE [ZOOM] [MODE] - create an event, ZOOM and MODE are optional\n6. event sunset X Y [ZOOM] [MODE] - create sunset event, ZOOM and MODE are optional\n7. event delete NUM - delete an event under number NUM\n8. event list - list your events\n9. guestpass - generate one-time guest password")
	case "photo":
		if !found {
			msg = tgbotapi.NewMessage(id, "invalid command usage")
			break
		}

		var x uint16
		var y, zoom, mode uint8
		if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
			log.Println("failed to get arguments for photo command", err)
			msg = tgbotapi.NewMessage(id, "invalid command usage")
			break
		}

		photo, err := newPhoto(id, x, y, zoom, mode)
		if err != nil {
			log.Println("failed to create new photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		if err := c.camera.queuePhoto(photo); err != nil {
			log.Println("failed to queue photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		msg = tgbotapi.NewMessage(id, fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode))
	case "modes":
		msg = tgbotapi.NewMessage(id, "0 - none\n1 - monochrome\n2 - negative\n3 - sepia\n4 - aqua\n5 - whiteboard\n6 - blackboard\n7 - nashville\n8 - hefe\n9 - valencia\n10 -xproll\n11 - lofi\n12 - sierra\n13 - walden")
	case "random":
		dice := tgbotapi.NewDice(id)
		if _, err := c.bot.Send(dice); err != nil {
			log.Println("failed to send dice", err)
		}

		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		photo, _ := newPhoto(id, uint16(x), uint8(y), uint8(zoom), uint8(mode))

		if err := c.camera.queuePhoto(photo); err != nil {
			log.Println("failed to queue photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		time.Sleep(time.Second * 3)
		msg = tgbotapi.NewMessage(id, fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode))
	case "event":
		if !found {
			msg = tgbotapi.NewMessage(id, "invalid command usage")
			break
		}
		cmd, params, found := strings.Cut(params, " ")

		switch cmd {
		case "create":
			if !found {
				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			if len(*c.events) >= 5 {
				msg = tgbotapi.NewMessage(id, "you can have maximum of 5 events")
				break
			}

			var x uint16
			var y, hour, minute, zoom, mode uint8
			if n, err := fmt.Sscanf(params, "%d %d %d %d %d %d", &x, &y, &hour, &minute, &zoom, &mode); err != nil && n < 4 {
				log.Println("failed to get arguments for event create command", err, n)

				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			photo, err := newPhoto(id, x, y, zoom, mode)
			if err != nil {
				log.Println("failed to create new photo", err)

				msg = tgbotapi.NewMessage(id, err.Error())
				break
			}

			event, err := newStaticEvent(photo, hour, minute)
			if err != nil {
				log.Println("failed to create new static event", err)

				msg = tgbotapi.NewMessage(id, err.Error())
				break
			}

			*c.events = append(*c.events, &event)

			msg = tgbotapi.NewMessage(id, "event created")
		case "sunset":
			if !found {
				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			if len(*c.events) >= 5 {
				msg = tgbotapi.NewMessage(id, "you can have maximum of 5 events")
				break
			}

			var x uint16
			var y, zoom, mode uint8
			if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
				log.Println("failed to get arguments for event sunset command", err, n)

				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			photo, err := newPhoto(id, x, y, zoom, mode)
			if err != nil {
				log.Println("failed to create new photo", err)

				msg = tgbotapi.NewMessage(id, err.Error())
				break
			}

			event := newSunsetEvent(photo, c.sunsetTime)
			*c.events = append(*c.events, &event)

			msg = tgbotapi.NewMessage(id, "event created")
		case "delete":
			if !found {
				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			var eventnum uint8
			if n, err := fmt.Sscanf(params, "%d", &eventnum); err != nil {
				log.Println("failed to get arguments for event delete commamd", err, n)

				msg = tgbotapi.NewMessage(id, "invalid command usage")
				break
			}

			if int(eventnum) > len(*c.events) || eventnum == 0 {
				msg = tgbotapi.NewMessage(id, "no event at this number")
				break
			}
			*c.events = append((*c.events)[:eventnum], (*c.events)[eventnum+1:]...)

			msg = tgbotapi.NewMessage(id, "event deleted")
		case "list":
			eventList := ""
			for i, event := range *c.events {
				photo := event.eventPhoto()
				eventList += fmt.Sprintf("%d. X:%d Y:%d HOUR:%d MINUTE:%d ZOOM:%d MODE:%d SUNSET:%v\n", i+1, photo.x, photo.y, event.hour(), event.minute(), photo.zoom, photo.mode, event.isSunset())
			}
			msg = tgbotapi.NewMessage(id, eventList)
		default:
			msg = tgbotapi.NewMessage(id, "invalid command usage")
		}
	case "guestpass":
		msg = tgbotapi.NewMessage(id, fmt.Sprint("guest password is", c.guestPassword))
	default:
		msg = tgbotapi.NewMessage(id, "unknown command")
	}

	if _, err := c.bot.Send(msg); err != nil {
		log.Println("failed to send message", err)
	}
	return c.handleOwner
}

func (c *Chat) handleGuest(update tgbotapi.Update) stateFn {
	cmd, params, found := strings.Cut(update.Message.Text, " ")
	log.Printf("command cut %s params %s\n", cmd, params)

	id := update.Message.Chat.ID

	var msg tgbotapi.MessageConfig
	switch cmd {
	case "/help", "help":
		msg = tgbotapi.NewMessage(id, "1. help - display list of commands\n2. photo X Y [ZOOM] [MODE] - take a photo, ZOOM and MODE are optional\n3. modes - list available modes\n4. random - take random photo")
	case "photo":
		if !found {
			msg = tgbotapi.NewMessage(id, "invalid command usage")
			break
		}

		var x uint16
		var y, zoom, mode uint8
		if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
			log.Println("failed to get arguments for photo command", err)
			msg = tgbotapi.NewMessage(id, "invalid command usage")
			break
		}

		photo, err := newPhoto(id, x, y, zoom, mode)
		if err != nil {
			log.Println("failed to create new photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		if err := c.camera.queuePhoto(photo); err != nil {
			log.Println("failed to queue photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		msg = tgbotapi.NewMessage(id, fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode))
	case "modes":
		msg = tgbotapi.NewMessage(id, "0 - none\n1 - monochrome\n2 - negative\n3 - sepia\n4 - aqua\n5 - whiteboard\n6 - blackboard\n7 - nashville\n8 - hefe\n9 - valencia\n10 -xproll\n11 - lofi\n12 - sierra\n13 - walden")
	case "random":
		dice := tgbotapi.NewDice(id)
		if _, err := c.bot.Send(dice); err != nil {
			log.Println("failed to send dice", err)
		}

		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		photo, _ := newPhoto(id, uint16(x), uint8(y), uint8(zoom), uint8(mode))

		if err := c.camera.queuePhoto(photo); err != nil {
			log.Println("failed to queue photo", err)

			msg = tgbotapi.NewMessage(id, err.Error())
			break
		}

		time.Sleep(time.Second * 3)
		msg = tgbotapi.NewMessage(id, fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode))
	default:
		msg = tgbotapi.NewMessage(id, "unknown command")
	}

	if _, err := c.bot.Send(msg); err != nil {
		log.Println("failed to send message", err)
	}
	return c.handleGuest
}
