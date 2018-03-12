package cmd

import (
	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
)

func NewSaveCommand(i core.Credulousier) cli.Command {
	return cli.Command{
		Name:  "save",
		Usage: "Save AWS credentials",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "key, k",
				Value: &cli.StringSlice{},
				Usage: "\n        SSH public keys for encryption",
			},
			cli.StringSliceFlag{
				Name:  "env, e",
				Value: &cli.StringSlice{},
				Usage: "\n        Environment variables to set in the form VAR=value",
			},
			cli.IntFlag{
				Name:  "lifetime, l",
				Value: 0,
				Usage: "\n        Credential lifetime in seconds (0 means forever)",
			},
			cli.BoolFlag{
				Name: "force, f",
				Usage: "\n        Force saving without validating username or account." +
					"\n        You MUST specify -u username -a account",
			},
			cli.StringFlag{
				Name:  "username, u",
				Value: "",
				Usage: "\n        Username (for use with '--force')",
			},
			cli.StringFlag{
				Name:  "account, a",
				Value: "",
				Usage: "\n        Account alias (for use with '--force')",
			},
			cli.StringFlag{
				Name:  "repo, r",
				Value: "local",
				Usage: "\n        Repository location ('local' by default)",
			},
		},
		Action: func(c *cli.Context) {
			s, err := i.ParseArgs(c)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			err = core.Save(i, s)
			handler.LogAndDieOnFatalError(err)
		},
	}
}
