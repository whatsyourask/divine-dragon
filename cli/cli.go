package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ModuleSettings struct {
	Name    string
	Options map[string]string
	Info    string
	Module  any
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
				"VERBOSE":        "",
				"SAFE_MODE":      "",
				"DOWNGRADE":      "",
				"USERNAMES_LIST": "",
				"LOG_FILE":       "",
				"THREADS":        "",
				"DELAY":          "",
			},
		},
		{
			Name: "remote_enum/smb_enum",
			Info: "Module to enumerate SMB shares and download all their content.",
			Options: map[string]string{
				"DOMAIN":      "",
				"REMOTE_HOST": "",
				"REMOTE_PORT": "",
				"USERNAME":    "",
				"PASSWORD":    "",
				"HASH":        "",
				"VERBOSE":     "",
				"LOG_FILE":    "",
			},
		},
	}
	fmt.Println(tcli.modulesSettings)
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
						fmt.Println(moduleName)
						tcli.showInfo(moduleName)
					}
				}
			} else {
				for c, tcliFunc := range tcli.generalCommandsMethods {
					if formattedCommand == c {
						tcliFunc()
						break
					}
				}
			}
		} else {
			fmt.Println(command, err)
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
			fmt.Println(fmt.Sprintf(formatCommand, moduleSettings.Name))
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
			fmt.Println(formattedCommand)
			if strings.Contains(formattedCommand, "set ") {
				optionValue := strings.Split(formattedCommand, " ")
				option := optionValue[1]
				value := optionValue[2]
				tcli.setModuleOption(option, value)
				fmt.Println(tcli.selectedModule.Options)
			}
			if strings.Contains(formattedCommand, "back") {
				tcli.label = defaultLabel + ">>> "
				tcli.selectedModule = ModuleSettings{}
				break
			}
			if strings.Contains(formattedCommand, "show options") {
				fmt.Println("here")
				tcli.showModuleOptions()
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
	commandSlice := strings.Split(command, " ")
	if len(commandSlice) > 2 {
		commandValue := commandSlice[2]
		for moduleOption, _ := range tcli.selectedModule.Options {
			fmt.Println(fmt.Sprintf(tcli.setOptionsFormatCommand, moduleOption, commandValue))
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
	fmt.Println(tcli.selectedModule.Options)
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

// TODO:
// 1. Code a function to Run a module with set parameters.
