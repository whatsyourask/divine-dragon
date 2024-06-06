// https://ldap.com/the-ldap-search-operation/
// https://cybernetist.com/2020/05/18/getting-started-with-go-ldap/

package transport

import (
	"divine-dragon/util"
	"fmt"

	"github.com/go-ldap/ldap"
)

func LDAPConnect(remoteHost string, remotePort string) (*ldap.Conn, error) {
	var ldapURL string
	if remotePort == "389" {
		ldapURL = fmt.Sprintf("ldap://%s:%s", remoteHost, remotePort)
	} else if remotePort == "636" {
		ldapURL = fmt.Sprintf("ldaps://%s:%s", remoteHost, remotePort)
	} else {
		return nil, fmt.Errorf("remotePort has incorrect value - %v", remotePort)
	}
	ldapCon, err := ldap.DialURL(ldapURL)
	if err != nil {
		return nil, fmt.Errorf("can't establish a remote with LDAP service: %v", err)
	}
	return ldapCon, nil
}

func LDAPUnAuthenticatedBind(ldapCon *ldap.Conn, domain string) error {
	baseDN, err := util.ConstructBaseDN(domain)
	if err != nil {
		return fmt.Errorf("can't determine baseDN: %v", err)
	}
	ldapUsername := fmt.Sprintf("cn=read-only-admin,%s", baseDN)
	err = ldapCon.UnauthenticatedBind(ldapUsername)
	if err != nil {
		return fmt.Errorf("can't do unauth bind: %v", err)
	}
	return nil
}

func LDAPAuthenticatedBind(ldapCon *ldap.Conn, username string, password string) error {
	simpleBindRequest := ldap.NewSimpleBindRequest(username, password, []ldap.Control{})
	_, err := ldapCon.SimpleBind(simpleBindRequest)
	if err != nil {
		return fmt.Errorf("can't do auth bind: %v", err)
	}
	return nil
}

func LDAPQuery(ldapCon *ldap.Conn, baseDN string, filter string, attributes []string) (*ldap.SearchResult, error) {
	searchReq := ldap.NewSearchRequest(baseDN, ldap.ScopeWholeSubtree, 0, 0, 0, false, filter, attributes, nil)
	resp, err := ldapCon.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("something wrong with your search: %v", err)
	}
	return resp, nil
}
