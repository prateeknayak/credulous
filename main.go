package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/realestate-com-au/credulous/cmd"
	"github.com/realestate-com-au/credulous/pkg/caws"
	"github.com/realestate-com-au/credulous/pkg/ccrypto"
	"github.com/realestate-com-au/credulous/pkg/cgit"
	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/creds"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/realestate-com-au/credulous/pkg/parser"
	"github.com/urfave/cli"
)

func main() {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		handler.LogAndDieOnFatalError(err)
	}
	crypto := ccrypto.NewCrypto()

	c := &core.Credulous{
		AccountInformer: caws.NewAWSIAMImpl(awsiam.New(sess)),
		ArgsParser:      parser.NewParser(),
		GitRepoDetector: cgit.NewGitImpl(),
		CryptoOperator:  crypto,
		CredsReadWriter: creds.NewEncodeDecodeCreds(),
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
