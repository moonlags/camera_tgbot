package main

type Event interface {
	IsReady() bool
	GetID() int64
	GetPos() (int, int)
}
