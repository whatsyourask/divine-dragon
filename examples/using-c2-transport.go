package main

import (
	"divine-dragon/transport"
	"fmt"
)

func main() {
	c2s := transport.NewC2Server("127.0.0.1", "8888")
	err := c2s.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
