package cmd

import (
	"fmt"
	"os"

	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
)

func NewListCommand(i core.Credulousier) cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "List available AWS credentials",
		Action: func(c *cli.Context) {
			rootDir, err := os.Open(handler.GetRootPath())
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			set, err := core.ListAvailableCredentials(i, rootDir)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			for _, cred := range set {
				fmt.Println(cred)
			}
		},
	}
}
