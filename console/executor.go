package console

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	topLevelCmds = make(map[string]*cobra.Command)
)

func listCmdsFromCtx(m map[string]*cobra.Command) {
	for _, cmd := range m {
		fmt.Println("\t", cmd.NameAndAliases())
	}
}

func listCmds() {
	fmt.Println("Available commands")
	listCmdsFromCtx(topLevelCmds)
	switch prefixCtx {
	case "admin":
		listCmdsFromCtx(adminTopCmds)
	case "config":
		listCmdsFromCtx(configTopCmds)
	}
}

func init() {

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

	exitCmd := &cobra.Command{
		Use:     "exit",
		Short:   "Exit the console",
		Long:    "Exit the console",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"quit"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Bye!")
			os.Exit(0)
			return
		},
	}

	topLevelCmds[exitCmd.Name()] = exitCmd
	topLevelCmds[switchCmd.Name()] = switchCmd
}

func Executor(in string) {
	args, _ := parseCommandLine(in)
	if len(args) == 0 {
		// No input provided. Provide list of commands available in this context and return
		listCmds()
		return
	}

	// Check whether the command is any of global commands exit, quit or switch
	if cmd, ok := topLevelCmds[args[0]]; ok {
		cmd.SetArgs(args[1:])
		cmd.Execute()
		return
	}

	// Here means context based commands
	switch prefixCtx {
	case "admin":
		handleCmd(args, adminTopCmds)
	case "config":
		handleCmd(args, configTopCmds)
	default:
		listCmds()
	}

}

func handleCmd(args []string, cmds map[string]*cobra.Command) {
	// No need to verify args length. Here means it's already validated
	if cmd, ok := cmds[args[0]]; ok {
		resetAllCmdFlags(cmd)
		cmd.SetArgs(args[1:])
		cmd.Execute()
	} else {
		listCmds()
	}
}

func resetCmdFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	if fs != nil {
		fs.VisitAll(func(f *pflag.Flag) {
			f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
}

func resetAllCmdFlags(p *cobra.Command) {
	for _, cmd := range p.Commands() {
		resetCmdFlags(cmd)
	}
}
