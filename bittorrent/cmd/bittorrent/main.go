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

	// ...
}
