package photo

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

type Photo struct {
	X    int
	Y    int
	ID   int64
	Zoom int
}

func New(x int, y int, zoom int, id int64) (*Photo, error) {
	if x < 0 || x > 360 {
		return nil, fmt.Errorf("x should be in range of 0 - 360, got: %d", x)
	}
	if y < 0 || y > 90 {
		return nil, fmt.Errorf("y should be in range of 0 - 90, got: %d", y)
	}
	if zoom < 0 || zoom > 10 {
		return nil, fmt.Errorf("zoom should be in range of 0 - 10, got: %d", zoom)
	}

	return &Photo{X: x, Y: y, Zoom: zoom, ID: id}, nil
}

func (p *Photo) Take(curr int) ([]byte, error) {
	p.setZoom()
	cmd := exec.Command("./motor_driver.bin", fmt.Sprint(p.X), fmt.Sprint(p.Y), "False", fmt.Sprint(curr), "3", "wget -N -P . http://127.0.0.1:8080/photoaf.jpg")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile("photoaf.jpg")
	if err != nil {
		return nil, err
	}
	os.Remove("photoaf.jpg")

	return data, err
}

func (p *Photo) setZoom() error {
	url := fmt.Sprintf("http://127.0.0.1:8080/ptz?zoom=%d", p.Zoom)
	if _, err := http.Get(url); err != nil {
		return err
	}
	return nil
}
