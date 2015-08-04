package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
)

// List of valid SSH format strings, taken from
// https://www.iana.org/assignments/ssh-parameters/ssh-parameters-19.csv
// on 2015-07-06
var validFormats = [11]string{
	"ssh-dss",
	"ssh-rsa",
	"spki-sign-rsa",
	"spki-sign-dss",
	"pgp-sign-rsa",
	"pgp-sign-dss",
	"null",
	"x509v3-ssh-dss",
	"x509v3-ssh-rsa",
	"x509v3-rsa2048-sha256",
}

// List of valid SSH format string prefixes, taken from
// https://www.iana.org/assignments/ssh-parameters/ssh-parameters-19.csv
// on 2015-07-06
var validFormatPrefixes = [2]string{
	"ecdsa-sha2-",
	"x509v3-ecdsa-sha2-",
}

// see flag calls in init()
var secretFile, username, format, fingerprint, key string
var add bool

const (
	// Kind to use for authorized keys when accessing datastore
	keyKind = "authorizedKeys"
	// Kind to use for coders when accessing datastore
	coderKind = "coders"
)

// Coder holds a coder with a nickname
type Coder struct {
	Nickname string
}

// AuthorizedKey holds a single public key (and it's format and fingerprint)
// for use with SSH. It refers to an object inside datastore that identifies
// the user in possession of the key.
type AuthorizedKey struct {
	Fingerprint []byte
	Format      string `datastore:",noindex"`
	Key         []byte `datastore:",noindex"`
	Coder       *datastore.Key
}

// String generates a string representation of an authorized key that conforms
// to the format of authorized key files as seen with sshd. It omits the key's
// fingerprint and does not resolve the coder.
func (key AuthorizedKey) String() string {
	return key.Format + " " + base64.StdEncoding.EncodeToString(key.Key)
}

// NewAuthorizedKey creates a new AuthorizedKey entry from it's fingerprint (bytes as
// hexadecimal numbers separated by colons), a format (registered at IANA) and a key
// (Base64 encoded).
func NewAuthorizedKey(fingerprint, format, key string) (result AuthorizedKey, err error) {
	if !isValidFormat(format) {
		err = errors.New("invalid format string")
		return
	}

	rawFingerprint, err := decodeFingerprint(fingerprint)
	if err != nil {
		return
	}

	rawKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return
	}

	result = AuthorizedKey{
		Fingerprint: rawFingerprint,
		Format:      format,
		Key:         rawKey,
	}
	return
}

// decodeFingerprint strips the colon from a fingerprint string
// and translates it to bytes.
func decodeFingerprint(fingerprint string) ([]byte, error) {
	return hex.DecodeString(strings.Replace(fingerprint, ":", "", -1))
}

// isValidFormat checks whether the passed SSH key format
// is registered with IANA, i.e. listed in
// http://www.iana.org/assignments/ssh-parameters/ssh-parameters.xhtml#ssh-parameters-19
func isValidFormat(format string) bool {
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	if strings.Contains(format, "@") {
		return false
	}
	for _, validFormatPrefix := range validFormatPrefixes {
		if strings.HasPrefix(format, validFormatPrefix) {
			return true
		}
	}
	return false
}

func init() {
	flag.StringVar(&secretFile, "s", "secret.json", "file name of secret")
	flag.StringVar(&username, "u", "", "the username being authenticated/added")
	flag.StringVar(&format, "t", "", "key type offered for authentication")
	flag.StringVar(&fingerprint, "f", "", "fingerprint of the key")
	flag.StringVar(&key, "k", "", "key being offered for authentication")
	flag.BoolVar(&add, "a", false, "add key")
	flag.Parse()
}

func main() {
	if add {
		key, err := NewAuthorizedKey(fingerprint, format, key)
		if err != nil {
			panic(err)
		}

		err = putAuthorizedKey(&key, username)
		if err != nil {
			panic(err)
		}
		return
	}

	if username != "git" {
		fmt.Fprintln(os.Stderr, "Won't accept any other user than \"git\".")
		os.Exit(1)
	}

	key, username, err := query()

	if err != nil {
		panic(err)
	}

	env := fmt.Sprintf(`environment="GITHUB_USERNAME=%s"`, username)
	fmt.Printf("%s %s\n", env, key.String())
}

func query() (key *AuthorizedKey, username string, err error) {
	ctx, err := connect()
	if err != nil {
		return
	}

	rawFingerprint, err := decodeFingerprint(fingerprint)
	if err != nil {
		return
	}

	key = new(AuthorizedKey)

	it := datastore.NewQuery(keyKind).Filter("Fingerprint =", rawFingerprint).Limit(1).Run(ctx)

	_, err = it.Next(key)
	if err != nil {
		return
	}

	var coder Coder
	err = datastore.Get(ctx, key.Coder, &coder)
	if err != nil {
		return
	}

	username = coder.Nickname
	return
}

func putAuthorizedKey(ak *AuthorizedKey, username string) (err error) {
	ctx, err := connect()
	if err != nil {
		return
	}

	var coder Coder
	it := datastore.NewQuery(coderKind).Filter("Nickname =", username).Limit(1).Run(ctx)
	ck, err := it.Next(&coder)
	if err != nil {
		return
	}

	ak.Coder = ck

	dk := datastore.NewIncompleteKey(ctx, keyKind, nil)
	_, err = datastore.Put(ctx, dk, ak)

	return
}

func connect() (ctx context.Context, err error) {
	secret, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return
	}

	config, err := google.JWTConfigFromJSON(secret, datastore.ScopeDatastore, datastore.ScopeUserEmail)
	if err != nil {
		return
	}

	ctx = cloud.NewContext("coduno", config.Client(oauth2.NoContext))
	return
}
