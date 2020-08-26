// Copyright 2020 Kevin Gentile
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
	"github.com/spf13/viper"
)

type BlockchainService struct {
	mutex  sync.Mutex
	chain  *blockchain.Blockchain
	daemon *daemon
}

var blockchainService *BlockchainService

func NewBlockchainService(daemon *daemon) (*BlockchainService, error) {
	bs := &BlockchainService{
		daemon: daemon,
	}

	//TODO remove with load ledger
	if viper.GetBool("genesis") {
		blockchainService.resetChain()
	}

	return bs, nil
}

func (service *BlockchainService) addBlock(content []byte) (*block.Block, error) {
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
	chain, err := blockchain.New(genesis)
	if err != nil {
		return err
	}

	service.chain = chain
	return nil
}

func (service *BlockchainService) GCI(other *blockchain.Blockchain) (int, error) {
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

func (service *BlockchainService) UpdateChain(other *blockchain.Blockchain) error {
	service.lock()
	defer service.unlock()
	newChain, err := blockchain.UpdateChain(service.chain, other)
	if err != nil {
		return err
	}

	service.chain = newChain
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

var ErrBlockNotFound = errors.New("blockchainService: block not found")

// FindBlockByIndex searches for a block by index
func (service *BlockchainService) FindBlockByIndex(index int) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.At(index)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

// FindBlockByHash searches for a block by hash
func (service *BlockchainService) FindBlockByHash(hash []byte) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByBlockHash(hash)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

//FindBlockByParentHash searches for a block by parent hash
func (service *BlockchainService) FindBlockByParentHash(hash []byte) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByParentHash(hash)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

// FindBlockByTimestamp searches for a block by timestamp
func (service *BlockchainService) FindBlockByTimestamp(timestamp int64) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByTimestamp(timestamp)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

func (service *BlockchainService) Chain() *blockchain.Blockchain {
	service.lock()
	defer service.unlock()

	return service.chain
}
