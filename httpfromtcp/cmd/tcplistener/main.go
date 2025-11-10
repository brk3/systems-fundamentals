package main

import (
	"fmt"
	"net"
	"io"
	"strings"
	"log"
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
		// flush
		c <- currentLine
		close(c)
	}(c)

	return c
}

func main() {
	/*
	f, err := os.OpenFile("messages.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for s := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", s)
	}
	*/

	ln, err := net.Listen("tcp", ":42069")
	defer ln.Close()
	log.Println("listening on :42069")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		log.Println("accepted conn")
		//go getLinesChannel(conn)
		for s := range getLinesChannel(conn) {
			fmt.Println(s)
		}
	}
}
