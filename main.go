package main

import (
	"divine-dragon/remote_enum"
)

func main() {
	// asrep_module := remote_exploit.SetupModule("htb.local", "10.129.40.155", false, false, false, "", "remote_exploit/names.txt", "", 10, 0)
	// asrep_module.Run()

	kerb_enum_module := remote_enum.SetupModule("htb.local", "10.129.40.155", false, false, false, "remote_exploit/names.txt", "", 80, 0)
	kerb_enum_module.Run()
}
