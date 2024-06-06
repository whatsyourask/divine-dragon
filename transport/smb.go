package transport

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

func SmbConnect(remoteHost string, remotePort string, domain string, username string, password string, hash string) (net.Conn, *smb2.Session, error) {
	conn, err := net.Dial("tcp", remoteHost+":"+remotePort)
	if err != nil {
		return nil, nil, fmt.Errorf("can't establish a remote connection with error: %v", err)
	}
	var hashBytes []byte
	if hash != "" {
		hashBytes, err = hex.DecodeString(hash)
		if err != nil {
			return conn, nil, fmt.Errorf("can't decode your hash: %v", err)
		}
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
			Domain:   domain,
			Hash:     hashBytes,
		},
	}
	s, err := d.Dial(conn)
	if err != nil {
		return conn, nil, fmt.Errorf("can't establish a session: %v", err)
	}
	return conn, s, nil
}

func SmbClose(conn net.Conn, s *smb2.Session) error {
	err := conn.Close()
	if err != nil {
		return fmt.Errorf("can't close a connection: %v", err)
	}
	err = s.Logoff()
	if err != nil {
		return fmt.Errorf("can't close a session: %v", err)
	}
	return nil
}

func SmbHandleAuthError(err error) (bool, string) {
	eString := err.Error()
	if strings.Contains(eString, "The attempted logon is invalid") {
		return true, "Invalid password"
	} else {
		return false, "Some unknown error"
	}
}
