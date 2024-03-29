package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

type Storer[T any] interface {
	Put(*T) error
	Get(string) (*T, error)
}

type MemoryUTXOStore struct {
	lock  *sync.RWMutex
	utxos map[string]*UTXO
}

func NewMemoryUTXOStore() Storer[UTXO] {
	return &MemoryUTXOStore{
		lock:  &sync.RWMutex{},
		utxos: make(map[string]*UTXO),
	}
}

func (mus *MemoryUTXOStore) Put(u *UTXO) error {
	mus.lock.Lock()
	defer mus.lock.Unlock()

	// key is "<Hash>_<outIndex>"
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

type MemoryTxStore struct {
	lock *sync.RWMutex
	txs  map[string]*proto.Transaction
}

func NewMemoryTxStore() Storer[proto.Transaction] {
	return &MemoryTxStore{
		lock: &sync.RWMutex{},
		txs:  make(map[string]*proto.Transaction),
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
	GetBlock(string) (*proto.Block, error)
	GetHeader(int64) (*proto.Header, error)
	Height() int64
}

type MemoryBlockStore struct {
	lock    *sync.RWMutex
	blocks  map[string]*proto.Block
	headers []*proto.Header
}

func NewMemoryBlockStore() BlockStorer {
	return &MemoryBlockStore{
		lock:    &sync.RWMutex{},
		blocks:  map[string]*proto.Block{},
		headers: []*proto.Header{},
	}
}

func (mbs *MemoryBlockStore) Put(b *proto.Block) error {
	mbs.lock.Lock()
	defer mbs.lock.Unlock()

	hash := hex.EncodeToString(types.HashBlock(b))
	mbs.headers = append(mbs.headers, b.Header)
	mbs.blocks[hash] = b
	return nil
}

func (mbs *MemoryBlockStore) GetBlock(hash string) (*proto.Block, error) {
	mbs.lock.RLock()
	defer mbs.lock.RUnlock()

	block, ok := mbs.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block [%s] not found", hash)
	}
	return block, nil
}

func (mbs *MemoryBlockStore) GetHeader(height int64) (*proto.Header, error) {
	mbs.lock.RLock()
	defer mbs.lock.RUnlock()

	if height > mbs.Height() {
		return nil, fmt.Errorf("height too high. height=%d currHeight=%d", height, mbs.Height())
	}
	return mbs.headers[height], nil
}

func (mbs *MemoryBlockStore) Height() int64 {
	mbs.lock.RLock()
	defer mbs.lock.RUnlock()
	return int64(len(mbs.headers) - 1)
}
