package main

import "time"

type Event interface {
	IsReady() bool
	GetID() int64
	GetPos() (int, int)
}

type StaticEvent struct {
	X        int
	Y        int
	ShotTime time.Time
	ID       int64
}

func (event *StaticEvent) IsReady() bool {
	return time.Now().Hour() == event.ShotTime.Hour() && time.Now().Minute() == event.ShotTime.Minute()
}

func (event *StaticEvent) GetID() int64 {
	return event.ID
}

func (event *StaticEvent) GetPos() (int, int) {
	return event.X, event.Y
}

type SunsetEvent struct {
	X      int
	Y      int
	Sunset *time.Time
	ID     int64
}

func (event *SunsetEvent) IsReady() bool {
	return time.Now().Hour() == event.Sunset.Hour() && time.Now().Minute() == event.Sunset.Minute()
}

func (event *SunsetEvent) GetID() int64 {
	return event.ID
}

func (event *SunsetEvent) GetPos() (int, int) {
	return event.X, event.Y
}
