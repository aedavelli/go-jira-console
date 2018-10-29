package console

import (
	"fmt"

	"github.com/aedavelli/go-jira"
	flag "github.com/spf13/pflag"
)

func handleAdminCmd(args []string) {

	switch args[0] {
	case "user":

		if args[1] == "list" {
			var q string
			var e bool
			if len(args) == 2 {
				q = `%%`
			} else {
				f := flag.NewFlagSet("user", flag.ContinueOnError)
				f.BoolVar(&e, "email", false, "List email")

				err := f.Parse(args[2:])
				if err != nil {
					fmt.Println(err)
				}
				if len(f.Args()) > 0 {
					q = f.Args()[0]
				} else {
					q = `%%`
				}
			}
			ul, _, err := jClient.User.Find(jira.WithQuery(q))
			if err != nil {
				fmt.Println(err)
			}

			if len(ul) == 0 {
				fmt.Println("No user found matching : ", q)
				return
			}
			for _, u := range ul {
				if e {

					fmt.Printf("%s - %s\n", u.Name, u.EmailAddress)
				} else {
					fmt.Printf("%s\n", u.Name)
				}
			}
		}
	}
}
