package remote_enum

import (
	"divine-dragon/transport"
	"divine-dragon/util"
	"strings"

	"github.com/go-ldap/ldap"
)

type LdapEnumModule struct {
	domain     string
	remoteHost string
	remotePort string
	username   string
	password   string
	baseDN     string
	verbose    bool

	conn   *ldap.Conn
	logger util.Logger
}

func NewLdapEnumModule(domainOpt string, remoteHostOpt string, remotePortOpt string,
	usernameOpt string, passwordOpt string, baseDNOpt string, verboseOpt bool, logFileNameOpt string) *LdapEnumModule {
	lem := LdapEnumModule{
		domain:     domainOpt,
		remoteHost: remoteHostOpt,
		remotePort: remotePortOpt,
		username:   usernameOpt,
		password:   passwordOpt,
		verbose:    verboseOpt,
	}
	if baseDNOpt != "" {
		lem.baseDN = baseDNOpt
	} else {
		baseDNTemp, err := util.ConstructBaseDN(domainOpt)
		if err != nil {
			return nil
		}
		lem.baseDN = baseDNTemp
	}
	lem.logger = util.LdapEnumLogger(verboseOpt, logFileNameOpt)
	return &lem
}

func (lem *LdapEnumModule) Run() {
	conn, err := transport.LDAPConnect(lem.remoteHost, lem.remotePort)
	lem.conn = conn
	if err != nil {
		lem.logger.Log.Error(err)
		lem.logger.Log.Info("Exiting...")
		return
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
		lem.logger.Log.Info("Exiting...")
		return
	} else {
		lem.logger.Log.Notice("Bound to LDAP service successfully.")
	}
	err = lem.queryAllDomainControllers()
	if err != nil {
		lem.logger.Log.Error(err)
		if strings.Contains(err.Error(), "000004DC") {
			lem.logger.Log.Info("For successfull enumeration you have to provide some credentials, these LDAP service doesn't support querying as anonymous.")
			return
		}
	}
	err = lem.queryAllWinServers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryAllOrgUnits()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryAllUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryAllComputers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryAllGroups()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryAllAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryDomainAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryEnterpriseAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.querySchemaAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryKeyAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryPrintOperators()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryBackupOperators()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryRemoteDesktopUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryRemoteManagementUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryASREPRoastableUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryKerberoastableUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	err = lem.queryPasswordsInDescription()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	defer lem.conn.Close()
}

// independent query
func (lem *LdapEnumModule) queryAllUsers() error {
	lem.logger.Log.Info("Querying for all active User Accounts...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(objectClass=%s)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))",
		[]any{"person", "user"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryAllOrgUnits() error {
	lem.logger.Log.Info("Querying for all Organizational Units (OU)...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(!(userAccountControl:1.2.840.113556.1.4.803:=%s)))", []any{"organizationalUnit", "2"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryAllComputers() error {
	lem.logger.Log.Info("Querying for all Computers...")
	filter := util.ConstructLDAPFilter("(sAMAccountType=%s)", []any{"805306369"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryAllGroups() error {
	lem.logger.Log.Info("Querying for all Groups...")
	filter := util.ConstructLDAPFilter("(objectCategory=%s)", []any{"group"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryAllAdmins() error {
	lem.logger.Log.Info("Querying for all Admins...")
	filter := util.ConstructLDAPFilter("(adminCount=%s)", []any{"1"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryDomainAdmins() error {
	lem.logger.Log.Info("Querying for all Domain Admins...")
	domainAdminsCN := "CN=Domain Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{domainAdminsCN})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryEnterpriseAdmins() error {
	lem.logger.Log.Info("Querying for all Enterprise Admins...")
	enterpriseAdminsCN := "CN=Enterprise Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{enterpriseAdminsCN})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) querySchemaAdmins() error {
	lem.logger.Log.Info("Querying for all Schema Admins...")
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{"CN=Schema Admins,CN=Users,DC=htb,DC=local"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryKeyAdmins() error {
	lem.logger.Log.Info("Querying for all Key Admins...")
	keyAdminsCN := "CN=Key Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{keyAdminsCN})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryPrintOperators() error {
	lem.logger.Log.Info("Querying for all Print Operators...")
	printOperatorsCN := "CN=CN=Print Operators,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{printOperatorsCN})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryBackupOperators() error {
	lem.logger.Log.Info("Querying for all Backup Operators...")
	backupOperatorsCN := "CN=CN=Backup Operators,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{backupOperatorsCN})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryASREPRoastableUsers() error {
	lem.logger.Log.Info("Querying for all ASREPRoastable Users...")
	filter := util.ConstructLDAPFilter("(&(UserAccountControl:1.2.840.113556.1.4.803:=%s)(!(UserAccountControl:1.2.840.113556.1.4.803:=%s))(objectCategory=%s))", []any{"4194304", "2", "person"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryKerberoastableUsers() error {
	lem.logger.Log.Info("Querying for all Kerberoastable Users...")
	filter := util.ConstructLDAPFilter("(&(objectClass=%s)(servicePrincipalName=%s)(!(cn=%s))(!(userAccountControl:1.2.840.113556.1.4.803:=%s)))", []any{"user", "*", "krbtgt", "2"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// independent query
func (lem *LdapEnumModule) queryPasswordsInDescription() error {
	lem.logger.Log.Info("Querying for passwords in description...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(|(description=%s)(description=%s)))", []any{"user", "*pass*", "*pwd*"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryRemoteDesktopUsers() error {
	lem.logger.Log.Info("Querying for all Remote Desktop Users...")
	remoteDesktopUsers := "CN=Remote Desktop Users,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{remoteDesktopUsers})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

// default group
func (lem *LdapEnumModule) queryRemoteManagementUsers() error {
	lem.logger.Log.Info("Querying for all Remote Management Users...")
	remoteDesktopUsers := "CN=Remote Management Users,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{remoteDesktopUsers})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

func (lem *LdapEnumModule) queryAllDomainControllers() error {
	lem.logger.Log.Info("Querying for all Domain Controllers (DCs)...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(userAccountControl:1.2.840.113556.1.4.803:=%s))", []any{"computer", "8192"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}

func (lem *LdapEnumModule) queryAllWinServers() error {
	lem.logger.Log.Info("Querying for all Windows Servers...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(operatingSystem=%s)(!userAccountControl:1.2.840.113556.1.4.803:=%s))", []any{"computer", "*server*", "8192"})
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return err
	}
	if len(resp.Entries) < 1 {
		lem.logger.Log.Error("Query was not successfull...")
		return nil
	}
	lem.logger.Log.Notice("Result of the query:\n")
	util.LDAPListObjectsInResult(resp)
	return nil
}
