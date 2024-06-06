package main

import (
	"divine-dragon/transport"
	"fmt"
)

func main() {
	c2s, err := transport.NewC2Server("127.0.0.1", "8888", "test")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = c2s.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
