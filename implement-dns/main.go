package main

import (
	"fmt"
	"implement-dns/internal/dns"
)

func main() {
	fmt.Printf("% x\n", dns.EncodeDNSName([]byte("google.com")))
}
