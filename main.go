package main

import (
	"os"

	"flag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/realestate-com-au/credulous/cmd"
	"github.com/realestate-com-au/credulous/pkg/caws"
	"github.com/realestate-com-au/credulous/pkg/ccrypto"
	"github.com/realestate-com-au/credulous/pkg/cgit"
	"github.com/realestate-com-au/credulous/pkg/cio"
	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/creds"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/realestate-com-au/credulous/pkg/parser"
	"github.com/urfave/cli"
)

/**
type Credulousier interface {
	AccountInformer
	ArgsParser
	CredentialStorer
	CryptoOperator
	CredsReadWriter
	Displayer
	Writer
	Reader
}
*/
type Credulous struct {
	core.AccountInformer
	core.ArgsParser
	core.CredentialStorer
	core.CryptoOperator
	core.CredsReadWriter
	core.Displayer
	core.Writer
	core.Reader
}

func main() {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		handler.LogAndDieOnFatalError(err)
	}
	crypto := ccrypto.NewCrypto()
	fileIO := cio.NewFileIO()

	c := &Credulous{
		AccountInformer:  caws.NewAWSIAMImpl(awsiam.New(sess)),
		ArgsParser:       parser.NewParser(),
		CredentialStorer: cgit.NewGitImpl(),
		CryptoOperator:   crypto,
		CredsReadWriter:  creds.NewEncodeDecodeCreds(),
		Displayer:        cio.NewConsoleWriter(os.Stdout),
		Writer:           fileIO,
		Reader:           fileIO,
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
	flag.Parse()
	app.Run(os.Args)
}
