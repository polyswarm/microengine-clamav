package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"time"

	"github.com/dutchcoders/go-clamd"
)

func connectToClamd(url string) (*clamd.Clamd, error) {
	ret := clamd.NewClamd(url)

	timeout := time.After(60 * time.Second)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-timeout:
			return nil, errors.New("timeout waiting for clamd")
		case <-tick:
			if err := ret.Ping(); err == nil {
				return ret, nil
			}
		}
	}
}

func main() {
	log.Println("Starting microengine")

	c, err := connectToClamd(os.Getenv("CLAMD_URL"))
	if err != nil {
		log.Fatalln(err)
	}

	reader := bytes.NewReader(clamd.EICAR)
	response, err := c.ScanStream(reader, make(chan bool))

	for s := range response {
		log.Printf("%v %v\n", s, err)
	}
}
