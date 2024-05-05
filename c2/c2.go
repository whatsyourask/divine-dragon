package c2

import (
	"bytes"
	"crypto/tls"
	"divine-dragon/transport"
	"divine-dragon/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type C2Module struct {
	localHost                string
	localPort                string
	c2s                      *transport.C2Server
	password                 string
	logger                   util.Logger
	apiUrl                   string
	authorizationToken       string
	authorizationTokenExpire string
}

func NewC2Module(localHostOpt, localPortOpt string) *C2Module {
	c2m := C2Module{
		localHost: localHostOpt,
		localPort: localPortOpt,
	}
	c2m.logger = util.C2Logger(false, "")
	c2m.logger.Log.Info("Initializing a C2 server...")
	c2m.password = util.RandString(24)
	c2m.logger.Log.Infof("Operator account has the following password - %s\n", c2m.password)
	c2, err := transport.NewC2Server(localHostOpt, localPortOpt, c2m.password)
	if err != nil {
		c2m.logger.Log.Error(err)
		return nil
	}
	c2m.c2s = c2
	c2m.apiUrl = "https://" + "127.0.0.1" + ":" + c2m.localPort
	return &c2m
}

func (c2m *C2Module) Run() {
	c2m.logger.Log.Infof("A new C2 server started on %s:%s", c2m.localHost, c2m.localPort)
	go c2m.protect(c2m.c2s.Run)
	// err := c2m.operatorLogin()
	// if err != nil {
	// 	c2m.logger.Log.Error(err)
	// }
	// fmt.Println(c2m.GetAgents())
}

func (c2m *C2Module) protect(f func() error) {
	defer func() {
		if err := recover(); err != nil {
			c2m.logger.Log.Noticef("Recovered C2 server: %v", err)
		}
	}()
	err := f()
	if err != nil {
		c2m.logger.Log.Error(err)
	}
}

func (c2m *C2Module) GetAgents() []transport.Agent {
	agents := c2m.getAgents()
	return agents
}

func (c2m *C2Module) getAgents() []transport.Agent {
	err := c2m.operatorLogin()
	if err != nil {
		c2m.logger.Log.Error(err)
		return nil
	}
	req, err := http.NewRequest("GET", c2m.apiUrl+"/operator/agents", nil)
	if err != nil {
		c2m.logger.Log.Errorf("can't create a new request: %v", err)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+c2m.authorizationToken)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		c2m.logger.Log.Errorf("can't perform a POST request: %v", err)
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c2m.logger.Log.Errorf("can't do io.ReadAll: %v", err)
		return nil
	}
	var respJson []transport.Agent
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		c2m.logger.Log.Errorf("can't unmarshal a JSON in response: %v", err)
		return nil
	}
	return respJson
}

func (c2m *C2Module) AddJob(agentUuid string) error {
	uuid := uuid.New()
	jobUuid := uuid.String()
	addJobPostBody := map[string]string{"agent-uuid": agentUuid, "job-uuid": jobUuid}
	addJobPostJson, err := json.Marshal(addJobPostBody)
	if err != nil {
		return fmt.Errorf("can't marshal a json for POST login: %v", err)
	}
	req, err := http.NewRequest("POST", c2m.apiUrl+"/agents/job/add", bytes.NewBuffer(addJobPostJson))
	if err != nil {
		return fmt.Errorf("can't create a POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c2m.authorizationToken)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		c2m.logger.Log.Errorf("can't perform a POST request: %v", err)
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c2m.logger.Log.Errorf("can't do io.ReadAll: %v", err)
		return nil
	}
	var respJson struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		c2m.logger.Log.Errorf("can't unmarshal a JSON in response: %v", err)
		return nil
	}
	if respJson.Status == "ok" {
		return nil
	} else if respJson.Status == "agent not found" {
		return fmt.Errorf("there is no such agent")
	} else {
		return fmt.Errorf("can't add a job to the agent")
	}
}

func (c2m *C2Module) operatorLogin() error {
	loginPostBody := map[string]string{"username": "c2operator", "password": c2m.password}
	loginPostJson, err := json.Marshal(loginPostBody)
	if err != nil {
		return fmt.Errorf("can't marshal a json for POST login: %v", err)
	}
	req, err := http.NewRequest("POST", c2m.apiUrl+"/login", bytes.NewBuffer(loginPostJson))
	if err != nil {
		return fmt.Errorf("can't create a POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't perform a POST request: %v", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't do io.ReadAll: %v", err)
	}
	var respJson struct {
		Code   int    `json:"code"`
		Expire string `json:"expire"`
		Token  string `json:"token"`
	}
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		return fmt.Errorf("can't unmarshal a JSON in response: %v", err)
	}
	if respJson.Code == 200 {
		c2m.authorizationToken = respJson.Token
		c2m.authorizationTokenExpire = respJson.Expire
	} else {
		return fmt.Errorf("can't login with this credentials: %v, %v", respJson.Code, err)
	}
	return nil
}
