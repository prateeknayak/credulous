package main

import (
	"testing"

	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/realestate-com-au/credulous/pkg/caws"
	"github.com/realestate-com-au/credulous/pkg/ccrypto"
	"github.com/realestate-com-au/credulous/pkg/cgit"
	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/creds"
	"github.com/realestate-com-au/credulous/pkg/models"
	"github.com/realestate-com-au/credulous/pkg/parser"
	"github.com/stretchr/testify/assert"
)

// Set of Integration tests;
// depends on ~/.aws/credentials
// Do not wan't to use ENV Vars
// Because of IDE shenanigans
// if modifying make sure it still works with
// ~/.aws/credentials

func TestSaveComand(t *testing.T) {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		t.Fail()
		os.Exit(1)
	}
	client := iam.New(sess)
	crypto := ccrypto.NewCrypto()

	c := &Credulous{
		AccountInformer:  caws.NewAWSIAMImpl(client),
		ArgsParser:       parser.NewParser(),
		CredentialStorer: cgit.NewGitImpl(),
		CryptoOperator:   crypto,
		CredsReadWriter:  creds.NewEncodeDecodeCreds(),
	}
	v, err := client.Config.Credentials.Get()
	if err != nil {
		t.Fail()
		os.Exit(1)
	}

	s := &models.SaveData{
		Cred: models.Credential{
			KeyId:     v.AccessKeyID,
			SecretKey: v.SecretAccessKey,
		},
		Repo: "local",
	}
	err = core.Save(c, s)

	assert.Nil(t, err)
}
