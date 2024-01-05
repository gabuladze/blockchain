package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

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
