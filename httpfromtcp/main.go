package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)

	out := make([]byte, 8)
	currentLine := ""

	go func(c chan string) {
		for {
			// read 8b
			n, err := f.Read(out)
			if err != nil {
				break
			}

			parts := strings.Split(string(out[:n]), "\n")

			if len(parts) == 1 {
				currentLine += parts[0]
			} else {
				for i := 0; i < len(parts)-1; i++ {
					currentLine += parts[i]
				}
				c <- currentLine
				currentLine = parts[len(parts)-1]
			}
		}
		close(c)
	}(c)

	return c
}

func main() {
	f, err := os.OpenFile("messages.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for s := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", s)
	}
}
