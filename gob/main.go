package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var buf bytes.Buffer

	err := gob.NewEncoder(&buf).Encode(gl)

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return buf.Bytes(), err
}

func decode(data []byte) (GameLog, error) {
	gl := GameLog{}

	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&gl)

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return gl, err
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
