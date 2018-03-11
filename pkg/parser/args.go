package parser

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/realestate-com-au/credulous/pkg/core"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

const pattern = "^[A-Za-z_][A-Za-z0-9_]*=.*"

type parser struct{}

func NewParser() *parser {
	return &parser{}
}
func (p *parser) ParseArgs(c *cli.Context) (*core.SaveData, error) {
	pubkeys, err := parseKeyArgs(c)
	if err != nil {
		return nil, err
	}

	username, alias, err := parseUserAndAlias(c)
	if err != nil {
		return nil, err
	}

	envmap, err := parseEnvironmentArgs(c)
	if err != nil {
		return nil, err
	}

	lifetime, err := parseLifetimeArgs(c)
	if err != nil {
		return nil, err
	}

	repo, err := parseRepoArgs(c)
	if err != nil {
		return nil, err
	}

	AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if AWSAccessKeyId == "" || AWSSecretAccessKey == "" {
		err = fmt.Errorf("can't save, no credentials in the environment")
		if err != nil {
			return nil, err
		}
	}
	cred := core.Credential{
		KeyId:     AWSAccessKeyId,
		SecretKey: AWSSecretAccessKey,
		EnvVars:   envmap,
	}

	return &core.SaveData{
		Cred:     cred,
		Username: username,
		Alias:    alias,
		Pubkeys:  pubkeys,
		Lifetime: lifetime,
		Repo:     repo,
	}, nil
}

func parseUserAndAlias(c *cli.Context) (username string, alias string, err error) {

	if (c.String("username") == "" || c.String("account") == "") && c.Bool("force") {
		err = fmt.Errorf("must specify both username and account with force")
		return
	}

	// if username OR account were specified, but not both, complain
	if (c.String("username") != "" && c.String("account") == "") || (c.String("username") == "" && c.String("account") != "") {

		if c.Bool("force") {
			err = fmt.Errorf("must specify both username and account for force save")
		} else {
			err = fmt.Errorf("must use force save when specifying username or account")
		}
		return
	}

	// if username/account were specified, but force wasn't set, complain
	if c.String("username") != "" && c.String("account") != "" {
		if !c.Bool("force") {
			err = fmt.Errorf("cannot specify username and/or account without force")
			return
		} else {
			log.Print("WARNING: saving credentials without verifying username or account alias")
			username = c.String("username")
			alias = c.String("account")
		}
	}
	return
}

func parseEnvironmentArgs(c *cli.Context) (map[string]string, error) {
	if len(c.StringSlice("env")) == 0 {
		return nil, nil
	}

	envMap := make(map[string]string)
	for _, arg := range c.StringSlice("env") {
		match, err := regexp.Match(pattern, []byte(arg))
		if err != nil {
			return nil, err
		}
		if !match {
			log.Print("WARNING: Skipping env argument " + arg + " -- not in NAME=value format")
			continue
		}
		parts := strings.SplitN(arg, "=", 2)
		envMap[parts[0]] = parts[1]
	}
	return envMap, nil
}

func readSSHPubkeyFile(filename string) (pubkey ssh.PublicKey, err error) {
	pubkeyString, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	pubkey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(pubkeyString))
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func parseKeyArgs(c *cli.Context) (pubkeys []ssh.PublicKey, err error) {
	// no args, so just use the default
	if len(c.StringSlice("key")) == 0 {
		pubkey, err := readSSHPubkeyFile(filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa.pub"))
		if err != nil {
			return nil, err
		}
		pubkeys = append(pubkeys, pubkey)
		return pubkeys, nil
	}

	for _, arg := range c.StringSlice("key") {
		pubkey, err := readSSHPubkeyFile(arg)
		if err != nil {
			return nil, err
		}
		pubkeys = append(pubkeys, pubkey)
	}
	return pubkeys, nil
}

// parseLifetimeArgs attempts to be a little clever in determining what credential
// lifetime you've chosen. It returns a number of hours and an error. It assumes that
// the argument was passed in as hours.
func parseLifetimeArgs(c *cli.Context) (lifetime int, err error) {
	// the default is zero, which is our default
	if c.Int("lifetime") < 0 {
		return 0, nil
	}

	return c.Int("lifetime"), nil
}

func parseRepoArgs(c *cli.Context) (repo string, err error) {
	// the default is 'local' which is set below, so not much to do here
	if c.String("repo") == "local" {
		repo = path.Join(getRootPath(), "local")
	} else {
		repo = c.String("repo")
	}
	return repo, nil
}

func getRootPath() string {
	home := os.Getenv("HOME")
	rootPath := home + "/.credulous"
	os.MkdirAll(rootPath, 0700)
	return rootPath
}
