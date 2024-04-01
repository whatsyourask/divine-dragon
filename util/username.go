package util

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func formatUsername(username string) (user string, err error) {
	if username == "" {
		return "", errors.New("bad username: blank")
	}
	parts := strings.Split(username, "@")
	if len(parts) > 2 {
		return "", errors.New("bad username: too many @ signs")
	}
	return parts[0], nil
}

func GetUsernames(usernameslist string) ([]string, error) {
	if usernameslist != "" {
		file, err := os.Open(usernameslist)
		if err != nil {
			// logger.Log.Error(err.Error())
			return nil, fmt.Errorf("can't open file with usernames: %v", err.Error())
		}
		defer file.Close()
		var usernames []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			usernameLine := scanner.Text()
			username, err := formatUsername(usernameLine)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			usernames = append(usernames, username)
		}
		return usernames, nil
	}
	return nil, errors.New("filename with usernames was not specified")
}
