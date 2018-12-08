package openassets

import "github.com/btcsuite/btcd/wire"

func isOpenAssetMarkerData(b []byte) bool {
	if b[0] == 0x4f && b[1] == 0x41 && b[2] == 0x01 && b[3] == 0x00 { // Open Asset 1.0
		return true
	}
	return false
}

func IsOpenAssetTransaction(tx *wire.MsgTx) bool {
	for _, txo := range tx.TxOut {
		if len(txo.PkScript) > 0 && txo.PkScript[0] == 0x6a {
			idx := 1
			for {
				opCode := txo.PkScript[idx]
				if opCode >= 0x05 && opCode <= 0x4b { // pushdata
					if isOpenAssetMarkerData(txo.PkScript[idx+1 : idx+1+int(opCode)]) {
						return true
					}
					idx += 1 + int(opCode)
				} else {
					idx++
				}
				if len(txo.PkScript) <= idx {
					break
				}
			}

		}
	}

	return false
}
