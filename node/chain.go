package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

const GodSeedStr = "97d3a71712a442f6345e16df34ecec93c3f6666dc84cee739c2e95a878ea99e6"

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() HeaderList {
	return HeaderList{
		headers: []*proto.Header{},
	}
}

func (hl *HeaderList) Add(h *proto.Header) {
	hl.headers = append(hl.headers, h)
}

func (hl *HeaderList) Get(height int) *proto.Header {
	if height > hl.Height() {
		panic("height it too high")
	}
	return hl.headers[height]
}

func (hl *HeaderList) Height() int {
	return hl.Len() - 1
}

func (hl *HeaderList) Len() int {
	return len(hl.headers)
}

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	blockStore   BlockStorer
	futureBlocks map[int32]*proto.Block
	fbLock       sync.RWMutex
	txStore      TxStorer
	utxoStore    UTXOStorer
	headers      HeaderList
}

func NewChain(bs BlockStorer, ts TxStorer) *Chain {
	chain := &Chain{
		blockStore:   bs,
		txStore:      ts,
		utxoStore:    NewMemoryUTXOStore(),
		headers:      NewHeaderList(),
		futureBlocks: make(map[int32]*proto.Block),
	}
	chain.addBlock(chain.createGenesisBlock())
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	// if there's a gap between node's chain height and block's height
	// save the block and process it in the future, once the node has it's parent
	if c.Height()+1 != int(b.Header.Height) {
		c.fbLock.Lock()
		defer c.fbLock.Unlock()
		_, ok := c.futureBlocks[b.Header.Height]
		if ok {
			return nil
		}
		c.futureBlocks[b.Header.Height] = b
		log.Printf(
			"Saved block for future processing currHeight=%d hash=%s height=%d numTxs=%d",
			c.Height(), hex.EncodeToString(types.HashBlock(b)), b.Header.Height, len(b.Transactions),
		)
		return nil
	}
	log.Printf("Adding block currHeight=%d hash=%s numTxs=%d", c.Height(), hex.EncodeToString(types.HashBlock(b)), len(b.Transactions))
	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	if err := c.addBlock(b); err != nil {
		return err
	}

	// try to process future blocks
	c.fbLock.Lock()
	defer c.fbLock.Unlock()
	for height, block := range c.futureBlocks {
		if int32(c.Height())+1 == height {
			if err := c.AddBlock(block); err != nil {
				return err
			}

			delete(c.futureBlocks, height)
		}
	}

	return nil
}

func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)

	for _, tx := range b.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}

		for _, input := range tx.Inputs {
			key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PrevTxHash), input.PrevOutIndex)
			utxo, err := c.utxoStore.Get(key)
			if err != nil {
				return err
			}
			utxo.Spent = true
		}

		hash := hex.EncodeToString(types.HashTransaction(tx))
		for i, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: i,
				Amount:   output.Amount,
				Spent:    false,
			}

			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}
	}

	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if height > c.headers.Height() {
		return nil, fmt.Errorf("height too high")
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validate block signature
	if !types.VerifyBlock(b) {
		return fmt.Errorf("signature verification failed for block: %+v", b)
	}

	// validate block hash
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	curentBlockHash := types.HashBlock(currentBlock)
	if !bytes.Equal(curentBlockHash, b.Header.PrevHash) {
		return fmt.Errorf("prevHash mismatch. expected: %s got: %s", curentBlockHash, b.Header.PrevHash)
	}

	// validate transactions
	for _, tx := range b.Transactions {
		if err := c.validateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) validateTransaction(tx *proto.Transaction) error {
	// verify signature
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid transaction signature. tx=%+v", tx)
	}

	// verify that inputs are not spent
	var (
		sumPrevOutputs int64
		sumOutputs     int64
	)
	numInputs := len(tx.Inputs)
	for i := 0; i < numInputs; i++ {
		in := tx.Inputs[i]
		key := fmt.Sprintf("%s_%d", hex.EncodeToString(in.PrevTxHash), in.PrevOutIndex) // double-check
		utxo, err := c.utxoStore.Get(key)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("utxo is already spent. prevHash=%s prevOutIndex=%d", hex.EncodeToString(in.PrevTxHash), in.PrevOutIndex)
		}
		sumPrevOutputs += utxo.Amount
	}

	// verify spendable amount
	for _, out := range tx.Outputs {
		sumOutputs += out.Amount
	}
	if sumPrevOutputs < sumOutputs {
		return fmt.Errorf("insufficient funds. need=%d have=%d", sumOutputs, sumPrevOutputs)
	}

	return nil
}

func (c *Chain) createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromString(GodSeedStr)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
			Height:  0,
		},
		Transactions: []*proto.Transaction{
			{
				Version: 1,
				Outputs: []*proto.TxOutput{
					{Amount: 1000, Address: privKey.Public().Address().Bytes()},
				},
			},
		},
	}
	types.SignBlock(privKey, block)
	types.GenerateRootHash(block)
	return block
}
