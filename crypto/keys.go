package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	PrivKeyLen   = 64
	PubKeyLen    = 32
	AddressLen   = 20
	SeedLen      = 32
	SignatureLen = 64
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKey() PrivateKey {
	seed := make([]byte, SeedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}

	return PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromString(seedStr string) PrivateKey {
	seed, err := hex.DecodeString(seedStr)
	if err != nil {
		panic("invalid seed string")
	}
	return NewPrivateKeyFromSeed(seed)
}

func NewPrivateKeyFromSeed(seed []byte) PrivateKey {
	if len(seed) != SeedLen {
		panic("invalid seed length")
	}
	return PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func (k PrivateKey) Bytes() []byte {
	return k.key
}

func (k PrivateKey) Sign(msg []byte) Signature {
	return Signature{
		value: ed25519.Sign(k.key, msg),
	}
}

func (k PrivateKey) Public() PublicKey {
	key, ok := k.key.Public().(ed25519.PublicKey)
	if !ok {
		panic("failed to get public key")
	}
	return PublicKey{
		key: key,
	}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func PublicKeyFromBytes(k []byte) PublicKey {
	if len(k) != PubKeyLen {
		panic("invalid length")
	}
	return PublicKey{
		key: k,
	}
}

func (k PublicKey) Bytes() []byte {
	return k.key
}

func (k PublicKey) Address() Address {
	return Address{
		value: k.key[len(k.key)-AddressLen:],
	}
}

type Signature struct {
	value []byte
}

func SignatureFromBytes(s []byte) Signature {
	if len(s) != SignatureLen {
		panic("invalid length")
	}
	return Signature{
		value: s,
	}
}

func (s Signature) Verify(pubKey PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

func (s Signature) Bytes() []byte {
	return s.value
}

type Address struct {
	value []byte
}

func AddressFromBytes(b []byte) Address {
	if len(b) != AddressLen {
		panic("invalid address length")
	}
	return Address{
		value: b,
	}
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

func (a Address) Bytes() []byte {
	return a.value
}
