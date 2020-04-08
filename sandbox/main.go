package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Open("/Users/dandrew/Desktop/andrew.txt")
	if err != nil {
		panic(err)
	}
	b := make([]byte, 100, 100)
	n, err := f.Read(b)
	fmt.Println("n: ", n)
	fmt.Println("err: ", err)
	fmt.Println(string(b))
}
