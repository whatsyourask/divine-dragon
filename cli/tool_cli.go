package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ModuleSettings struct {
	Name    string
	Options []string
	Info    string
	Module  any
	Run     func()
}

const defaultLabel = "(divine-dragon) / "

type ToolCommandLineInterface struct {
	generalCommandsMethods map[string]func()
	label                  string
	selectedModule         string
	commandBuffer          []string
	// moduleCommandsMethods  map[string]func()
	modulesSettings   []ModuleSettings
	useFormatCommand  string
	infoFormatCommand string
	// setModuleOptionsCommandsMethods map[string]func()
}

func NewToolCommandLineInterface() (*ToolCommandLineInterface, error) {
	tcli := ToolCommandLineInterface{}
	tcli.label = defaultLabel + ">>> "
	tcli.modulesSettings = []ModuleSettings{
		{
			Name: "remote_enum/kerberos_enumusers",
			Info: "Module to enumerate AD users via Kerberos.",
			Options: []string{
				"DOMAIN",
				"DC",
				"VERBOSE",
				"SAFE_MODE",
				"DOWNGRADE",
				"USERNAMES_LIST",
				"LOG_FILE",
				"THREADS",
				"DELAY",
			},
		},
		{
			Name: "remote_enum/smb_enum",
			Info: "Module to enumerate SMB shares and download all their content.",
			Options: []string{
				"DOMAIN",
				"REMOTE_HOST",
				"REMOTE_PORT",
				"USERNAME",
				"PASSWORD",
				"HASH",
				"VERBOSE",
				"LOG_FILE",
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
	tcli.infoFormatCommand = "show info %s"
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
		formatedCommand := tcli.formatCommand()
		err := tcli.validateGeneralCommand(formatedCommand)
		if err == nil {
			// fmt.Println("formated command: " + formatedCommand)
			if strings.Contains(formatedCommand, "use") {
				for _, moduleSettings := range tcli.modulesSettings {
					if formatedCommand == fmt.Sprintf(tcli.useFormatCommand, moduleSettings.Name) {
						splitedCommand := strings.Split(command, " ")
						tcli.selectedModule = splitedCommand[1]
						tcli.label = defaultLabel + "(" + tcli.selectedModule + ") >>> "
						tcli.configModule()
					}
				}
			} else if strings.Contains(formatedCommand, "show info") {
				fmt.Println("hello")
				for _, moduleSettings := range tcli.modulesSettings {
					if formatedCommand == fmt.Sprintf(tcli.infoFormatCommand, moduleSettings.Name) {
						splitedCommand := strings.Split(command, " ")
						moduleName := splitedCommand[2]
						fmt.Println(moduleName)
						tcli.showInfo(moduleName)
					}
				}
			} else {
				for c, tcliFunc := range tcli.generalCommandsMethods {
					if formatedCommand == c {
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

func (tcli *ToolCommandLineInterface) formatCommand(command string) {
	formattedCommand := strings.Trim(command, " ")
	formattedCommandSlice := strings.Split(formattedCommand, " ")
}

func (tcli *ToolCommandLineInterface) validateGeneralCommand(command string) error {
	for generalCommandMethod, _ := range tcli.generalCommandsMethods {
		if command == generalCommandMethod {
			return nil
		}
	}
	for _, formatCommand := range []string{tcli.infoFormatCommand, tcli.useFormatCommand} {
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

// func (tcli *ToolCommandLineInterface) useModule() {
// }

// func (tcli *ToolCommandLineInterface) validateModuleCommand(command string) error {
// 	fmt.Println("Command 1: " + command)
// 	formatedCommand := strings.Trim(command, " ")
// 	for _, availableCommand := range availableModuleCommands {
// 		if formatedCommand == availableCommand {
// 			return nil
// 		}
// 	}
// 	fmt.Println("Command 2: " + formatedCommand)
// 	for _, availableModuleFormatCommand := range availableModuleFormatCommands {
// 		for _, availableSetOption := range availableModulesSetOptions[tcli.selectedModule] {
// 			splitedCommand := strings.Split(formatedCommand, " ")
// 			if len(splitedCommand) < 3 {
// 				return fmt.Errorf("Command's not recognized.\n")
// 			}
// 			setValue := splitedCommand[2]
// 			fmt.Println(fmt.Sprintf(availableModuleFormatCommand, availableSetOption, setValue))
// 			if formatedCommand == fmt.Sprintf(availableModuleFormatCommand, availableSetOption, setValue) {
// 				fmt.Println("ok")
// 				return nil
// 			}
// 		}
// 	}
// 	fmt.Println("not ok")
// 	return fmt.Errorf("Command's not recognized.\n")
// }

func (tcli *ToolCommandLineInterface) configModule() {
	// fmt.Print(tcli.label)
	// scanner := bufio.NewScanner(os.Stdin)
	// for scanner.Scan() {
	// 	command := scanner.Text()
	// 	formatedCommand := strings.Trim(command, " ")
	// 	err := tcli.validateModuleCommand(formatedCommand)
	// 	if err == nil {
	// 		if strings.Contains(formatedCommand, "set") {
	// 			break
	// 		}
	// 	}
	// 	fmt.Print(tcli.label)
	// }
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
