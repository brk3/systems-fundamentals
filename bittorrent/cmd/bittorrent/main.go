package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Open("debian-11.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Println("error opening torrent file: ", err)
		os.Exit(1)
	}
	defer f.Close()

	// r := bufio.NewReader(f)
	// torrent := bittorrent.bencodeTorrent{}
	// err = bencode.Unmarshal(r, &torrent)
	// if err != nil {
	// 	fmt.Println("error unmarshalling torrent file: ", err)
	// 	os.Exit(1)
	// }

	// infoHash, err := torrent.Info.infoHash()
	// if err != nil {
	// 	fmt.Println("error generating infoHash: ", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("infoHash: %x\n", infoHash)

	// trackerURL, err := torrent.trackerURL("paulstorrentclient!", 8080)
	// if err != nil {
	// 	fmt.Println("error generating trackerURL: ", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("trackerURL: ", trackerURL)
}
