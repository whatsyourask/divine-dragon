// https://ldap.com/the-ldap-search-operation/
// https://cybernetist.com/2020/05/18/getting-started-with-go-ldap/

package transport

import (
	"errors"
	"fmt"
	"strings"

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
	sld, tld, err := constructBaseDN(domain)
	if err != nil {
		return fmt.Errorf("can't create a base DN for LDAP Authentication: %v", err)
	}
	ldapUsername := fmt.Sprintf("cn=read-only-admin,dc=%s,dc=%s", sld, tld)
	err = ldapCon.UnauthenticatedBind(ldapUsername)
	if err != nil {
		return fmt.Errorf("can't do unauth bind: %v", err)
	}
	return nil
}

func constructBaseDN(domain string) (string, string, error) {
	domainName := strings.Split(domain, ".")
	if len(domainName) != 2 {
		return "", "", errors.New("invalid domain name")
	}
	tld := domainName[1]
	sld := domainName[0]
	return sld, tld, nil
}

func LDAPAuthenticatedBind(ldapCon *ldap.Conn, domain string, username string, password string) error {
	sld, tld, err := constructBaseDN(domain)
	if err != nil {
		return fmt.Errorf("can't create a base DN for LDAP Authentication: %v", err)
	}
	ldapUsername := fmt.Sprintf("cn=%s,dc=%s,dc=%s", username, sld, tld)
	err = ldapCon.Bind(ldapUsername, password)
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
