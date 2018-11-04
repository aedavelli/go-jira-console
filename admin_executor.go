package main

import (
	"fmt"
	"os"

	"github.com/aedavelli/go-console"
	"github.com/aedavelli/go-jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	userCmd := &cobra.Command{
		Use:   "user [command]",
		Short: "Perform user related operations",
		Long:  `Perform user related operations like "add" "list" "list_groups" etc`,
		Args:  cobra.MinimumNArgs(1),
	}

	usCmd := &cobra.Command{
		Use:   "search <string>",
		Short: "Perform user search operation",
		Long:  `Perform user search operation based on given string`,
		Args:  cobra.ExactArgs(1),
		Run:   userSearchCmd,
	}

	usCmd.Flags().Bool("list-all", false, "List all users override search string")
	usCmd.Flags().Bool("list-inactive", false, "List inactive users")
	usCmd.Flags().Int16P("limit", "l", 500, "Limit the results")
	userCmd.AddCommand(usCmd)

	console.RegisterCommandWithCtx(userCmd, "admin", nil)
}

func userSearchCmd(cmd *cobra.Command, args []string) {
	q := args[0]
	l, _ := cmd.Flags().GetInt16("limit")
	i, _ := cmd.Flags().GetBool("list-inactive")

	if cmd.Flags().Changed("list-all") {
		q = `%%`
		// Override remaining flags
		i = true
		l = 500
	}

	ul, _, err := jClient.User.Find(jira.WithQuery(q),
		jira.WithMaxResults(int(l)), jira.WithInactive(i))
	if err != nil {
		fmt.Println(err)
	}

	if len(ul) == 0 {
		fmt.Println("No user found matching : ", q)
		return
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Name", "DisplayName", "Email", "Active"})

	for _, u := range ul {
		if u.Active {
			t.Append([]string{u.Name, u.DisplayName, u.EmailAddress, "Active"})
		} else {
			t.Append([]string{u.Name, u.DisplayName, u.EmailAddress, "Inactive"})
		}
	}
	t.Render()
}
