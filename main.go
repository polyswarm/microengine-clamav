package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/dutchcoders/go-clamd"

	"github.com/ipfs/go-ipfs-api"
)

const MAX_DATA_SIZE = 50 * 1024 * 1024

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
	time.Sleep(15 * time.Second)
	log.Println("Starting microengine")

	ipfssh := shell.NewShell(os.Getenv("IPFS_HOST"))
	if ipfssh == nil {
		log.Fatalln("could not connect to ipfs")
	}
	log.Println(ipfssh)

	ipfs_uri := "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/readme"
	//	stat, err := ipfssh.ObjectStat(ipfs_uri)
	//	log.Println(stat)
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//
	//	if stat.DataSize > MAX_DATA_SIZE || stat.NumLinks == 0 {
	//		log.Fatalln("invalid artifact at uri")
	//	}

	r, err := ipfssh.Cat(ipfs_uri)
	defer r.Close()

	c, err := connectToClamd(os.Getenv("CLAMD_HOST"))
	if err != nil {
		log.Fatalln(err)
	}

	//reader := bytes.NewReader(clamd.EICAR)
	response, err := c.ScanStream(r, make(chan bool))

	for s := range response {
		log.Printf("%v %v\n", s, err)
	}
}
