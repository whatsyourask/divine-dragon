package remote_enum

import (
	"divine-dragon/transport"
	"divine-dragon/util"

	"github.com/go-ldap/ldap"
)

type LdapEnumModule struct {
	domain     string
	remoteHost string
	remotePort string
	username   string
	password   string
	verbose    bool

	conn   *ldap.Conn
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
	lem.conn = conn
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		lem.logger.Log.Noticef("Connected to LDAP service successfully - %s:%s", lem.remoteHost, lem.remotePort)
	}
	if lem.username != "" {
		err = transport.LDAPAuthenticatedBind(lem.conn, lem.domain, lem.username, lem.password)
	} else {
		err = transport.LDAPUnAuthenticatedBind(lem.conn, lem.domain)
	}
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		lem.logger.Log.Notice("Bound to LDAP service successfully.")
	}
	// //filter := transport.ConstructFilter("(&(objectCategory=%s)(objectClass=%s))", []any{"person", "user"})
	// filter := transport.ConstructFilter("(objectCategory=%s)", []any{"computer"})
	// err = transport.LDAPQuery(conn, "DC=htb,DC=local", filter, []string{})
	// err = lem.QueryAllUsers()
	err = lem.QueryAllOrgUnits("DC=htb,DC=local")
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.QueryAllUsers("DC=htb,DC=local")
	if err != nil {
		lem.logger.Log.Error(err)
	}
}

// ldapsearch -x -H "ldap://10.129.5.70:389/" -D ” -w ” -b "DC=htb,DC=local" "(&(objectCategory=person)(objectClass=user)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))" | grep dn:
func (lem *LdapEnumModule) QueryAllUsers(baseDN string) error {
	lem.logger.Log.Info("Querying for All Active User Accounts...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(objectClass=%s)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))",
		[]any{"person", "user"})
	resp, err := transport.LDAPQuery(lem.conn, baseDN, filter, []string{})
	if err != nil {
		return err
	}
	lem.logger.Log.Info("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

func (lem *LdapEnumModule) QueryAllOrgUnits(baseDN string) error {
	lem.logger.Log.Info("Querying for All Organizational Units (OU)...")
	filter := util.ConstructLDAPFilter("(&(!(objectCategory=%s))(OU=*)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))", []any{"computer"})
	resp, err := transport.LDAPQuery(lem.conn, baseDN, filter, []string{})
	if err != nil {
		return err
	}
	lem.logger.Log.Info("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	// util.LDAPGetEntryAttributes(resp)
	return nil
}

func (lem *LdapEnumModule) QueryAllComputers() {
}

func (lem *LdapEnumModule) QueryDomainAdmins() {

}

func (lem *LdapEnumModule) QueryEnterpriseAdmins() {

}

func (lem *LdapEnumModule) QueryAdmins() {

}

func (lem *LdapEnumModule) QueryRemoteDesktopGroup() {

}

func (lem *LdapEnumModule) QueryRemoteManagementGroup() {

}
