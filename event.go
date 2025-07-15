package main

import (
	"database/sql"
	"errors"
	"time"
)

type Event interface {
	isReady() bool
	eventPhoto() Photo
	hour() int
	minute() int
	isSunset() bool
	removeFromDB(db *sql.DB) error
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

func (p *StaticEvent) ToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO events (userid, x, y, zoom, mode, hour, minute, sunset) VALUES (?, ?, ?, ?, ?, ?, ?, FALSE)", p.photo.reciever, p.photo.x, p.photo.y, p.photo.zoom, p.photo.mode, p.hour(), p.minute())
	return err
}

func (p *StaticEvent) removeFromDB(db *sql.DB) error {
	_, err := db.Exec("REMOVE FROM events WHERE userid = ? AND x = ? AND y = ? AND zoom = ? AND mode = ? AND hour = ? AND minute = ? AND sunset = FALSE", p.photo.reciever, p.photo.x, p.photo.y, p.photo.zoom, p.photo.mode, p.hour(), p.minute())
	return err
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

func (p *SunsetEvent) ToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO events (userid, x, y, zoom, mode, sunset) VALUES (?, ?, ?, ?, ?, TRUE)", p.photo.reciever, p.photo.x, p.photo.y, p.photo.zoom, p.photo.mode)
	return err
}

func (p *SunsetEvent) removeFromDB(db *sql.DB) error {
	_, err := db.Exec("REMOVE FROM events WHERE userid = ? AND x = ? AND y = ? AND zoom = ? AND mode = ? AND sunset = TRUE", p.photo.reciever, p.photo.x, p.photo.y, p.photo.zoom, p.photo.mode)
	return err
}

type EventDB struct {
	Userid int64
	X      uint16
	Y      uint8
	Zoom   uint8
	Mode   uint8
	Hour   *uint8
	Minute *uint8
	Sunset bool
}
