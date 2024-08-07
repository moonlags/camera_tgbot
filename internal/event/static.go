package event

import (
	"fmt"
	"time"

	"example/sashaTelegram/internal/photo"
)

type StaticEvent struct {
	shotTime time.Time
	x        int
	y        int
	id       int64
}

func NewStatic(hour int, minute int, x int, y int, id int64) (*StaticEvent, error) {
	if x < 0 || x > 360 {
		return nil, fmt.Errorf("x should be in range of 0 - 360, got: %d", x)
	}
	if y < 0 || y > 90 {
		return nil, fmt.Errorf("y should be in range of 0 - 90, got: %d", y)
	}
	if hour < 0 || hour > 24 {
		return nil, fmt.Errorf("hour should be in range of 0 - 24, got: %d", hour)
	}
	if minute < 0 || minute > 60 {
		return nil, fmt.Errorf("minute should be in range of 0 - 60, got: %d", minute)
	}

	return &StaticEvent{
		shotTime: time.Date(0, 0, 0, hour, minute, 0, 0, time.Local),
		x:        x,
		y:        y,
		id:       id,
	}, nil
}

func (event *StaticEvent) IsReady() bool {
	return time.Now().Hour() == event.shotTime.Hour() && time.Now().Minute() == event.shotTime.Minute()
}

func (event *StaticEvent) Photo() *photo.Photo {
	img, _ := photo.New(event.x, event.y, 0, 0, event.id)
	return img
}

func (e *StaticEvent) ID() int64 {
	return e.id
}
