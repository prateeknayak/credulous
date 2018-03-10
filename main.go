package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/prateeknayak/credulous/cmd"
	"github.com/prateeknayak/credulous/pkg/caws"
	"github.com/prateeknayak/credulous/pkg/ccrypto"
	"github.com/prateeknayak/credulous/pkg/cgit"
	"github.com/prateeknayak/credulous/pkg/core"
	"github.com/prateeknayak/credulous/pkg/creds"
	"github.com/prateeknayak/credulous/pkg/handler"
	"github.com/prateeknayak/credulous/pkg/parser"
	"github.com/urfave/cli"
)

/*
	AccountInformer
	ArgsParser
	GitRepoDetector
	CryptoOperator
	CredsReadWriter
*/
type Credulous struct {
	core.AccountInformer
	core.ArgsParser
	core.GitRepoDetector
	core.CryptoOperator
	core.CredsReadWriter
}

func saveSetup() (core.Credulousier, error) {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		return nil, err
	}
	crypto := ccrypto.NewCrypto()

	c := &Credulous{
		caws.NewAWSIAMImpl(awsiam.New(sess)),
		parser.NewParser(),
		cgit.NewGitImpl(),
		crypto,
		creds.NewEncodeDecodeCreds(),
	}
	return c, nil
}

func main() {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		handler.LogAndDieOnFatalError(err)
	}
	crypto := ccrypto.NewCrypto()

	c := &Credulous{
		caws.NewAWSIAMImpl(awsiam.New(sess)),
		parser.NewParser(),
		cgit.NewGitImpl(),
		crypto,
		creds.NewEncodeDecodeCreds(),
	}
	app := cli.NewApp()
	app.Name = "credulous"
	app.Usage = "Secure AWS Credential Management"
	app.Version = "0.2.2"

	app.Commands = []cli.Command{
		cmd.NewSaveCommand(c),
		cmd.NewSourceCommand(c),
		cmd.NewCurrentCommand(c),
		cmd.NewDisplayCommand(),
		cmd.NewListCommand(c),
		cmd.NewRotateCommand(c),
	}

	app.Run(os.Args)
}
