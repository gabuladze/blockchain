package node

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/gabuladze/blockchain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version    string
	listenAddr string

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	return &Node{
		peers: make(map[proto.NodeClient]*proto.Version),
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.listenAddr = listenAddr
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	grpcServerOpts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(grpcServerOpts...)

	proto.RegisterNodeServer(grpcServer, n)
	log.Printf("grpc server running on: %s", listenAddr)

	go n.bootstrapNetwork(bootstrapNodes)

	return grpcServer.Serve(ln)
}

func (n *Node) AddPeer(nc proto.NodeClient, cv *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	log.Printf("[%s] new peer added. addr = %s peers = %+v\n", n.listenAddr, cv.ListenAddr, cv.Peers)
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
	peer, _ := peer.FromContext(ctx)
	log.Printf("received HandleTransaction call from %+v\n", peer)
	return &proto.None{}, nil
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

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.version,
		ListenAddr: n.listenAddr,
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
