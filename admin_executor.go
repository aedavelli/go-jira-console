package main

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

func init() {
	userCmd := &cobra.Command{
		Use:   "user [command]",
		Short: "Perform user related operations",
		Long:  `Perform user related operations like "add" "list" "search" etc`,
		Args:  cobra.MinimumNArgs(1),
	}

	usCmd := &cobra.Command{
		Use:   "search <string>",
		Short: "Perform user search operation",
		Long:  `Perform user search operation based on given string`,
		Args:  cobra.ExactArgs(1),
		Run:   userSearchExec,
	}

	usCmd.Flags().Bool("list-all", false, "List all users override search string")
	usCmd.Flags().Bool("list-inactive", false, "List inactive users")
	usCmd.Flags().Int16P("limit", "l", 500, "Limit the results")
	userCmd.AddCommand(usCmd)

	console.RegisterCommandWithCtx(userCmd, "admin", nil)

	grpCmd := &cobra.Command{
		Use:   "group [command]",
		Short: "Perform group related operations",
		Long:  `Perform group related operations like "add" "list" etc`,
		Args:  cobra.MinimumNArgs(1),
	}

	grpListCmd := &cobra.Command{
		Use:   "list",
		Short: "List the available groups",
		Long:  "List the available groups",
		Args:  cobra.ExactArgs(0),
		Run:   grpListExec,
	}
	grpListCmd.Flags().Int16P("limit", "l", 20, "Limit the results")
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

	grpCmd.AddCommand(grpListCmd, grpShowCmd)
	console.RegisterCommandWithCtx(grpCmd, "admin", nil)
	console.RegisterArgCompleter(grpShowCmd, "admin", grpListCompleter)

}

func userSearchExec(cmd *cobra.Command, args []string) {
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
		return
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

func grpListExec(cmd *cobra.Command, args []string) {
	l, _ := cmd.Flags().GetInt16("limit")

	qp := make(url.Values, 0)
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

var grpSet = map[string]bool{}

func grpListCompleter() []string {
	s := []string{}
	for key, _ := range grpSet {
		g := key
		if strings.ContainsRune(key, ' ') {
			g = fmt.Sprintf("%q", key)
		}
		s = append(s, g)
	}
	return s
}

func grpList(v url.Values) (*jira.GroupList, error) {
	gl, _, err := jClient.Group.GetListWithOptions(v)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, g := range gl.Groups {
		grpSet[g.Name] = true
	}
	return gl, nil
}

func grpShowExec(cmd *cobra.Command, args []string) {
	l, _ := cmd.Flags().GetInt16("limit")
	s, _ := cmd.Flags().GetInt16("start-at")
	i, _ := cmd.Flags().GetBool("ist-inactive")

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
					buf.WriteString(" ,")
				}
				buf.WriteString(m.DisplayName + "(" + m.Name + ")")
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
