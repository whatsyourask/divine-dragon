package transport

import (
	"divine-dragon/util"
	"fmt"

	"github.com/gin-gonic/gin"
)

type C2Server struct {
	host string
	port string
	r    *gin.Engine
}

func NewC2Server(hostOpt, portOpt string) *C2Server {
	c2s := C2Server{
		host: hostOpt,
		port: portOpt,
	}
	c2s.r = gin.Default()
	return &c2s
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
