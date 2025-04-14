package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/NicoNex/echotron/v3"
)

type stateFn func(*echotron.Update) stateFn

type botConfig struct {
	password      string
	guestPassword *string
	sunset        *time.Time
}

type bot struct {
	chatID int64
	config botConfig
	camera *camera
	events map[int64]event
	state  stateFn
	echotron.API
}

func (b *bot) Update(update *echotron.Update) {
	slog.Info("message", "id", b.chatID, "firstname", update.Message.From.FirstName)
	if update.Message != nil {
		b.state = b.state(update)
	}
}

func (b *bot) handleMessage(update *echotron.Update) stateFn {
	switch update.Message.Text {
	case b.config.password:
		slog.Info("owner login", "id", b.chatID, "firstname", update.Message.From.FirstName)
		if _, err := b.SendMessage("welcome, see /help for commands", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
		return b.handleOwner
	case *b.config.guestPassword:
		slog.Info("guest login", "id", b.chatID, "firstname", update.Message.From.FirstName)
		if _, err := b.SendMessage("welcome, you have entered as a guest, see /help for commands", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
		*b.config.guestPassword = fmt.Sprint(rand.Uint32())

		return b.handleGuest
	default:
		if _, err := b.SendMessage("please send a password", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
		return b.handleMessage
	}
}

func (b *bot) handleOwner(update *echotron.Update) stateFn {
	cmd, params, found := strings.Cut(update.Message.Text, " ")
	slog.Info("command cut", "cmd", cmd, "params", params)

	switch cmd {
	case "/help", "help":
		if _, err := b.SendMessage("1. help - display list of commands\n2. photo X Y [ZOOM] [MODE] - take a photo, ZOOM and MODE are optional\n3. modes - list available modes\n4. random - take random photo\n5. event create X Y HOUR MINUTE [ZOOM] [MODE] - create an event, ZOOM and MODE are optional\n6. event sunset X Y [ZOOM] [MODE] - create sunset event, ZOOM and MODE are optional\n7. event delete - delete an event\n8. guestpass - generate one-time guest password", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	case "photo":
		if !found {
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		var x, y, zoom, mode int
		if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
			slog.Error("failed to get arguments for photo command", "err", err, "n", n)
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		photo, err := newPhoto(b.chatID, x, y, zoom, mode)
		if err != nil {
			slog.Error("failed to call newPhoto", "err", err)
			if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		if err := b.camera.queuePhoto(photo); err != nil {
			if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		if _, err := b.SendMessage(fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode), b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	case "modes":
		if _, err := b.SendMessage("0 - none\n1 - monochrome\n2 - negative\n3 - sepia\n4 - aqua\n5 - whiteboard\n6 - blackboard\n7 - nashville\n8 - hefe\n9 - valencia\n10 -xproll\n11 - lofi\n12 - sierra\n13 - walden", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	case "random":
		if _, err := b.SendDice(b.chatID, echotron.Die, nil); err != nil {
			slog.Error("failed to send dice", "err", err)
		}

		time.Sleep(time.Second * 3)
		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		if _, err := b.SendMessage(fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode), b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}

		photo, _ := newPhoto(b.chatID, x, y, zoom, mode)
		b.camera.queuePhoto(photo)
	case "event":
		if !found {
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}
		cmd, params, found := strings.Cut(params, " ")

		switch cmd {
		case "create":
			if !found {
				if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			if _, ok := b.events[b.chatID]; ok {
				if _, err := b.SendMessage("event already exists, delete it with event delete", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			var hour, minute, x, y, zoom, mode int
			if n, err := fmt.Sscanf(params, "%d %d %d %d %d %d", &hour, &minute, &x, &y, &zoom, &mode); err != nil && n < 4 {
				slog.Error("failed to get arguments for eventcreate command", "err", err, "n", n)
				if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			photo, err := newPhoto(b.chatID, x, y, zoom, mode)
			if err != nil {
				slog.Error("failed to call newPhoto", "err", err)
				if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			event, err := newStaticEvent(photo, hour, minute)
			if err != nil {
				slog.Error("failed to call newStaticEvent", "err", err)
				if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			b.events[b.chatID] = &event

			if _, err := b.SendMessage("event created", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
		case "sunset":
			if !found {
				if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			if _, ok := b.events[b.chatID]; ok {
				if _, err := b.SendMessage("event already exists, delete it with eventdelete", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			var x, y, zoom, mode int
			if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
				slog.Error("failed to get arguments for eventsunset command", "err", err, "n", n)
				if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			photo, err := newPhoto(b.chatID, x, y, zoom, mode)
			if err != nil {
				slog.Error("failed to call newPhoto", "err", err)
				if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
					slog.Error("failed to send the message", "err", err)
				}
				break
			}

			event := newSunsetEvent(photo, b.config.sunset)
			b.events[b.chatID] = &event

			if _, err := b.SendMessage("event created", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
		case "delete":
			delete(b.events, b.chatID)

			if _, err := b.SendMessage("event deleted", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
		default:
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
		}
	case "guestpass":
		if _, err := b.SendMessage(fmt.Sprintf("guest password is %s", *b.config.guestPassword), b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	default:
		if _, err := b.SendMessage("unknown command", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	}
	return b.handleOwner
}

func (b *bot) handleGuest(update *echotron.Update) stateFn {
	cmd, params, found := strings.Cut(update.Message.Text, " ")
	slog.Info("command cut", "cmd", cmd, "params", params)

	switch cmd {
	case "/help", "help":
		if _, err := b.SendMessage("1. help - display list of commands\n2. photo X Y [ZOOM] [MODE] - take a photo, ZOOM and MODE are optional\n3. modes - list available modes\n4. random - take random photo", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	case "photo":
		if !found {
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		var x, y, zoom, mode int
		if n, err := fmt.Sscanf(params, "%d %d %d %d", &x, &y, &zoom, &mode); err != nil && n < 2 {
			slog.Error("failed to get arguments for photo command", "err", err, "n", n)
			if _, err := b.SendMessage("invalid command usage", b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		photo, err := newPhoto(b.chatID, x, y, zoom, mode)
		if err != nil {
			slog.Error("failed to call newPhoto", "err", err)
			if _, err := b.SendMessage(err.Error(), b.chatID, nil); err != nil {
				slog.Error("failed to send the message", "err", err)
			}
			break
		}

		if _, err := b.SendMessage(fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode), b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
		b.camera.queuePhoto(photo)
	case "modes":
		if _, err := b.SendMessage("0 - none\n1 - monochrome\n2 - negative\n3 - sepia\n4 - aqua\n5 - whiteboard\n6 - blackboard\n7 - nashville\n8 - hefe\n9 - valencia\n10 -xproll\n11 - lofi\n12 - sierra\n13 - walden", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	case "random":
		if _, err := b.SendDice(b.chatID, echotron.Die, nil); err != nil {
			slog.Error("failed to send dice", "err", err)
		}

		time.Sleep(time.Second * 3)
		x, y, zoom, mode := rand.Intn(361), rand.Intn(91), rand.Intn(11), rand.Intn(14)
		if _, err := b.SendMessage(fmt.Sprintf("taking photo on x: %d y: %d zoom: %d mode: %d", x, y, zoom, mode), b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}

		photo, _ := newPhoto(b.chatID, x, y, zoom, mode)
		b.camera.queuePhoto(photo)
	default:
		if _, err := b.SendMessage("unknown command", b.chatID, nil); err != nil {
			slog.Error("failed to send the message", "err", err)
		}
	}
	return b.handleGuest
}
