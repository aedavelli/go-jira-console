package jiraconsole

import (
	"fmt"

	"github.com/aedavelli/go-console"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	jInstance = ""
	jUser     = ""
	jPswd     = ""
)

func init() {
	// username for JIRA
	unCmd := &cobra.Command{
		Use:   "username <string>",
		Short: "Set JIRA username",
		Long:  `Set JIRA instance username`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jUser = args[0]
			return
		},
	}
	console.RegisterCommandWithCtx(unCmd, "config", nil)

	// password for JIRA
	pswdCmd := &cobra.Command{
		Use:   "password <string>",
		Short: "Set JIRA password",
		Long:  `Set JIRA instance password`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jPswd = args[0]
			return
		},
	}

	console.RegisterCommandWithCtx(pswdCmd, "config", nil)

	// JIRA instance command
	instCmd := &cobra.Command{
		Use:   "instance <string>",
		Short: "Set JIRA instance URL",
		Long:  `Set JIRA instance URL`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jInstance = args[0]
			return
		},
	}
	console.RegisterCommandWithCtx(instCmd, "config", nil)

	// JIRA instance update
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update JIRA client with latest config",
		Long:  `Update JIRA client with latest config`,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			updateJiraClient()
			return
		},
	}
	console.RegisterCommandWithCtx(updateCmd, "config", nil)
	readConfig()
	updateJiraClient()
}

// Read the application configuration and update JIRA
func readConfig() {

	viper.SetConfigName("jiraconsole")
	viper.AddConfigPath("$HOME/.jiraconsole")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Unable to read configuration file")
	}

	// function for updating the JIRA config params
	f := func() {
		jInstance = viper.GetString("instance")
		jUser = viper.GetString("username")
		jPswd = viper.GetString("password")
	}
	f()

	// Watch the config file
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		f()
		updateJiraClient()
	})
}
