package ccrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"

	"encoding/pem"

	"crypto/x509"

	"os"

	"github.com/howeyc/gopass"
	"github.com/realestate-com-au/credulous/pkg/models"
	"golang.org/x/crypto/ssh"
)

type aesEncryption struct {
	EncodedKey string
	Ciphertext string
}

type Crypto struct{}

func NewCrypto() *Crypto {
	return &Crypto{}
}

const saltLength = 8

func (c *Crypto) GenerateSalt() (string, error) {

	b := make([]byte, saltLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoder := base64.StdEncoding
	encoded := make([]byte, encoder.EncodedLen(len(b)))
	encoder.Encode(encoded, b)

	return string(encoded), nil
}

// returns a base64 encoded ciphertext.
// OAEP can only encrypt plaintexts that are smaller than the key length; for
// a 1024-bit key, about 117 bytes. So instead, this function:
// * generates a random 32-byte symmetric key (randKey)
// * encrypts the plaintext with AES256 using that random symmetric key -> cipherText
// * encrypts the random symmetric key with the ssh PublicKey -> cipherKey
// * returns the base64-encoded marshalled JSON for the ciphertext and key
func (c *Crypto) Encode(plaintext string, pubkey ssh.PublicKey) (ciphertext string, err error) {
	rsaKey := sshPubkeyToRsaPubkey(pubkey)
	randKey := make([]byte, 32)
	_, err = rand.Read(randKey)
	if err != nil {
		return "", err
	}

	encoded, err := encodeAES(randKey, plaintext)
	if err != nil {
		return "", err
	}

	out, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &rsaKey, []byte(randKey), []byte("Credulous"))
	if err != nil {
		return "", err
	}
	cipherKey := base64.StdEncoding.EncodeToString(out)

	cipherStruct := aesEncryption{
		EncodedKey: cipherKey,
		Ciphertext: encoded,
	}

	tmp, err := json.Marshal(cipherStruct)
	if err != nil {
		return "", err
	}

	ciphertext = base64.StdEncoding.EncodeToString(tmp)

	return ciphertext, nil
}

func (c *Crypto) DecodeAES(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// pull apart the layers of base64-encoded JSON
	var encrypted aesEncryption
	err = json.Unmarshal(in, &encrypted)
	if err != nil {
		return "", err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(encrypted.EncodedKey)
	if err != nil {
		return "", err
	}

	// decrypt the AES key using the ssh private key
	aesKey, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, encryptedKey, []byte("Credulous"))
	if err != nil {
		return "", err
	}

	plaintext, err = decodeAES(aesKey, encrypted.Ciphertext)

	return plaintext, nil
}

func (c *Crypto) DecodePureRSA(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, in, []byte("Credulous"))
	if err != nil {
		return "", err
	}
	plaintext = string(out)
	return plaintext, nil
}

func (c *Crypto) DecodeWithSalt(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, in, []byte("Credulous"))
	if err != nil {
		return "", err
	}
	plaintext = strings.Replace(string(out), salt, "", 1)
	return plaintext, nil
}

func (c *Crypto) SSHFingerprint(pubkey ssh.PublicKey) (fingerprint string) {
	binary := pubkey.Marshal()
	hash := md5.Sum(binary)
	// now add the colons
	fingerprint = fmt.Sprintf("%02x", (hash[0]))
	for i := 1; i < len(hash); i += 1 {
		fingerprint += ":" + fmt.Sprintf("%02x", (hash[i]))
	}
	return fingerprint
}

func (c *Crypto) SSHPrivateFingerprint(privkey *rsa.PrivateKey) (fingerprint string, err error) {
	sshPubkey, err := rsaPubkeyToSSHPubkey(privkey.PublicKey)
	if err != nil {
		return "", err
	}
	fingerprint = c.SSHFingerprint(sshPubkey)
	return fingerprint, nil
}

func (c *Crypto) ParseKey(key models.PrivateKey) (*rsa.PrivateKey, error) {
	var keyInBytes []byte
	var err error

	pemblock, _ := pem.Decode(key.Bytes)

	if x509.IsEncryptedPEMBlock(pemblock) {

		if _, err = fmt.Fprintf(os.Stderr, "Enter passphrase for %s: ", key.Name); err != nil {
			return nil, err
		}
		passwd, err := gopass.GetPasswd()
		decryptedBytes, err := x509.DecryptPEMBlock(pemblock, passwd)
		if err != nil {
			return nil, err
		}

		pemBytes := pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: decryptedBytes,
		}
		keyInBytes = pem.EncodeToMemory(&pemBytes)

	} else {
		keyInBytes = key.Bytes
	}

	k, err := ssh.ParseRawPrivateKey(keyInBytes)
	if err != nil {
		return nil, err
	}
	return k.(*rsa.PrivateKey), nil
}

func sshPubkeyToRsaPubkey(pubkey ssh.PublicKey) rsa.PublicKey {
	s := reflect.ValueOf(pubkey).Elem()
	rsaKey := rsa.PublicKey{
		N: s.Field(0).Interface().(*big.Int),
		E: s.Field(1).Interface().(int),
	}
	return rsaKey
}

func rsaPubkeyToSSHPubkey(rsakey rsa.PublicKey) (sshkey ssh.PublicKey, err error) {
	sshkey, err = ssh.NewPublicKey(&rsakey)
	if err != nil {
		return nil, err
	}
	return sshkey, nil
}

func encodeAES(key []byte, plaintext string) (ciphertext string, err error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// We need an unique IV to go at the front of the ciphertext
	out := make([]byte, aes.BlockSize+len(plaintext))
	iv := out[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(cipherBlock, iv)
	stream.XORKeyStream(out[aes.BlockSize:], []byte(plaintext))
	encoded := base64.StdEncoding.EncodeToString(out)
	return encoded, nil
}

// takes a base64-encoded AES-encrypted ciphertext
func decodeAES(key []byte, ciphertext string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	decrypter, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := encrypted[:aes.BlockSize]
	msg := encrypted[aes.BlockSize:]
	aesDecrypter := cipher.NewCFBDecrypter(decrypter, iv)
	aesDecrypter.XORKeyStream(msg, msg)
	return string(msg), nil
}
