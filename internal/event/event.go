package event

import (
	"example/sashaTelegram/internal/photo"
)

type Event interface {
	IsReady() bool
	Photo() *photo.Photo
	ID() int64
}
