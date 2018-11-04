package console

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func getSuggestions(d prompt.Document, m map[string]*cobra.Command) []prompt.Suggest {

	s := []prompt.Suggest{}

	tok, err := parseCommandLine(d.Text)
	if err != nil {
		return s
	}

	if len(tok) == 1 {
		//Iterate over top level commands
		for _, lm := range []map[string]*cobra.Command{topLevelCmds, m} {
			for _, cmd := range lm {
				s = cmdSuggestions(cmd, s, true)
			}
		}
		return s
	}

	for _, lm := range []map[string]*cobra.Command{topLevelCmds, m} {
		if cmd, ok := lm[tok[0]]; ok {
			if c := findLastValidCommand(cmd, tok[1:]); c != nil {
				if c.HasSubCommands() {
					s = cmdSuggestions(c, s, false)
				} else {
					s = cmdSuggestions(c, s, true)
				}
			}
			return s
		}
	}
	return s
}

func cmdSuggestions(c *cobra.Command, s []prompt.Suggest, this bool) []prompt.Suggest {
	var cmds []*cobra.Command
	if this {
		cmds = make([]*cobra.Command, 0)
		cmds = append(cmds, c)
	} else {
		cmds = c.Commands()
	}

	for _, c := range cmds {
		// Here we've matching subcommand
		s = append(s, prompt.Suggest{Text: c.Name(), Description: c.Short})
		// Look for aliases
		for _, a := range c.Aliases {
			s = append(s, prompt.Suggest{Text: a, Description: c.Short})
		}

		// Populate flags
		if fs := c.Flags(); fs != nil {
			fs.VisitAll(func(f *pflag.Flag) {
				if f.Shorthand != "" {
					s = append(s, prompt.Suggest{Text: "-" + f.Shorthand, Description: f.Usage})
				}
				if f.Name != "" {
					s = append(s, prompt.Suggest{Text: "--" + f.Name, Description: f.Usage})
				}
			})
		}
	}
	return s
}

func findLastValidCommand(c *cobra.Command, args []string) *cobra.Command {
	if !c.HasSubCommands() || len(args) < 2 {
		return c
	}

	if cmd := findNext(c, args[0]); cmd != nil {
		return findLastValidCommand(cmd, args[1:])
	}
	return nil
}

func findNext(c *cobra.Command, next string) *cobra.Command {
	for _, cmd := range c.Commands() {
		if cmd.Name() == next || cmd.HasAlias(next) {
			return cmd
		}
	}
	return nil
}

func Completer(d prompt.Document) []prompt.Suggest {
	var s []prompt.Suggest
	w := d.GetWordBeforeCursor()
	if w == "" {
		return s
	}

	switch prefixCtx {
	case "config":
		s = getSuggestions(d, configTopCmds)
	case "admin":
		s = getSuggestions(d, adminTopCmds)
	default:
		s = getSuggestions(d, nil)
	}

	return prompt.FilterHasPrefix(s, w, true)
}
