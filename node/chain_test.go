package node

import (
	"reflect"
	"testing"

	"github.com/gabuladze/blockchain/types"
	"github.com/gabuladze/blockchain/utils"
)

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {
		var (
			block     = utils.RandomBlock()
			blockHash = types.HashBlock(block)
		)

		err := chain.AddBlock(block)
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

		blockByHeight, err := chain.GetBlockByHeight(i)
		if err != nil {
			t.Fatal("failed to get block by hash", err)
		}
		if !reflect.DeepEqual(block, blockByHeight) {
			t.Fatalf("inserted & fetched blocks don't match. inserted=%+v fetched=%+v", block, blockByHeight)
		}
	}
}
