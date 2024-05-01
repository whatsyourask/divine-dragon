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
	c2, err := transport.NewC2Server(localHostOpt, localPortOpt)
	if err != nil {
		c2m.logger.Log.Error(err)
		return nil
	}
	c2m.c2s = c2
	return &c2m
}

func (c2m *C2Module) Run() {
	go c2m.c2s.Run()
}
