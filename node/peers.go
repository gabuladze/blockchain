package node

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gabuladze/blockchain/proto"
)

type PeerStore struct {
	peerLock  *sync.RWMutex
	peers     map[proto.NodeClient]*proto.Version
	peerIndex map[string]bool
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peerLock:  &sync.RWMutex{},
		peers:     make(map[proto.NodeClient]*proto.Version),
		peerIndex: make(map[string]bool),
	}
}

func (ps *PeerStore) Add(nc proto.NodeClient, cv *proto.Version) error {
	ps.peerLock.Lock()
	defer ps.peerLock.Unlock()
	if _, ok := ps.peerIndex[cv.Addr]; ok {
		return fmt.Errorf("peer already known. %v", cv.Addr)
	}

	slog.Info("new peer added", slog.String("addr", cv.Addr))
	ps.peers[nc] = cv
	ps.peerIndex[cv.Addr] = true

	return nil
}

func (ps *PeerStore) GetAddressList() []string {
	ps.peerLock.RLock()
	defer ps.peerLock.RUnlock()

	pList := []string{}
	for _, version := range ps.peers {
		pList = append(pList, version.Addr)
	}

	return pList
}

func (ps *PeerStore) GetPeers() map[proto.NodeClient]*proto.Version {
	ps.peerLock.RLock()
	defer ps.peerLock.RUnlock()

	return ps.peers
}

func (ps *PeerStore) Has(addr string) bool {
	ps.peerLock.RLock()
	defer ps.peerLock.RUnlock()
	_, ok := ps.peerIndex[addr]
	return ok
}
