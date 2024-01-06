package types

import (
	"crypto/sha256"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

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
	return sig.Verify(pubKey, hash)
}
