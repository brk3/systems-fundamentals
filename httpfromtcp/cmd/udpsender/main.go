package main

import (
  "os"
  "fmt"
  "net"
  "bufio"
)

func main() {
  addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
  if err != nil {
    fmt.Errorf("error resolving udp addr: %v", err)
  }

  conn, err := net.DialUDP("udp", nil, addr)
  defer conn.Close()
  if err != nil {
    fmt.Errorf("error dialing udp addr: %v", err)
  }

  stdin := bufio.NewReader(os.Stdin)
  for {
    fmt.Printf("> ")
    input, err := stdin.ReadString('\n')
    if err != nil {
      fmt.Errorf("error reading line from stdin")
    }
    _, err = conn.Write([]byte(input))
    if err != nil {
      fmt.Errorf("error writing input to udp conn")
    }
  }
}
