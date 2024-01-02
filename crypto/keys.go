package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	privKeyLen = 64
	pubKeyLen  = 32
	addressLen = 20
	seedLen    = 32
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKey() PrivateKey {
	seed := make([]byte, seedLen)
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
	if len(seed) != seedLen {
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

func (k PublicKey) Bytes() []byte {
	return k.key
}

func (k PublicKey) Address() Address {
	return Address{
		value: k.key[len(k.key)-addressLen:],
	}
}

type Signature struct {
	value []byte
}

func (s Signature) Verify(pubKey PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

func (a Address) Bytes() []byte {
	return a.value
}
