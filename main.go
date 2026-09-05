package main

import (
	"fmt"
)

func main() {
	fmt.Println("V1")
	network := GetNetworkInfo()

	network.Display()
}
