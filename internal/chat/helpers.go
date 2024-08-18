package chat

import (
	"log/slog"
	"os"

	"example/sashaTelegram/internal/photo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CantSendMessage(err error) {
	slog.Error("Can not send a message", "err", err)
	os.Exit(1)
}

func (c *Chat) QueuePhoto(update tgbotapi.Update, photo *photo.Photo) {
	if len(c.Photos) >= 5 {
		msg := tgbotapi.NewMessage(c.ID, "No place for photo in queue, try again later")
		if _, err := c.Bot.Send(msg); err != nil {
			CantSendMessage(err)
		}

		slog.Info("No place for photo in queue", "photo", photo)
		return
	}
	c.Photos <- photo
}
