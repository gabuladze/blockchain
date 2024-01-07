package node

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
	"github.com/gabuladze/blockchain/utils"
)

func randomBlock(t *testing.T, c *Chain) *proto.Block {
	var (
		block = utils.RandomBlock()
	)

	prevBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		t.Fatal("Failed to get prev block", err)
	}
	block.Header.PrevHash = types.HashBlock(prevBlock)

	return block
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())

	chainHeight := chain.Height()
	if chainHeight != 0 {
		t.Fatalf("invalid chain height. expected: %d got: %d", 0, chainHeight)
	}

	_, err := chain.GetBlockByHeight(0)
	if err != nil {
		t.Fatal("failed to get genesis block: ", err)
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())

	for i := 0; i < 100; i++ {
		privKey := crypto.NewPrivateKey()
		block := randomBlock(t, chain)
		types.SignBlock(privKey, block)

		if err := chain.AddBlock(block); err != nil {
			t.Fatal("Failed to add block", err)
		}

		blockByHash, err := chain.GetBlockByHash(types.HashBlock(block))
		if err != nil {
			t.Fatal("failed to get block by hash", err)
		}
		if !reflect.DeepEqual(block, blockByHash) {
			t.Fatalf("inserted & fetched blocks don't match. inserted=%+v fetched=%+v", block, blockByHash)
		}

		blockByHeight, err := chain.GetBlockByHeight(i + 1)
		if err != nil {
			t.Fatal("failed to get block by height", err)
		}
		if !reflect.DeepEqual(block, blockByHeight) {
			t.Fatalf("inserted & fetched blocks don't match. inserted=%+v fetched=%+v", block, blockByHeight)
		}
	}
}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromString(godSeedStr)
		recipient = crypto.NewPrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("b40a25e867b748d2d07401885b936bd6997a5338dfb0cd2e85bba2f6b60e4486")
	if err != nil {
		t.Fatal("Failed to fetch tx", err)
	}

	inputs := []*proto.TxInput{{
		PrevTxHash:   types.HashTransaction(prevTx),
		PrevOutIndex: 0,
		PubKey:       privKey.Public().Bytes(),
	}}
	outputs := []*proto.TxOutput{
		{Amount: 100, Address: recipient},
		{Amount: 900, Address: privKey.Public().Address().Bytes()},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.GenerateRootHash(block)
	types.SignBlock(privKey, block)

	if err := chain.AddBlock(block); err != nil {
		t.Fatal("Failed to add block", err)
	}

	txHash := hex.EncodeToString(types.HashTransaction(tx))
	fetchedTx, err := chain.txStore.Get(txHash)
	if err != nil {
		t.Fatal("Failed to fetch tx", err)
	}

	if !reflect.DeepEqual(tx, fetchedTx) {
		t.Fatalf("tx mismatch. expected: %+v got: %+v", tx, fetchedTx)
	}
}

func TestAddBlockWithTxInsufficientFunds(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromString(godSeedStr)
		recipient = crypto.NewPrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("b40a25e867b748d2d07401885b936bd6997a5338dfb0cd2e85bba2f6b60e4486")
	if err != nil {
		t.Fatal("Failed to fetch tx", err)
	}

	inputs := []*proto.TxInput{{
		PrevTxHash:   types.HashTransaction(prevTx),
		PrevOutIndex: 0,
		PubKey:       privKey.Public().Bytes(),
	}}
	outputs := []*proto.TxOutput{{Amount: 1001, Address: recipient}}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.GenerateRootHash(block)
	types.SignBlock(privKey, block)

	if err := chain.AddBlock(block); err == nil {
		t.Fatal("expected add block to fail", err)
	}
}
