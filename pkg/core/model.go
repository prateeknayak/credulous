package core

import (
	"time"

	"golang.org/x/crypto/ssh"
)

const ROTATE_TIMEOUT int = 30
const FORMAT_VERSION string = "2014-06-12"

type AccessKey struct {
	Username   string
	KeyId      string
	Status     string
	CreateDate time.Time
	Secret     string
}
type CLIArgs struct {
	Cred     Credential
	Username string
	Account  string
	Pubkeys  []ssh.PublicKey
	Lifetime int
	Repo     string
}

type Credential struct {
	KeyId     string
	SecretKey string
	EnvVars   map[string]string
}

type OldCredential struct {
	CreateTime       string
	LifeTime         int
	KeyId            string
	SecretKey        string
	Salt             string
	AccountAliasOrId string
	IamUsername      string
	FingerPrint      string
}

type SaveData struct {
	Cred     Credential
	Username string
	Alias    string
	Pubkeys  []ssh.PublicKey
	Lifetime int
	Force    bool
	Repo     string
	IsRepo   bool
}

type RepoConfig struct {
	Name  string
	Email string
}

type Credentials struct {
	Version          string
	IamUsername      string
	AccountAliasOrId string
	CreateTime       string
	LifeTime         int
	Encryptions      []Encryption
}
type Encryption struct {
	Fingerprint string
	Ciphertext  string
	Decoded     Credential
}
