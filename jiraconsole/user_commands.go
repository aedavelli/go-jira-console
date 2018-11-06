package jiraconsole

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/aedavelli/go-console"
	"github.com/aedavelli/go-jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	userSet = map[string]string{}
)

func init() {
	// user top level command
	userCmd := &cobra.Command{
		Use:   "user [command]",
		Short: "Perform user related operations",
		Long:  `Perform user related operations like "add" "list" "search" etc`,
		Args:  cobra.MinimumNArgs(1),
	}

	// user search command
	userSearchCmd := &cobra.Command{
		Use:   "search <string>",
		Short: "Perform user search operation",
		Long:  `Perform user search operation based on given string`,
		Args:  cobra.ExactArgs(1),
		Run:   userSearchExec,
	}

	userSearchCmd.Flags().Bool("list-all", false, "List all users override search string")
	userSearchCmd.Flags().Bool("list-inactive", false, "List inactive users")
	userSearchCmd.Flags().Int16P("limit", "l", 500, "Limit the results")

	// user show command
	userShowCmd := &cobra.Command{
		Use:     "show <string>",
		Short:   "Show details about the user",
		Long:    `Show details about the user`,
		Aliases: []string{"list"},
		Args:    cobra.MinimumNArgs(1),
		Run:     userShowExec,
	}

	userShowCmd.Flags().BoolP("include-groups", "g", true, "Include user groups")

	// user create command
	userCreateCmd := &cobra.Command{
		Use:     "create <string>",
		Short:   "Create a user",
		Long:    `Create a user on JIRA instance`,
		Aliases: []string{"add"},
		Args:    cobra.ExactArgs(0),
		Run:     userCreateExec,
	}
	userCreateCmd.Flags().StringP("email", "e", "", "Email address of the user")
	userCreateCmd.MarkFlagRequired("email")

	userCreateCmd.Flags().StringP("display-name", "n", "", "Display name of the user")
	userCreateCmd.MarkFlagRequired("display-name")

	// Add user to the groups
	userAddGroupCmd := &cobra.Command{
		Use:   "add_groups <username> <groupnames>",
		Short: "Add users to group",
		Long:  "Add users to group",
		Args:  cobra.MinimumNArgs(2),
		Run:   userAddGroupExec,
	}

	// Remove user from groups
	userRemoveGroupCmd := &cobra.Command{
		Use:   "remove_groups <username> <groupnames>",
		Short: "Add users from group",
		Long:  "Add users from group",
		Args:  cobra.MinimumNArgs(2),
		Run:   userRemoveGroupExec,
	}

	// Add sub-commands to user command
	userCmd.AddCommand(userSearchCmd, userShowCmd, userCreateCmd, userAddGroupCmd, userRemoveGroupCmd)

	console.RegisterCommandWithCtx(userCmd, "admin", nil)
	console.RegisterArgCompleter(userShowCmd, "admin", userListCompleter)
	console.RegisterArgCompleter(userAddGroupCmd, "admin", userAndGroupListCompleter)
	console.RegisterArgCompleter(userRemoveGroupCmd, "admin", userAndGroupListCompleter)

	// Update user list map
	go updateUserListComplter()
}

// Get the user list completer map
func userListCompleter() map[string]string {
	return userSet
}

// Update the user list completer map
func updateUserListComplter() {
	qp := url.Values{}
	qp["query"] = []string{`%%`}
	qp["includeInactive"] = []string{"true"}
	qp["maxResults"] = []string{"500"}
	getUserDetails(qp)
}

// Get the use details by querying the JIRA instance
func getUserDetails(qp url.Values) ([]jira.User, error) {
	ul, _, err := jClient.User.FindWithQueryParams(qp)
	if err != nil {
		return nil, err
	}
	for _, u := range ul {
		userSet[u.Name] = fmt.Sprintf("user: %s(%t)", u.DisplayName, u.Active)
	}
	return ul, nil
}

// Print the user list table
func printUsersTable(ul []jira.User, g bool) {
	t := tablewriter.NewWriter(os.Stdout)
	t.NumLines()
	t.SetRowLine(true)
	th := []string{"Name", "DisplayName", "Email", "Active"}
	if g {
		th = append(th, "Groups")
	}
	t.SetHeader(th)

	for _, u := range ul {
		tr := []string{u.Name, u.DisplayName, u.EmailAddress}
		if u.Active {
			tr = append(tr, "Active")
		} else {
			tr = append(tr, "Inactive")
		}

		if g {
			first := true
			var buf bytes.Buffer
			for _, grp := range u.Groups.Items {
				if !first {
					buf.WriteString(", ")
				}
				buf.WriteString(grp.Name)
				first = false
			}
			tr = append(tr, buf.String())
		}
		t.Append(tr)
	}
	t.Render()
}

// Execute user search
func userSearchExec(cmd *cobra.Command, args []string) {
	q := args[0]
	l, _ := cmd.Flags().GetInt16("limit")
	i, _ := cmd.Flags().GetBool("list-inactive")

	qp := url.Values{}
	qp["query"] = []string{q}
	// Override remaining flags
	qp["includeInactive"] = []string{fmt.Sprintf("%t", i)}
	qp["maxResults"] = []string{fmt.Sprintf("%d", l)}

	if cmd.Flags().Changed("list-all") {
		qp["query"] = []string{`%%`}
		// Override remaining flags
		qp["includeInactive"] = []string{"true"}
		qp["maxResults"] = []string{"500"}
	}

	ul, err := getUserDetails(qp)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(ul) == 0 {
		fmt.Println("No user found matching : ", q)
		return
	}
	printUsersTable(ul, false)
}

// Execute the user creation
func userCreateExec(cmd *cobra.Command, args []string) {
	e, _ := cmd.Flags().GetString("email")
	n, _ := cmd.Flags().GetString("display-name")

	u := &jira.User{DisplayName: n, EmailAddress: e, Notification: true}

	if u, _, err := jClient.User.Create(u); err != nil {
		fmt.Println("Unable to create user : ", n)
		fmt.Println(err)
	} else {
		userSet[u.Name] = fmt.Sprintf("user: %s(%t)", u.DisplayName, u.Active)
		fmt.Println("User : ", u.DisplayName, "Created")
	}

}

// Execute use listing
func userShowExec(cmd *cobra.Command, args []string) {
	g, _ := cmd.Flags().GetBool("include-groups")
	wg := sync.WaitGroup{}

	ch := make(chan jira.User)

	for _, u := range args {
		wg.Add(1)
		user := u
		go func() {
			qp := url.Values{}
			qp["username"] = []string{user}
			if g {
				qp["expand"] = []string{"groups"}
			}
			ud, _, err := jClient.User.GetWithQueryParams(qp)
			if err != nil {
				ch <- jira.User{Name: user + " (not available)"}
			} else {
				ch <- *ud
			}
		}()
	}

	ul := []jira.User{}

	go func() {
		for ud := range ch {
			ul = append(ul, ud)
			wg.Done()
		}
	}()

	wg.Wait()
	close(ch)
	printUsersTable(ul, g)
}

// Execute user addition to the groups
func userAddGroupExec(cmd *cobra.Command, args []string) {
	wg := sync.WaitGroup{}

	ch := make(chan string)

	for _, g := range args[1:] {
		wg.Add(1)
		grp := g
		go func() {
			_, _, err := jClient.Group.AddUser(grp, args[0])
			if err != nil {
				ch <- fmt.Sprintf("user: %s not added to group %s. Err: %v", args[0], grp, err)
			} else {
				ch <- fmt.Sprintf("user: %s added to group: %s", args[0], grp)
			}
		}()
	}

	go func() {
		for s := range ch {
			fmt.Println(s)
			wg.Done()
		}
	}()

	wg.Wait()
	close(ch)
	return
}

// Execute user removal from the groups
func userRemoveGroupExec(cmd *cobra.Command, args []string) {
	wg := sync.WaitGroup{}

	ch := make(chan string)

	for _, g := range args[1:] {
		wg.Add(1)
		grp := g
		go func() {
			_, err := jClient.Group.RemoveUser(grp, args[0])
			if err != nil {
				ch <- fmt.Sprintf("user: %s not removed from group %s. Err: %v", args[0], grp, err)
			} else {
				ch <- fmt.Sprintf("user: %s removed from group: %s", args[0], grp)
			}
		}()
	}

	go func() {
		for s := range ch {
			fmt.Println(s)
			wg.Done()
		}
	}()

	wg.Wait()
	close(ch)
	return
}
