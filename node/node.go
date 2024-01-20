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

type Node struct {
	version    string
	listenAddr string
	privKey    *crypto.PrivateKey

	peerLock *sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool
	chain    *Chain

	addBlockCh       chan *proto.Block
	broadcastBlockCh chan *proto.Block
	errCh            chan error

	proto.UnimplementedNodeServer
}

func NewNode(version, listenAddr string, privKey *crypto.PrivateKey) *Node {
	return &Node{
		version:    version,
		listenAddr: listenAddr,
		privKey:    privKey,
		peerLock:   &sync.RWMutex{},
		peers:      make(map[proto.NodeClient]*proto.Version),
		mempool:    NewMempool(),
		chain:      NewChain(NewMemoryBlockStore(), NewMemoryTxStore()),

		addBlockCh:       make(chan *proto.Block),
		broadcastBlockCh: make(chan *proto.Block),
		errCh:            make(chan error),
	}
}

func (n *Node) Start(bootstrapNodes []string) error {
	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return err
	}

	grpcServerOpts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(grpcServerOpts...)

	proto.RegisterNodeServer(grpcServer, n)
	log.Printf("grpc server running on: %s", n.listenAddr)

	go n.bootstrapNetwork(bootstrapNodes)

	if n.privKey != nil {
		go n.validatorLoop()
	}

	go n.chain.StartBlockReceiver(n.addBlockCh, n.broadcastBlockCh, n.errCh)

	go func() {
		for {
			select {
			case broadcastBlk := <-n.broadcastBlockCh:
				go n.broadcast(broadcastBlk)
			case errMsg := <-n.errCh:
				log.Fatalf("[%s] error: %s\n", n.listenAddr, errMsg)
			}
		}
	}()

	return grpcServer.Serve(ln)
}

func (n *Node) AddPeer(nc proto.NodeClient, cv *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	log.Printf("[%s] new peer added. addr = %s\n", n.listenAddr, cv.ListenAddr)
	n.peers[nc] = cv

	go n.bootstrapNetwork(cv.Peers)
}

func (n *Node) DeletePeer(nc proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	log.Printf("[%s] peer deleted. version = %+v", n.listenAddr, n.peers[nc])
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
	if n.mempool.Add(tx) {
		peer, _ := peer.FromContext(ctx)
		hash := hex.EncodeToString(types.HashTransaction(tx))
		log.Printf("[%s] received tx peer=%s hash=%s\n", n.listenAddr, peer.LocalAddr, hash)
		go func() {
			err := n.broadcast(tx)
			if err != nil {
				log.Fatalf("[%s] broadcast error %v", n.listenAddr, err)
			}
		}()
	}

	return &proto.None{}, nil
}

func (n *Node) HandleBlock(ctx context.Context, b *proto.Block) (*proto.None, error) {
	n.addBlockCh <- b
	return &proto.None{}, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.version,
		Height:     int64(n.chain.Height()),
		ListenAddr: n.listenAddr,
		Peers:      n.getPeers(),
	}
}

func (n *Node) broadcast(msg any) error {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			if _, err := peer.HandleTransaction(context.Background(), v); err != nil {
				return err
			}
		case *proto.Block:
			if _, err := peer.HandleBlock(context.Background(), v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) validatorLoop() {
	log.Printf("[%s] starting validator loop", n.listenAddr)
	ticker := time.NewTicker(blockTime)

	for {
		<-ticker.C

		txs := n.mempool.Clear()

		lastBlock, err := n.chain.GetBlockByHeight(n.chain.Height())
		if err != nil {
			log.Fatalf("[%s] error fetching last block: %v", n.listenAddr, err)
			continue
		}
		header := types.BuildHeader(1, int32(n.chain.Height())+1, types.HashBlock(lastBlock), time.Now().UnixNano())
		newBlock := types.BuildAndSignBlock(header, txs, *n.privKey)

		log.Printf("[%s] validated block. hash=%s height=%d txs=%d", n.listenAddr, hex.EncodeToString(types.HashBlock(newBlock)), newBlock.Header.Height, len(newBlock.Transactions))

		if _, err := n.chain.AddBlock(newBlock); err != nil {
			log.Fatalf("[%s] error when adding block %v", n.listenAddr, err)
			continue
		}

		if err := n.broadcast(newBlock); err != nil {
			log.Fatalf("[%s] error when broadcasting block %v", n.listenAddr, err)
			continue
		}
	}
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectToPeer(addr) {
			continue
		}

		nc, cv, err := n.dialRemoteNode(addr)
		if err != nil {
			log.Printf("[%s] dial error %v", n.listenAddr, err)
		}

		n.AddPeer(nc, cv)
	}
	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	log.Printf("[%s] DIALING %s\n", n.listenAddr, addr)
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
	if n.listenAddr == addr {
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
