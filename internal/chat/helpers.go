package chat

import (
	"log/slog"
	"os"
)

func CantSendMessage(err error) {
	slog.Error("Can not send a message", "err", err)
	os.Exit(1)
}
