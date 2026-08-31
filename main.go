package main

import (
	"fmt"
)

func main() {
	fmt.Println("Version: vDev 1")
	network := GetNetworkInfo()

	network.Display()
}
