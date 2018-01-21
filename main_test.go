package main_test

import (
	"os"
	"testing"

	"github.com/ipfs/go-ipfs-api"
)

const (
	Directory = 1
	File      = 2
)

func connectToIpfs(t *testing.T) *shell.Shell {
	ipfssh := shell.NewShell(os.Getenv("Ipfs_HOST"))
	if ipfssh == nil {
		t.Fatal("could not connect to ipfs")
	}
	return ipfssh
}

func TestSingleFileIpfs(t *testing.T) {
	ipfssh := connectToIpfs(t)
	links, err := ipfssh.List("QmYNmQKp6SuaVrpgWRsPTgCQCnpxUYGq76YEKBXuj2N4H6")
	if err != nil {
		t.Fatal(err)
	}

	if len(links) != 0 {
		t.Fatal("single object has links")
	}
}

func TestFlatDirectoryIpfs(t *testing.T) {
	ipfssh := connectToIpfs(t)
	links, err := ipfssh.List("QmRvxukeoLxSYY2VdnGaa3nyqwDpGN8Nydp9rBUVJ2y48P")
	if err != nil {
		t.Fatal(err)
	}

	for _, link := range links {
		if link.Type != File {
			t.Fatal("expected a file")
		}
	}
}

func TestMultiLevelDirectoryIpfs(t *testing.T) {
	ipfssh := connectToIpfs(t)
	links, err := ipfssh.List("QmQnHj3vp4LGnXeHS58YjGwbanG6YF6hqqqNDmMmRqFXmy")
	if err != nil {
		t.Fatal(err)
	}

	for _, link := range links {
		t.Log(link)
	}
}
