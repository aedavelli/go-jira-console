package jiraconsole

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aedavelli/go-console"
	"github.com/aedavelli/go-jira"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	grpSet = map[string]string{}
)

func init() {
	// group top level command
	grpCmd := &cobra.Command{
		Use:   "group [command]",
		Short: "Perform group related operations",
		Long:  `Perform group related operations like "add" "list" etc`,
		Args:  cobra.MinimumNArgs(1),
	}

	// group list command
	grpListCmd := &cobra.Command{
		Use:   "list",
		Short: "List the available groups",
		Long:  "List the available groups",
		Args:  cobra.ExactArgs(0),
		Run:   grpListExec,
	}
	grpListCmd.Flags().Int16P("limit", "l", 20, "Limit the results")

	// group show command
	grpShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show the groups members",
		Long:  "Show the groups members",
		Args:  cobra.MinimumNArgs(1),
		Run:   grpShowExec,
	}
	grpShowCmd.Flags().Int16P("limit", "l", 50, "The maximum number of users to return per page.")
	grpShowCmd.Flags().Bool("list-inactive", false, "Include inactive users.")
	grpShowCmd.Flags().Int16("start-at", 0, "The index of the first user to return.")

	// group creation command
	grpCreateCmd := &cobra.Command{
		Use:     "create <groupnames>",
		Short:   "Create groups",
		Long:    "Create groups",
		Aliases: []string{"add"},
		Args:    cobra.MinimumNArgs(1),
		Run:     grpCreateExec,
	}

	// users addition to the group command
	grpAddUserCmd := &cobra.Command{
		Use:   "add_users <groupname> <user names>",
		Short: "Add users to group",
		Long:  "Add users to group",
		Args:  cobra.MinimumNArgs(2),
		Run:   grpAddUserExec,
	}

	// users removal from the group command
	grpRemoveUserCmd := &cobra.Command{
		Use:   "remove_users <groupname> <user names>",
		Short: "Add users from group",
		Long:  "Add users from group",
		Args:  cobra.MinimumNArgs(2),
		Run:   grpRemoveUserExec,
	}

	grpCmd.AddCommand(grpListCmd, grpShowCmd, grpCreateCmd, grpAddUserCmd, grpRemoveUserCmd)
	console.RegisterCommandWithCtx(grpCmd, "admin", nil)
	console.RegisterArgCompleter(grpShowCmd, "admin", grpListCompleter)
	console.RegisterArgCompleter(grpAddUserCmd, "admin", userAndGroupListCompleter)
	console.RegisterArgCompleter(grpRemoveUserCmd, "admin", userAndGroupListCompleter)

	go updateGroupListComplter()
}

// Update the group list completer
func updateGroupListComplter() {
	qp := url.Values{}
	qp["maxResults"] = []string{"500"}
	grpList(qp)
}

// Get the user and group completer map
func userAndGroupListCompleter() map[string]string {
	s := userListCompleter()
	for key, val := range grpListCompleter() {
		s[key] = val
	}
	return s
}

// Get the group completer map
func grpListCompleter() map[string]string {
	s := map[string]string{}
	for key, _ := range grpSet {
		g := key
		if strings.ContainsRune(key, ' ') {
			g = fmt.Sprintf("%q", key)
		}
		s[g] = fmt.Sprintf("group: %s", g)
	}
	return s
}

// Get the group list
func grpList(v url.Values) (*jira.GroupList, error) {
	gl, _, err := jClient.Group.GetListWithOptions(v)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, g := range gl.Groups {
		grpSet[g.Name] = fmt.Sprintf("group: %s", g.Name)
	}
	return gl, nil
}

// Execute group listing
func grpListExec(cmd *cobra.Command, args []string) {
	l, _ := cmd.Flags().GetInt16("limit")

	qp := url.Values{}
	qp["maxResults"] = []string{strconv.Itoa(int(l))}
	if gl, err := grpList(qp); err == nil {
		fmt.Println("Total groups", gl.Total)
		t := tablewriter.NewWriter(os.Stdout)
		t.SetHeader([]string{"Group Name"})
		for _, g := range gl.Groups {
			t.Append([]string{g.Name})
		}
		t.Render()
	}
}

// Execute group detailing
func grpShowExec(cmd *cobra.Command, args []string) {
	l, _ := cmd.Flags().GetInt16("limit")
	s, _ := cmd.Flags().GetInt16("start-at")
	i, _ := cmd.Flags().GetBool("list-inactive")

	wg := sync.WaitGroup{}

	type groupMembers struct {
		GroupName string
		Members   []jira.GroupMember
	}

	ch := make(chan groupMembers)

	for _, g := range args {
		wg.Add(1)
		grp := g
		go func() {
			gm, _, err := jClient.Group.GetWithOptions(grp,
				&jira.GroupSearchOptions{
					MaxResults:           int32(l),
					StartAt:              int64(s),
					IncludeInactiveUsers: i,
				})
			if err != nil {
				ch <- groupMembers{GroupName: grp, Members: []jira.GroupMember{}}
			} else {
				ch <- groupMembers{GroupName: grp, Members: gm}
			}
		}()
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Group Name", "Members"})
	t.NumLines()
	t.SetRowLine(true)
	go func() {
		for gm := range ch {
			buf := bytes.Buffer{}
			first := true
			for _, m := range gm.Members {
				if !first {
					buf.WriteString(", ")
				}
				buf.WriteString(m.DisplayName + " (" + m.Name + ")")
				first = false
			}
			t.Append([]string{gm.GroupName, buf.String()})
			wg.Done()
		}
	}()

	wg.Wait()
	t.Render()
	close(ch)
	return
}

// Execute group creation
func grpCreateExec(cmd *cobra.Command, args []string) {
	wg := sync.WaitGroup{}

	ch := make(chan *jira.GroupDetails)

	for _, g := range args {
		wg.Add(1)
		grp := g
		go func() {
			gd, _, err := jClient.Group.Create(
				&jira.GroupDetails{
					Name: grp,
				})
			if err != nil {
				ch <- &jira.GroupDetails{Name: grp + " (Not)"}
			} else {
				ch <- gd
			}
		}()
	}

	go func() {
		for gd := range ch {
			fmt.Println(gd.Name, "Created")
			if gd.Html != "" {
				grpSet[gd.Name] = fmt.Sprintf("group: %s", gd.Name)
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(ch)
	return
}

// Execute users addition to the group
func grpAddUserExec(cmd *cobra.Command, args []string) {
	wg := sync.WaitGroup{}

	ch := make(chan string)

	for _, u := range args[1:] {
		wg.Add(1)
		user := u
		go func() {
			_, _, err := jClient.Group.AddUser(args[0], user)
			if err != nil {
				ch <- fmt.Sprintf("user: %s not added to group %s. Err: %v", user, args[0], err)
			} else {
				ch <- fmt.Sprintf("user: %s added to group: %s", user, args[0])
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

// Execute users removal from the group
func grpRemoveUserExec(cmd *cobra.Command, args []string) {
	wg := sync.WaitGroup{}

	ch := make(chan string)

	for _, u := range args[1:] {
		wg.Add(1)
		user := u
		go func() {
			_, err := jClient.Group.RemoveUser(args[0], user)
			if err != nil {
				ch <- fmt.Sprintf("user: %s not removed from group %s. Err: %v", user, args[0], err)
			} else {
				ch <- fmt.Sprintf("user: %s removed from group: %s", user, args[0])
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
