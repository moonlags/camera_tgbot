package event

import (
	"fmt"
	"time"

	"example/sashaTelegram/internal/photo"
)

type SunsetEvent struct {
	sunset *time.Time
	x      int
	y      int
	id     int64
}

func NewSunset(sunset *time.Time, x int, y int, id int64) (*SunsetEvent, error) {
	if x < 0 || x > 360 {
		return nil, fmt.Errorf("x should be in range of 0 - 360, got: %d", x)
	}
	if y < 0 || y > 90 {
		return nil, fmt.Errorf("y should be in range of 0 - 90, got: %d", y)
	}

	return &SunsetEvent{
		x:      x,
		y:      y,
		id:     id,
		sunset: sunset,
	}, nil
}

func (event *SunsetEvent) IsReady() bool {
	return time.Now().Hour() == event.sunset.Hour() && time.Now().Minute() == event.sunset.Minute()
}

func (e *SunsetEvent) Photo() *photo.Photo {
	img, _ := photo.New(e.x, e.y, 0, 0, e.id)
	return img
}

func (e *SunsetEvent) ID() int64 {
	return e.id
}
