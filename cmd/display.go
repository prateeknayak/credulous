package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func NewDisplayCommand() cli.Command {
	return cli.Command{
		Name:  "display",
		Usage: "Display loaded AWS credentials",
		Action: func(c *cli.Context) {
			AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
			AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			fmt.Printf("AWS_ACCESS_KEY_ID: %s\n", AWSAccessKeyId)
			fmt.Printf("AWS_SECRET_ACCESS_KEY: %s\n", AWSSecretAccessKey)
		},
	}
}
