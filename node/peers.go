package node

import (
	"log/slog"
	"sync"

	"github.com/gabuladze/blockchain/proto"
)

type PeerStore struct {
	peerLock *sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peerLock: &sync.RWMutex{},
		peers:    make(map[proto.NodeClient]*proto.Version),
	}
}

func (ps *PeerStore) Add(nc proto.NodeClient, cv *proto.Version) error {
	ps.peerLock.Lock()
	defer ps.peerLock.Unlock()
	slog.Info("new peer added", slog.String("addr", cv.Addr))
	ps.peers[nc] = cv
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
