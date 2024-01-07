package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/cbergoon/merkletree"
	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

type TXHash struct {
	hash []byte
}

func NewTXHash(hash []byte) TXHash {
	return TXHash{hash: hash}
}

func (t TXHash) CalculateHash() ([]byte, error) {
	return t.hash, nil
}
func (t TXHash) Equals(other merkletree.Content) (bool, error) {
	return bytes.Equal(t.hash, other.(TXHash).hash), nil
}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func SignTransaction(pk crypto.PrivateKey, tx *proto.Transaction) crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		// we have to remove signature from inputs
		// because the original tx didn't have signature
		// when it was hashed.
		sigBytes := input.Signature
		input.Signature = nil
		pubKey := crypto.PublicKeyFromBytes(input.PubKey)
		sig := crypto.SignatureFromBytes(sigBytes)
		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
		input.Signature = sigBytes
	}
	return true
}
