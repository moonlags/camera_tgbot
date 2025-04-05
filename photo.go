package main

import "fmt"

type photoError struct {
	object   string
	from, to int
	got      int
}

func (e *photoError) Error() string {
	return fmt.Sprintf("photo error: expected %s to be in range from %d to %d. Recieved %d!", e.object, e.from, e.to, e.got)
}

type photo struct {
	reciever         int64
	x, y, zoom, mode int
	retry            bool
}

func newPhoto(id int64, x, y, zoom, mode int) (photo, error) {
	if x < 0 || x > 360 {
		return photo{}, &photoError{"x", 0, 360, x}
	}
	if y < 0 || y > 90 {
		return photo{}, &photoError{"y", 0, 90, y}
	}
	if zoom < 0 || zoom > 10 {
		return photo{}, &photoError{"zoom", 0, 10, zoom}
	}
	if mode < 0 || mode > 13 {
		return photo{}, &photoError{"mode", 0, 13, mode}
	}

	return photo{id, x, y, zoom, mode, false}, nil
}
