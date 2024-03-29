package types

import (
	"bytes"
	"crypto/sha256"
	"log/slog"

	"github.com/cbergoon/merkletree"
	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

func BuildHeader(version int32, height int64, prevHash []byte, timestamp int64) *proto.Header {
	return &proto.Header{
		Version:   version,
		Height:    height,
		PrevHash:  prevHash,
		Timestamp: timestamp,
	}
}

func BuildBlock(header *proto.Header, txs []*proto.Transaction) *proto.Block {
	return &proto.Block{
		Header:       header,
		Transactions: txs,
	}
}

func BuildAndSignBlock(header *proto.Header, txs []*proto.Transaction, privKey crypto.PrivateKey) *proto.Block {
	b := BuildBlock(header, txs)
	GenerateRootHash(b)
	SignBlock(privKey, b)
	return b
}

// HashBlock returns sha256 of the header.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(h *proto.Header) []byte {
	b, err := pb.Marshal(h)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func SignBlock(pk crypto.PrivateKey, b *proto.Block) crypto.Signature {
	hash := HashBlock(b)
	sig := pk.Sign(hash)
	b.Signature = sig.Bytes()
	b.PubKey = pk.Public().Bytes()
	return sig
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.PubKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PubKey)
	hash := HashBlock(b)
	if !sig.Verify(pubKey, hash) {
		return false
	}

	if len(b.Transactions) > 0 {
		if len(b.Header.RootHash) == 0 {
			return false
		}
		t, err := GenerateMerkleTree(b)
		if err != nil {
			slog.Error("merkle tree generation failed", err)
			return false
		}
		mr := t.MerkleRoot()
		return bytes.Equal(mr, b.Header.RootHash)
	}

	return true
}

func GenerateMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))
	for i, tx := range b.Transactions {
		h := NewTXHash(HashTransaction(tx))
		list[i] = h
	}

	return merkletree.NewTree(list)
}

func GenerateRootHash(b *proto.Block) ([]byte, error) {
	t, err := GenerateMerkleTree(b)
	if err != nil {
		return nil, err
	}

	mr := t.MerkleRoot()
	b.Header.RootHash = mr

	return mr, nil
}
