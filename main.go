package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/gorilla/websocket"
	"github.com/ipfs/go-ipfs-api"
	"github.com/polyswarm/polyswarm/bounty"
)

const MAX_DATA_SIZE = 50 * 1024 * 1024

func connectToClamd(host string) (*clamd.Clamd, error) {
	u := url.URL{Scheme: "tcp", Host: host}
	ret := clamd.NewClamd(u.String())

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

func connectToIpfs(host string) (*shell.Shell, error) {
	ret := shell.NewShell(host)

	timeout := time.After(60 * time.Second)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-timeout:
			return nil, errors.New("timeout waiting for clamd")
		case <-tick:
			if ret.IsUp() {
				return ret, nil
			}
		}
	}
}

func connectToPolyswarm(host string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/events"}
	log.Println(u.String())

	timeout := time.After(60 * time.Second)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-timeout:
			return nil, errors.New("timeout waiting for polyswarm")
		case <-tick:
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err == nil {
				return conn, nil
			}
		}
	}
}

func main() {
	time.Sleep(15 * time.Second)
	log.Println("Starting microengine")

	clam, err := connectToClamd(os.Getenv("CLAMD_HOST"))
	if err != nil {
		log.Fatalln(err)
	}
	_ = clam

	ipfssh, err := connectToIpfs(os.Getenv("IPFS_HOST"))
	if err != nil {
		log.Fatalln(err)
	}
	_ = ipfssh

	conn, err := connectToPolyswarm(os.Getenv("POLYSWARM_HOST"))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading from websocket:", err)
		}

		var event bounty.Event
		json.Unmarshal(message, &event)

		log.Println("recv:", event)
	}

}
