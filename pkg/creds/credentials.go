package creds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"path/filepath"

	"github.com/howeyc/gopass"
	"github.com/prateeknayak/credulous/pkg/core"
	"github.com/prateeknayak/credulous/pkg/handler"
	"golang.org/x/crypto/ssh"
)

type EncodeDecodeCreds struct{}

func NewEncodeDecodeCreds() *EncodeDecodeCreds {
	return &EncodeDecodeCreds{}
}

func (e *EncodeDecodeCreds) ReadCredentialFile(c core.CryptoOperator, fileName string, keyfile string) (*core.Credentials, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var creds core.Credentials
	err = json.Unmarshal(b, &creds)
	if err != nil {
		return nil, err
	}

	privKey, err := loadPrivateKey(keyfile)
	if err != nil {
		return nil, err
	}

	fp, err := c.SSHPrivateFingerprint(*privKey)
	if err != nil {
		return nil, err
	}

	var offset int = -1
	for i, enc := range creds.Encryptions {
		if enc.Fingerprint == fp {
			offset = i
			break
		}
	}

	if offset < 0 {
		err := errors.New("the SSH key specified cannot decrypt those credentials")
		return nil, err
	}

	var tmp string
	switch {
	case creds.Version == "2014-06-12":
		tmp, err = c.CredulousDecodeAES(creds.Encryptions[offset].Ciphertext, privKey)
		if err != nil {
			return nil, err
		}
	}

	var cred core.Credential
	err = json.Unmarshal([]byte(tmp), &cred)
	if err != nil {
		return nil, err
	}

	creds.Encryptions[0].Decoded = cred

	return &creds, nil
}

func (e *EncodeDecodeCreds) GetDirs(fl core.FileLister) ([]os.FileInfo, error) {
	dirents, err := fl.Readdir(0) // get all the entries
	if err != nil {
		return nil, err
	}

	dirs := []os.FileInfo{}
	for _, dirent := range dirents {
		if dirent.IsDir() {
			dirs = append(dirs, dirent)
		}
	}

	return dirs, nil
}

func (e *EncodeDecodeCreds) FindDefaultDir(fl core.FileLister) (string, error) {
	dirs, err := e.GetDirs(fl)
	if err != nil {
		return "", err
	}

	switch {
	case len(dirs) == 0:
		return "", fmt.Errorf("no saved credentials found; please run 'credulous save' first")
	case len(dirs) > 1:
		return "", fmt.Errorf("more than one account found; please specify account and user")
	}

	return dirs[0].Name(), nil
}

func (e *EncodeDecodeCreds) LatestFileInDir(dir string) (os.FileInfo, error) {
	entries, err := ioutil.ReadDir(dir)
	handler.LogAndDieOnFatalError(err)
	if len(entries) == 0 {
		return nil, fmt.Errorf("no credentials have been saved for that user and account; please run 'credulous save' first")
	}
	return entries[len(entries)-1], nil
}

func (e *EncodeDecodeCreds) GetPrivateKey(name string) (filename string) {
	if name == "" {
		filename = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
	} else {
		filename = name
	}
	return filename
}

func loadPrivateKey(filename string) (privateKey *rsa.PrivateKey, err error) {
	var tmp []byte

	if tmp, err = ioutil.ReadFile(filename); err != nil {
		return &rsa.PrivateKey{}, err
	}

	pemblock, _ := pem.Decode([]byte(tmp))
	if x509.IsEncryptedPEMBlock(pemblock) {
		if tmp, err = decryptPEM(pemblock, filename); err != nil {
			return &rsa.PrivateKey{}, err
		}
	} else {
		log.Print("WARNING: Your private SSH key has no passphrase!")
	}

	key, err := ssh.ParseRawPrivateKey(tmp)
	if err != nil {
		return &rsa.PrivateKey{}, err
	}
	privateKey = key.(*rsa.PrivateKey)
	return privateKey, nil
}

func decryptPEM(pemblock *pem.Block, filename string) ([]byte, error) {
	var err error
	if _, err = fmt.Fprintf(os.Stderr, "Enter passphrase for %s: ", filename); err != nil {
		return []byte(""), err
	}

	// we already emit the prompt to stderr; GetPass only emits to stdout
	passwd, err := gopass.GetPasswd()
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return []byte(""), err
	}

	var decryptedBytes []byte
	if decryptedBytes, err = x509.DecryptPEMBlock(pemblock, passwd); err != nil {
		return []byte(""), err
	}

	pemBytes := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: decryptedBytes,
	}
	decryptedPEM := pem.EncodeToMemory(&pemBytes)
	return decryptedPEM, nil
}
