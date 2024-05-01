package main

import "divine-dragon/c2"

func main() {
	c2m := c2.NewC2Module(
		"127.0.0.1",
		"8888",
	)
	c2m.Run()
}
