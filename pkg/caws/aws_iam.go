package caws

import (
	"fmt"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/realestate-com-au/credulous/pkg/core"
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
		username, err := a.GetAWSUsername()
		if err != nil {
			return "", err
		}
		return username, nil
	}
	return *output.AccountAliases[0], nil
}

func (a *AWSIAMImpl) GetAllAccessKeys(username string) ([]core.AccessKey, error) {

	input := &awsiam.ListAccessKeysInput{UserName: aws.String(username)}
	output, err := a.client.ListAccessKeys(input)
	if err != nil {
		return nil, err
	}

	// wtf?
	if len(output.AccessKeyMetadata) == 0 {
		return nil, fmt.Errorf("cannot find any access key for username: %s", username)
	}

	// only one key
	if len(output.AccessKeyMetadata) == 1 {
		return nil, nil
	}
	var allKeys []core.AccessKey
	for _, v := range output.AccessKeyMetadata {
		allKeys = append(allKeys, core.AccessKey{
			Username:   *v.UserName,
			Status:     *v.Status,
			CreateDate: *v.CreateDate,
			KeyId:      *v.AccessKeyId,
		})
	}
	return allKeys, nil
}

func (a *AWSIAMImpl) DeleteAccessKey(key *core.AccessKey) error {
	input := &awsiam.DeleteAccessKeyInput{
		AccessKeyId: &key.KeyId,
		UserName:    &key.Username,
	}
	_, err := a.client.DeleteAccessKey(input)
	if err != nil {
		return err
	}
	return nil
}

func (a *AWSIAMImpl) CreateNewAccessKey(username string) (*core.AccessKey, error) {
	input := &awsiam.CreateAccessKeyInput{
		UserName: aws.String(username),
	}
	output, err := a.client.CreateAccessKey(input)
	if err != nil {
		return nil, err
	}
	key := &core.AccessKey{
		Username:   *output.AccessKey.UserName,
		KeyId:      *output.AccessKey.AccessKeyId,
		Status:     *output.AccessKey.Status,
		CreateDate: *output.AccessKey.CreateDate,
		Secret:     *output.AccessKey.SecretAccessKey,
	}

	return key, nil
}

func (a *AWSIAMImpl) GetKeyCreateDate(username string) (time.Time, error) {
	keys, err := a.GetAllAccessKeys(username)
	if err != nil {
		return time.Time{}, err
	}

	value, err := a.client.Config.Credentials.Get()
	if err != nil {
		return time.Time{}, err
	}
	keyid := value.AccessKeyID
	for _, key := range keys {
		if key.KeyId == keyid {
			return key.CreateDate, nil
		}
	}
	return time.Time{}, fmt.Errorf("couldn't find this key")
}
