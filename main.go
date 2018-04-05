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
	"strconv"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type Success struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"`
}

type ArtifactStats struct {
	Hash           string
	BlockSize      int
	CumulativeSize int
	DataSize       int
	NumLinks       int
}

type Event struct {
	Type string      `json:"event"`
	Data interface{} `json:"data"`
}

const TIMEOUT = 3000 * time.Second
const MAX_DATA_SIZE = 50 * 1024 * 1024
const BID_AMOUNT = 62500000000000000
const PASSWORD = "password"

func login(polyswarmHost string) error {
	accountsURL := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("accounts")}
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}

	accountsResp, err := client.Get(accountsURL.String())
	if err != nil {
		return err
	}

	var success Success
	json.NewDecoder(accountsResp.Body).Decode(&success)

	accounts, ok := success.Result.([]interface{})
	if !ok {
		return errors.New("invalid accounts")
	}

	var p struct {
		Password string `json:"password"`
	}
	p.Password = PASSWORD

	j, err := json.Marshal(p)
	if err != nil {
		return err
	}

	var account string
	if len(accounts) == 0 {
		newAccountResp, err := client.Post(accountsURL.String(), "application/json", bytes.NewBuffer(j))
		if err != nil {
			return err
		}

		json.NewDecoder(newAccountResp.Body).Decode(&success)

		account, ok = success.Result.(string)
		if !ok {
			return errors.New("invalid account")
		}
	} else {
		account, ok = accounts[0].(string)
		if !ok {
			return errors.New("invalid account")
		}
	}

	unlockURL := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("accounts", account, "unlock")}
	client.Post(unlockURL.String(), "application/json", bytes.NewBuffer(j))

	log.Println("using account:", account)
	return nil
}

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

func retrieveFileFromIpfs(host, resource string, id int) (io.ReadCloser, error) {
	client := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	artifactURL := url.URL{Scheme: "http", Host: host, Path: path.Join("artifacts", resource, strconv.Itoa(id))}
	statResp, err := client.Get(artifactURL.String() + "/stat")
	if err != nil {
		return nil, err
	}
	defer statResp.Body.Close()

	var success Success
	json.NewDecoder(statResp.Body).Decode(&success)

	stats, ok := success.Result.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid ipfs artifact stats")
	}

	dataSize, ok := stats["data_size"].(float64)
	if !ok {
		return nil, errors.New("invalid ipfs artifact stats")
	}

	if uint(dataSize) == 0 || uint(dataSize) > MAX_DATA_SIZE {
		return nil, errors.New("invalid ipfs artifact")
	}

	resp, err := client.Get(artifactURL.String())
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func scanBounty(polyswarmHost string, clamd *clamd.Clamd, uri string) ([]bool, string) {
	verdicts := make([]bool, 0, 256)
	var metadata bytes.Buffer

	log.Println("retrieving artifacts:", uri)
	for i := 0; i < 256; i++ {
		r, err := retrieveFileFromIpfs(polyswarmHost, uri, i)
		if err != nil {
			log.Println("error retrieving artifact", i, ":", err)
			break
		}
		defer r.Close()

		log.Println("got artifact, scanning:", uri)

		response, err := clamd.ScanStream(r, make(chan bool))
		if err != nil {
			log.Println("error scanning artifact:", err)
			continue
		}

		log.Println("scanned artifact:", uri, i)

		verdict := false
		for s := range response {
			verdict = verdict || s.Status == "FOUND"

			verdicts = append(verdicts, s.Status == "FOUND")
			metadata.WriteString(s.Description)
			metadata.WriteString(";")
		}
	}

	return verdicts, metadata.String()
}

func makeBoolMask(len int) []bool {
	ret := make([]bool, len)
	for i := 0; i < len; i++ {
		ret[i] = true
	}
	return ret
}

func main() {
	time.Sleep(30 * time.Second)
	log.Println("Starting microengine")

	clamd, err := connectToClamd(os.Getenv("CLAMD_HOST"))
	if err != nil {
		log.Fatalln(err)
	}

	polyswarmHost := os.Getenv("POLYSWARM_HOST")
	if err := login(polyswarmHost); err != nil {
		log.Fatalln(err)
	}

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

		var event Event
		json.Unmarshal(message, &event)

		if event.Type == "bounty" {
			data, ok := event.Data.(map[string]interface{})
			if !ok {
				log.Println("invalid bounty object")
				continue
			}

			log.Println("got bounty:", data)

			guid, ok := data["guid"].(string)
			if !ok {
				log.Println("invalid guid")
				continue
			}

			uuid, err := uuid.FromString(guid)
			if err != nil {
				log.Println("invalid uuid:", err)
				continue
			}

			uri, ok := data["uri"].(string)
			if !ok {
				log.Println("invalid uri")
				continue
			}

			verdicts, metadata := scanBounty(polyswarmHost, clamd, uri)

			var a struct {
				Verdicts []bool `json:"verdicts"`
				Mask     []bool `json:"mask"`
				Bid      string `json:"bid"`
				Metadata string `json:"metadata"`
			}

			a.Verdicts = verdicts
			a.Mask = makeBoolMask(len(verdicts))
			a.Bid = strconv.Itoa(BID_AMOUNT)
			a.Metadata = metadata

			j, err := json.Marshal(a)
			if err != nil {
				log.Println("error marshaling assertion:", err)
				continue
			}

			assertionURL := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("bounties", uuid.String(), "assertions")}
			client := http.Client{
				Timeout: time.Duration(10 * time.Second),
			}

			client.Post(assertionURL.String(), "application/json", bytes.NewBuffer(j))
			log.Println("posted assertion")
		} else if event.Type == "block" {
			data, ok := event.Data.(map[string]interface{})
			if !ok {
				log.Println("invalid block event")
				continue
			}

			number, ok := data["number"].(float64)
			if !ok {
				log.Println("invalid block event")
				continue
			}

			if int(number)%100 == 0 {
				log.Println("scanning pending bounties")

				pendingURL := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("bounties", "pending")}
				client := http.Client{
					Timeout: time.Duration(10 * time.Second),
				}

				pendingResp, err := client.Get(pendingURL.String())
				if err != nil {
					log.Println("error fetching pending bounties:", err)
					continue
				}

				var success Success
				json.NewDecoder(pendingResp.Body).Decode(&success)

				pending, ok := success.Result.([]interface{})
				if !ok {
					log.Println("invalid pending bounties")
					continue
				}

				for i := 0; i < len(pending); i++ {
					bounty, ok := pending[i].(map[string]interface{})
					if !ok {
						log.Println("invalid pending bounties")
						continue
					}

					guid, ok := bounty["guid"].(string)
					if !ok {
						log.Println("invalid pending bounties")
						continue
					}

					uuid, err := uuid.FromString(guid)
					if err != nil {
						log.Println("invalid uuid:", err)
						continue
					}

					uri, ok := bounty["uri"].(string)
					if !ok {
						log.Println("invalid pending bounties")
						continue
					}

					verdicts, _ := scanBounty(polyswarmHost, clamd, uri)

					var v struct {
						Verdicts []bool `json:"verdicts"`
					}

					v.Verdicts = verdicts

					j, err := json.Marshal(v)
					if err != nil {
						log.Println("error marshaling settlement:", err)
						continue
					}

					settleURL := url.URL{Scheme: "http", Host: polyswarmHost, Path: path.Join("bounties", uuid.String(), "settle")}
					client := http.Client{
						Timeout: time.Duration(10 * time.Second),
					}

					client.Post(settleURL.String(), "application/json", bytes.NewBuffer(j))
					log.Println("posted settlement")

				}
			}
		}

		log.Println("recv:", event)
	}

}
