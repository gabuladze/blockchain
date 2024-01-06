package types

import (
	"bytes"
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

func TestSignVerifyBlock(t *testing.T) {
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

	if !bytes.Equal(block.PubKey, pubKey.Bytes()) {
		t.Fatalf("invalid block pubKey. expected: %v got: %v", pubKey.Bytes(), block.PubKey)
	}
	if !bytes.Equal(block.Signature, sig.Bytes()) {
		t.Fatalf("invalid block signature. expected: %v got: %v", sig.Bytes(), block.Signature)
	}

	if !VerifyBlock(block) {
		t.Fatal("signature verification failed")
	}

	incorrectPrivKey := crypto.NewPrivateKey()
	block.PubKey = incorrectPrivKey.Public().Bytes()
	if VerifyBlock(block) {
		t.Fatal("expected signature verification to fail with incorrect validator pub key")
	}
}
