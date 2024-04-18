package transport

import (
	"fmt"
	"html/template"
	"strings"

	"divine-dragon/util"

	"github.com/ropnop/gokrb5/v8/iana/errorcode"

	kclient "github.com/ropnop/gokrb5/v8/client"
	kconfig "github.com/ropnop/gokrb5/v8/config"
	"github.com/ropnop/gokrb5/v8/messages"
)

const krb5ConfigTemplateDNS = `[libdefaults]
dns_lookup_kdc = true
default_realm = {{.Realm}}
`

const krb5ConfigTemplateKDC = `[libdefaults]
default_realm = {{.Realm}}
[realms]
{{.Realm}} = {
	kdc = {{.DomainController}}
	admin_server = {{.DomainController}}
}
`

type KerberosSession struct {
	Domain       string
	Realm        string
	Kdcs         map[int]string
	ConfigString string
	Config       *kconfig.Config
	Verbose      bool
	SafeMode     bool
	Logger       *util.Logger
}

type KerberosSessionOptions struct {
	Domain           string
	DomainController string
	Verbose          bool
	SafeMode         bool
	Downgrade        bool
	logger           *util.Logger
}

func NewKerberosSession(options KerberosSessionOptions) (k KerberosSession, err error) {
	if options.Domain == "" {
		return k, fmt.Errorf("domain must not be empty")
	}

	realm := strings.ToUpper(options.Domain)
	configstring := buildKrb5Template(realm, options.DomainController)
	Config, err := kconfig.NewFromString(configstring)
	if options.Downgrade {
		Config.LibDefaults.DefaultTktEnctypeIDs = []int32{23} // downgrade to arcfour-hmac-md5 for crackable AS-REPs
		options.logger.Log.Info("Using downgraded encryption: arcfour-hmac-md5")
	}
	if err != nil {
		panic(err)
	}
	_, kdcs, err := Config.GetKDCs(realm, false)
	if err != nil {
		err = fmt.Errorf("Couldn't find any KDCs for realm %s. Please specify a Domain Controller", realm)
	}
	k = KerberosSession{
		Domain:       options.Domain,
		Realm:        realm,
		Kdcs:         kdcs,
		ConfigString: configstring,
		Config:       Config,
		Verbose:      options.Verbose,
		SafeMode:     options.SafeMode,
		Logger:       options.logger,
	}
	return k, err

}

func buildKrb5Template(realm, domainController string) string {
	data := map[string]interface{}{
		"Realm":            realm,
		"DomainController": domainController,
	}
	var kTemplate string
	if domainController == "" {
		kTemplate = krb5ConfigTemplateDNS
	} else {
		kTemplate = krb5ConfigTemplateKDC
	}
	t := template.Must(template.New("krb5ConfigString").Parse(kTemplate))
	builder := &strings.Builder{}
	if err := t.Execute(builder, data); err != nil {
		panic(err)
	}
	return builder.String()
}

func (k KerberosSession) TestLogin(username, password string) (bool, error) {
	Client := kclient.NewWithPassword(username, k.Realm, password, k.Config, kclient.DisablePAFXFAST(true), kclient.AssumePreAuthentication(true))
	defer Client.Destroy()
	if ok, err := Client.IsConfigured(); !ok {
		return false, err
	}
	err := Client.Login()
	if err == nil {
		return true, err
	}
	success, err := k.TestLoginError(err)
	return success, err
}

func (k KerberosSession) TestUsername(username string) (bool, []byte, error) {
	// client here does NOT assume preauthentication (as opposed to the one in TestLogin)

	cl := kclient.NewWithPassword(username, k.Realm, "foobar", k.Config, kclient.DisablePAFXFAST(true))

	req, err := messages.NewASReqForTGT(cl.Credentials.Domain(), cl.Config, cl.Credentials.CName())
	if err != nil {
		fmt.Printf(err.Error())
	}
	b, err := req.Marshal()
	if err != nil {
		return false, nil, err
	}
	rb, err := cl.SendToKDC(b, k.Realm)

	if err == nil {
		return true, rb, nil
	}
	e, ok := err.(messages.KRBError)
	if !ok {
		return false, nil, err
	}
	switch e.ErrorCode {
	case errorcode.KDC_ERR_PREAUTH_REQUIRED:
		return true, nil, nil
	default:
		return false, nil, err

	}
}

// TestLoginError returns true for certain KRB Errors that only happen when the password is correct
// The correct credentials we're passed, but the error prevented a successful TGT from being retrieved
func (k KerberosSession) TestLoginError(err error) (bool, error) {
	eString := err.Error()
	if strings.Contains(eString, "Password has expired") {
		// user's password expired, but it's valid!
		return true, fmt.Errorf("User's password has expired")
	}
	if strings.Contains(eString, "Clock skew too great") {
		// clock skew off, but that means password worked since PRE-AUTH was successful
		return true, fmt.Errorf("Clock skew is too great")
	}
	return false, err
}

func (k KerberosSession) HandleKerbError(err error) (bool, string) {
	eString := err.Error()

	// handle non KRB errors
	if strings.Contains(eString, "client does not have a username") {
		return true, "Skipping blank username"
	}
	if strings.Contains(eString, "Networking_Error: AS Exchange Error") {
		return false, "NETWORK ERROR - Can't talk to KDC. Aborting..."
	}
	if strings.Contains(eString, " AS_REP is not valid or client password/keytab incorrect") {
		return true, "Got AS-REP (no pre-auth) but couldn't decrypt - bad password"
	}

	// handle KRB errors
	if strings.Contains(eString, "KDC_ERR_WRONG_REALM") {
		return false, "KDC ERROR - Wrong Realm. Try adjusting the domain? Aborting..."
	}
	if strings.Contains(eString, "KDC_ERR_C_PRINCIPAL_UNKNOWN") {
		return true, "User does not exist"
	}
	if strings.Contains(eString, "KDC_ERR_PREAUTH_FAILED") {
		return true, "Invalid password"
	}
	if strings.Contains(eString, "KDC_ERR_CLIENT_REVOKED") {
		if k.SafeMode {
			return false, "USER LOCKED OUT and safe mode on! Aborting..."
		}
		return true, "USER LOCKED OUT"
	}
	if strings.Contains(eString, " AS_REP is not valid or client password/keytab incorrect") {
		return true, "Got AS-REP (no pre-auth) but couldn't decrypt - bad password"
	}
	if strings.Contains(eString, "KRB_AP_ERR_SKEW Clock skew too great") {
		return true, "Clock skew too great"
	}

	return false, eString
}

func (k KerberosSession) ReceiveTGS(username string, password string, spn string) (messages.Ticket, error) {
	Client := kclient.NewWithPassword(username, k.Realm, password, k.Config, kclient.DisablePAFXFAST(true), kclient.AssumePreAuthentication(true))
	defer Client.Destroy()
	if ok, err := Client.IsConfigured(); !ok {
		return messages.Ticket{}, err
	}
	err := Client.Login()
	if err == nil {
		tkt, _, err := Client.GetServiceTicket(spn)
		if err != nil {
			return messages.Ticket{}, fmt.Errorf("can't get TGS: %v", err)
		}
		return tkt, nil
	} else {
		_, err := k.TestLoginError(err)
		return messages.Ticket{}, fmt.Errorf("can't login with given credentials: %v", err)
	}
}
