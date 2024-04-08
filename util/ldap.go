package util

import (
	"fmt"

	"github.com/go-ldap/ldap"
)

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
