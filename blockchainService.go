package main

import (
	"encoding/json"
	"sync"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
)

type BlockchainService struct {
	mutex sync.Mutex
	chain blockchain.Blockchain
}

var blockchainService *BlockchainService

func (service *BlockchainService) addBlock(content []byte) (block.Block, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.chain.AddSHA512(content)
	block := service.chain.At(service.chain.Length() - 1)
	return block, nil
}

func (service *BlockchainService) resetChain() error {
	genesis := block.NewSHA512Genesis()
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.chain = blockchain.New(genesis)
	return nil
}

func (service *BlockchainService) GCI(other blockchain.Blockchain) (int, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	gci, err := service.chain.GetGCI(other)
	if err != nil {
		return -1, err
	}

	return gci, nil
}

func (service *BlockchainService) ChainLength() int {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	return service.chain.Length()
}

func (service *BlockchainService) lock() {
	service.mutex.Lock()
}

func (service *BlockchainService) unlock() {
	service.mutex.Unlock()
}

func (service *BlockchainService) UpdateChain(other blockchain.Blockchain) error {
	service.lock()
	defer service.unlock()
	if err := service.chain.UpdateChain(other); err != nil {
		return err
	}

	return nil
}

func (service *BlockchainService) ChainJSON() ([]byte, error) {
	service.lock()
	defer service.unlock()

	json, err := json.Marshal(service.chain)
	if err != nil {
		return nil, err
	}

	return json, nil
}

// FindBlockByIndex searches for a block by index
func (service *BlockchainService) FindBlockByIndex(index int) block.Block {
	service.lock()
	defer service.unlock()

	return service.chain.At(index)
}

// FindBlockByHash searches for a block by hash
func (service *BlockchainService) FindBlockByHash(hash []byte) block.Block {
	service.lock()
	defer service.unlock()

	return service.chain.FindByBlockHash(hash)
}

//FindBlockByParentHash searches for a block by parent hash
func (service *BlockchainService) FindBlockByParentHash(hash []byte) block.Block {
	service.lock()
	defer service.unlock()

	return service.chain.FindByParentHash(hash)
}

// FindBlockByTimestamp searches for a block by timestamp
func (service *BlockchainService) FindBlockByTimestamp(timestamp int64) block.Block {
	service.lock()
	defer service.unlock()

	return service.chain.FindByTimestamp(timestamp)
}

func (service *BlockchainService) Chain() blockchain.Blockchain {
	service.lock()
	defer service.unlock()

	return service.chain
}
