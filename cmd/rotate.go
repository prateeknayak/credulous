package cmd

import (
	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
)

func NewRotateCommand(i core.Credulousier) cli.Command {
	return cli.Command{
		Name:  "rotate",
		Usage: "Rotate current AWS credentials, deleting the oldest",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "lifetime, l",
				Value: 0,
				Usage: "\n        New credential lifetime in seconds (0 means forever)",
			},
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
			cli.StringFlag{
				Name:  "repo, r",
				Value: "local",
				Usage: "\n        Repository location ('local' by default)",
			},
		},
		Action: func(c *cli.Context) {
			s, err := i.ParseArgs(c)
			handler.LogAndDieOnFatalError(err)

			username, _, err := core.GetAWSUsernameAndAlias(i)
			handler.LogAndDieOnFatalError(err)

			err = core.RotateCredentials(i, username)
			handler.LogAndDieOnFatalError(err)

			s.Force = c.Bool("force")
			err = core.Save(i, *s)
			handler.LogAndDieOnFatalError(err)
		},
	}
}
