package main

import "divine-dragon/cli"

func main() {
	tcli, _ := cli.NewToolCommandLineInterface()
	tcli.Run()
}
