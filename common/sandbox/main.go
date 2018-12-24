package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.Mkdir("/Users/andrewdawson/Desktop/hello", 0777)

	// it is an error to make a directory which already exists
	fmt.Println(err)
}
