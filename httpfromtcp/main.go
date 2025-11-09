package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("messages.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	out := make([]byte, 8)
	for {
		_, err := f.Read(out)
		if err != nil {
			//fmt.Printf("err %v", err)
			break
		}
		fmt.Printf("read: %s\n", out)
	}
}
