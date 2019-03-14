package wallet

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gertjaap/vertcoin/logging"

	"github.com/btcsuite/btcd/chaincfg"

	"github.com/btcsuite/btcutil/hdkeychain"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

type Key struct {
	passwordChannel       chan PasswordPrompt     // A channel passed in to allow password prompts when needed
	rootKey               []byte                  // Stores the root key, encrypted. Need to unlock it with relevant operations
	plainRootPub          *hdkeychain.ExtendedKey // Stores the pubkey of the root plainly so we can derive pubkeys without needing the password
	nonInteractiveRootKey *hdkeychain.ExtendedKey // Stores a derived root key from the root plain for online operations such as Lightning
}

type PasswordPrompt struct {
	Reason          string
	Confirm         bool
	ResponseChannel chan string
}

func NewKey(keyFile string, passChan chan PasswordPrompt) (*Key, error) {
	key := new(Key)
	key.passwordChannel = passChan
	var err error
	var privKey [32]byte
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		rand.Read(privKey[:])
		salt := new([24]byte) // salt for scrypt / nonce for secretbox
		dk32 := new([32]byte) // derived key from scrypt

		//get 24 random bytes for scrypt salt (and secretbox nonce)
		_, err := rand.Read(salt[:])
		if err != nil {
			return nil, err
		}

		// Get password to use
		pass := ""
		passwordPrompt := PasswordPrompt{
			ResponseChannel: make(chan string),
			Reason:          "Encrypt new keyfile",
			Confirm:         true,
		}

		passChan <- passwordPrompt
		for {
			pass = <-passwordPrompt.ResponseChannel
			if pass == "" {
				passwordPrompt.Reason += " - Password cannot be empty!"
				passChan <- passwordPrompt
			} else {
				break
			}
		}

		// next use the pass and salt to make a 32-byte derived key
		dk, err := scrypt.Key([]byte(pass), salt[:], 16384, 8, 1, 32)
		if err != nil {
			return nil, err
		}
		copy(dk32[:], dk[:])

		enckey := append(salt[:], secretbox.Seal(nil, privKey[:], salt, dk32)...)
		key.rootKey = make([]byte, len(enckey))
		copy(key.rootKey[:], enckey[:])
		err = ioutil.WriteFile(keyFile, enckey[:], 0600)
		if err != nil {
			return nil, err
		}
	} else {
		bytes, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return nil, err
		}
		key.rootKey = make([]byte, len(bytes))
		copy(key.rootKey[:], bytes[:])
	}

	_, err = key.UnlockedKeyStatement("Open Wallet", func(rootKey *hdkeychain.ExtendedKey) (interface{}, error) {
		var err error
		key.nonInteractiveRootKey, err = rootKey.Child(20987398)
		if err != nil {
			return nil, err
		}
		key.plainRootPub, err = rootKey.Neuter()
		return nil, err
	})

	if err != nil {
		return nil, err
	}

	return key, nil
}

func (k *Key) UnlockedKeyStatement(reason string, f func(rootKey *hdkeychain.ExtendedKey) (interface{}, error)) (interface{}, error) {
	passwordPrompt := PasswordPrompt{
		ResponseChannel: make(chan string),
		Reason:          reason,
	}

	k.passwordChannel <- passwordPrompt
	pass := <-passwordPrompt.ResponseChannel

	logging.Debugf("Encrypted key length: %d", len(k.rootKey))

	salt := new([24]byte)
	copy(salt[:], k.rootKey[:24])

	// enckey is actually encrypted, get derived key from pass and salt
	dk, err := scrypt.Key([]byte(pass), salt[:], 16384, 8, 1, 32) // derive key
	if err != nil {
		return nil, err
	}

	dk32 := new([32]byte)
	copy(dk32[:], dk)

	// nonce for secretbox is the same as scrypt salt.  Seems fine.  Really.
	priv, worked := secretbox.Open(nil, k.rootKey[24:], salt, dk32)
	if worked != true {
		return nil, fmt.Errorf("Unable to decrypt private key with the given password")
	}

	// TODO change this to a non-coin
	rootPrivKey, err := hdkeychain.NewMaster(priv[:], &chaincfg.TestNet3Params)
	if err != nil {
		return nil, err
	}

	priv = nil // this probably doesn't do anything but... eh why not
	returnValue, err := f(rootPrivKey)
	rootPrivKey.Zero()
	rootPrivKey = nil

	return returnValue, err

}
