package main

type Chat struct {
	guest bool
	id    int64
	state stateFn
}

type stateFn func(string) error

// todo
