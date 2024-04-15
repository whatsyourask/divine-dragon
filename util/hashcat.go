package util

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ropnop/gokrb5/v8/messages"
)

func ASRepToHashcat(asrep messages.ASRep) (string, error) {
	return fmt.Sprintf("$krb5asrep$%d$%s@%s:%s$%s",
		asrep.EncPart.EType,
		asrep.CName.PrincipalNameString(),
		asrep.CRealm,
		hex.EncodeToString(asrep.EncPart.Cipher[:16]),
		hex.EncodeToString(asrep.EncPart.Cipher[16:])), nil
}

func TGSEncPartToHashcat(tkt messages.Ticket, username string, spn string) (string, error) {
	if tkt.EncPart.EType == 23 {
		return fmt.Sprintf("$krb5tgs$%d$*%s$%s$%s*$%s$%s",
			tkt.EncPart.EType,
			username,
			tkt.Realm,
			strings.Replace(spn, ":", "~", 1),
			hex.EncodeToString(tkt.EncPart.Cipher[:16]),
			hex.EncodeToString(tkt.EncPart.Cipher[16:]),
		), nil
	} else if tkt.EncPart.EType == 17 {
		return fmt.Sprintf("$krb5tgs$%d$%s$%s$*%s*$%s$%s",
			tkt.EncPart.EType,
			username,
			tkt.Realm,
			strings.Replace(spn, ":", "~", 1),
			hex.EncodeToString(tkt.EncPart.Cipher[len(tkt.EncPart.Cipher)-12:]),
			hex.EncodeToString(tkt.EncPart.Cipher[:len(tkt.EncPart.Cipher)-12]),
		), nil
	} else if tkt.EncPart.EType == 18 {
		return fmt.Sprintf("$krb5tgs$%d$*%s$%s$%s*$%s$%s",
			tkt.EncPart.EType,
			username,
			tkt.Realm,
			strings.Replace(spn, ":", "~", 1),
			hex.EncodeToString(tkt.EncPart.Cipher[len(tkt.EncPart.Cipher)-12:]),
			hex.EncodeToString(tkt.EncPart.Cipher[:len(tkt.EncPart.Cipher)-12]),
		), nil
	} else if tkt.EncPart.EType == 3 {
		return fmt.Sprintf("$krb5tgs$%d$*%s$%s$%s*$%s$%s",
			tkt.EncPart.EType,
			username,
			tkt.Realm,
			strings.Replace(spn, ":", "~", 1),
			hex.EncodeToString(tkt.EncPart.Cipher[:16]),
			hex.EncodeToString(tkt.EncPart.Cipher[16:]),
		), nil
	} else {
		return "", fmt.Errorf("can't get hash of encrypted part of TGS")
	}
}
