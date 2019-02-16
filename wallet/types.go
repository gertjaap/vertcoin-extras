package wallet

import (
	"bytes"
	"encoding/binary"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type Utxo struct {
	TxHash   chainhash.Hash
	Outpoint uint32
	Value    uint64
	PkScript []byte
}

func (u Utxo) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(u.TxHash[:])
	binary.Write(&buf, binary.BigEndian, u.Outpoint)
	binary.Write(&buf, binary.BigEndian, u.Value)
	buf.Write(u.PkScript)
	return buf.Bytes()
}

func UtxoFromBytes(b []byte) Utxo {
	buf := bytes.NewBuffer(b)
	u := Utxo{}
	hash, _ := chainhash.NewHash(buf.Next(32))
	u.TxHash = *hash
	binary.Read(buf, binary.BigEndian, &u.Outpoint)
	binary.Read(buf, binary.BigEndian, &u.Value)
	copy(u.PkScript, buf.Bytes())
	return u
}

type StealthUtxo struct {
	Utxo   Utxo
	EncOTK []byte
}

func (su StealthUtxo) Bytes() []byte {
	var buf bytes.Buffer
	wire.WriteVarBytes(&buf, 0, su.EncOTK)
	buf.Write(su.Utxo.Bytes())
	return buf.Bytes()
}

func StealthUtxoFromBytes(b []byte) StealthUtxo {
	buf := bytes.NewBuffer(b)
	su := StealthUtxo{}

	encOtk, _ := wire.ReadVarBytes(buf, 0, 150, "encotk")
	su.EncOTK = encOtk
	su.Utxo = UtxoFromBytes(buf.Bytes())

	return su
}

type OpenAssetUtxo struct {
	AssetID    []byte
	Utxo       Utxo
	AssetValue uint64
	Ours       bool
}

func (oau OpenAssetUtxo) Bytes() []byte {
	var buf bytes.Buffer
	wire.WriteVarBytes(&buf, 0, oau.AssetID)
	binary.Write(&buf, binary.BigEndian, oau.AssetValue)
	binary.Write(&buf, binary.BigEndian, oau.Ours)
	buf.Write(oau.Utxo.Bytes())
	return buf.Bytes()
}

func OpenAssetUtxoFromBytes(b []byte) OpenAssetUtxo {
	buf := bytes.NewBuffer(b)
	oau := OpenAssetUtxo{}

	assetId, _ := wire.ReadVarBytes(buf, 0, 80, "assetId")
	oau.AssetID = assetId
	binary.Read(buf, binary.BigEndian, &oau.AssetValue)
	binary.Read(buf, binary.BigEndian, &oau.Ours)
	oau.Utxo = UtxoFromBytes(buf.Bytes())

	return oau
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
	Follow   bool
}

func (oa *OpenAsset) Bytes() []byte {
	var buf bytes.Buffer
	wire.WriteVarBytes(&buf, 0, oa.AssetID)
	wire.WriteVarBytes(&buf, 0, []byte(oa.Metadata.Ticker))
	binary.Write(&buf, binary.BigEndian, oa.Metadata.Decimals)
	binary.Write(&buf, binary.BigEndian, oa.Follow)
	return buf.Bytes()
}

func OpenAssetFromBytes(b []byte) *OpenAsset {
	buf := bytes.NewBuffer(b)
	oa := new(OpenAsset)

	assetId, _ := wire.ReadVarBytes(buf, 0, 80, "assetId")
	oa.AssetID = assetId
	tickerBytes, _ := wire.ReadVarBytes(buf, 0, 80, "ticker")
	oa.Metadata.Ticker = string(tickerBytes)

	binary.Read(buf, binary.BigEndian, &oa.Metadata.Decimals)
	binary.Read(buf, binary.BigEndian, &oa.Follow)

	return oa
}

type StealthTransaction struct {
	RecipientPubKey [33]byte
	Amount          uint64
}

type SendTransaction struct {
	RecipientPkh [20]byte
	Amount       uint64
}
