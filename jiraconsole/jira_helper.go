package jiraconsole

import (
	"fmt"

	"github.com/aedavelli/go-console"
	"github.com/aedavelli/go-jira"
)

var (
	jClient *jira.Client
)

func init() {
	console.SetAppName("jira")
}

// Update the JIRA client instance
func updateJiraClient() {
	tp := jira.BasicAuthTransport{
		Username: jUser,
		Password: jPswd,
	}
	var err error
	jClient, err = jira.NewClient(tp.Client(), jInstance)
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
}
