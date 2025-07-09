package main

import (
	"errors"
	"time"
)

type Event interface {
	isReady() bool
	eventPhoto() Photo
	hour() int
	minute() int
	isSunset() bool
}

type StaticEvent struct {
	time  time.Time
	photo Photo
}

func newStaticEvent(p Photo, hour uint8, minute uint8) (StaticEvent, error) {
	if hour < 0 || hour > 23 {
		return StaticEvent{}, errors.New("Invalid value for HOUR")
	}
	if minute < 0 || minute > 59 {
		return StaticEvent{}, errors.New("Invalid value for MINUTE")
	}

	return StaticEvent{time.Date(0, 0, 0, int(hour), int(minute), 0, 0, time.Local), p}, nil
}

func (p *StaticEvent) isReady() bool {
	return time.Now().Hour() == p.time.Hour() && time.Now().Minute() == p.time.Minute()
}

func (p *StaticEvent) eventPhoto() Photo {
	return p.photo
}

func (p *StaticEvent) hour() int {
	return p.time.Hour()
}

func (p *StaticEvent) minute() int {
	return p.time.Minute()
}

func (p *StaticEvent) isSunset() bool {
	return false
}

type SunsetEvent struct {
	time  *time.Time
	photo Photo
}

func newSunsetEvent(p Photo, sunset *time.Time) SunsetEvent {
	return SunsetEvent{sunset, p}
}

func (p *SunsetEvent) isReady() bool {
	return time.Now().Hour() == p.time.Hour() && time.Now().Minute() == p.time.Minute()
}

func (p *SunsetEvent) eventPhoto() Photo {
	return p.photo
}

func (p *SunsetEvent) hour() int {
	return p.time.Hour()
}

func (p *SunsetEvent) minute() int {
	return p.time.Minute()
}

func (p *SunsetEvent) isSunset() bool {
	return true
}
