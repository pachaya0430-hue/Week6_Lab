package main

import (
	"fmt"
	"time"

	"moviehub/auth"
)

func main() {
	password := "popcorn2569"

	hashA, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}

	hashB, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}

	fmt.Println(hashA)
	fmt.Println(hashB)
	fmt.Println(password, "ตรงกับ a:", auth.CheckPassword(password, hashA))
	fmt.Println(password, "ตรงกับ b:", auth.CheckPassword(password, hashB))
	fmt.Println("popcorn2568 ตรงกับ a:", auth.CheckPassword("popcorn2568", hashA))

	for _, cost := range []int{8, 10, 12} {
		start := time.Now()
		_, err := auth.HashPasswordWithCost(password, cost)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("cost %d ใช้เวลา %v\n", cost, time.Since(start).Round(time.Millisecond))
	}
}
