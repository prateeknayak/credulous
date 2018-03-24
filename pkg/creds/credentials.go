package creds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"crypto/rsa"

	"path/filepath"

	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/realestate-com-au/credulous/pkg/handler"
	"github.com/realestate-com-au/credulous/pkg/models"
)

type EncodeDecodeCreds struct{}

func NewEncodeDecodeCreds() *EncodeDecodeCreds {
	return &EncodeDecodeCreds{}
}

func (e *EncodeDecodeCreds) ReadCredentials(c core.CredsRetriever, b []byte, fp string, privKey *rsa.PrivateKey) (*models.Credentials, error) {
	var creds models.Credentials
	err := json.Unmarshal(b, &creds)
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
		tmp, err = c.DecodeAES(creds.Encryptions[offset].Ciphertext, privKey)
		if err != nil {
			return nil, err
		}
	}

	var cred models.Credential
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

func (e *EncodeDecodeCreds) FindDefaultDir(rootPath string) (string, error) {
	rootDir, err := os.Open(rootPath)
	defer rootDir.Close()

	if err != nil {
		handler.LogAndDieOnFatalError(err)
	}

	dirs, err := e.GetDirs(rootDir)
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

func (e *EncodeDecodeCreds) GetKey(name string) (filename string) {
	if name == "" {
		filename = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
	} else {
		filename = name
	}
	return filename
}
