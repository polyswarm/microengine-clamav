package main_test

import (
	"errors"
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	clamd "github.com/dutchcoders/go-clamd"
)

const (
	Directory = 1
	File      = 2
)

//func connectToIpfs(t *testing.T) *shell.Shell {
//	ipfssh := shell.NewShell(os.Getenv("IPFS_HOST"))
//	if ipfssh == nil {
//		t.Fatal("could not connect to ipfs")
//	}
//	return ipfssh
//}
//
//func TestSingleFileIpfs(t *testing.T) {
//	ipfssh := connectToIpfs(t)
//	links, err := ipfssh.List("QmYNmQKp6SuaVrpgWRsPTgCQCnpxUYGq76YEKBXuj2N4H6")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if len(links) != 0 {
//		t.Fatal("single object has links")
//	}
//}
//
//func TestFlatDirectoryIpfs(t *testing.T) {
//	ipfssh := connectToIpfs(t)
//	links, err := ipfssh.List("QmRvxukeoLxSYY2VdnGaa3nyqwDpGN8Nydp9rBUVJ2y48P")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for _, link := range links {
//		if link.Type != File {
//			t.Fatal("expected a file")
//		}
//	}
//}
//
//func TestMultiLevelDirectoryIpfs(t *testing.T) {
//	ipfssh := connectToIpfs(t)
//	links, err := ipfssh.List("QmQnHj3vp4LGnXeHS58YjGwbanG6YF6hqqqNDmMmRqFXmy")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for _, link := range links {
//		t.Log(link)
//	}
//}

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

func TestMalware(t *testing.T) {
	clamd, err := connectToClamd(os.Getenv("CLAMD_HOST"))
	if err != nil {
		log.Fatalln(err)
	}

	r, _ := os.Open("./exe")
	defer r.Close()
	response, _ := clamd.ScanStream(r, make(chan bool))
	for s := range response {
		log.Println(s)
	}
}
