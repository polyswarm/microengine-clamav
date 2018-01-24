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
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
	"github.com/polyswarm/polyswarm/bounty"
	uuid "github.com/satori/go.uuid"
)

const TIMEOUT = 3000 * time.Second
const MAX_DATA_SIZE = 50 * 1024 * 1024
const POSTER_ADDRESS = "0x6B68D0bf6b983C3662D503eD2D44E0DF4a9BB874"

func connectToClamd(host string) (*clamd.Clamd, error) {
	u := url.URL{Scheme: "tcp", Host: host}
	ret := clamd.NewClamd(u.String())

	timeout := time.After(TIMEOUT)
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

func connectToPolyswarm(host string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/events"}

	timeout := time.After(TIMEOUT)
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

func retrieveFileFromIpfs(host, resource string) (io.ReadCloser, error) {
	artifactUrl := url.URL{Scheme: "http", Host: host, Path: path.Join("artifacts", resource)}
	statResp, err := http.Get(artifactUrl.String() + "/stat")
	if err != nil {
		return nil, err
	}
	defer statResp.Body.Close()

	var stat bounty.ArtifactStats
	json.NewDecoder(statResp.Body).Decode(&stat)

	if stat.DataSize == 0 || stat.DataSize > MAX_DATA_SIZE {
		return nil, errors.New("invalid ipfs artifact")
	}

	resp, err := http.Get(artifactUrl.String())
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func main() {
	time.Sleep(15 * time.Second)
	log.Println("Starting microengine")

	clamd, err := connectToClamd(os.Getenv("CLAMD_HOST"))
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

			log.Println("got bounty:", body)

			author, ok := body["Author"].(string)
			if !ok {
				log.Println("invalid author address")
				continue
			}

			if common.HexToAddress(author) != common.HexToAddress(POSTER_ADDRESS) {
				log.Println("unrecognized poster:", author)
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

			r, err := retrieveFileFromIpfs(polyswarmHost, uri)
			if err != nil {
				log.Println("error retrieving artifact:", err)
				continue
			}
			defer r.Close()

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
