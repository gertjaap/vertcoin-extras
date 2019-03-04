package util

import (
	"strings"

	"github.com/adiabat/bech32"
	"github.com/btcsuite/btcd/txscript"
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

	return false
}
