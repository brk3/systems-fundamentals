package main

import (
	"fmt"
	"os"

	"go-bt-learning.brk3.github.io/internal/client"
	"go-bt-learning.brk3.github.io/internal/torrent"
	"go-bt-learning.brk3.github.io/internal/torrentfile"
)

func main() {
	f, err := os.Open("debian-11.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Printf("error opening torrent file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	tf, err := torrentfile.NewTorrentFile(f)
	if err != nil {
		fmt.Printf("error loading torrent file: %v\n", err)
		os.Exit(1)
	}
	t := torrent.NewTorrent(tf)
	err = t.Announce(client.PeerID, 6881)
	if err != nil {
		fmt.Printf("error announcing ourselves to tracker: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("received %d peers from tracker\n", len(t.Peers))
	data := t.Download()
	f, err = os.Create(t.File.Name)
	if err != nil {
		fmt.Printf("error opening output file: %v\n", err)
		os.Exit(1)
	}
	_, err = f.Write(data)
	if err != nil {
		fmt.Printf("error writing output file: %v\n", err)
		os.Exit(1)
	}
}
