package node

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const blockTime = 3 * time.Second

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

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivKey    *crypto.PrivateKey
}

type Node struct {
	ServerConfig

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool

	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {
	return &Node{
		ServerConfig: cfg,
		peers:        make(map[proto.NodeClient]*proto.Version),
		mempool:      NewMempool(),
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	grpcServerOpts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(grpcServerOpts...)

	proto.RegisterNodeServer(grpcServer, n)
	log.Printf("grpc server running on: %s", listenAddr)

	go n.bootstrapNetwork(bootstrapNodes)

	if n.PrivKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

func (n *Node) AddPeer(nc proto.NodeClient, cv *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	log.Printf("[%s] new peer added. addr = %s peers = %+v\n", n.ListenAddr, cv.ListenAddr, cv.Peers)
	n.peers[nc] = cv

	go n.bootstrapNetwork(cv.Peers)
}

func (n *Node) DeletePeer(nc proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	log.Printf("[%s] peer deleted. version = %+v", n.ListenAddr, n.peers[nc])
	delete(n.peers, nc)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	nc, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.AddPeer(nc, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.None, error) {
	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {
		log.Printf("[%s] received tx peer=%s hash=%s\n", n.ListenAddr, peer.LocalAddr, hash)
		go func() {
			err := n.broadcast(tx)
			if err != nil {
				log.Fatalf("[%s] broadcast error %v", n.ListenAddr, err)
			}
		}()
	}

	return &proto.None{}, nil
}

func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) validatorLoop() {
	log.Printf("[%s] starting validator loop", n.ListenAddr)
	ticker := time.NewTicker(blockTime)

	for {
		<-ticker.C

		txs := n.mempool.Clear()
		log.Println("VALIDATE BLOCK", "len(txs)=", len(txs))
	}
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectToPeer(addr) {
			continue
		}

		nc, cv, err := n.dialRemoteNode(addr)
		if err != nil {
			log.Printf("[%s] dial error %v", n.ListenAddr, err)
		}

		n.AddPeer(nc, cv)
	}
	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	log.Printf("[%s] DIALING %s\n", n.ListenAddr, addr)
	nc, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := nc.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return nc, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.Version,
		ListenAddr: n.ListenAddr,
		Peers:      n.getPeers(),
	}
}

func (n *Node) getPeers() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	pList := []string{}
	for _, version := range n.peers {
		pList = append(pList, version.ListenAddr)
	}

	return pList
}

func (n *Node) canConnectToPeer(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}

	for _, version := range n.peers {
		if version.ListenAddr == addr {
			return false
		}
	}

	return true
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	client, err := grpc.Dial(listenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(client), nil
}
