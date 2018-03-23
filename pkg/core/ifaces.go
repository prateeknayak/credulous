package core

import (
	"crypto/rsa"
	"os"
	"time"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

type FileLister interface {
	Readdir(int) ([]os.FileInfo, error)
	Name() string
}

// Account informer interfaces
type UsernameGetter interface {
	GetUsername() (string, error)
}

type AliasGetter interface {
	GetAlias() (string, error)
}

type AllAccessKeyGetter interface {
	GetAllAccessKeys(username string) ([]AccessKey, error)
}

type AccessKeyDeleter interface {
	DeleteAccessKey(key *AccessKey) error
}

type AccessKeyCreater interface {
	CreateAccessKey(username string) (*AccessKey, error)
}

type KeyCreationDateGetter interface {
	GetKeyCreationDate(username string) (time.Time, error)
}

type AccountInformer interface {
	UsernameGetter
	AliasGetter
	AllAccessKeyGetter
	AccessKeyDeleter
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

type SaltDecoder interface {
	DecodeWithSalt(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error)
}

type FingerprintGetter interface {
	SSHFingerprint(pubkey ssh.PublicKey) (fingerprint string)
}

type PrivateFingerprintGetter interface {
	SSHPrivateFingerprint(privkey rsa.PrivateKey) (fingerprint string, err error)
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
}

// Credentials Reader Wirter

type DefualtDirFinder interface {
	FindDefaultDir(fl FileLister) (string, error)
}

type DirGetter interface {
	GetDirs(fl FileLister) ([]os.FileInfo, error)
}

type FileGetter interface {
	LatestFileInDir(dir string) (os.FileInfo, error)
}

type CredsReader interface {
	ReadCredentials(c CryptoOperator, fileName string, keyfile string) (*Credentials, error)
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

// Git operation
type StoreDetector interface {
	IsValidStore(checkpath string) (bool, error)
}

type Persister interface {
	Persist(repopath, filename, message string, config RepoConfig) (commitId string, err error)
}

type CredentialStorer interface {
	StoreDetector
	Persister
}

// Args Parser interface

type ArgsParser interface {
	ParseArgs(c *cli.Context) (*SaveData, error)
}

type Credulousier interface {
	AccountInformer
	ArgsParser
	CredentialStorer
	CryptoOperator
	CredsReadWriter
}
