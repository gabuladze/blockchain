package node

import (
	"encoding/hex"
	"fmt"

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
	return &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
}

func (c *Chain) AddBlock(b *proto.Block) error {
	// TODO: validation
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
