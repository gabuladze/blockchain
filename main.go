package main

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/node"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeNode(":3000", []string{}, true)
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}, false)
	// time.Sleep(time.Second)
	// makeNode(":5000", []string{":4000"}, false)

	// time.Sleep(3 * time.Second)
	// makeTransaction()
	select {}
}

func makeNode(listenAddr string, addrs []string, isValidator bool) *node.Node {
	var privKey *crypto.PrivateKey
	if isValidator {
		p := crypto.NewPrivateKey()
		privKey = &p
	}
	n := node.NewNode("chain-0.1", listenAddr, privKey)
	go n.Start(addrs)

	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	nc := proto.NewNodeClient(client)

	privKey := crypto.NewPrivateKeyFromString(node.GodSeedStr)
	prevHash, _ := hex.DecodeString("b40a25e867b748d2d07401885b936bd6997a5338dfb0cd2e85bba2f6b60e4486")
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   prevHash,
				PrevOutIndex: 0,
				PubKey:       privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	_, err = nc.HandleTransaction(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
