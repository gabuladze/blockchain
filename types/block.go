package types

import (
	"crypto/sha256"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	pb "google.golang.org/protobuf/proto"
)

// HashBlock returns sha256 of the header.
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func SignBlock(pk crypto.PrivateKey, b *proto.Block) crypto.Signature {
	return pk.Sign(HashBlock(b))
}
