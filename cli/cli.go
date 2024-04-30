package cli

import (
	"bufio"
	"divine-dragon/remote_enum"
	"divine-dragon/remote_exploit"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ModuleSettings struct {
	Name    string
	Options map[string]string
	Info    string
	Run     func()
}

const defaultLabel = "(divine-dragon) / "

type ToolCommandLineInterface struct {
	generalCommandsMethods map[string]func()
	label                  string
	selectedModule         ModuleSettings
	// commandBuffer          []string
	// moduleCommandsMethods  map[string]func()
	modulesSettings          []ModuleSettings
	useFormatCommand         string
	infoGeneralFormatCommand string
	backModuleCommand        string
	infoModuleFormatCommand  string
	showModuleOptionsCommand string
	setOptionsFormatCommand  string
	runModuleCommand         string
	// setModuleOptionsCommandsMethods map[string]func()
}

func NewToolCommandLineInterface() (*ToolCommandLineInterface, error) {
	tcli := ToolCommandLineInterface{}
	tcli.label = defaultLabel + ">>> "
	tcli.modulesSettings = []ModuleSettings{
		{
			Name: "remote_enum/kerberos_enumusers",
			Info: "Module to enumerate AD users via Kerberos.",
			Options: map[string]string{
				"DOMAIN":         "",
				"DC":             "",
				"VERBOSE":        "false",
				"SAFE_MODE":      "false",
				"DOWNGRADE":      "true",
				"USERNAMES_LIST": "",
				"LOG_FILE":       "",
				"THREADS":        "10",
				"DELAY":          "0",
			},
			Run: tcli.runKerberosEnumUsersModule,
		},
		{
			Name: "remote_enum/smb_enum",
			Info: "Module to enumerate SMB shares and download all their content.",
			Options: map[string]string{
				"DOMAIN":      "",
				"REMOTE_HOST": "",
				"REMOTE_PORT": "445",
				"USERNAME":    "guest",
				"PASSWORD":    "",
				"HASH":        "",
				"VERBOSE":     "false",
				"LOG_FILE":    "",
			},
			Run: tcli.runSmbEnumModule,
		},
		{
			Name: "remote_enum/ldap_enum",
			Info: "Module to enumerate LDAP for group memberships, users, computers, and potential flaws.",
			Options: map[string]string{
				"DOMAIN":      "",
				"REMOTE_HOST": "",
				"REMOTE_PORT": "389",
				"USERNAME":    "",
				"PASSWORD":    "",
				"BASE_DN":     "",
				"VERBOSE":     "false",
				"LOG_FILE":    "",
			},
			Run: tcli.runLdapEnumModule,
		},
		{
			Name: "remote_exploit/asreproasting",
			Info: "Module to execute ASREPRoasting attack against DC with usernames list.",
			Options: map[string]string{
				"DOMAIN":         "",
				"DC":             "",
				"VERBOSE":        "false",
				"SAFE_MODE":      "false",
				"DOWNGRADE":      "true",
				"USERNAMES_LIST": "",
				"HASH_FILE":      "",
				"LOG_FILE":       "",
				"THREADS":        "10",
				"DELAY":          "0",
			},
			Run: tcli.runASREPRoastingModule,
		},
		{
			Name: "remote_exploit/kerberoasting",
			Info: "Module to execute Kerberoasting attack against DC with username and password.",
			Options: map[string]string{
				"DOMAIN":    "",
				"DC":        "",
				"VERBOSE":   "false",
				"SAFE_MODE": "false",
				"DOWNGRADE": "true",
				"USERNAME":  "",
				"PASSWORD":  "",
				"HASH_FILE": "",
				"LOG_FILE":  "",
			},
			Run: tcli.runKerberoastingModule,
		},
		{
			Name: "remote_exploit/kerberos_password_spraying",
			Info: "Module to execute Password Spraying attack against domain machine with given usernames list and a single password via Kerberos protocol.",
			Options: map[string]string{
				"DOMAIN":         "",
				"DC":             "",
				"VERBOSE":        "false",
				"SAFE_MODE":      "false",
				"DOWNGRADE":      "true",
				"USERNAMES_LIST": "",
				"PASSWORD":       "",
				"LOG_FILE":       "",
				"THREADS":        "10",
				"DELAY":          "0",
			},
			Run: tcli.runKerberosPasswordSprayingModule,
		},
		{
			Name: "remote_exploit/smb_password_spraying",
			Info: "Module to execute Password Spraying attack against domain machine with given usernames list and a single password via SMB protocol.",
			Options: map[string]string{
				"DOMAIN":         "",
				"REMOTE_HOST":    "",
				"REMOTE_PORT":    "445",
				"USERNAMES_LIST": "",
				"PASSWORD":       "",
				"VERBOSE":        "false",
				"LOG_FILE":       "",
				"THREADS":        "10",
				"DELAY":          "0",
			},
			Run: tcli.runSmbPasswordSprayingModule,
		},
		{
			Name: "payload_generator/payload_generator",
			Info: "Module to generate Bind/Reverse shell payload in go binary.",
			Options: map[string]string{
				"DOMAIN":         "",
				"REMOTE_HOST":    "",
				"REMOTE_PORT":    "445",
				"USERNAMES_LIST": "",
				"PASSWORD":       "",
				"VERBOSE":        "false",
				"LOG_FILE":       "",
				"THREADS":        "10",
				"DELAY":          "0",
			},
			Run: tcli.runStageOnePayloadGenerator,
		},
	}
	}
	// fmt.Println(tcli.modulesSettings)
	tcli.generalCommandsMethods = map[string]func(){
		"list modules": tcli.listModules,
		"quit":         tcli.exit,
		"exit":         tcli.exit,
	}
	tcli.useFormatCommand = "use %s"
	tcli.infoGeneralFormatCommand = "show info %s"
	tcli.backModuleCommand = "back"
	tcli.infoModuleFormatCommand = "info"
	tcli.showModuleOptionsCommand = "show options"
	tcli.setOptionsFormatCommand = "set %s %s"
	tcli.runModuleCommand = "run"
	return &tcli, nil
}

func (tcli *ToolCommandLineInterface) Run() {
	tcli.printLogo()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(tcli.label)
	for scanner.Scan() {
		// // fmt.Print(tcli.label)
		// // fmt.Scanln(&command)
		command := scanner.Text()
		// fmt.Println("unformated command: ", command)
		formattedCommand := tcli.formatCommand(command)
		err := tcli.validateGeneralCommand(formattedCommand)
		if err == nil {
			// fmt.Println("formated command: " + formatedCommand)
			if strings.Contains(formattedCommand, "use ") {
				for _, moduleSettings := range tcli.modulesSettings {
					if formattedCommand == fmt.Sprintf(tcli.useFormatCommand, moduleSettings.Name) {
						splitedCommand := strings.Split(command, " ")
						for _, moduleSettings := range tcli.modulesSettings {
							if splitedCommand[1] == moduleSettings.Name {
								tcli.selectedModule = moduleSettings
							}
						}
						tcli.label = defaultLabel + "(" + tcli.selectedModule.Name + ") >>> "
						tcli.configModule()
					}
				}
			} else if strings.Contains(formattedCommand, "show info ") {
				// fmt.Println("hello")
				for _, moduleSettings := range tcli.modulesSettings {
					if formattedCommand == fmt.Sprintf(tcli.infoGeneralFormatCommand, moduleSettings.Name) {
						splitedCommand := strings.Split(command, " ")
						moduleName := splitedCommand[2]
						// fmt.Println(moduleName)
						tcli.showInfo(moduleName)
					}
				}
			} else {
				for c, tcliFunc := range tcli.generalCommandsMethods {
					if formattedCommand == c {
						tcliFunc()
					}
				}
			}
		}
		fmt.Print(tcli.label)
	}
}

func (tcli *ToolCommandLineInterface) printLogo() {
	fmt.Println()
	fmt.Println(`######  ### #     # ### #     # #######    ######  ######     #     #####  ####### #     #
#     #  #  #     #  #  ##    # #          #     # #     #   # #   #     # #     # ##    # 
#     #  #  #     #  #  # #   # #          #     # #     #  #   #  #       #     # # #   # 
#     #  #  #     #  #  #  #  # #####      #     # ######  #     # #  #### #     # #  #  # 
#     #  #   #   #   #  #   # # #          #     # #   #   ####### #     # #     # #   # # 
#     #  #    # #    #  #    ## #          #     # #    #  #     # #     # #     # #    ## 
######  ###    #    ### #     # #######    ######  #     # #     #  #####  ####### #     #`)
	fmt.Println("==========================================================================================")
}

func (tcli *ToolCommandLineInterface) formatCommand(command string) string {
	formattedCommand := strings.Trim(command, " ")
	formattedCommandSlice := strings.Split(formattedCommand, " ")
	newSlice := []string{}
	for _, part := range formattedCommandSlice {
		if part != " " {
			newSlice = append(newSlice, part)
		}
	}
	return strings.Join(newSlice, " ")
}

func (tcli *ToolCommandLineInterface) validateGeneralCommand(command string) error {
	for generalCommandMethod, _ := range tcli.generalCommandsMethods {
		if command == generalCommandMethod {
			return nil
		}
	}
	for _, formatCommand := range []string{tcli.infoGeneralFormatCommand, tcli.useFormatCommand} {
		for _, moduleSettings := range tcli.modulesSettings {
			// fmt.Println(fmt.Sprintf(formatCommand, moduleSettings.Name))
			if command == fmt.Sprintf(formatCommand, moduleSettings.Name) {
				return nil
			}
		}
	}
	return fmt.Errorf("Command's not recognized.\n")
}

func (tcli *ToolCommandLineInterface) listModules() {
	fmt.Println()
	fmt.Println("Available Modules:")
	fmt.Println("==================")
	for _, moduleSettings := range tcli.modulesSettings {
		fmt.Println(moduleSettings.Name)
	}
	fmt.Println()
}

func (tcli *ToolCommandLineInterface) exit() {
	os.Exit(0)
}

func (tcli *ToolCommandLineInterface) showInfo(moduleName string) {
	fmt.Println()
	fmt.Println("Module information")
	fmt.Println("------------------")
	for _, moduleSettings := range tcli.modulesSettings {
		if moduleName == moduleSettings.Name {
			fmt.Println(moduleSettings.Info)
		}
	}
	fmt.Println("------------------")
	fmt.Println()
}

func (tcli *ToolCommandLineInterface) configModule() {
	fmt.Print(tcli.label)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		formattedCommand := tcli.formatCommand(command)
		err := tcli.validateModuleCommand(formattedCommand)
		// fmt.Println(err.Error())
		if err == nil {
			// fmt.Println(formattedCommand)
			if strings.Contains(formattedCommand, "set ") {
				optionValue := strings.Split(formattedCommand, " ")
				option := optionValue[1]
				value := optionValue[2]
				tcli.setModuleOption(option, value)
				// fmt.Println(tcli.selectedModule.Options)
			}
			if strings.Contains(formattedCommand, "back") {
				tcli.label = defaultLabel + ">>> "
				tcli.selectedModule = ModuleSettings{}
				break
			}
			if strings.Contains(formattedCommand, "show options") {
				// fmt.Println("here")
				tcli.showModuleOptions()
			}
			if strings.Contains(formattedCommand, "run") {
				tcli.runModule()
			}
		}
		fmt.Print(tcli.label)
	}
}

func (tcli *ToolCommandLineInterface) validateModuleCommand(command string) error {
	if command == tcli.backModuleCommand {
		return nil
	}
	if command == tcli.infoModuleFormatCommand {
		return nil
	}
	if command == tcli.showModuleOptionsCommand {
		return nil
	}
	if command == tcli.runModuleCommand {
		return nil
	}
	commandSlice := strings.Split(command, " ")
	if len(commandSlice) > 2 {
		commandValue := commandSlice[2]
		for moduleOption, _ := range tcli.selectedModule.Options {
			// fmt.Println(fmt.Sprintf(tcli.setOptionsFormatCommand, moduleOption, commandValue))
			if command == fmt.Sprintf(tcli.setOptionsFormatCommand, moduleOption, commandValue) {
				return nil
			}
		}
	}
	return fmt.Errorf("Command's not recognized.\n")
}

func (tcli *ToolCommandLineInterface) setModuleOption(option string, value string) {
	tcli.selectedModule.Options[option] = value
}

func (tcli *ToolCommandLineInterface) showModuleOptions() {
	// fmt.Println(tcli.selectedModule.Options)
	fmt.Println()
	fmt.Printf("%15s: %15s\n\n", "OPTION", "VALUE")
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		var moduleValueStr string
		if moduleValue != "" {
			moduleValueStr = moduleValue
		} else {
			moduleValueStr = ""
		}
		fmt.Printf("%15s: %15s\n\n", moduleOption, moduleValueStr)
	}
}

func (tcli *ToolCommandLineInterface) runModule() {
	tcli.selectedModule.Run()
}

func (tcli *ToolCommandLineInterface) runKerberosEnumUsersModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" {
			allSet = false
			break
		}
	}
	fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		safemode, _ := strconv.ParseBool(tcli.selectedModule.Options["SAFE_MODE"])
		downgrade, _ := strconv.ParseBool(tcli.selectedModule.Options["DOWNGRADE"])
		threads, _ := strconv.ParseInt(tcli.selectedModule.Options["THREADS"], 10, 32)
		delay, _ := strconv.ParseInt(tcli.selectedModule.Options["DELAY"], 10, 32)
		keum := remote_enum.NewKerberosEnumUsersModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["DC"],
			verbose,
			safemode,
			downgrade,
			tcli.selectedModule.Options["USERNAMES_LIST"],
			tcli.selectedModule.Options["LOG_FILE"],
			int(threads),
			int(delay),
		)
		fmt.Println("Running")
		keum.Run()
	}
}

func (tcli *ToolCommandLineInterface) runSmbEnumModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" && moduleOption != "PASSWORD" && moduleOption != "HASH" {
			allSet = false
			break
		}
	}
	fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		sem := remote_enum.NewSmbEnumModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["REMOTE_HOST"],
			tcli.selectedModule.Options["REMOTE_PORT"],
			tcli.selectedModule.Options["USERNAME"],
			tcli.selectedModule.Options["PASSWORD"],
			tcli.selectedModule.Options["HASH"],
			verbose,
			tcli.selectedModule.Options["LOG_FILE"],
		)
		fmt.Println("Running")
		sem.Run()
	}
}

func (tcli *ToolCommandLineInterface) runLdapEnumModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" && moduleOption != "USERNAME" && moduleOption != "PASSWORD" && moduleOption != "BASE_DN" {
			allSet = false
			break
		}
	}
	fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		lem := remote_enum.NewLdapEnumModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["REMOTE_HOST"],
			tcli.selectedModule.Options["REMOTE_PORT"],
			tcli.selectedModule.Options["USERNAME"],
			tcli.selectedModule.Options["PASSWORD"],
			tcli.selectedModule.Options["BASE_DN"],
			verbose,
			tcli.selectedModule.Options["LOG_FILE"],
		)
		fmt.Println("Running")
		lem.Run()
	}
}

func (tcli *ToolCommandLineInterface) runASREPRoastingModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" && moduleOption != "HASH_FILE" {
			allSet = false
			break
		}
	}
	fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		safemode, _ := strconv.ParseBool(tcli.selectedModule.Options["SAFE_MODE"])
		downgrade, _ := strconv.ParseBool(tcli.selectedModule.Options["DOWNGRADE"])
		threads, _ := strconv.ParseInt(tcli.selectedModule.Options["THREADS"], 10, 32)
		delay, _ := strconv.ParseInt(tcli.selectedModule.Options["DELAY"], 10, 32)
		arm := remote_exploit.NewASREPRoastingModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["DC"],
			verbose,
			safemode,
			downgrade,
			tcli.selectedModule.Options["HASH_FILE"],
			tcli.selectedModule.Options["USERNAMES_LIST"],
			tcli.selectedModule.Options["LOG_FILE"],
			int(threads),
			int(delay),
		)
		fmt.Println("Running")
		arm.Run()
	}
}

func (tcli *ToolCommandLineInterface) runKerberoastingModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" && moduleOption != "HASH_FILE" {
			allSet = false
			break
		}
	}
	// fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		safemode, _ := strconv.ParseBool(tcli.selectedModule.Options["SAFE_MODE"])
		downgrade, _ := strconv.ParseBool(tcli.selectedModule.Options["DOWNGRADE"])
		km := remote_exploit.NewKerberoastingModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["DC"],
			tcli.selectedModule.Options["USERNAME"],
			tcli.selectedModule.Options["PASSWORD"],
			safemode,
			downgrade,
			tcli.selectedModule.Options["HASH_FILE"],
			verbose,
			tcli.selectedModule.Options["LOG_FILE"],
		)
		// fmt.Println("Running")
		km.Run()
	}
}

func (tcli *ToolCommandLineInterface) runKerberosPasswordSprayingModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" {
			allSet = false
			break
		}
	}
	// fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		safemode, _ := strconv.ParseBool(tcli.selectedModule.Options["SAFE_MODE"])
		downgrade, _ := strconv.ParseBool(tcli.selectedModule.Options["DOWNGRADE"])
		threads, _ := strconv.ParseInt(tcli.selectedModule.Options["THREADS"], 10, 32)
		delay, _ := strconv.ParseInt(tcli.selectedModule.Options["DELAY"], 10, 32)
		ksm := remote_exploit.NewKerberosSprayingModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["DC"],
			verbose,
			safemode,
			downgrade,
			tcli.selectedModule.Options["USERNAMES_LIST"],
			tcli.selectedModule.Options["PASSWORD"],
			tcli.selectedModule.Options["LOG_FILE"],
			int(threads),
			int(delay),
		)
		// fmt.Println("Running")
		ksm.Run()
	}
}

func (tcli *ToolCommandLineInterface) runSmbPasswordSprayingModule() {
	var allSet bool = true
	for moduleOption, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" && moduleOption != "LOG_FILE" {
			allSet = false
			break
		}
	}
	// fmt.Println("allSet: ", allSet)
	if allSet {
		verbose, _ := strconv.ParseBool(tcli.selectedModule.Options["VERBOSE"])
		threads, _ := strconv.ParseInt(tcli.selectedModule.Options["THREADS"], 10, 32)
		delay, _ := strconv.ParseInt(tcli.selectedModule.Options["DELAY"], 10, 32)
		ssm := remote_exploit.NewSmbSprayModule(
			tcli.selectedModule.Options["DOMAIN"],
			tcli.selectedModule.Options["REMOTE_HOST"],
			tcli.selectedModule.Options["REMOTE_PORT"],
			tcli.selectedModule.Options["USERNAMES_LIST"],
			tcli.selectedModule.Options["PASSWORD"],
			verbose,
			tcli.selectedModule.Options["LOG_FILE"],
			int(threads),
			int(delay),
		)
		// fmt.Println("Running")
		ssm.Run()
	}
}
