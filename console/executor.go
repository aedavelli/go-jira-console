package console

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Executor(in string) {
	args, _ := parseCommandLine(in)
	if len(args) == 0 {
		// No input provided. Simply return
		return
	}

	// Check whether the command is any of global commands exit, quit or switch
	switch args[0] {
	case "quit", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
		return
	case "switch":
		switchCmd := &cobra.Command{
			Use:   "switch [mode]",
			Short: "Change the console mode",
			Long:  "Change the console mode to one of the following \"admin\", \"config\"",
			Args:  cobra.ExactArgs(1),
		}

		switchAdminCmd := &cobra.Command{
			Use:   "admin",
			Short: "Change to admin mode",
			Long:  "Change to admin mode to perform admin operation like adding group, user etc",
			Args:  cobra.ExactArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				prefixCtx = "admin"
				prefixPmt = "# "
			},
		}

		switchConfigCmd := &cobra.Command{
			Use:   "config",
			Short: "Change to config mode",
			Long:  "Change to config mode to configure JIRA username and password",
			Args:  cobra.ExactArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				prefixCtx = "config"
				prefixPmt = "$ "
			},
		}

		switchCmd.AddCommand(switchAdminCmd, switchConfigCmd)
		switchCmd.SetArgs(args[1:])
		switchCmd.Execute()
		return
	}

	// Here means context based commands
	switch prefixCtx {
	case "admin":
		handleAdminCmd(args)
	case "config":
		handleConfigCmd(args)
	}

}
