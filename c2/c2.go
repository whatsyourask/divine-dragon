package c2

import (
	"divine-dragon/transport"
	"divine-dragon/util"
)

type C2Module struct {
	localHost string
	localPort string
	c2s       *transport.C2Server
	logger    util.Logger
}

func NewC2Module(localHostOpt, localPortOpt string) *C2Module {
	c2m := C2Module{
		localHost: localHostOpt,
		localPort: localPortOpt,
	}
	c2m.logger = util.C2Logger(false, "")
	c2m.logger.Log.Info("Initializing a C2 server...")
	password := util.RandString(24)
	c2m.logger.Log.Infof("Operator account has the following password - %s\n", password)
	c2, err := transport.NewC2Server(localHostOpt, localPortOpt, password)
	if err != nil {
		c2m.logger.Log.Error(err)
		return nil
	}
	c2m.c2s = c2
	return &c2m
}

func (c2m *C2Module) Run() {
	c2m.logger.Log.Infof("A new C2 server started on %s:%s", c2m.localHost, c2m.localPort)
	go c2m.protect(c2m.c2s.Run)
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
