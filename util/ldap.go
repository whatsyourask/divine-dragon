package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap"
)

func ConstructBaseDN(domain string) (string, error) {
	domainName := strings.Split(domain, ".")
	if len(domainName) != 2 {
		return "", errors.New("invalid domain name")
	}
	tld := domainName[1]
	sld := domainName[0]
	return fmt.Sprintf("dc=%s,dc=%s", sld, tld), nil
}

func ConstructLDAPFilter(template string, values []any) string {
	filter := fmt.Sprintf(template, values...)
	return filter
}

func LDAPListObjectsInResult(resp *ldap.SearchResult) {
	for _, entry := range resp.Entries {
		fmt.Printf("%s\n", entry.DN)
	}
	fmt.Println()
}

func LDAPGetEntryAttributes(resp *ldap.SearchResult) {
	for _, entry := range resp.Entries {
		for _, attribute := range entry.Attributes {
			fmt.Printf("%s: %v\n", attribute.Name, attribute.Values)
		}
	}
}
