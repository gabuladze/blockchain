package types

import (
	"testing"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/utils"
)

func TestHashTransaction(t *testing.T) {
	var (
		fromPrivKey = crypto.NewPrivateKey()
		fromPubKey  = fromPrivKey.Public()
		fromAddress = fromPubKey.Address()
		toPrivKey   = crypto.NewPrivateKey()
		toAddress   = toPrivKey.Public().Address()
	)

	input := &proto.TxInput{
		PrevTxHash:   utils.RandomHash(),
		PrevOutIndex: 0,
		PubKey:       fromPubKey.Bytes(),
	}
	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress.Bytes(),
	}
	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress.Bytes(),
	}
	tx := proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	sig := SignTransaction(fromPrivKey, &tx)
	input.Signature = sig.Bytes()

	if !VerifyTransaction(&tx) {
		t.Fatal("tx verification failed")
	}
}
