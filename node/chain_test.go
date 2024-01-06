package node

import (
	"reflect"
	"testing"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/types"
	"github.com/gabuladze/blockchain/utils"
)

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

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
	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {
		prevBlock, err := chain.GetBlockByHeight(chain.Height())
		if err != nil {
			t.Fatal("Failed to get prev block", err)
		}

		privKey := crypto.NewPrivateKey()
		block := utils.RandomBlock()
		block.Header.PrevHash = types.HashBlock(prevBlock)
		types.SignBlock(privKey, block)
		blockHash := types.HashBlock(block)

		err = chain.AddBlock(block)
		if err != nil {
			t.Fatal("Failed to add block", err)
		}

		blockByHash, err := chain.GetBlockByHash(blockHash)
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
