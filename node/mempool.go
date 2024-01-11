package node

import (
	"encoding/hex"
	"sync"

	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

type Mempool struct {
	lock sync.RWMutex
	txs  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txs: make(map[string]*proto.Transaction),
	}
}

func (m *Mempool) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.txs)
}

func (m *Mempool) Add(tx *proto.Transaction) bool {
	if m.Has(tx) {
		return false
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	m.txs[hash] = tx
	return true
}

func (m *Mempool) Has(tx *proto.Transaction) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := m.txs[hash]
	return ok
}

func (m *Mempool) Clear() []*proto.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()

	txs := make([]*proto.Transaction, len(m.txs))
	i := 0
	for k, v := range m.txs {
		txs[i] = v
		delete(m.txs, k)
		i++
	}
	return txs
}
