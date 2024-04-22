package main

import (
	"divine-dragon/cli"
)

func main() {
	// asrep_module := remote_exploit.NewASREPRoastingModule("htb.local", "10.129.95.210", false, false, true, "svc-alfresco.txt", "remote_exploit/names2.txt", "", 50, 0)
	// asrep_module.Run()

	// kerberos_enum_module := remote_enum.NewKerberosEnumUsersModule("htb.local", "10.129.95.210", false, true, false, "remote_exploit/names.txt", "", 50, 0)
	// kerberos_enum_module.Run()

	// ldap_enum_module := remote_enum.NewLdapEnumModule("intelligence.htb", "10.129.15.150", "389", "Tiffany.Molina@intelligence.htb", "NewIntelligenceCorpUser9876", "", true, "")
	// ldap_enum_module.Run()

	// kerberoasting_module := remote_exploit.NewKerberoastingModule("active.htb", "10.129.15.134", "SVC_TGS", "GPPstillStandingStrong2k18", false, false, "administrator.txt", true, "")
	// kerberoasting_module.Run()

	// smb_enum_module := remote_enum.NewSmbModuleNewLdapEnumModule("timelapse.htb", "10.129.227.113", "445", "guest", "", "", true, "")
	// smb_enum_module.Run()

	// kerberosPasswordSprayingModule := remote_exploit.NewKerberosSprayingModule("intelligence.htb", "10.129.15.150", true, false, false, "users.txt", "NewIntelligenceCorpUser9876", "", 50, 0)
	// kerberosPasswordSprayingModule.Run()

	// smbPasswordSprayModule := remote_exploit.NewSmbSprayModule("intelligence.htb", "10.129.15.150", "445", "users.txt", "NewIntelligenceCorpUser9876", true, "", 50, 0)
	// smbPasswordSprayModule.Run()
	tcli, _ := cli.NewToolCommandLineInterface()
	tcli.Run()
}
