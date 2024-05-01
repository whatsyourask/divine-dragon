package cli

import (
	"bufio"
	"divine-dragon/c2"
	"divine-dragon/util"
	"fmt"
	"os"
)

type C2CommandLineInterface struct {
	label string
}

const defaultC2Label = "(c2-console) >>> "

func NewC2CommandLineInterface() (*C2CommandLineInterface, error) {
	c2cli := C2CommandLineInterface{}
	c2cli.label = defaultC2Label
	return &c2cli, nil
}

func (c2cli *C2CommandLineInterface) Run(host, port string) {
	c2s := c2.NewC2Module(host, port)
	c2s.Run()
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(c2cli.label)
	for scanner.Scan() {
		command := scanner.Text()
		formattedCommand := util.FormatCommand(command)
		fmt.Println("you entered: " + formattedCommand)
		// // err := c2cli.validateCommand(formattedCommand)
		// if err == nil {

		// }
	}
}

func (c2cli *C2CommandLineInterface) validateCommand(command string) {

}
