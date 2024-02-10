package main

import (
	"flag"
	"strings"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/node"
)

func main() {
	grpcAddr := flag.String("grpc.addr", "0.0.0.0", "Address for GRPC server.")
	grpcPort := flag.String("grpc.port", "3000", "Port for GRPC server.")
	bootnodes := flag.String("grpc.bootnodes", "", "Comma separated list of bootnode addresses.")
	cryptoSeed := flag.String("crypto.seed", "97d3a71712a442f6345e16df34ecec93c3f6666dc84cee739c2e95a878ea99e1", "Seed string for generating private key.")
	isValidator := flag.Bool("validator", false, "Start node as a validator.")
	flag.Parse()

	var bootnodeList []string = []string{}
	if *bootnodes != "" {
		bootnodeList = strings.Split(*bootnodes, ",")
	}

	makeNode(*grpcAddr, *grpcPort, bootnodeList, *cryptoSeed, *isValidator)
	select {}
}

func makeNode(grpcAddr, grpcPort string, addrs []string, cryptoSeed string, isValidator bool) *node.Node {
	var privKey *crypto.PrivateKey
	if isValidator {
		p := crypto.NewPrivateKeyFromString(cryptoSeed)
		privKey = &p
	}
	n := node.NewNode("chain-0.1", grpcAddr, grpcPort, privKey)
	go n.Start(addrs)

	return n
}
