package cmd

import (
	"os"
	"path"

	"fmt"
	"strings"

	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
)

func splitUserAndAccount(arg string) (string, string, error) {
	atpos := strings.LastIndex(arg, "@")
	if atpos < 1 {
		err := fmt.Errorf("invalid account format; please specify <username>@<account>")
		return "", "", err
	}
	// pull off everything before the last '@'
	return arg[atpos+1:], arg[0:atpos], nil
}

func getAccountAndUserName(c *cli.Context) (string, string, error) {
	if len(c.Args()) > 0 {
		user, acct, err := splitUserAndAccount(c.Args()[0])
		if err != nil {
			return "", "", err
		}
		return user, acct, nil
	}
	if c.String("credentials") != "" {
		user, acct, err := splitUserAndAccount(c.String("credentials"))
		if err != nil {
			return "", "", err
		}
		return user, acct, nil
	} else {
		return c.String("account"), c.String("username"), nil
	}
}

func parseRepoArgs(c *cli.Context) (repo string, err error) {
	// the default is 'local' which is set below, so not much to do here
	if c.String("repo") == "local" {
		repo = path.Join(handler.GetRootPath(), "local")
	} else {
		repo = c.String("repo")
	}
	return repo, nil
}

func NewSourceCommand(i core.Credulousier) cli.Command {
	return cli.Command{
		Name:  "source",
		Usage: "Source AWS credentials",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "account, a",
				Value: "",
				Usage: "\n        AWS Account alias or id",
			},
			cli.StringFlag{
				Name:  "key, k",
				Value: "",
				Usage: "\n        SSH private key",
			},
			cli.StringFlag{
				Name:  "username, u",
				Value: "",
				Usage: "\n        IAM User",
			},
			cli.StringFlag{
				Name:  "credentials, c",
				Value: "",
				Usage: "\n        Credentials, for example username@account",
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "\n        Force sourcing of credentials without validating username or account",
			},
			cli.StringFlag{
				Name:  "repo, r",
				Value: "local",
				Usage: "\n        Repository location ('local' by default)",
			},
		},
		Action: func(c *cli.Context) {
			keyfile := i.GetPrivateKey(c.String("key"))
			account, username, err := getAccountAndUserName(c)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			repo, err := parseRepoArgs(c)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}
			creds, err := core.RetrieveCredentials(i, repo, account, username, keyfile)
			if err != nil {
				handler.LogAndDieOnFatalError(err)
			}

			if !c.Bool("force") {
				err = core.ValidateCredentials(i, creds, account, username)
				if err != nil {
					handler.LogAndDieOnFatalError(err)
				}
			}
			core.DisplayCreds(os.Stdout, creds)
		},
	}
}
