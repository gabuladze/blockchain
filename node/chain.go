package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/gabuladze/blockchain/crypto"
	"github.com/gabuladze/blockchain/proto"
	"github.com/gabuladze/blockchain/types"
)

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

type Chain struct {
	blockStore BlockStorer
	headers    HeaderList
}

func NewChain(bs BlockStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
	chain.addBlock(chain.createGenesisBlock())
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	c.headers.Add(b.Header)
	return c.blockStore.Put(b)
}

func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)
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
	// validate signature
	if !types.VerifyBlock(b) {
		return fmt.Errorf("signature verification failed for block: %+v", b)
	}

	// validate hash
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	curentBlockHash := types.HashBlock(currentBlock)
	if !bytes.Equal(curentBlockHash, b.Header.PrevHash) {
		return fmt.Errorf("prevHash mismatch. expected: %s got: %s", curentBlockHash, b.Header.PrevHash)
	}

	return nil
}

func (c *Chain) createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKey()
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
			Height:  0,
		},
	}
	types.SignBlock(privKey, block)
	return block
}
