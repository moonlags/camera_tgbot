package main

import (
	"fmt"
	"time"
)

type eventError struct {
	object   string
	from, to int
	got      int
}

func (e *eventError) Error() string {
	return fmt.Sprintf("event error: expected %s to be in range from %d to %d. Recieved %d!", e.object, e.from, e.to, e.got)
}

type event interface {
	isReady() bool
	eventPhoto() photo
}

type staticEvent struct {
	time  time.Time
	photo photo
}

func newStaticEvent(p photo, hour int, minute int) (staticEvent, error) {
	if hour < 0 || hour > 23 {
		return staticEvent{}, &eventError{"hour", 0, 23, hour}
	}
	if minute < 0 || minute > 59 {
		return staticEvent{}, &eventError{"minute", 0, 59, minute}
	}

	return staticEvent{time.Date(0, 0, 0, hour, minute, 0, 0, time.Local), p}, nil
}

func (p *staticEvent) isReady() bool {
	return time.Now().Hour() == p.time.Hour() && time.Now().Minute() == p.time.Minute()
}

func (p *staticEvent) eventPhoto() photo {
	return p.photo
}

type sunsetEvent struct {
	time  *time.Time
	photo photo
}

func newSunsetEvent(p photo, sunset *time.Time) sunsetEvent {
	return sunsetEvent{sunset, p}
}

func (p *sunsetEvent) isReady() bool {
	return time.Now().Hour() == p.time.Hour() && time.Now().Minute() == p.time.Minute()
}

func (p *sunsetEvent) eventPhoto() photo {
	return p.photo
}
