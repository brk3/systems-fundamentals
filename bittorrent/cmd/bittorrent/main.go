package main

import (
	"fmt"
	"os"

	"go-bt-learning.brk3.github.io/internal/bittorrent"
)

const (
	peerID = "paulsbittorentclient" // 20 chars
)

func main() {
	f, err := os.Open("debian-11.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Printf("error opening torrent file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	// TODO: add state machine?
	tf, err := bittorrent.NewTorrentFile(f)
	if err != nil {
		fmt.Printf("error loading torrent file: %v\n", err)
		os.Exit(1)
	}
	t := bittorrent.NewTorrent(tf)
	err = t.Announce(peerID, 6881)
	if err != nil {
		fmt.Printf("error announcing ourselves to tracker: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("peers: %d", t.Peers)
}
