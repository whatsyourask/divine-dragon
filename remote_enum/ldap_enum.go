package remote_enum

import (
	"divine-dragon/transport"
	"divine-dragon/util"
	"fmt"
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
	err := lem.ConnectAndBind()
	if err != nil {
		lem.logger.Log.Info("Module completed with error. Exiting...")
		return
	}
	resp, err := lem.queryAllDomainControllers()
	if err != nil {
		lem.logger.Log.Error(err)
		if strings.Contains(err.Error(), "000004DC") {
			lem.logger.Log.Info("For successfull enumeration you have to provide some credentials, these LDAP service doesn't support querying as anonymous.")
			return
		}
	}
	util.LDAPListObjectsInResult(resp)
	resp, err = lem.queryAllWinServers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryAllOrgUnits()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryAllUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryAllComputers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryAllGroups()
	if err != nil {
		lem.logger.Log.Error(err)
	}
	util.LDAPListObjectsInResult(resp)
	resp, err = lem.queryAllAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryDomainAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryEnterpriseAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.querySchemaAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryKeyAdmins()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryPrintOperators()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryBackupOperators()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryRemoteDesktopUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryRemoteManagementUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryASREPRoastableUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryKerberoastableUsers()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	resp, err = lem.queryPasswordsInDescription()
	if err != nil {
		lem.logger.Log.Error(err)
	} else {
		util.LDAPListObjectsInResult(resp)
	}
	defer lem.conn.Close()
}

func (lem *LdapEnumModule) ConnectAndBind() error {
	conn, err := transport.LDAPConnect(lem.remoteHost, lem.remotePort)
	lem.conn = conn
	if err != nil {
		lem.logger.Log.Error(err)
		return err
	} else {
		lem.logger.Log.Noticef("Connected to LDAP service successfully - %s:%s", lem.remoteHost, lem.remotePort)
	}
	if lem.username != "" {
		lem.baseDN, err = util.ConstructBaseDN(lem.domain)
		if err != nil {
			lem.logger.Log.Error(err)
			return err
		}
		err = transport.LDAPAuthenticatedBind(lem.conn, lem.username, lem.password)
	} else {
		err = transport.LDAPUnAuthenticatedBind(lem.conn, lem.domain)
	}
	if err != nil {
		lem.logger.Log.Error(err)
		return err
	} else {
		lem.logger.Log.Notice("Bound to LDAP service successfully.")
	}
	return nil
}

//
// TODO:
// Refactoring of query functions
// Return the result, don't just print it!
//

func (lem *LdapEnumModule) doQuery(filter string) (*ldap.SearchResult, error) {
	resp, err := transport.LDAPQuery(lem.conn, lem.baseDN, filter, []string{})
	if err != nil {
		return nil, fmt.Errorf("[-] query was not successful: %v", err)
	}
	if len(resp.Entries) < 1 {
		return nil, fmt.Errorf("[!] query was successful, but has no results.")
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryAllUsers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("[!] Querying for all active User Accounts...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(objectClass=%s)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))",
		[]any{"person", "user"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryAllOrgUnits() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Organizational Units (OU)...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(!(userAccountControl:1.2.840.113556.1.4.803:=%s)))", []any{"organizationalUnit", "2"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryAllComputers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Computers...")
	filter := util.ConstructLDAPFilter("(sAMAccountType=%s)", []any{"805306369"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryAllGroups() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Groups...")
	filter := util.ConstructLDAPFilter("(objectCategory=%s)", []any{"group"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryAllAdmins() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Admins...")
	filter := util.ConstructLDAPFilter("(adminCount=%s)", []any{"1"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryDomainAdmins() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Domain Admins...")
	domainAdminsCN := "CN=Domain Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{domainAdminsCN})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryEnterpriseAdmins() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Enterprise Admins...")
	enterpriseAdminsCN := "CN=Enterprise Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{enterpriseAdminsCN})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) querySchemaAdmins() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Schema Admins...")
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{"CN=Schema Admins,CN=Users,DC=htb,DC=local"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryKeyAdmins() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Key Admins...")
	keyAdminsCN := "CN=Key Admins,CN=Users," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{keyAdminsCN})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryPrintOperators() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Print Operators...")
	printOperatorsCN := "CN=CN=Print Operators,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{printOperatorsCN})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryBackupOperators() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Backup Operators...")
	backupOperatorsCN := "CN=CN=Backup Operators,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{backupOperatorsCN})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryASREPRoastableUsers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all ASREPRoastable Users...")
	filter := util.ConstructLDAPFilter("(&(UserAccountControl:1.2.840.113556.1.4.803:=%s)(!(UserAccountControl:1.2.840.113556.1.4.803:=%s))(objectCategory=%s))", []any{"4194304", "2", "person"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryKerberoastableUsers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Kerberoastable Users...")
	filter := util.ConstructLDAPFilter("(&(objectClass=%s)(servicePrincipalName=%s)(!(cn=%s))(!(userAccountControl:1.2.840.113556.1.4.803:=%s)))", []any{"user", "*", "krbtgt", "2"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// independent query
func (lem *LdapEnumModule) queryPasswordsInDescription() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for passwords in description...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(|(description=%s)(description=%s)))", []any{"user", "*pass*", "*pwd*"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryRemoteDesktopUsers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Remote Desktop Users...")
	remoteDesktopUsers := "CN=Remote Desktop Users,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{remoteDesktopUsers})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

// default group
func (lem *LdapEnumModule) queryRemoteManagementUsers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Remote Management Users...")
	remoteDesktopUsers := "CN=Remote Management Users,CN=Builtin," + lem.baseDN
	filter := util.ConstructLDAPFilter("(memberOf:1.2.840.113556.1.4.1941:=%s)", []any{remoteDesktopUsers})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

func (lem *LdapEnumModule) queryAllDomainControllers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Domain Controllers (DCs)...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(userAccountControl:1.2.840.113556.1.4.803:=%s))", []any{"computer", "8192"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

func (lem *LdapEnumModule) queryAllWinServers() (*ldap.SearchResult, error) {
	lem.logger.Log.Info("Querying for all Windows Servers...")
	filter := util.ConstructLDAPFilter("(&(objectCategory=%s)(operatingSystem=%s)(!userAccountControl:1.2.840.113556.1.4.803:=%s))", []any{"computer", "*server*", "8192"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		lem.logger.Log.Error(err)
		return nil, err
	}
	return resp, nil
}

func (lem *LdapEnumModule) RunAndQueryOnlyKerberoastableUsers() (map[string]string, error) {
	lem.ConnectAndBind()
	lem.logger.Log.Info("Querying for all Kerberoastable Users...")
	filter := util.ConstructLDAPFilter("(&(objectClass=%s)(servicePrincipalName=%s)(!(cn=%s))(!(userAccountControl:1.2.840.113556.1.4.803:=%s)))", []any{"user", "*", "krbtgt", "2"})
	resp, err := lem.doQuery(filter)
	if err != nil {
		return nil, err
	}
	SPNs := make(map[string]string)
	for _, entry := range resp.Entries {
		sAMAccountName := entry.GetAttributeValue("sAMAccountName")
		if !strings.Contains(sAMAccountName, "$") {
			SPN := entry.GetAttributeValue("servicePrincipalName")
			SPNs[sAMAccountName] = SPN
		}
	}
	defer lem.conn.Close()
	return SPNs, nil
}
