package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

const QUEUE_SIZE = 10

type Photo struct {
	reciever      int64
	x             uint16
	y, zoom, mode uint8
	retry         bool
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

	return Photo{id, x, y, zoom, mode, false}, nil
}

type Camera struct {
	queue    chan Photo
	currentX uint16
}

func newCamera() (*Camera, error) {
	cmd := exec.Command("./motor_driver.bin", "0", "0", "True", "0", "3", "")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &Camera{
		queue: make(chan Photo, QUEUE_SIZE),
	}, nil
}

func (c *Camera) queuePhoto(p Photo) error {
	if len(c.queue) >= QUEUE_SIZE {
		return errors.New("queue is full")
	}
	c.queue <- p

	log.Printf("added photo to queue X: %d Y:%d ZOOM:%d MODE:%d RETRY:%v\n", p.x, p.y, p.zoom, p.mode, p.retry)
	return nil
}

func (c *Camera) take() (Photo, error) {
	p := <-c.queue
	log.Printf("taking photo X: %d Y:%d ZOOM:%d MODE:%d RETRY:%v\n", p.x, p.y, p.zoom, p.mode, p.retry)

	c.setModeAndZoom(p.mode, p.zoom)

	cmd := exec.Command("./motor_driver.bin", fmt.Sprint(p.x), fmt.Sprint(p.y), "False", fmt.Sprint(c.currentX), "3", "wget -O photoaf.jpg http://127.0.0.1:8080/photoaf.jpg")
	if err := cmd.Run(); err != nil {
		return p, err
	}

	c.currentX = p.x

	return p, nil
}

func (c Camera) setModeAndZoom(mode uint8, zoom uint8) {
	modes := []string{"none", "mono", "negative", "sepia", "aqua", "whiteboard", "blackboard", "nashville", "hefe", "valencia", "xproll", "lofi", "sierra", "walden"}
	url := "http://127.0.0.1:8080/settings/coloreffect?set=" + modes[mode]
	if _, err := http.Get(url); err != nil {
		log.Println("failed to set mode", err)
	}

	url = fmt.Sprintf("http://127.0.0.1:8080/ptz?zoom=%d", zoom)
	if _, err := http.Get(url); err != nil {
		log.Println("failed to set zoom", err)
	}
}

func (c Camera) phoneInit() error {
	if err := exec.Command("./phone_init.sh").Run(); err != nil {
		return err
	}
	return nil
}
