package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

type UTXOStorer interface {
	// key is "<Hash>_<outIndex>"
	Put(utxo *UTXO) error
	Get(string) (*UTXO, error)
}

type MemoryUTXOStore struct {
	lock  sync.RWMutex
	utxos map[string]*UTXO
}

func NewMemoryUTXOStore() UTXOStorer {
	return &MemoryUTXOStore{
		utxos: make(map[string]*UTXO),
	}
}

func (mus *MemoryUTXOStore) Put(u *UTXO) error {
	mus.lock.Lock()
	defer mus.lock.Unlock()

	key := fmt.Sprintf("%s_%d", u.Hash, u.OutIndex)
	mus.utxos[key] = u

	return nil
}

func (mus *MemoryUTXOStore) Get(key string) (*UTXO, error) {
	mus.lock.RLock()
	defer mus.lock.RUnlock()

	utxo, ok := mus.utxos[key]
	if !ok {
		return nil, fmt.Errorf("utxo not found. key=%s", key)
	}

	return utxo, nil
}

type TxStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}

type MemoryTxStore struct {
	lock sync.RWMutex
	txs  map[string]*proto.Transaction
}

func NewMemoryTxStore() TxStorer {
	return &MemoryTxStore{
		txs: make(map[string]*proto.Transaction),
	}
}

func (mts *MemoryTxStore) Put(tx *proto.Transaction) error {
	mts.lock.Lock()
	defer mts.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	mts.txs[hash] = tx
	return nil
}

func (mts *MemoryTxStore) Get(hash string) (*proto.Transaction, error) {
	mts.lock.RLock()
	defer mts.lock.RUnlock()

	tx, ok := mts.txs[hash]
	if !ok {
		return nil, fmt.Errorf("tx not found. hash=%s", hash)
	}

	return tx, nil
}

type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*proto.Block
}

func NewMemoryBlockStore() BlockStorer {
	return &MemoryBlockStore{
		blocks: map[string]*proto.Block{},
	}
}

func (mbs *MemoryBlockStore) Put(b *proto.Block) error {
	mbs.lock.Lock()
	defer mbs.lock.Unlock()

	hash := hex.EncodeToString(types.HashBlock(b))
	mbs.blocks[hash] = b
	return nil
}

func (mbs *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	mbs.lock.RLock()
	defer mbs.lock.RUnlock()

	block, ok := mbs.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block [%s] not found", hash)
	}
	return block, nil
}
