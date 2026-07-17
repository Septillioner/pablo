package main

import (
	"fmt"
	"os"
)

const version = "1.2.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	fmt.Println("myservice ready")
}
