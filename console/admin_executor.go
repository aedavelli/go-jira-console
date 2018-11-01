package console

import (
	"fmt"
	"os"
	// "text/tabwriter"

	"github.com/aedavelli/go-jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func handleAdminCmd(args []string) {
	// No need to verify args length. Here means it's already validated
	switch args[0] {
	case "user":
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
		var ulAll bool

		usCmd.Flags().BoolVar(&ulAll, "list-all", false, "List all users override search string")

		userCmd.AddCommand(usCmd)
		userCmd.SetArgs(args[1:])
		userCmd.Execute()

	}
}

func userSearchCmd(cmd *cobra.Command, args []string) {
	flAll := cmd.Flag("list-all")
	q := args[0]

	if flAll.Changed {
		q = `%%`
	}

	ul, _, err := jClient.User.Find(jira.WithQuery(q),
		jira.WithMaxResults(500), jira.WithInactive(true))
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
