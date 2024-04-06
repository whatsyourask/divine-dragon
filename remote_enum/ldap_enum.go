package remote_enum

import (
	"divine-dragon/transport"
	"divine-dragon/util"
)

type LdapEnumModule struct {
	domain     string
	remoteHost string
	remotePort string
	username   string
	password   string
	verbose    bool

	logger util.Logger
}

func NewLdapEnumModule(domainOpt string, remoteHostOpt string, remotePortOpt string,
	usernameOpt string, passwordOpt string, verboseOpt bool, logFileNameOpt string) *LdapEnumModule {
	lem := LdapEnumModule{
		domain:     domainOpt,
		remoteHost: remoteHostOpt,
		remotePort: remotePortOpt,
		username:   usernameOpt,
		password:   passwordOpt,
		verbose:    verboseOpt,
	}
	lem.logger = util.LdapEnumLogger(verboseOpt, logFileNameOpt)
	return &lem
}

func (lem *LdapEnumModule) Run() {
	conn, err := transport.LDAPConnect(lem.remoteHost, lem.remotePort)
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		lem.logger.Log.Noticef("Connected to LDAP service successfully - %s:%s", lem.remoteHost, lem.remotePort)
	}
	if lem.username != "" {
		err = transport.LDAPAuthenticatedBind(conn, lem.domain, lem.username, lem.password)
	} else {
		err = transport.LDAPUnAuthenticatedBind(conn, lem.domain)
	}
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		lem.logger.Log.Notice("Bound to LDAP service successfully.")
	}
	//filter := transport.ConstructFilter("(&(objectCategory=%s)(objectClass=%s))", []any{"person", "user"})
	filter := transport.ConstructFilter("(objectCategory=%s)", []any{"computer"})
	err = transport.LDAPQuery(conn, "DC=htb,DC=local", filter, []string{})
	if err != nil {
		lem.logger.Log.Error(err)
	}
}
