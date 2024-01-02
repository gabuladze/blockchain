package crypto

import (
	"testing"
)

func TestNewPrivateKey(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()
	if len(privKey.Bytes()) != privKeyLen {
		t.Fatalf("invalid private key length. expected: %v got: %v", privKeyLen, len(privKey.Bytes()))
	}

	if len(pubKey.Bytes()) != pubKeyLen {
		t.Fatalf("invalid public key length. expected: %v got: %v", privKeyLen, len(pubKey.Bytes()))
	}
}

func TestNewPrivateKeyFromString(t *testing.T) {
	seedStr := "cb11735c9a35e641acb394b7b60b2f18dadc5e8a4595be338e61cd89791514d1"
	privKey := NewPrivateKeyFromString(seedStr)
	addressStr := "e7ff2dd4aa81ee62c84916cfb9e7cab6ff6e1a70"

	if len(privKey.Bytes()) != privKeyLen {
		t.Fatalf("invalid private key length. expected: %v got: %v", privKeyLen, len(privKey.Bytes()))
	}

	addr := privKey.Public().Address()
	if addr.String() != addressStr {
		t.Fatalf("invalid address. expected: %s got %s", addressStr, addr)
	}
}

func TestPrivateKeySign(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()
	msg := []byte("test msg")
	sig := privKey.Sign(msg)

	if !sig.Verify(pubKey, msg) {
		t.Fatalf("failed to verify signature with key: %v", pubKey)
	}

	if sig.Verify(pubKey, []byte("test")) {
		t.Fatal("expected to fail verification with incorrect msg")
	}

	privKey1 := NewPrivateKey()
	if sig.Verify(privKey1.Public(), msg) {
		t.Fatalf("expected to fail verification with incorrect pub key")
	}
}

func TestPrivateKeyToAddress(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()

	if len(address.Bytes()) != addressLen {
		t.Fatalf("expected address length to be: %v got: %v", addressLen, len(address.Bytes()))
	}
}
