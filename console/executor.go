package console

import (
	"fmt"
	"os"
)

func Executor(in string) {

	fmt.Println("Your input: " + in)
	args, _ := parseCommandLine(in)

	// Check whether the command is any of global commands exit, quit or switch
	switch args[0] {
	case "quit", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
		return
	case "switch":
		if len(args) == 1 {
			prefixCtx = ""
			prefixPmt = " "
			return
		}
		switch args[1] {
		case "admin":
			prefixCtx = "admin"
			prefixPmt = "# "
		case "config":
			prefixCtx = "config"
			prefixPmt = "$ "
		}
		return
	}

	switch prefixCtx {
	case "admin":
		handleAdminCmd(args)
	case "config":
		handleConfigCmd(args)
	}

}
