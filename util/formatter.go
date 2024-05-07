package util

import "strings"

func FormatCommand(command string) string {
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
