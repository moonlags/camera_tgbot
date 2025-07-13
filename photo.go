package main

import (
	"errors"
)

type Photo struct {
	reciever      int64
	x             uint16
	y, zoom, mode uint8
}

type PhotoConfig struct {
	X             uint16
	Y, Zoom, Mode uint8
}

func newPhoto(id int64, x uint16, y, zoom, mode uint8) (Photo, error) {
	if x > 360 {
		return Photo{}, errors.New("Invalid value for X")
	}
	if y > 90 {
		return Photo{}, errors.New("Invalid value for Y")
	}
	if zoom > 10 {
		return Photo{}, errors.New("Invalid value for ZOOM")
	}
	if mode > 13 {
		return Photo{}, errors.New("Invalid value for MODE")
	}

	return Photo{id, x, y, zoom, mode}, nil
}

func (p Photo) toConfig() PhotoConfig {
	return PhotoConfig{X: p.x, Y: p.y, Zoom: p.zoom, Mode: p.mode}
}

func queuePhoto(photoRequests chan Photo, p Photo) {
	photoRequests <- p
}
