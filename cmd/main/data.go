package main

import (
	"encoding/gob"
	"os"
)

func (s *server) loadGobData() error {
	file, err := os.Open("events.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewDecoder(file).Decode(&s.events)
}

func (s *server) saveGobData() error {
	file, err := os.Create("events.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(s.events)
}
