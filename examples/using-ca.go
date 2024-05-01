package main

import (
	"divine-dragon/util"
	"fmt"
)

func main() {
	rca, err := util.NewRootCertificateAuthority()
	if err != nil {
		fmt.Println(err.Error())
	}
	err = rca.CreateTLSCert("127.0.0.1")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = rca.DumpAll()
	if err != nil {
		fmt.Println(err.Error())
	}
}
