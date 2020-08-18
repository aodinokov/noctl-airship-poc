package crypter

import (
	"fmt"
	"time"

	"crypto/sha256"
	"encoding/base64"

	"github.com/fernet/fernet-go"
	"golang.org/x/crypto/pbkdf2"
)

func Key(password string) (string, error) {
	dk := pbkdf2.Key([]byte(password), []byte("airshipit.org/salt"), 100000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(dk), nil
}

func Decrypt(in string, key string) (string, error) {
	k := fernet.MustDecodeKeys(key)

	tok, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "", fmt.Errorf("wasn't able to decode64 string %s", in)
	}

	b := fernet.VerifyAndDecrypt(tok, 0*time.Second, k)
	if b == nil {
		return "", fmt.Errorf("wasn't able to decrypt string %s", in)
	}

	db, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return "", fmt.Errorf("wasn't able to decode base64 string %s", db)
	}

	return string(db), nil
}

func Encrypt(in string, key string) (string, error) {
	k := fernet.MustDecodeKeys(key)

	eb := base64.StdEncoding.EncodeToString([]byte(in))

	tok, err := fernet.EncryptAndSign([]byte(eb), k[0])
	if err != nil {
		return "", fmt.Errorf("wasn't able to encrypt string %s", in)
	}

	return base64.StdEncoding.EncodeToString(tok), nil
}
