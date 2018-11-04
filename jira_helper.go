package main

import (
	"fmt"

	jira "github.com/aedavelli/go-jira"
)

var (
	jClient *jira.Client
)

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
