package ldapclient

import (
	// "context"
	"crypto/tls"

	"fmt"
	"log"

	// "os/user"

	// "net"
	"testing"

	"github.com/go-ldap/ldap/v3"
)

func TestConnect(t *testing.T) {
	addr := "<Ad serverIP>:636"
	// d := net.Dialer{Timeout: ldap.DefaultTimeout}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tcpcn, err := ldap.DialTLS("tcp", addr, tlsConfig)
	if err != nil {
		t.Log(err)
	}
	t.Log(tcpcn)
	// IsTLS := false
	// if IsTLS {
	// 	tlscn, err := tls.DialWithDialer(&d, "tcp", addr, nil)
	// 	if err != nil {
	// 		t.Log(err)
	// 	}
	// 	tcpcn = tlscn
	// }

	// ldapcn := ldap.NewConn(tcpcn, IsTLS)

	// ldapcn.Start()
	t.Log("in connect")
	t.Log("Bind started")
	BaseDN := "dc=example,dc=com"
	bindDN := "abc@example.com"
	password := "p@ssw0rd12#"
	t.Log(tcpcn.Bind(bindDN, password))
	if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
		t.Log("invalid credentials")
	}
	t.Log(err)
	user := "abc@example.com"
	query := fmt.Sprintf("(&(|(objectClass=organizationalPerson)(objectClass=inetOrgPerson))"+
		"(|(uid=%[1]s)(mail=%[1]s)(userPrincipalName=%[1]s)(sAMAccountName=%[1]s)))", user)

	req := ldap.NewSearchRequest(BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, query, []string{"dn"}, nil)
	res, err := tcpcn.Search(req)
	if err != nil {
		t.Log(err)
	}

	var entries []map[string]interface{}
	for _, v := range res.Entries {
		entry := map[string]interface{}{"dn": v.DN}
		for _, attr := range v.Attributes {
			// We need the first value only for the named attribute.
			entry[attr.Name] = attr.Values[0]
		}
		entries = append(entries, entry)
	}
	t.Log(entries)
}

func TestOidc(t *testing.T) {
	username := "abc@example.com"
	if username == "" {
		log.Print("username is empty")
	}

	addr := "<Ad serveIP>:636"
	// d := net.Dialer{Timeout: ldap.DefaultTimeout}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tcpcn, err := ldap.DialTLS("tcp", addr, tlsConfig)
	if err != nil {
		t.Log(err)
	}
	t.Log(tcpcn)
	AttrClaims := map[string]string{"name": "name", "sn": "family_name", "givenName": "given_name", "mail": "email"}
	// We need to find LDAP attribute's names for all required claims.
	attrs := []string{"dn"}
	for k := range AttrClaims {
		attrs = append(attrs, k)
	}
	t.Log(attrs)

	// Find the attributes in the LDAP server.

	t.Log("Bind started")
	BaseDN := "dc=example,dc=com"
	bindDN := "abc@example.com"
	password := "p@ssw0rd12#"
	t.Log(tcpcn.Bind(bindDN, password))
	if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
		t.Log("invalid credentials")
	}
	t.Log(err)
	user := username
	query := fmt.Sprintf("(&(|(objectClass=organizationalPerson)(objectClass=inetOrgPerson))"+
		"(|(uid=%[1]s)(mail=%[1]s)(userPrincipalName=%[1]s)(sAMAccountName=%[1]s)))", user)

	req := ldap.NewSearchRequest(BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, query, []string{"dn"}, nil)
	res, err := tcpcn.Search(req)
	if err != nil {
		t.Log(err)
	}

	var entries []map[string]interface{}
	for _, v := range res.Entries {
		entry := map[string]interface{}{"dn": v.DN}
		for _, attr := range v.Attributes {
			// We need the first value only for the named attribute.
			entry[attr.Name] = attr.Values[0]
		}
		entries = append(entries, entry)
	}
	t.Log(entries)
	var (
		entry   = entries[0]
		details = make(map[string]interface{})
	)
	for _, attr := range attrs {
		t.Log(attr)
		if v, ok := entry[attr]; ok {
			details[attr] = v
		}
	}
	t.Log(details)

	// Transform the retrieved attributes to corresponding claims.
	claims := make(map[string]interface{})
	for attr, v := range details {
		t.Log(AttrClaims[attr])
		if claim, ok := AttrClaims[attr]; ok {
			claims[claim] = v
		}
	}
	t.Log(claims)

	query2 := fmt.Sprintf("(|"+"(&(|(objectClass=group)(objectClass=groupOfNames))(member=%[1]s))"+"(&(objectClass=groupOfUniqueNames)(uniqueMember=%[1]s))"+")", user)

	req2 := ldap.NewSearchRequest(BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, query2, []string{"dn"}, nil)
	res2, err := tcpcn.Search(req2)
	if err != nil {
		t.Log(err)
	}

	var entries2 []map[string]interface{}
	for _, v := range res2.Entries {
		entry := map[string]interface{}{"dn": v.DN}
		for _, attr := range v.Attributes {
			// We need the first value only for the named attribute.
			entry[attr.Name] = attr.Values[0]
		}
		entries2 = append(entries2, entry)
	}
	t.Log(entries2)
}
