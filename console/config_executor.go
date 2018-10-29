package console

import (
	"fmt"
	"strings"
)

var (
	jInstance = ""
	jUser     = ""
	jPswd     = ""
)

func init() {
	updateJiraClient()
}

func handleConfigCmd(args []string) {
	fmt.Println(strings.Join(args, " "))
	switch args[0] {
	case "instance":
		jInstance = args[1]
	case "username":
		jUser = args[1]
	case "passwd":
		jPswd = args[1]
	case "update":
		updateJiraClient()
	}
}
