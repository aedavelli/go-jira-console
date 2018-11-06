package main

import (
	"fmt"

	"github.com/aedavelli/go-console"
	"github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"

	_ "github.com/aedavelli/go-jira-console/jiraconsole"
)

var (
	version  = "0.1"
	revision = "alpha"
)

func main() {
	fmt.Printf("JIRA console %s (rev-%s)\n", version, revision)
	fmt.Println("Please use `exit` or `Ctrl-D` to exit this program.")
	defer fmt.Println("Bye!")
	p := prompt.New(
		console.Executor,
		console.Completer,
		prompt.OptionTitle("JIRA-console: interactive JIRA client"),
		prompt.OptionPrefix("jira> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionLivePrefix(console.Prefix),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)
	p.Run()
}
