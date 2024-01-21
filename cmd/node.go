package main

import (
	"flag"
	"strings"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/node"
)

func main() {
	listenAddr := flag.String("grpc.addr", ":3000", "Listen address for GRPC server.")
	bootnodes := flag.String("grpc.bootnodes", "", "Comma separated list of bootnode addresses.")
	cryptoSeed := flag.String("crypto.seed", "97d3a71712a442f6345e16df34ecec93c3f6666dc84cee739c2e95a878ea99e1", "Seed string for generating private key.")
	isValidator := flag.Bool("validator", false, "Start node as a validator.")
	flag.Parse()

	var bootnodeList []string = []string{}
	if *bootnodes != "" {
		bootnodeList = strings.Split(*bootnodes, ",")
	}

	makeNode(*listenAddr, bootnodeList, *cryptoSeed, *isValidator)
	select {}
}

func makeNode(listenAddr string, addrs []string, cryptoSeed string, isValidator bool) *node.Node {
	var privKey *crypto.PrivateKey
	if isValidator {
		p := crypto.NewPrivateKeyFromString(cryptoSeed)
		privKey = &p
	}
	n := node.NewNode("chain-0.1", listenAddr, privKey)
	go n.Start(addrs)

	return n
}
