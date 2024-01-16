package utils

import (
	randc "crypto/rand"
	"io"
	"time"

	"github.com/gabuladze/blockchain/proto"
)

func RandomHash() []byte {
	hash := make([]byte, 32)
	io.ReadFull(randc.Reader, hash)
	return hash
}

func RandomBlock() *proto.Block {
	header := proto.Header{
		Version:   1,
		PrevHash:  RandomHash(),
		RootHash:  RandomHash(),
		Timestamp: time.Now().UnixNano(),
	}
	return &proto.Block{
		Header: &header,
	}
}
