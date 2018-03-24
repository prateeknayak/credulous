package core

import (
	"testing"

	"fmt"

	"time"

	"github.com/realestate-com-au/credulous/pkg/models"
	"github.com/stretchr/testify/assert"
)

type unaHappy struct{}

func (u *unaHappy) GetUsername() (string, error) { return "user", nil }
func (u *unaHappy) GetAlias() (string, error)    { return "alias", nil }

func TestGetAWSUsernameAndAlias_Happy(t *testing.T) {
	u := &unaHappy{}

	s1, s2, err := GetUsernameAndAlias(u)
	assert.Nil(t, err)
	assert.NotEmpty(t, s1)
	assert.NotEmpty(t, s2)
}

type unaUError struct{}

func (u *unaUError) GetUsername() (string, error) { return "", fmt.Errorf("username empty") }
func (u *unaUError) GetAlias() (string, error)    { return "", nil }

func TestGetAWSUsernameAndAlias_UError(t *testing.T) {
	u := &unaUError{}

	s1, s2, err := GetUsernameAndAlias(u)
	assert.NotNil(t, err)
	assert.Empty(t, s1)
	assert.Empty(t, s2)
}

type unaAError struct{}

func (u *unaAError) GetUsername() (string, error) { return "", nil }
func (u *unaAError) GetAlias() (string, error)    { return "", fmt.Errorf("alias empty") }

func TestGetAWSUsernameAndAlias_AError(t *testing.T) {
	u := &unaAError{}

	s1, s2, err := GetUsernameAndAlias(u)
	assert.NotNil(t, err)
	assert.Empty(t, s1)
	assert.Empty(t, s2)
}

type aliasHappy struct{}

func (u *aliasHappy) GetAlias() (string, error) { return "alias", nil }

func TestVerifyAccount_Happy(t *testing.T) {
	a := &aliasHappy{}
	err := VerifyAccount(a, "alias")
	assert.Nil(t, err)
}

type aliasError struct{}

func (u *aliasError) GetAlias() (string, error) { return "", fmt.Errorf("alias not found") }

func TestVerifyAccount_AError(t *testing.T) {
	a := &aliasError{}
	err := VerifyAccount(a, "alias")
	assert.NotNil(t, err)
}

type aliasNoMatch struct{}

func (u *aliasNoMatch) GetAlias() (string, error) { return "alias", nil }

func TestVerifyAlias_NoMatch(t *testing.T) {
	a := &aliasNoMatch{}
	err := VerifyAccount(a, "nomatch")
	assert.NotNil(t, err)
}

type userHappy struct{}

func (u *userHappy) GetUsername() (string, error) { return "user", nil }

func TestVerifyUsername_Happy(t *testing.T) {
	a := &userHappy{}
	err := VerifyUser(a, "user")
	assert.Nil(t, err)
}

type userError struct{}

func (u *userError) GetUsername() (string, error) { return "", fmt.Errorf("user not found") }

func TestVerifyUsername_AError(t *testing.T) {
	a := &userError{}
	err := VerifyUser(a, "user")
	assert.NotNil(t, err)
}

type userNoMatch struct{}

func (u *userNoMatch) GetUsername() (string, error) { return "user", nil }

func TestVerifyAccount_NoMatch(t *testing.T) {
	a := &userNoMatch{}
	err := VerifyUser(a, "nomatch")
	assert.NotNil(t, err)
}

type vC struct{}

func (u *vC) GetUsername() (string, error) { return "user", nil }
func (u *vC) GetAlias() (string, error)    { return "alias", nil }

func TestValidateCredentials_Happy(t *testing.T) {
	v := &vC{}
	c := models.Credentials{
		IamUsername:      "user",
		AccountAliasOrId: "alias",
	}
	err := ValidateCredentials(v, c, "alias", "user")
	assert.Nil(t, err)
}

func TestValidateCredentials_UsernameNoMatch(t *testing.T) {
	v := &vC{}
	c := models.Credentials{
		IamUsername: "nomatch",
	}
	err := ValidateCredentials(v, c, "alias", "user")
	assert.NotNil(t, err)
}

func TestValidateCredentials_AliasNoMatch(t *testing.T) {
	v := &vC{}
	c := models.Credentials{
		IamUsername:      "user",
		AccountAliasOrId: "nomatch",
	}
	err := ValidateCredentials(v, c, "alias", "user")
	assert.NotNil(t, err)
}

type vCErr struct {
	u, a string
	uErr error
	aErr error
}

func (u *vCErr) GetUsername() (string, error) { return u.u, u.uErr }
func (u *vCErr) GetAlias() (string, error)    { return u.a, u.aErr }

func TestValidateCredentials_VerifyAccountError(t *testing.T) {
	v := &vCErr{
		aErr: fmt.Errorf("alias not found"),
	}
	c := models.Credentials{
		IamUsername:      "user",
		AccountAliasOrId: "alias",
	}
	err := ValidateCredentials(v, c, "alias", "user")
	assert.NotNil(t, err)
}

func TestValidateCredentials_VerifyUserError(t *testing.T) {
	v := &vCErr{
		uErr: fmt.Errorf("user not found"),
	}
	c := models.Credentials{
		IamUsername:      "user",
		AccountAliasOrId: "alias",
	}
	err := ValidateCredentials(v, c, "alias", "user")
	assert.NotNil(t, err)
}

func TestValidateCredentials_SameUsernameAndAliasSuccess(t *testing.T) {
	v := &vCErr{
		a: "alias",
	}
	c := models.Credentials{
		IamUsername:      "alias",
		AccountAliasOrId: "alias",
	}
	err := ValidateCredentials(v, c, "alias", "alias")
	assert.Nil(t, err)
}

func TestValidateCredentials_SameUsernameAndAliasErr(t *testing.T) {
	v := &vCErr{
		uErr: fmt.Errorf("user not found"),
	}
	c := models.Credentials{
		IamUsername:      "alias",
		AccountAliasOrId: "alias",
	}
	err := ValidateCredentials(v, c, "alias", "alias")
	assert.NotNil(t, err)
}

type fakeDeleteOneAccessKey struct {
	keys      []models.AccessKey
	keyErr    error
	deleteErr error
}

func getAccessKeys() []models.AccessKey {
	return []models.AccessKey{
		{"user", "banana1", "active", time.Now(), "morebanana1"},
		{"user", "banana2", "active", time.Now(), "morebanana2"},
		{"user", "banana3", "Inactive", time.Now(), "morebanana3"},
	}
}
func (f *fakeDeleteOneAccessKey) GetAllAccessKeys(username string) ([]models.AccessKey, error) {
	return f.keys, f.keyErr
}

func (f *fakeDeleteOneAccessKey) DeleteAccessKey(key *models.AccessKey) error { return f.deleteErr }

func TestDeleteOneKey_GetAllAccessKeyError(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keyErr: fmt.Errorf("error in getting access keys"),
	}
	err := DeleteOneKey(f, "user")
	assert.NotNil(t, err)
}

func TestDeleteOneKey_LenKeysIsOne(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keys: []models.AccessKey{{"user", "b", "active", time.Now(), "a"}},
	}
	err := DeleteOneKey(f, "user")
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("only one key in the account; cannot delete"))
}

func TestDeleteOneKey_KeysIsEmpty(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keys: []models.AccessKey{},
	}
	err := DeleteOneKey(f, "user")
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("cannot find oldest key for this account, will not rotate"))
}

func TestDeleteOneKey_InactiveKeyFound(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keys: getAccessKeys(),
	}
	err := DeleteOneKey(f, "user")
	assert.Nil(t, err)
}

func TestDeleteOneKey_InactiveKeyNotFound(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keys: []models.AccessKey{
			{"user", "banana1", "active", time.Now(), "morebanana1"},
			{"user", "banana2", "active", time.Now(), "morebanana2"},
			{"user", "banana3", "active", time.Now(), "morebanana3"},
		},
	}
	err := DeleteOneKey(f, "user")
	assert.Nil(t, err)
}

func TestDeleteOneKey_DeleteErr(t *testing.T) {
	customErr := fmt.Errorf("some err")
	f := &fakeDeleteOneAccessKey{
		keys: []models.AccessKey{
			{"user", "banana1", "active", time.Now(), "morebanana1"},
			{"user", "banana2", "active", time.Now(), "morebanana2"},
			{"user", "banana3", "active", time.Now(), "morebanana3"},
		},
		deleteErr: customErr,
	}
	err := DeleteOneKey(f, "user")
	assert.NotNil(t, err)
	assert.Equal(t, err, customErr)
}

func TestDeleteOneKey_Happy(t *testing.T) {
	f := &fakeDeleteOneAccessKey{
		keys: getAccessKeys(),
	}
	err := DeleteOneKey(f, "user")
	assert.Nil(t, err)

}

type fakeCredsRetriever struct{}
