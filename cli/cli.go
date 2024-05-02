package cli

import (
	"bufio"
	"divine-dragon/c2"
	"divine-dragon/payload_generator"
	"divine-dragon/remote_enum"
	"divine-dragon/remote_exploit"
	"divine-dragon/transport"
	"divine-dragon/util"
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
	generalCommandsMethods   map[string]func()
	label                    string
	selectedModule           ModuleSettings
	modulesSettings          []ModuleSettings
	useFormatCommand         string
	infoGeneralFormatCommand string
	backModuleCommand        string
	infoModuleFormatCommand  string
	showModuleOptionsCommand string
	setOptionsFormatCommand  string
	runModuleCommand         string
	moduleGeneralCommand     []string
	c2m                      *c2.C2Module
	agents                   []transport.Agent
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
				"HOST":            "",
				"PORT":            "4444",
				"SHELL_TYPE":      "",
				"PLATFORM":        "windows",
				"ARCH":            "amd64",
				"EXECUTABLE_NAME": "payload.exe",
			},
			Run: tcli.runPayloadGeneratorModule,
		},
		{
			Name: "c2",
			Info: "Module to start C2 server.",
			Options: map[string]string{
				"HOST": "127.0.0.1",
				"PORT": "8888",
			},
			Run: tcli.runC2Module,
		},
	}
	// fmt.Println(tcli.modulesSettings)
	tcli.generalCommandsMethods = map[string]func(){
		"list modules": tcli.listModules,
		"quit":         tcli.exit,
		"exit":         tcli.exit,
		"sessions":     tcli.checkSessions,
	}
	tcli.useFormatCommand = "use %s"
	tcli.infoGeneralFormatCommand = "show info %s"
	tcli.moduleGeneralCommand = []string{"back", "info", "show options", "run", "quit", "exit", "sessions"}
	// tcli.backModuleCommand = "back"
	// tcli.infoModuleFormatCommand = "info"
	// tcli.showModuleOptionsCommand = "show options"
	tcli.setOptionsFormatCommand = "set %s %s"
	// tcli.runModuleCommand = "run"
	tcli.c2m = nil
	tcli.agents = nil
	return &tcli, nil
}

func (tcli *ToolCommandLineInterface) Run() {
	tcli.printLogo()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(tcli.label)
	for scanner.Scan() {
		command := scanner.Text()
		formattedCommand := util.FormatCommand(command)
		err := tcli.validateGeneralCommand(formattedCommand)
		if err == nil {
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
				for _, moduleSettings := range tcli.modulesSettings {
					if formattedCommand == fmt.Sprintf(tcli.infoGeneralFormatCommand, moduleSettings.Name) {
						splitedCommand := strings.Split(command, " ")
						moduleName := splitedCommand[2]
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
		formattedCommand := util.FormatCommand(command)
		err := tcli.validateModuleCommand(formattedCommand)
		if err == nil {
			if strings.Contains(formattedCommand, "set ") {
				optionValue := strings.Split(formattedCommand, " ")
				option := optionValue[1]
				value := optionValue[2]
				tcli.setModuleOption(option, value)
			}
			if strings.Contains(formattedCommand, "back") {
				tcli.label = defaultLabel + ">>> "
				tcli.selectedModule = ModuleSettings{}
				break
			}
			if strings.Contains(formattedCommand, "show options") {
				tcli.showModuleOptions()
			}
			if strings.Contains(formattedCommand, "run") {
				tcli.runModule()
			}
			if strings.Contains(formattedCommand, "exit") || strings.Contains(formattedCommand, "quit") {
				tcli.exit()
			}
			if strings.Contains(formattedCommand, "info") {
				tcli.showInfo(tcli.selectedModule.Name)
			}
			if strings.Contains(formattedCommand, "sessions") {
				tcli.checkSessions()
			}
		}
		fmt.Print(tcli.label)
	}
}

func (tcli *ToolCommandLineInterface) validateModuleCommand(command string) error {
	for _, moduleCommand := range tcli.moduleGeneralCommand {
		if command == moduleCommand {
			return nil
		}
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

func (tcli *ToolCommandLineInterface) runPayloadGeneratorModule() {
	var allSet bool = true
	for _, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" {
			allSet = false
			break
		}
	}
	if allSet {
		pg := payload_generator.NewPayloadGeneratorModule(
			tcli.selectedModule.Options["HOST"],
			tcli.selectedModule.Options["PORT"],
			tcli.selectedModule.Options["SHELL_TYPE"],
			tcli.selectedModule.Options["PLATFORM"],
			tcli.selectedModule.Options["ARCH"],
			tcli.selectedModule.Options["EXECUTABLE_NAME"],
		)
		// fmt.Println("Running")
		pg.Run()
	}
}

func (tcli *ToolCommandLineInterface) runC2Module() {
	var allSet bool = true
	for _, moduleValue := range tcli.selectedModule.Options {
		if moduleValue == "" {
			allSet = false
			break
		}
	}
	if allSet {
		c2m := c2.NewC2Module(
			tcli.selectedModule.Options["HOST"],
			tcli.selectedModule.Options["PORT"],
		)
		c2m.Run()
		tcli.c2m = c2m
	}
}

func (tcli *ToolCommandLineInterface) checkSessions() {
	if tcli.c2m != nil {
		agents := tcli.c2m.GetAgents()
		tcli.agents = agents
		out := ""
		if len(agents) > 0 {
			for _, agent := range agents {
				out += fmt.Sprintf("%16s| %15s| %s", agent.Uuid, agent.Hostname, agent.Username)
			}
			fmt.Println()
			fmt.Println("\tActive agent sessions")
			fmt.Println("=================================================")
			fmt.Println("  Agent UUID    |     Hostname   |     Username    ")
			fmt.Println(out)
			fmt.Println()
		} else {
			fmt.Println()
			fmt.Println("C2 server instance is active. But nobody is connected yet.")
			fmt.Println()
		}
	} else {
		fmt.Println()
		fmt.Println("No active C2 server instances. So, you don't have any sessions.")
		fmt.Println()
	}
}
