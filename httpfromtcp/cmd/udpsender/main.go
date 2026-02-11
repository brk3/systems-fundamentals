package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Printf("error resolving udp addr: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	defer conn.Close()
	if err != nil {
		log.Printf("error dialing udp addr: %v", err)
	}

	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		input, err := stdin.ReadString('\n')
		if err != nil {
			log.Printf("error reading line from stdin")
		}
		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Printf("error writing input to udp conn")
		}
	}
}
