package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Utxo struct {
	TxHash   chainhash.Hash
	Outpoint uint32
	Value    uint64
	PkScript []byte
}

type OpenAssetUtxo struct {
	AssetID    []byte
	Utxo       Utxo
	AssetValue uint64
	Ours       bool
}

type OpenAssetIssuanceOutput struct {
	RecipientPkh [20]byte
	Value        uint64
}

type OpenAssetTransferOutput struct {
	AssetID      []byte
	RecipientPkh [20]byte
	Value        uint64
}

type OpenAssetTransaction struct {
	NormalInputs  []Utxo
	AssetInputs   []OpenAssetUtxo
	Issuances     []OpenAssetIssuanceOutput
	Transfers     []OpenAssetTransferOutput
	ChangeAddress [20]byte
	Metadata      OpenAssetMetadata
}

type OpenAssetMetadata struct {
	Ticker   string
	Decimals uint8
}

type OpenAsset struct {
	AssetID  []byte
	Metadata OpenAssetMetadata
}
