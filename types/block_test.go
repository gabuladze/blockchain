package types

import (
	"testing"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/utils"
)

func TestHashBlock(t *testing.T) {
	block := utils.RandomBlock()
	hash := HashBlock(block)

	if len(hash) != 32 {
		t.Fatalf("invalid hash length. expected: %d got: %d", 32, len(hash))
	}
}

func TestSignBlock(t *testing.T) {
	var (
		block   = utils.RandomBlock()
		privKey = crypto.NewPrivateKey()
		pubKey  = privKey.Public()
	)

	sig := SignBlock(privKey, block)
	if len(sig.Bytes()) != 64 {
		t.Fatalf("invalid signature length. expected: %d got: %d", 64, len(sig.Bytes()))
	}
	if !sig.Verify(pubKey, HashBlock(block)) {
		t.Fatal("invalid signature")
	}
}
