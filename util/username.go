package util

import (
	"errors"
	"strings"
)

func FormatUsername(username string) (user string, err error) {
	if username == "" {
		return "", errors.New("bad username: blank")
	}
	parts := strings.Split(username, "@")
	if len(parts) > 2 {
		return "", errors.New("bad username: too many @ signs")
	}
	return parts[0], nil
}
