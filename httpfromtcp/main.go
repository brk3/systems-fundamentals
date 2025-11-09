package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.OpenFile("messages.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	out := make([]byte, 8)
	currentLine := ""

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
			fmt.Printf("read: %s\n", currentLine)
			currentLine = parts[len(parts)-1]
		}
	}
}
