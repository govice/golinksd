package blockchain

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/govice/golinks/block"
	"github.com/govice/golinks/blockchain"
)

type Service struct {
	mutex sync.Mutex
	chain *blockchain.Blockchain
}

func New() (*Service, error) {
	return &Service{}, nil
}

func (service *Service) AddBlock(content []byte) (*block.Block, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.chain.AddSHA512(content)
	block := service.chain.At(service.chain.Length() - 1)
	return block, nil
}

func (service *Service) ResetChain() error {
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

func (service *Service) GCI(other *blockchain.Blockchain) (int, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	gci, err := service.chain.GetGCI(other)
	if err != nil {
		return -1, err
	}

	return gci, nil
}

func (service *Service) ChainLength() int {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	return service.chain.Length()
}

func (service *Service) lock() {
	service.mutex.Lock()
}

func (service *Service) unlock() {
	service.mutex.Unlock()
}

func (service *Service) UpdateChain(other *blockchain.Blockchain) error {
	service.lock()
	defer service.unlock()
	newChain, err := blockchain.UpdateChain(service.chain, other)
	if err != nil {
		return err
	}

	service.chain = newChain
	return nil
}

func (service *Service) ChainJSON() ([]byte, error) {
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
func (service *Service) FindBlockByIndex(index int) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.At(index)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

// FindBlockByHash searches for a block by hash
func (service *Service) FindBlockByHash(hash []byte) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByBlockHash(hash)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

//FindBlockByParentHash searches for a block by parent hash
func (service *Service) FindBlockByParentHash(hash []byte) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByParentHash(hash)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

// FindBlockByTimestamp searches for a block by timestamp
func (service *Service) FindBlockByTimestamp(timestamp int64) (*block.Block, error) {
	service.lock()
	defer service.unlock()

	block := service.chain.FindByTimestamp(timestamp)
	if block == nil {
		return nil, ErrBlockNotFound
	}

	return block, nil
}

func (service *Service) Chain() *blockchain.Blockchain {
	service.lock()
	defer service.unlock()

	return service.chain
}
