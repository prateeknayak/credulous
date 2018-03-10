package caws

import (
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type AWSIAMImpl struct {
	client *awsiam.IAM
}

func NewAWSIAMImpl(c *awsiam.IAM) *AWSIAMImpl {
	return &AWSIAMImpl{
		client: c,
	}
}
func (a *AWSIAMImpl) GetAWSUsername() (string, error) {
	input := &awsiam.GetUserInput{}
	output, err := a.client.GetUser(input)
	if err != nil {
		return "", err
	}

	return *output.User.UserName, nil
}

func (a *AWSIAMImpl) GetAWSAccountAlias() (string, error) {

	input := &awsiam.ListAccountAliasesInput{}
	output, err := a.client.ListAccountAliases(input)

	if err != nil {
		return "", err
	}

	// There really is only one alias
	if len(output.AccountAliases) == 0 {
		// we have to do a getuser instead and parse out the
		// account ID from the ARN
		username, err := a.GetAWSUsername()
		if err != nil {
			return "", err
		}
		return username, nil
	}
	return *output.AccountAliases[0], nil
}
