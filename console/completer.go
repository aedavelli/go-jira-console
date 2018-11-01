package console

import (
	"github.com/c-bata/go-prompt"
)

var gs = []prompt.Suggest{
	{Text: "exit", Description: "Exit JIRA console"},
	{Text: "quit", Description: "Exit JIRA console"},
	{Text: "switch", Description: "Switch the mode"},
}

func Completer(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	switch prefixCtx {
	case "config":
		s = configSuggestions(d)
	case "admin", "":
		s = adminSuggestions(d)
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func configSuggestions(d prompt.Document) []prompt.Suggest {
	return append(gs, prompt.Suggest{Text: "instance", Description: "JIRA instance base URL"},
		prompt.Suggest{Text: "passwd", Description: "Password for the user"},
		prompt.Suggest{Text: "username", Description: "Username for querying JIRA"},
		prompt.Suggest{Text: "update", Description: "Update configuration"})
}

func adminSuggestions(d prompt.Document) []prompt.Suggest {
	return gs
}
