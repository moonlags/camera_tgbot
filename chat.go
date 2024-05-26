package main

type Chat struct {
	logged  bool
	guest   bool
	id      int64
	photos  chan Photo
	events  map[int64]Event
	handler handlerFn
}

type handlerFn func(string) error // todo

func (chat *Chat) UnathorizedHandler(string) error // todo
