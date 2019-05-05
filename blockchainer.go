package main

import (
	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
)

func chainAddNewBlock(content []byte) error {
	chainMutex.Lock()
	defer chainMutex.Unlock()
	chain.AddSHA512(content)
	return nil
}

func chainReset() error {
	genesis := block.NewSHA512Genesis()
	chainMutex.Lock()
	defer chainMutex.Unlock()
	chain = blockchain.New(genesis)
	return nil
}
