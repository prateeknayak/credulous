package caws

import (
	"testing"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/stretchr/testify/assert"
)

func setup() (*AWSIAMImpl, error) {
	sess, err := session.NewSession(aws2.NewConfig())
	if err != nil {
		return nil, err
	}
	return NewAWSIAMImpl(awsiam.New(sess)), nil
}
func TestAWSIAMImpl_GetAWSUsernameAndAlias(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Fail()
	}

	username, err := i.GetUsername()
	assert.NotEmpty(t, username)
	assert.Nil(t, err)

}

func TestAWSIAMImpl_GetAWSAccountAlias(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Fail()
	}

	alias, err := i.GetAlias()
	assert.NotEmpty(t, alias)
	assert.Nil(t, err)
}

func TestAWSIAMImpl_GetKeyCreateDate(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Fail()
	}
	name, err := i.GetUsername()
	if err != nil {
		t.Fail()
	}
	alias, err := i.GetKeyCreationDate(name)

	assert.NotEmpty(t, alias)
	assert.Nil(t, err)
}
