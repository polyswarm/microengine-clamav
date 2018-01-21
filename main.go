package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/gorilla/websocket"
	"github.com/ipfs/go-ipfs-api"
	"github.com/polyswarm/polyswarm/bounty"
	uuid "github.com/satori/go.uuid"
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
			return nil, errors.New("timeout waiting for ipfs")
		case <-tick:
			if ret.IsUp() {
				return ret, nil
			}
		}
	}
}

func connectToPolyswarm(host string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/events"}

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

// Eventually do this from polyswarm, we need a better artifact API though
func retrieveFileFromIpfs(ipfssh *shell.Shell, resource string) (io.ReadCloser, error) {
	return ipfssh.Cat(resource)
}

func main() {
	time.Sleep(15 * time.Second)
	log.Println("Starting microengine")

	clamd, err := connectToClamd(os.Getenv("CLAMD_HOST"))
	if err != nil {
		log.Fatalln(err)
	}

	ipfssh, err := connectToIpfs(os.Getenv("IPFS_HOST"))
	if err != nil {
		log.Fatalln(err)
	}

	polyswarmHost := os.Getenv("POLYSWARM_HOST")
	conn, err := connectToPolyswarm(polyswarmHost)
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

		if event.Type == "Bounty" {
			body, ok := event.Body.(map[string]interface{})
			if !ok {
				log.Println("invalid bounty object")
				continue
			}

			guid, ok := body["Guid"].(string)
			if !ok {
				log.Println("invalid guid")
				continue
			}

			uuid, err := uuid.FromString(guid)
			if err != nil {
				log.Println("invalid uuid:", err)
				continue
			}

			uri, ok := body["ArtifactURI"].(string)
			if !ok {
				log.Println("invalid uri:", err)
				continue
			}

			r, err := retrieveFileFromIpfs(ipfssh, uri)
			if err != nil {
				log.Println("error retrieving artifact:", err)
				continue
			}

			response, err := clamd.ScanStream(r, make(chan bool))
			if err != nil {
				log.Println("error scanning artifact:", err)
				continue
			}

			verdicts := make([]bool, 0)
			var metadata bytes.Buffer
			for s := range response {
				verdicts = append(verdicts, s.Status == "FOUND")
				metadata.WriteString(s.Description)
				metadata.WriteString(";")
			}

			var a struct {
				Verdicts []bool `json:"verdicts"`
				Bid      int    `json:"bid"`
				Metadata string `json:"metadata"`
			}

			a.Verdicts = verdicts
			a.Bid = 10
			a.Metadata = metadata.String()

			j, err := json.Marshal(a)
			if err != nil {
				log.Println("error marshaling assertion:", err)
				continue
			}

			assertionUrl := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("bounties", uuid.String(), "assertions")}
			http.Post(assertionUrl.String(), "application/json", bytes.NewBuffer(j))
		}

		log.Println("recv:", event)
	}

}
