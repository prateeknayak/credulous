package cmd

import (
	"fmt"

	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
)

func NewCurrentCommand(i core.Credulousier) cli.Command {
	return cli.Command{
		Name:  "current",
		Usage: "Show the username and alias of the currently-loaded credentials",
		Action: func(c *cli.Context) {
			username, alias, err := core.GetUsernameAndAlias(i)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			fmt.Printf("%s@%s\n", username, alias)
		},
	}
}
