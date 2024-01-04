package main

import (
	"context"
	"log"
	"time"

	"github.com/gabuladze/blockchain/node"
	"github.com/gabuladze/blockchain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeNode(":3000", []string{})
	time.Sleep(2 * time.Second)
	makeNode(":4000", []string{":3000"})
	time.Sleep(3 * time.Second)
	makeNode(":5000", []string{":4000"})

	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 3)
	// 		// makeTransaction()
	// 		handshake()
	// 	}
	// }()
	select {}
}

func makeNode(listenAddr string, addrs []string) *node.Node {
	n := node.NewNode()
	go n.Start(listenAddr, addrs)

	return n
}

func handshake() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	nc := proto.NewNodeClient(client)

	clientVer := proto.Version{
		Version: "blockchain-0.1",
		Height:  20,
	}
	serverVer, err := nc.Handshake(context.TODO(), &clientVer)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received server version: %+v\n", serverVer)
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	nc := proto.NewNodeClient(client)

	tx := proto.Transaction{
		Version: 1,
	}
	_, err = nc.HandleTransaction(context.TODO(), &tx)
	if err != nil {
		log.Fatal(err)
	}
}
