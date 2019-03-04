package util

import (
	"strings"

	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/adiabat/bech32"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gertjaap/vertcoin-openassets/ecies"
)

func KeyHashFromPkScript(pkscript []byte) []byte {
	// match p2wpkh
	if len(pkscript) == 22 && pkscript[0] == 0x00 && pkscript[1] == 0x14 {
		return pkscript[2:]
	}

	// match p2wsh
	if len(pkscript) == 34 && pkscript[0] == 0x00 && pkscript[1] == 0x20 {
		return pkscript[2:]
	}

	return nil
}

func PrintTx(tx *wire.MsgTx) {
	var buf bytes.Buffer

	tx.Serialize(&buf)
	fmt.Printf("TX: %x\n", buf.Bytes())
}

func DirectWPKHScriptFromPKH(pkh [20]byte) []byte {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0).AddData(pkh[:])
	b, _ := builder.Script()
	return b
}

func DirectWSHScriptFromSH(sh [32]byte) []byte {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0).AddData(sh[:])
	b, _ := builder.Script()
	return b
}

func DirectWSHScriptFromAddress(adr string) ([]byte, error) {
	var scriptHash [32]byte
	decoded, err := bech32.SegWitAddressDecode(adr)
	if err != nil {
		return []byte{}, err
	}
	copy(scriptHash[:], decoded[2:]) // skip version and pushdata byte returned by SegWitAddressDecode
	return DirectWSHScriptFromSH(scriptHash), nil
}

func DirectWPKHScriptFromAddress(adr string) ([]byte, error) {
	var pubkeyHash [20]byte
	decoded, err := bech32.SegWitAddressDecode(adr)
	if err != nil {
		return []byte{}, err
	}
	copy(pubkeyHash[:], decoded[2:]) // skip version and pushdata byte returned by SegWitAddressDecode
	return DirectWPKHScriptFromPKH(pubkeyHash), nil
}

func IsConnectionError(err error) bool {
	if strings.Contains(err.Error(), "connection refused") {
		return true
	}

	if strings.Contains(err.Error(), "401") {
		return true
	}
	return false
}

func EncryptECIES(pubKey [33]byte, msg []byte) ([]byte, error) {
	key, err := btcec.ParsePubKey(pubKey[:], btcec.S256())
	if err != nil {
		return nil, err
	}

	eKey := ecies.ImportECDSAPublic(key.ToECDSA())
	if eKey == nil {
		return nil, fmt.Errorf("Unable to import ECDSA key")
	}

	return ecies.Encrypt(rand.Reader, eKey, msg, nil, nil)
}

func DecryptECIES(priv *btcec.PrivateKey, cipherText []byte) ([]byte, error) {
	eKey := ecies.ImportECDSA(priv.ToECDSA())
	if eKey == nil {
		return nil, fmt.Errorf("Unable to import ECDSA key")
	}

	return eKey.Decrypt(cipherText, nil, nil)
}

// CombinePrivateKeys takes a set of private keys and combines them in the same way
// as done for public keys.  This only works if you know *all* of the private keys.
// If you don't, we'll do something with returning a scalar coefficient...
// I don't know how that's going to work.  Schnorr stuff isn't decided yet.
func CombinePrivateKeys(keys ...*btcec.PrivateKey) *btcec.PrivateKey {

	if keys == nil || len(keys) == 0 {
		return nil
	}
	if len(keys) == 1 {
		return keys[0]
	}
	// bunch of keys
	var pubs CombinablePubKeySlice
	for _, k := range keys {
		pubs = append(pubs, k.PubKey())
	}
	z := pubs.ComboCommit()
	sum := new(big.Int)

	for _, k := range keys {
		h := chainhash.HashH(append(z[:], k.PubKey().SerializeCompressed()...))
		// turn coefficient hash h into a bigint
		hashInt := new(big.Int).SetBytes(h[:])
		// multiply the hash by the private scalar for this particular key
		hashInt.Mul(hashInt, k.D)
		// reduce mod curve N
		hashInt.Mod(hashInt, btcec.S256().N)
		// add this scalar to the aggregate and reduce the sum mod N again
		sum.Add(sum, hashInt)
		sum.Mod(sum, btcec.S256().N)
	}

	// kindof ugly that it's converting the bigint to bytes and back but whatever
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), sum.Bytes())
	return priv
}

// PubKeySlice are slices of pubkeys, which can be combined (and sorted)
type CombinablePubKeySlice []*btcec.PublicKey

// Make PubKeySlices sortable
func (p CombinablePubKeySlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p CombinablePubKeySlice) Len() int      { return len(p) }
func (p CombinablePubKeySlice) Less(i, j int) bool {
	return bytes.Compare(p[i].SerializeCompressed(), p[j].SerializeCompressed()) == -1
}

// ComboCommit generates the "combination commitment" which contributes to the
// hash-coefficient for every key being combined.
func (p CombinablePubKeySlice) ComboCommit() chainhash.Hash {
	// sort the pubkeys, smallest first
	sort.Sort(p)
	// feed em into the hash
	combo := make([]byte, len(p)*33)
	for i, k := range p {
		copy(combo[i*33:(i+1)*33], k.SerializeCompressed())
	}
	return chainhash.HashH(combo)
}
