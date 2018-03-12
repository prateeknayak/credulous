package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"time"

	"crypto/rsa"

	"io/ioutil"
	"os"
	"path/filepath"

	"sort"

	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

type FileLister interface {
	Readdir(int) ([]os.FileInfo, error)
	Name() string
}

type AccountInformer interface {
	GetAWSUsername() (string, error)
	GetAWSAccountAlias() (string, error)
	GetAllAccessKeys(username string) ([]AccessKey, error)
	DeleteAccessKey(key *AccessKey) error
	CreateNewAccessKey(username string) (*AccessKey, error)
	GetKeyCreateDate(username string) (time.Time, error)
}

type CryptoOperator interface {
	CredulousEncode(plaintext string, pubkey ssh.PublicKey) (ciphertext string, err error)
	CredulousDecodeAES(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error)
	CredulousDecodePureRSA(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error)
	CredulousDecodeWithSalt(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error)
	SSHFingerprint(pubkey ssh.PublicKey) (fingerprint string)
	SSHPrivateFingerprint(privkey rsa.PrivateKey) (fingerprint string, err error)
	GenerateSalt() (string, error)
}

type ArgsParser interface {
	ParseArgs(c *cli.Context) (*SaveData, error)
}

type CredsReadWriter interface {
	FindDefaultDir(fl FileLister) (string, error)
	GetDirs(fl FileLister) ([]os.FileInfo, error)
	LatestFileInDir(dir string) (os.FileInfo, error)
	ReadCredentialFile(c CryptoOperator, fileName string, keyfile string) (*Credentials, error)
	GetPrivateKey(key string) (filename string)
}

type GitRepoDetector interface {
	IsGitRepo(checkpath string) (bool, error)
	GitAddCommitFile(repopath, filename, message string, config RepoConfig) (commitId string, err error)
}

type Credulousier interface {
	AccountInformer
	ArgsParser
	GitRepoDetector
	CryptoOperator
	CredsReadWriter
}

// Credulous APIs called from main.go
func Save(i Credulousier, data *SaveData) error {

	var keyCreateDate int64

	if data.Force {
		keyCreateDate = time.Now().Unix()
	} else {
		var err error
		if data.Username == "" {
			data.Username, err = i.GetAWSUsername()
			if err != nil {
				return err
			}
		}
		if data.Alias == "" {
			data.Alias, err = i.GetAWSAccountAlias()
			if err != nil {
				return err
			}
		}

		date, err := i.GetKeyCreateDate(data.Username)
		if err != nil {
			return err
		}

		t, err := time.Parse(time.RFC3339, date.Format(time.RFC3339))
		if err != nil {
			return err
		}
		keyCreateDate = t.Unix()
	}

	fmt.Printf("saving credentials for %s@%s\n", data.Username, data.Alias)
	plaintext, err := json.Marshal(data.Cred)
	if err != nil {
		return err
	}

	encSlice := []Encryption{}
	for _, pubkey := range data.Pubkeys {
		encoded, err := i.CredulousEncode(string(plaintext), pubkey)
		if err != nil {
			return err
		}

		encSlice = append(encSlice, Encryption{
			Ciphertext:  encoded,
			Fingerprint: i.SSHFingerprint(pubkey),
		})
	}
	creds := Credentials{
		Version:          FORMAT_VERSION,
		AccountAliasOrId: data.Alias,
		IamUsername:      data.Username,
		CreateTime:       fmt.Sprintf("%d", keyCreateDate),
		Encryptions:      encSlice,
		LifeTime:         data.Lifetime,
	}

	filename := fmt.Sprintf("%v-%v.json", keyCreateDate, data.Cred.KeyId[12:])
	err = WriteToDisk(creds, data.Repo, filename, i)
	if err != nil {
		return err
	}

	return nil
}

func GetAWSUsernameAndAlias(i AccountInformer) (string, string, error) {
	username, err := i.GetAWSUsername()
	if err != nil {
		return "", "", err
	}

	alias, err := i.GetAWSAccountAlias()
	if err != nil {
		return "", "", err
	}

	return username, alias, nil
}

func VerifyAccount(i AccountInformer, alias string) error {
	acctAlias, err := i.GetAWSAccountAlias()
	if err != nil {
		return err
	}
	if acctAlias == alias {
		return nil
	}
	return fmt.Errorf("cannot verify account, does not match alias: %s", alias)
}

func VerifyUser(i AccountInformer, username string) error {
	name, err := i.GetAWSUsername()
	if err != nil {
		return err
	}
	if username == name {
		return nil
	}
	return fmt.Errorf("cannot verify user, does not match access keys: %s", username)
}

func ValidateCredentials(i AccountInformer, cred Credentials, alias string, username string) error {
	if cred.IamUsername != username {
		err := errors.New("FATAL: username in credential does not match requested username")
		return err
	}
	if cred.AccountAliasOrId != alias {
		err := errors.New("FATAL: account alias in credential does not match requested alias")
		return err
	}

	// Make sure the account is who we expect
	err := VerifyAccount(i, cred.AccountAliasOrId)
	if err != nil {
		return err
	}

	// Make sure the user is who we expect
	// If the username is the same as the account name, then it's the root user
	// and there's actually no username at all (oddly)
	if cred.IamUsername == cred.AccountAliasOrId {
		err = VerifyUser(i, "")
	} else {
		err = VerifyUser(i, cred.IamUsername)
	}
	if err != nil {
		return err
	}

	return nil
}

func DisplayCreds(output io.Writer, cred Credentials) {
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=\"%v\"\nexport AWS_SECRET_ACCESS_KEY=\"%v\"\n",
		cred.Encryptions[0].Decoded.KeyId, cred.Encryptions[0].Decoded.SecretKey)
	for key, val := range cred.Encryptions[0].Decoded.EnvVars {
		fmt.Fprintf(output, "export %s=\"%s\"\n", key, val)
	}
}

func DeleteOneKey(i AccountInformer, username string) error {
	keys, err := i.GetAllAccessKeys(username)
	if err != nil {
		return err
	}

	if len(keys) == 1 {
		return fmt.Errorf("only one key in the account; cannot delete")
	}
	// Find out which key to delete.
	var oldestId string
	var oldest int64
	var oldestIndex int
	for k, key := range keys {
		t, err := time.Parse(time.RFC3339, key.CreateDate.Format(time.RFC3339))
		if err != nil {
			return err
		}
		createDate := t.Unix()
		// If we find an inactive one, just delete it
		if key.Status == "Inactive" {
			oldestId = key.KeyId
			break
		}
		if oldest == 0 || createDate < oldest {
			oldest = createDate
			oldestId = key.KeyId
			oldestIndex = k
		}
	}

	if oldestId == "" {
		return fmt.Errorf("cannot find oldest key for this account, will not rotate")
	}

	err = i.DeleteAccessKey(&keys[oldestIndex])
	if err != nil {
		return err
	}

	return nil
}

// Potential conditions to handle here:
// * AWS has one key
//     * only generate a new key, do not delete the old one
// * AWS has two keys
//     * both are active and valid
//     * new one is inactive
//     * old one is inactive
// * We successfully delete the oldest key, but fail in creating the new key (eg network, permission issues)
func RotateCredentials(i AccountInformer, username string) error {
	err := DeleteOneKey(i, username)
	if err != nil {
		return err
	}
	_, err = i.CreateNewAccessKey(username)
	if err != nil {
		return err
	}
	return nil
}

func WriteToDisk(cred Credentials, repo, filename string, g GitRepoDetector) (err error) {
	b, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	path := filepath.Join(repo, cred.AccountAliasOrId, cred.IamUsername)
	os.MkdirAll(path, 0700)
	err = ioutil.WriteFile(filepath.Join(path, filename), b, 0600)
	if err != nil {
		return err
	}
	isrepo, err := g.IsGitRepo(repo)
	if err != nil {
		return err
	}
	if !isrepo {
		return nil
	}
	relpath := filepath.Join(cred.AccountAliasOrId, cred.IamUsername, filename)
	_, err = g.GitAddCommitFile(repo, relpath, "Added by Credulous", RepoConfig{Name: cred.IamUsername})
	if err != nil {
		return err
	}
	return nil
}

func RetrieveCredentials(i Credulousier, rootPath string, alias string, username string, keyfile string) (Credentials, error) {
	rootDir, err := os.Open(rootPath)
	if err != nil {
		handler.LogAndDieOnFatalError(err)
	}

	if alias == "" {
		if alias, err = i.FindDefaultDir(rootDir); err != nil {
			handler.LogAndDieOnFatalError(err)
		}
	}

	if username == "" {
		aliasDir, err := os.Open(filepath.Join(rootPath, alias))
		if err != nil {
			handler.LogAndDieOnFatalError(err)
		}
		username, err = i.FindDefaultDir(aliasDir)
		if err != nil {
			handler.LogAndDieOnFatalError(err)
		}
	}

	fullPath := filepath.Join(rootPath, alias, username)
	latest, err := i.LatestFileInDir(fullPath)
	if err != nil {
		return Credentials{}, err
	}
	filePath := filepath.Join(fullPath, latest.Name())
	cred, err := i.ReadCredentialFile(i, filePath, keyfile)
	if err != nil {
		return Credentials{}, err
	}

	return *cred, nil
}

func ListAvailableCredentials(c CredsReadWriter, rootDir FileLister) ([]string, error) {
	creds := make(map[string]int)

	repoDirs, err := c.GetDirs(rootDir)
	if err != nil {
		return nil, err
	}

	if len(repoDirs) == 0 {
		return nil, fmt.Errorf("no saved credentials found; please run 'credulous save' first")
	}

	for _, repoDirent := range repoDirs {
		repoPath := filepath.Join(rootDir.Name(), repoDirent.Name())
		repoDir, err := os.Open(repoPath)
		if err != nil {
			return nil, err
		}

		aliasDirs, err := c.GetDirs(repoDir)
		if err != nil {
			return nil, err
		}

		for _, aliasDirent := range aliasDirs {
			if aliasDirent.Name() == ".cgit" {
				continue
			}
			aliasPath := filepath.Join(repoPath, aliasDirent.Name())
			aliasDir, err := os.Open(aliasPath)
			if err != nil {
				return nil, err
			}

			userDirs, err := c.GetDirs(aliasDir)
			if err != nil {
				return nil, err
			}

			for _, userDirent := range userDirs {
				userPath := filepath.Join(aliasPath, userDirent.Name())
				latest, err := c.LatestFileInDir(userPath)
				if err != nil {
					return nil, err
				}
				if latest.Name() != "" {
					creds[userDirent.Name()+"@"+aliasDirent.Name()] += 1
				}
			}
		}
	}

	names := make([]string, len(creds))
	i := 0
	for k := range creds {
		names[i] = k
		i++
	}
	sort.Strings(names)
	return names, nil
}
