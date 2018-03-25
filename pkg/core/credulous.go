package core

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"os"
	"path/filepath"

	"sort"

	"encoding/pem"

	"github.com/realestate-com-au/credulous/pkg/models"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

type Writer interface {
	Write(b []byte, path, filename string) error
}

type Reader interface {
	Read(filename string) ([]byte, error)
}

type FileLister interface {
	Readdir(int) ([]os.FileInfo, error)
	Name() string
}

type Displayer interface {
	Display(message string)
}

// Account informer interfaces
type UsernameGetter interface {
	GetUsername() (string, error)
}

type AliasGetter interface {
	GetAlias() (string, error)
}

type AllAccessKeyGetter interface {
	GetAllAccessKeys(username string) ([]models.AccessKey, error)
}

type AccessKeyDeleter interface {
	DeleteAccessKey(key *models.AccessKey) error
}

type AccessKeyCreater interface {
	CreateAccessKey(username string) (*models.AccessKey, error)
}

type KeyCreationDateGetter interface {
	GetKeyCreationDate(username string) (time.Time, error)
}
type UsernameAliasGetter interface {
	UsernameGetter
	AliasGetter
}

type OneKeyDeleter interface {
	AllAccessKeyGetter
	AccessKeyDeleter
}

type AccountInformer interface {
	UsernameAliasGetter
	OneKeyDeleter
	AccessKeyCreater
	KeyCreationDateGetter
}

// Crypto Operator interfaces

type Encoder interface {
	Encode(plaintext string, pubkey ssh.PublicKey) (ciphertext string, err error)
}

type AESDecoder interface {
	DecodeAES(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error)
}

type PureRSADecoder interface {
	DecodePureRSA(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error)
}

type PemBlockDecoder interface {
	DecodePemBlock(pemblock *pem.Block, passwd []byte) ([]byte, error)
}
type SaltDecoder interface {
	DecodeWithSalt(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error)
}

type FingerprintGetter interface {
	SSHFingerprint(pubkey ssh.PublicKey) (fingerprint string)
}

type PrivateFingerprintGetter interface {
	SSHPrivateFingerprint(privkey *rsa.PrivateKey) (fingerprint string, err error)
}

type RawKeyParser interface {
	ParseKey(key models.PrivateKey) (*rsa.PrivateKey, error)
}
type SaltGenerator interface {
	GenerateSalt() (string, error)
}

type CryptoOperator interface {
	Encoder
	AESDecoder
	PureRSADecoder
	SaltDecoder
	FingerprintGetter
	PrivateFingerprintGetter
	SaltGenerator
	RawKeyParser
}

// Credentials Reader Wirter

type DefualtDirFinder interface {
	FindDefaultDir(rootPath string) (string, error)
}

type DirGetter interface {
	GetDirs(fl FileLister) ([]os.FileInfo, error)
}

type FileGetter interface {
	LatestFileInDir(dir string) (*models.CredsInfo, error)
}

type CredsReader interface {
	ReadCredentials(c CredsRetriever, b []byte, fp string, privKey *rsa.PrivateKey) (*models.Credentials, error)
}

type KeyGetter interface {
	GetKey(key string) (filename string)
}

type CredsReadWriter interface {
	DefualtDirFinder
	DirGetter
	FileGetter
	CredsReader
	KeyGetter
}

type CredsRetriever interface {
	FileGetter
	CredsReader
	PrivateFingerprintGetter
	AESDecoder
	Reader
	RawKeyParser
}

// Git operation
type StoreDetector interface {
	IsValidStore(checkpath string) (bool, error)
}

type Persister interface {
	Persist(repopath, filename, message string, config models.RepoConfig) (commitId string, err error)
}

type CredentialStorer interface {
	StoreDetector
	Persister
}

// Args Parser interface

type ArgsParser interface {
	ParseArgs(c *cli.Context) (*models.SaveData, error)
}

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

// Credulous APIs called from main.go
func Save(i Credulousier, data *models.SaveData) error {

	var keyCreateDate int64

	if data.Force {
		keyCreateDate = time.Now().Unix()
	} else {
		var err error
		if data.Username == "" {
			data.Username, err = i.GetUsername()
			if err != nil {
				return err
			}
		}
		if data.Alias == "" {
			data.Alias, err = i.GetAlias()
			if err != nil {
				return err
			}
		}

		date, err := i.GetKeyCreationDate(data.Username)
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

	encSlice := []models.Encryption{}
	for _, pubkey := range data.Pubkeys {
		encoded, err := i.Encode(string(plaintext), pubkey)
		if err != nil {
			return err
		}

		encSlice = append(encSlice, models.Encryption{
			Ciphertext:  encoded,
			Fingerprint: i.SSHFingerprint(pubkey),
		})
	}
	creds := models.Credentials{
		Version:          models.FORMAT_VERSION,
		AccountAliasOrId: data.Alias,
		IamUsername:      data.Username,
		CreateTime:       fmt.Sprintf("%d", keyCreateDate),
		Encryptions:      encSlice,
		LifeTime:         data.Lifetime,
	}

	filename := fmt.Sprintf("%v-%v.json", keyCreateDate, data.Cred.KeyId[12:])

	err = writeCredentialsTofile(i, creds, data.Repo, filename)
	if err != nil {
		return err
	}

	err = storeCreds(i, data.Username, data.Repo, filepath.Join(data.Username, data.Alias, filename))
	return err
}

func GetUsernameAndAlias(i UsernameAliasGetter) (string, string, error) {
	username, err := i.GetUsername()
	if err != nil {
		return "", "", err
	}

	alias, err := i.GetAlias()
	if err != nil {
		return "", "", err
	}

	return username, alias, nil
}

func VerifyAccount(i AliasGetter, alias string) error {
	acctAlias, err := i.GetAlias()
	if err != nil {
		return err
	}
	if acctAlias == alias {
		return nil
	}
	return fmt.Errorf("cannot verify account, does not match alias: %s", alias)
}

func VerifyUser(i UsernameGetter, username string) error {
	name, err := i.GetUsername()
	if err != nil {
		return err
	}
	if username == name {
		return nil
	}
	return fmt.Errorf("cannot verify user, does not match access keys: %s", username)
}

func ValidateCredentials(i UsernameAliasGetter, cred models.Credentials, alias string, username string) error {
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

func DisplayCredentials(i Displayer, cred models.Credentials) {

	i.Display(fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=\"%v\"\nexport AWS_SECRET_ACCESS_KEY=\"%v\"\n",
		cred.Encryptions[0].Decoded.KeyId,
		cred.Encryptions[0].Decoded.SecretKey,
	))

	for key, val := range cred.Encryptions[0].Decoded.EnvVars {
		i.Display(fmt.Sprintf(
			"export %s=\"%s\"\n",
			key,
			val,
		))
	}
}

func DeleteOneKey(i OneKeyDeleter, username string) error {
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
func CreateKey(i AccessKeyCreater, username string) error {
	_, err := i.CreateAccessKey(username)
	return err
}

func RetrieveCredentials(i CredsRetriever, req *models.RetrieveRequest) (c *models.Credentials, err error) {

	latest, err := i.LatestFileInDir(req.FullPath)
	if err != nil {
		return
	}

	rawKey, err := i.Read(req.Keyfile)
	if err != nil {
		return
	}

	privKey, err := i.ParseKey(models.PrivateKey{Bytes: rawKey, Name: req.Keyfile})
	if err != nil {
		return
	}

	fp, err := i.SSHPrivateFingerprint(privKey)
	if err != nil {
		return
	}

	filePath := filepath.Join(req.FullPath, latest.Name)

	b, err := i.Read(filePath)
	if err != nil {
		return
	}

	cred, err := i.ReadCredentials(i, b, fp, privKey)
	if err != nil {
		return
	}

	return cred, nil
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
				if latest.Name != "" {
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

func storeCreds(i CredentialStorer, user, repo, path string) (err error) {
	isRepo, err := i.IsValidStore(repo)
	if err != nil {
		return err
	}
	if !isRepo {
		return nil
	}
	_, err = i.Persist(repo, path, "Added by Credulous", models.RepoConfig{Name: user})
	return err
}

func writeCredentialsTofile(i Writer, cred models.Credentials, repo, filename string) error {
	b, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	path := filepath.Join(repo, cred.AccountAliasOrId, cred.IamUsername)
	return i.Write(b, path, filename)
}
