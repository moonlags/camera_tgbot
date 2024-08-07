package main

import (
	"log/slog"
	"os/exec"

	"example/sashaTelegram/internal/photo"
)

func (s *server) phoneInit(photo *photo.Photo, err error) {
	if err := exec.Command("./phone_init.sh").Run(); err != nil {
		slog.Error("Can not run phone_init.sh", "err", err)
	}
	s.photos <- photo

	slog.Warn("Can not take a photo", "err", err)
}
