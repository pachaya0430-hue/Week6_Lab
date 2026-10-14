package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/peek <token>")
		return
	}

	parts := strings.Split(os.Args[1], ".")
	if len(parts) != 3 {
		fmt.Println("Invalid JWT format")
		return
	}

	fmt.Println("JWT has 3 parts:")
	fmt.Println("Header:", parts[0])
	fmt.Println("Payload:", parts[1])
	fmt.Println("Signature: [hidden]")
}
