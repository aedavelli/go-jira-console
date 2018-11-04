package main

import (
	"github.com/aedavelli/go-console"
	"github.com/spf13/cobra"
)

var (
	jInstance = ""
	jUser     = ""
	jPswd     = ""
)

func init() {
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

	updateJiraClient()
}
