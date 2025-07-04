package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
)

const QUEUE_SIZE = 5

type camera struct {
	queue chan photo
	currX int
}

func newCamera() (*camera, error) {
	cmd := exec.Command("./motor_driver.bin", "0", "0", "True", "0", "3", "")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &camera{
		queue: make(chan photo, QUEUE_SIZE),
	}, nil
}

func (c *camera) queuePhoto(p photo) error {
	slog.Info("placing photo in queue", "x", p.x, "y", p.y, "zoom", p.zoom, "mode", p.mode, "id", p.reciever)
	if len(c.queue) >= 5 {
		slog.Info("no place for photo in queue", "x", p.x, "y", p.y, "zoom", p.zoom, "mode", p.mode, "id", p.reciever)
		return errors.New("no place for photo in queue")
	}
	c.queue <- p
	return nil
}

func (c *camera) take() (photo, error) {
	p := <-c.queue
	slog.Info("taking photo", "x", p.x, "y", p.y, "zoom", p.zoom, "mode", p.mode, "id", p.reciever)

	if err := c.setModeAndZoom(p.mode, p.zoom); err != nil {
		return p, err
	}

	cmd := exec.Command("./motor_driver.bin", fmt.Sprint(p.x), fmt.Sprint(p.y), "False", fmt.Sprint(c.currX), "3", "wget -O photoaf.jpg http://127.0.0.1:8080/photoaf.jpg")
	if err := cmd.Run(); err != nil {
		return p, err
	}

	c.currX = p.x

	return p, nil
}

func (c *camera) setModeAndZoom(mode int, zoom int) error {
	modes := []string{"none", "mono", "negative", "sepia", "aqua", "whiteboard", "blackboard", "nashville", "hefe", "valencia", "xproll", "lofi", "sierra", "walden"}
	url := "http://127.0.0.1:8080/settings/coloreffect?set=" + modes[mode]
	if _, err := http.Get(url); err != nil {
		return err
	}

	url = fmt.Sprintf("http://127.0.0.1:8080/ptz?zoom=%d", zoom)
	if _, err := http.Get(url); err != nil {
		return err
	}
	return nil
}

func (c *camera) phoneInit() error {
	if err := exec.Command("./phone_init.sh").Run(); err != nil {
		return err
	}
	return nil
}
