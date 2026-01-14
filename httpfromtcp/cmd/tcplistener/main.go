package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"httpfromtcp/internal/request"
)

func main() {
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

		r, err := request.RequestFromReader(conn)
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for k, v := range r.Headers {
			fmt.Printf("%s: %s\n", strings.ToUpper(k), v)
		}

		fmt.Println("Body:")
		fmt.Println(string(r.Body))
	}
}
